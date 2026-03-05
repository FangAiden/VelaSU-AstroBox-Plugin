package plugin

import (
	"fmt"
	"sort"
	"strconv"
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
		state.FileRenameInput = ""
		state.FilePreviewSeq++
		state.FileEditorText = ""
		state.FileEditorHexPreview = ""
		state.FileEditorIsBinary = false
		state.TransferProgress = "listing"
	})

	EnqueueRpcTask("list directory", func() error {
		cmd := "ls -l " + ShellQuote(dir)
		return RpcShellExec(cmd, DefaultShellTimeoutMs, func(result ShellExecResult, err error) {
			if err != nil {
				appendLogf("ERROR", "读取目录失败: %v", err)
				withState(func(state *DebugState) {
					state.TransferProgress = "list_failed"
				})
				return
			}

			entries := parseLsEntries(result.Output, dir)
			withState(func(state *DebugState) {
				sortFileEntries(entries, state.FileSortMode, state.FileSortAsc)
				state.FileEntries = entries
				state.FileVisibleCount = minInt(DefaultDirPageSize, len(entries))
				if len(entries) == 0 {
					state.TransferProgress = "list_empty"
				} else {
					state.TransferProgress = fmt.Sprintf("list_ready %d", len(entries))
				}
			})
			appendLogf("INFO", "目录加载完成: %s (items=%d)", dir, len(entries))
		})
	})

	return nil
}

func parseLsEntries(output string, dir string) []FileEntry {
	lines := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	seen := make(map[string]bool)
	entries := make([]FileEntry, 0, len(lines))

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "total ") || strings.HasSuffix(line, ":") {
			continue
		}

		name := ""
		isDir := false
		size := int64(-1)
		metaReady := false

		fields := strings.Fields(line)
		if len(fields) > 1 && looksLikeLsMode(fields[0]) {
			name = fields[len(fields)-1]
			name = strings.TrimSuffix(name, "/")
			isDir = fields[0][0] == 'd'
			if !isDir {
				size = parseLsSize(fields)
			}
			metaReady = true
		} else {
			// Fallback for plain `ls` output: keep clickable, resolve type on click if needed.
			name = strings.TrimSuffix(line, "/")
			if strings.HasSuffix(line, "/") {
				isDir = true
				metaReady = true
			}
		}

		if name == "" || name == "." || name == ".." {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true

		entries = append(entries, FileEntry{
			Name:      name,
			Path:      JoinPath(dir, name),
			Exists:    true,
			IsDir:     isDir,
			Size:      size,
			MetaReady: metaReady,
		})
	}

	return entries
}

func looksLikeLsMode(mode string) bool {
	if len(mode) < 2 {
		return false
	}
	switch mode[0] {
	case '-', 'd', 'l', 'c', 'b', 'p', 's':
		return true
	default:
		return false
	}
}

func parseLsSize(fields []string) int64 {
	for i := len(fields) - 2; i >= 1; i-- {
		if n, err := strconv.ParseInt(fields[i], 10, 64); err == nil {
			return n
		}
	}
	return -1
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

func sortFileEntries(entries []FileEntry, mode FileSortMode, asc bool) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}

		// File size sorting only applies to files. Directories are always grouped first.
		if mode == FileSortBySize && !entries[i].IsDir && !entries[j].IsDir {
			if entries[i].Size != entries[j].Size {
				if asc {
					return entries[i].Size < entries[j].Size
				}
				return entries[i].Size > entries[j].Size
			}
		}

		nameI := strings.ToLower(entries[i].Name)
		nameJ := strings.ToLower(entries[j].Name)
		if nameI == nameJ {
			if asc {
				return entries[i].Path < entries[j].Path
			}
			return entries[i].Path > entries[j].Path
		}
		if asc {
			return nameI < nameJ
		}
		return nameI > nameJ
	})
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
