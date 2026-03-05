package plugin

import (
	"fmt"
	"strings"
)

func actionFileRefresh() {
	dir := readState(func(state DebugState) string { return state.FileCurrentDir })
	if err := RefreshDirectory(dir); err != nil {
		appendLogf("ERROR", "刷新目录失败: %v", err)
	}
}

func actionFileGoDir() {
	dirInput := readState(func(state DebugState) string { return state.FileDirInput })
	snapshot := readStateSnapshot()
	normalized := NormalizePath(snapshot.FileCurrentDir, dirInput)
	if err := RefreshDirectory(normalized); err != nil {
		appendLogf("ERROR", "切换目录失败: %v", err)
	}
}

func actionFileParent() {
	current := readState(func(state DebugState) string { return state.FileCurrentDir })
	parent := ParentDir(current)
	if err := RefreshDirectory(parent); err != nil {
		appendLogf("ERROR", "返回上级目录失败: %v", err)
	}
}

func actionFileLoadMore() {
	withState(func(state *DebugState) {
		state.FileVisibleCount = minInt(state.FileVisibleCount+DefaultDirPageSize, len(state.FileEntries))
	})
}

func actionFileSort(mode FileSortMode) {
	withState(func(state *DebugState) {
		if state.FileSortMode == mode {
			state.FileSortAsc = !state.FileSortAsc
		} else {
			state.FileSortMode = mode
			state.FileSortAsc = true
		}
		sortFileEntries(state.FileEntries, state.FileSortMode, state.FileSortAsc)
	})
}

func actionFileOpenEntry(path string) {
	path = NormalizePath(readState(func(state DebugState) string { return state.FileCurrentDir }), path)
	snapshot := readStateSnapshot()

	var selected *FileEntry
	for i := range snapshot.FileEntries {
		if snapshot.FileEntries[i].Path == path {
			entry := snapshot.FileEntries[i]
			selected = &entry
			break
		}
	}
	if selected == nil {
		appendLog("WARN", "目标条目不存在，请先刷新目录")
		return
	}

	if selected.IsDir {
		if err := RefreshDirectory(selected.Path); err != nil {
			appendLogf("ERROR", "进入目录失败: %v", err)
		}
		return
	}

	// 删除“打开文件内容”能力后，点击文件仅用于选择。
	toggleFileSelection(selected.Path)
}

func toggleFileSelection(path string) {
	withState(func(state *DebugState) {
		idx := -1
		for i, p := range state.FileSelectedPaths {
			if p == path {
				idx = i
				break
			}
		}

		if idx >= 0 {
			state.FileSelectedPaths = append(state.FileSelectedPaths[:idx], state.FileSelectedPaths[idx+1:]...)
		} else {
			state.FileSelectedPaths = append(state.FileSelectedPaths, path)
		}

		if len(state.FileSelectedPaths) == 1 {
			state.FileRenameInput = BaseName(state.FileSelectedPaths[0])
		} else {
			state.FileRenameInput = ""
		}

		// Do not auto-load preview when clicking list items.
		state.FilePreviewSeq++
		state.FileEditorIsBinary = false
		state.FileEditorText = ""
		state.FileEditorHexPreview = ""
		state.TransferProgress = ""
	})
}

func findEntryByPath(entries []FileEntry, path string) (FileEntry, bool) {
	for _, it := range entries {
		if it.Path == path {
			return it, true
		}
	}
	return FileEntry{}, false
}

func actionFileNewDir() {
	baseDir := readState(func(state DebugState) string { return state.FileCurrentDir })
	name, ok := promptInputDialog("新建目录", "请输入目录名", "new_dir")
	if !ok {
		return
	}
	path := JoinPath(baseDir, name)
	if !confirmDialog("确认创建", "确认创建目录？\n"+path) {
		return
	}

	enqueueShellCommandTask("new dir", "mkdir -p "+ShellQuote(path), func(_ ShellExecResult) {
		appendLogf("INFO", "创建目录成功: %s", path)
		_ = RefreshDirectory(baseDir)
	})
}

func actionFileDelete() {
	targets := readState(func(state DebugState) []string { return state.FileSelectedPaths })
	if len(targets) == 0 {
		appendLog("WARN", "请先选中文件或目录")
		return
	}

	msg := fmt.Sprintf("确认删除选中的 %d 个项目？", len(targets))
	if len(targets) == 1 {
		msg = "确认删除？\n" + targets[0]
	}
	if !confirmDialog("确认删除", msg) {
		return
	}

	parts := make([]string, 0, len(targets)+2)
	parts = append(parts, "rm", "-r")
	for _, target := range targets {
		parts = append(parts, ShellQuote(target))
	}
	cmd := strings.Join(parts, " ")
	baseDir := readState(func(state DebugState) string { return state.FileCurrentDir })

	enqueueShellCommandTask("delete", cmd, func(_ ShellExecResult) {
		appendLogf("INFO", "删除完成: 数量=%d", len(targets))
		withState(func(state *DebugState) {
			state.FileSelectedPaths = nil
			state.FileRenameInput = ""
			state.FileEditorIsBinary = false
			state.FileEditorText = ""
			state.FileEditorHexPreview = ""
			state.TransferProgress = ""
		})
		_ = RefreshDirectory(baseDir)
	})
}

