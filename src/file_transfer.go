package plugin

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"
)

func ReadRemoteFile(path string, maxBytes int) ([]byte, bool, error) {
	_ = path
	_ = maxBytes
	return nil, false, fmt.Errorf("ReadRemoteFile 采用异步队列模式，请通过文件列表打开")
}

func WriteRemoteFile(path string, data []byte) error {
	path = NormalizePath(readState(func(state DebugState) string { return state.FileCurrentDir }), path)
	if path == "" || path == "/" {
		return fmt.Errorf("写入路径无效")
	}

	chunks := chunkBytes(data, DefaultUploadChunk)
	if len(chunks) == 0 {
		chunks = [][]byte{{}}
	}

	total := len(chunks)
	for i, raw := range chunks {
		chunk := append([]byte(nil), raw...)
		mode := "append"
		if i == 0 {
			mode = "truncate"
		}
		idx := i + 1
		EnqueueRpcTask(fmt.Sprintf("write chunk %d/%d", idx, total), func() error {
			encoded := base64.StdEncoding.EncodeToString(chunk)
			return RpcFsWriteChunk(path, mode, encoded, func(_ FsWriteResult, err error) {
				if err != nil {
					appendLogf("ERROR", "写入分块失败 %d/%d: %v", idx, total, err)
					return
				}
				withState(func(state *DebugState) {
					state.TransferProgress = fmt.Sprintf("upload %d/%d", idx, total)
				})
			})
		})
	}

	EnqueueRpcTask("write complete", func() error {
		withState(func(state *DebugState) {
			state.TransferProgress = "upload done"
		})
		appendLogf("INFO", "远端写入完成: %s", path)
		_ = RefreshDirectory(readState(func(state DebugState) string { return state.FileCurrentDir }))
		return nil
	})

	return nil
}

func UploadLocalFileToRemote(remotePath string) error {
	name, data, err := pickLocalFile()
	if err != nil {
		return err
	}
	currentDir := readState(func(state DebugState) string { return state.FileCurrentDir })
	remotePath = strings.TrimSpace(remotePath)
	if remotePath == "" {
		remotePath = JoinPath(currentDir, name)
	} else if strings.HasSuffix(remotePath, "/") {
		remotePath = JoinPath(remotePath, name)
	}
	withState(func(state *DebugState) {
		state.TransferProgress = fmt.Sprintf("upload queued (%d bytes)", len(data))
	})
	appendLogf("INFO", "准备上传文件: %s -> %s", name, remotePath)
	return WriteRemoteFile(remotePath, data)
}

func DownloadRemoteFileToLocal(remotePath string) (string, error) {
	remotePath = NormalizePath(readState(func(state DebugState) string { return state.FileCurrentDir }), remotePath)
	if remotePath == "" || remotePath == "/" {
		return "", fmt.Errorf("下载路径无效")
	}

	base := BaseName(remotePath)
	sessionID, saveName, err := startLocalSaveSession(base)
	if err != nil {
		return "", err
	}
	downloadedBytes := 0

	var enqueueRead func(offset int)
	enqueueRead = func(offset int) {
		offsetCopy := offset
		EnqueueRpcTask(fmt.Sprintf("download offset %d", offsetCopy), func() error {
			return RpcFsReadChunk(remotePath, offsetCopy, DefaultFsChunkSize, func(result FsReadResult, err error) {
				if err != nil {
					appendLogf("ERROR", "下载失败: %v", err)
					withState(func(state *DebugState) {
						state.TransferProgress = "download failed"
					})
					abortLocalSaveSession(sessionID)
					return
				}
				data, err := base64.StdEncoding.DecodeString(result.Data)
				if err != nil {
					appendLogf("ERROR", "下载解码失败: %v", err)
					withState(func(state *DebugState) {
						state.TransferProgress = "download decode failed"
					})
					abortLocalSaveSession(sessionID)
					return
				}
				if err := writeLocalSaveChunk(sessionID, data); err != nil {
					appendLogf("ERROR", "写入本地分块失败: %v", err)
					withState(func(state *DebugState) {
						state.TransferProgress = "download write failed"
					})
					abortLocalSaveSession(sessionID)
					return
				}
				downloadedBytes += len(data)
				withState(func(state *DebugState) {
					state.TransferProgress = fmt.Sprintf("download %d bytes", downloadedBytes)
				})
				if result.Eof {
					if err := finishLocalSaveSession(sessionID); err != nil {
						appendLogf("ERROR", "完成本地保存失败: %v", err)
						withState(func(state *DebugState) {
							state.TransferProgress = "download finish failed"
						})
						abortLocalSaveSession(sessionID)
						return
					}
					withState(func(state *DebugState) {
						state.TransferLastLocalPath = saveName
						state.TransferProgress = fmt.Sprintf("download done (%d bytes)", downloadedBytes)
					})
					appendLogf("INFO", "下载完成: %s", saveName)
					return
				}
				enqueueRead(result.NextOffset)
			})
		})
	}

	enqueueRead(0)
	return saveName, nil
}

