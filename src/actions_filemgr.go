package plugin

import "fmt"

func actionFileRefresh() {
	dir := readState(func(state DebugState) string {
		return state.FileCurrentDir
	})
	if err := RefreshDirectory(dir); err != nil {
		appendLogf("ERROR", "刷新目录失败: %v", err)
	}
}

func actionFileGoDir() {
	dirInput := readState(func(state DebugState) string {
		return state.FileDirInput
	})
	snapshot := readStateSnapshot()
	normalized := NormalizePath(snapshot.FileCurrentDir, dirInput)
	if err := RefreshDirectory(normalized); err != nil {
		appendLogf("ERROR", "切换目录失败: %v", err)
	}
}

func actionFileParent() {
	current := readState(func(state DebugState) string {
		return state.FileCurrentDir
	})
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
		appendLog("WARN", "找不到目标条目，请先刷新目录")
		return
	}

	if selected.IsDir {
		if err := RefreshDirectory(selected.Path); err != nil {
			appendLogf("ERROR", "进入目录失败: %v", err)
		}
		return
	}

	withState(func(state *DebugState) {
		idx := -1
		for i, p := range state.FileSelectedPaths {
			if p == selected.Path {
				idx = i
				break
			}
		}
		if idx >= 0 {
			state.FileSelectedPaths = append(state.FileSelectedPaths[:idx], state.FileSelectedPaths[idx+1:]...)
			if len(state.FileSelectedPaths) != 1 {
				state.FileEditorText = ""
				state.FileEditorHexPreview = ""
			}
		} else {
			state.FileSelectedPaths = append(state.FileSelectedPaths, selected.Path)
		}
	})

	snapshot = readStateSnapshot()
	if len(snapshot.FileSelectedPaths) == 1 {
		if err := LoadRemoteFilePreview(snapshot.FileSelectedPaths[0]); err != nil {
			appendLogf("ERROR", "打开文件失败: %v", err)
		}
	}
}

func actionFileNewFile() {
	baseDir := readState(func(state DebugState) string { return state.FileCurrentDir })
	name, ok := promptInputDialog("新建文件", "请输入文件名", "new.txt")
	if !ok {
		return
	}
	path := JoinPath(baseDir, name)
	if !confirmDialog("确认创建", "确认创建文件？\n"+path) {
		return
	}
	EnqueueRpcTask("new file", func() error {
		return RpcFsWriteChunk(path, "truncate", "", func(_ FsWriteResult, err error) {
			if err != nil {
				appendLogf("ERROR", "创建文件失败: %v", err)
				return
			}
			appendLogf("INFO", "创建文件成功: %s", path)
			_ = RefreshDirectory(baseDir)
		})
	})
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
		msg = "确认删除?\n" + targets[0]
	}
	if !confirmDialog("确认删除", msg) {
		return
	}

	baseDir := readState(func(state DebugState) string { return state.FileCurrentDir })
	for _, target := range targets {
		enqueueShellCommandTask("delete", "rm -rf "+ShellQuote(target), nil)
	}

	enqueueShellCommandTask("delete_done", "echo done", func(_ ShellExecResult) {
		appendLogf("INFO", "删除完成: 数量=%d", len(targets))
		withState(func(state *DebugState) {
			state.FileSelectedPaths = nil
			state.FileEditorText = ""
			state.FileEditorHexPreview = ""
		})
		_ = RefreshDirectory(baseDir)
	})
}

func actionFileRename() {
	targets := readState(func(state DebugState) []string { return state.FileSelectedPaths })
	if len(targets) != 1 {
		appendLog("WARN", "重命名操作需要且仅支持 1 个文件")
		return
	}
	target := targets[0]
	newName, ok := promptInputDialog("重命名", "请输入新名称", BaseName(target))
	if !ok {
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
		})
		_ = RefreshDirectory(baseDir)
	})
}

func actionFileCopy() {
	targets := readState(func(state DebugState) []string { return state.FileSelectedPaths })
	if len(targets) != 1 {
		appendLog("WARN", "复制操作需要且仅支持 1 个文件")
		return
	}
	target := targets[0]
	defaultDst := target + "_copy"
	dst, ok := promptInputDialog("复制", "请输入目标路径", defaultDst)
	if !ok {
		return
	}
	dst = NormalizePath(readState(func(state DebugState) string { return state.FileCurrentDir }), dst)
	if !confirmDialog("确认复制", fmt.Sprintf("%s\n->\n%s", target, dst)) {
		return
	}
	baseDir := readState(func(state DebugState) string { return state.FileCurrentDir })
	enqueueShellCommandTask("copy", "cp -r "+ShellQuote(target)+" "+ShellQuote(dst), func(_ ShellExecResult) {
		appendLogf("INFO", "复制成功: %s -> %s", target, dst)
		_ = RefreshDirectory(baseDir)
	})
}

func actionFileMove() {
	targets := readState(func(state DebugState) []string { return state.FileSelectedPaths })
	if len(targets) != 1 {
		appendLog("WARN", "移动操作需要且仅支持 1 个文件")
		return
	}
	target := targets[0]
	dst, ok := promptInputDialog("移动", "请输入目标路径", target)
	if !ok {
		return
	}
	dst = NormalizePath(readState(func(state DebugState) string { return state.FileCurrentDir }), dst)
	if !confirmDialog("确认移动", fmt.Sprintf("%s\n->\n%s", target, dst)) {
		return
	}
	baseDir := readState(func(state DebugState) string { return state.FileCurrentDir })
	enqueueShellCommandTask("move", "mv "+ShellQuote(target)+" "+ShellQuote(dst), func(_ ShellExecResult) {
		appendLogf("INFO", "移动成功: %s -> %s", target, dst)
		_ = RefreshDirectory(baseDir)
	})
}

func actionFileUpload() {
	baseDir := readState(func(state DebugState) string { return state.FileCurrentDir })
	remotePath, ok := promptInputDialog("上传", "请输入远端保存路径（目录 + 文件名）", baseDir+"/")
	if !ok {
		return
	}
	if err := UploadLocalFileToRemote(remotePath); err != nil {
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
		appendLog("WARN", "暂不支持多个文件同时下载，只能选择一个文件")
		return
	}
	target := targets[0]
	localPath, err := DownloadRemoteFileToLocal(target)
	if err != nil {
		appendLogf("ERROR", "下载失败: %v", err)
		return
	}
	appendLogf("INFO", "下载已开始，目标本地路径: %s", localPath)
}

func actionFileSave() {
	snapshot := readStateSnapshot()
	if len(snapshot.FileSelectedPaths) != 1 {
		appendLog("WARN", "请且仅选中一个文件以保存")
		return
	}
	target := snapshot.FileSelectedPaths[0]
	if snapshot.FileEditorIsBinary {
		appendLog("WARN", "二进制预览模式不支持直接保存")
		return
	}
	if !confirmDialog("确认保存", "确认覆盖远端文件?\n"+target) {
		return
	}
	if err := WriteRemoteFile(target, []byte(snapshot.FileEditorText)); err != nil {
		appendLogf("ERROR", "保存失败: %v", err)
		return
	}
	appendLogf("INFO", "保存已开始: %s", target)
}