func actionFileRename() {
	snapshot := readStateSnapshot()
	if len(snapshot.FileSelectedPaths) != 1 {
		appendLog("WARN", "重命名仅支持单个文件/目录")
		return
	}

	target := snapshot.FileSelectedPaths[0]
	base := BaseName(target)
	newName := strings.TrimSpace(snapshot.FileRenameInput)

	if newName == "" {
		appendLog("WARN", "请先输入新名称")
		return
	}
	if strings.Contains(newName, "/") || strings.Contains(newName, "\\") {
		appendLog("WARN", "新名称不能包含路径分隔符")
		return
	}
	if newName == base {
		appendLog("WARN", "名称未变化")
		return
	}

	dst := JoinPath(ParentDir(target), newName)
	if !confirmDialog("确认重命名", fmt.Sprintf("%s\n->\n%s", target, dst)) {
		return
	}

	baseDir := readState(func(state DebugState) string { return state.FileCurrentDir })
	enqueueShellCommandTask("rename", "mv "+ShellQuote(target)+" "+ShellQuote(dst), func(_ ShellExecResult) {
		appendLogf("INFO", "重命名成功: %s -> %s", target, dst)
		withState(func(state *DebugState) {
			state.FileSelectedPaths = []string{dst}
			state.FileRenameInput = BaseName(dst)
		})
		_ = RefreshDirectory(baseDir)
	})
}

func actionFileCopy() {
	targets := readState(func(state DebugState) []string { return state.FileSelectedPaths })
	if len(targets) == 0 {
		appendLog("WARN", "请先选中要复制的文件或目录")
		return
	}
	if len(targets) != 1 {
		appendLog("WARN", "复制仅支持单个文件/目录")
		return
	}

	target := targets[0]
	snapshot := readStateSnapshot()
	entry, ok := findEntryByPath(snapshot.FileEntries, target)
	isDir := ok && entry.IsDir
	withState(func(state *DebugState) {
		state.FileClipboard = ClipboardState{SourcePath: target, SourceIsDir: isDir, Mode: "copy"}
	})
	appendLogf("INFO", "已复制到剪贴板: %s（切换到目标目录后点击“粘贴”）", target)
}

func actionFileMove() {
	targets := readState(func(state DebugState) []string { return state.FileSelectedPaths })
	if len(targets) == 0 {
		appendLog("WARN", "请先选中要移动的文件或目录")
		return
	}
	if len(targets) != 1 {
		appendLog("WARN", "移动仅支持单个文件/目录")
		return
	}

	target := targets[0]
	snapshot := readStateSnapshot()
	entry, ok := findEntryByPath(snapshot.FileEntries, target)
	isDir := ok && entry.IsDir
	withState(func(state *DebugState) {
		state.FileClipboard = ClipboardState{SourcePath: target, SourceIsDir: isDir, Mode: "move"}
	})
	appendLogf("INFO", "已剪切到剪贴板: %s（切换到目标目录后点击“粘贴”）", target)
}

func actionFilePaste() {
	snapshot := readStateSnapshot()
	clip := snapshot.FileClipboard
	if strings.TrimSpace(clip.SourcePath) == "" {
		appendLog("WARN", "剪贴板为空，请先复制或移动")
		return
	}
	if clip.Mode != "copy" && clip.Mode != "move" {
		appendLog("WARN", "剪贴板状态无效")
		return
	}

	src := NormalizePath("/", clip.SourcePath)
	dst := JoinPath(snapshot.FileCurrentDir, BaseName(src))
	if dst == src {
		if clip.Mode == "move" {
			appendLog("WARN", "源路径和目标路径相同，无法移动")
			return
		}
		dst = JoinPath(snapshot.FileCurrentDir, BaseName(src)+"_copy")
	}

	actionText := "复制"
	qSrc := ShellQuote(src)
	qDst := ShellQuote(dst)
	cmd := "cp " + qSrc + " " + qDst
	if clip.SourceIsDir {
		cmd = "mkdir -p " + qDst + "; cp -r " + qSrc + " " + qDst
	}
	if clip.Mode == "move" {
		actionText = "移动"
		cmd = "mv " + qSrc + " " + qDst
	}

	if !confirmDialog("确认粘贴", fmt.Sprintf("%s\n%s\n->\n%s", actionText, src, dst)) {
		return
	}

	baseDir := snapshot.FileCurrentDir
	enqueueShellCommandTask("paste", cmd, func(_ ShellExecResult) {
		appendLogf("INFO", "%s成功: %s -> %s", actionText, src, dst)
		withState(func(state *DebugState) {
			state.FileSelectedPaths = []string{dst}
			state.FileRenameInput = BaseName(dst)
			if clip.Mode == "move" {
				state.FileClipboard = ClipboardState{}
			}
		})
		_ = RefreshDirectory(baseDir)
	})
}

func actionFileClipboardClear() {
	withState(func(state *DebugState) {
		state.FileClipboard = ClipboardState{}
	})
	appendLog("INFO", "已清空文件剪贴板")
}

func actionFileUpload() {
	if err := UploadLocalFileToRemote(""); err != nil {
		appendLogf("ERROR", "上传失败: %v", err)
	}
}

func actionFileDownload() {
	targets := readState(func(state DebugState) []string { return state.FileSelectedPaths })
	if len(targets) == 0 {
		appendLog("WARN", "请先选中要下载的文件")
		return
	}
	if len(targets) > 1 {
		appendLog("WARN", "暂不支持多文件同时下载，请只选择一个文件")
		return
	}

	target := targets[0]
	localPath, err := DownloadRemoteFileToLocal(target)
	if err != nil {
		appendLogf("ERROR", "下载失败: %v", err)
		return
	}
	appendLogf("INFO", "下载已开始，本地路径: %s", localPath)
}