func LoadRemoteFilePreview(remotePath string) error {
	remotePath = NormalizePath(readState(func(state DebugState) string { return state.FileCurrentDir }), remotePath)
	if remotePath == "" || remotePath == "/" {
		return fmt.Errorf("invalid remote file path")
	}

	var previewSeq uint64
	withState(func(state *DebugState) {
		state.FilePreviewSeq++
		previewSeq = state.FilePreviewSeq
		state.FileSelectedPaths = []string{remotePath}
		state.TransferProgress = "preview reading"
	})

	isPreviewActive := func() bool {
		return readState(func(state DebugState) bool {
			return state.FilePreviewSeq == previewSeq &&
				len(state.FileSelectedPaths) == 1 &&
				state.FileSelectedPaths[0] == remotePath
		})
	}

	EnqueueRpcTask("preview chunk 0", func() error {
		if !isPreviewActive() {
			return nil
		}
		return RpcFsReadChunk(remotePath, 0, DefaultFsChunkSize, func(result FsReadResult, err error) {
			if !isPreviewActive() {
				return
			}
			if err != nil {
				appendLogf("ERROR", "preview read failed: %v", err)
				withState(func(state *DebugState) {
					state.TransferProgress = "preview failed"
				})
				return
			}

			data, err := base64.StdEncoding.DecodeString(result.Data)
			if err != nil {
				appendLogf("ERROR", "preview decode failed: %v", err)
				withState(func(state *DebugState) {
					state.TransferProgress = "preview decode failed"
				})
				return
			}

			truncated := !result.Eof
			if len(data) > MaxEditorPreviewSize {
				data = data[:MaxEditorPreviewSize]
				truncated = true
			}
			applyEditorBytes(remotePath, data, truncated)
		})
	})
	return nil
}

func applyEditorBytes(path string, data []byte, truncated bool) {
	isBinary := !utf8.Valid(data)
	text := ""
	hexPreview := ""
	if isBinary {
		hexPreview = BytesToHexPreview(data, HexPreviewBytes)
		text = ""
	} else {
		text = string(data)
		hexPreview = ""
	}
	if truncated {
		if !isBinary {
			text += "\n\n[预览已截断]"
		} else {
			hexPreview += "\n[预览已截断]"
		}
	}

	withState(func(state *DebugState) {
		// Only apply preview if this file is the ONLY one selected currently
		if len(state.FileSelectedPaths) == 1 && state.FileSelectedPaths[0] == path {
			state.FileEditorIsBinary = isBinary
			state.FileEditorText = text
			state.FileEditorHexPreview = hexPreview
			state.TransferProgress = "preview done"
		} else {
			// Selection changed while reading chunk
			state.TransferProgress = ""
		}
	})
	appendLogf("INFO", "文件预览完成: %s (bytes=%d, binary=%t)", path, len(data), isBinary)
}

func chunkBytes(data []byte, chunkSize int) [][]byte {
	if chunkSize <= 0 {
		chunkSize = DefaultUploadChunk
	}
	if len(data) == 0 {
		return nil
	}
	chunks := make([][]byte, 0, (len(data)+chunkSize-1)/chunkSize)
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		part := make([]byte, end-i)
		copy(part, data[i:end])
		chunks = append(chunks, part)
	}
	return chunks
}
