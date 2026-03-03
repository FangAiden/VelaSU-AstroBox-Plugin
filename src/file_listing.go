package plugin

import (
	"fmt"
	"sort"
	"strings"
)

func BuildDirectoryEntries(dir string) ([]FileEntry, error) {
	snapshot := readStateSnapshot()
	dir = NormalizePath(snapshot.FileCurrentDir, dir)
	if dir != snapshot.FileCurrentDir {
		return nil, fmt.Errorf("directory not loaded: %s", dir)
	}
	entries := append([]FileEntry(nil), snapshot.FileEntries...)
	return entries, nil
}

func RefreshDirectory(dir string) error {
	snapshot := readStateSnapshot()
	dir = NormalizePath(snapshot.FileCurrentDir, dir)

	withState(func(state *DebugState) {
		state.FileCurrentDir = dir
		state.FileDirInput = dir
		state.FileEntries = nil
		state.FileVisibleCount = DefaultDirPageSize
		state.FileSelectedPaths = nil
		state.FileEditorText = ""
		state.FileEditorHexPreview = ""
		state.FileEditorIsBinary = false
		state.TransferProgress = "listing"
	})

	EnqueueRpcTask("list directory", func() error {
		cmd := "ls " + ShellQuote(dir)
		return RpcShellExec(cmd, DefaultShellTimeoutMs, func(result ShellExecResult, err error) {
			if err != nil {
				appendLogf("ERROR", "读取目录失败: %v", err)
				withState(func(state *DebugState) {
					state.TransferProgress = "list_failed"
				})
				return
			}

			names := ParseLsOutput(result.Output)
			entries := make([]FileEntry, 0, len(names))
			for _, name := range names {
				entries = append(entries, FileEntry{
					Name:      name,
					Path:      JoinPath(dir, name),
					Exists:    true,
					MetaReady: false,
				})
			}

			withState(func(state *DebugState) {
				state.FileEntries = entries
				state.FileVisibleCount = minInt(DefaultDirPageSize, len(entries))
				if len(entries) == 0 {
					state.TransferProgress = "list_empty"
				} else {
					state.TransferProgress = fmt.Sprintf("listing %d items", len(entries))
				}
			})

			for _, item := range entries {
				pathCopy := item.Path
				EnqueueRpcTask("stat "+pathCopy, func() error {
					return RpcFsStat(pathCopy, func(stat FsStatResult, err error) {
						if err != nil {
							markEntryMetaError(pathCopy, err.Error())
							return
						}
						withState(func(state *DebugState) {
							for i := range state.FileEntries {
								if state.FileEntries[i].Path == pathCopy {
									state.FileEntries[i].Exists = stat.Exists
									state.FileEntries[i].IsDir = stat.IsDir
									state.FileEntries[i].Size = stat.Size
									state.FileEntries[i].MetaReady = true
									state.FileEntries[i].MetaErr = ""
									break
								}
							}
						})
					})
				})
			}

			EnqueueRpcTask("sort entries", func() error {
				withState(func(state *DebugState) {
					sortFileEntries(state.FileEntries, state.FileSortMode)
					state.FileVisibleCount = minInt(DefaultDirPageSize, len(state.FileEntries))
					state.TransferProgress = fmt.Sprintf("list_ready %d", len(state.FileEntries))
				})
				appendLogf("INFO", "目录加载完成: %s", dir)
				return nil
			})
		})
	})

	return nil
}

func markEntryMetaError(path string, errText string) {
	errText = normalizeInline(errText)
	withState(func(state *DebugState) {
		for i := range state.FileEntries {
			if state.FileEntries[i].Path == path {
				state.FileEntries[i].MetaReady = true
				state.FileEntries[i].MetaErr = errText
				break
			}
		}
	})
}

func sortFileEntries(entries []FileEntry, mode FileSortMode) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		switch mode {
		case FileSortBySize:
			if entries[i].Size != entries[j].Size {
				return entries[i].Size > entries[j].Size
			}
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
