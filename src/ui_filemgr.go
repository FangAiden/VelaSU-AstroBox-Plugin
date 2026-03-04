package plugin

import (
	"fmt"

	ui "astroboxplugin/bindings/astrobox_psys_host_ui_v3"
)

func buildFileManagerPanel(snapshot DebugState) *ui.Element {
	panel := makeColumn().WidthFull().Gap(10)

	panel = panel.Child(buildFileHeader(snapshot))

	filtered := filterEntries(snapshot.FileEntries, snapshot.FileSearchQuery)
	visible := minInt(snapshot.FileVisibleCount, len(filtered))
	showEntries := filtered[:visible]

	if snapshot.FileViewMode == FileViewList {
		panel = panel.Child(buildFileListView(snapshot, showEntries))
	} else {
		panel = panel.Child(buildFileGridView(snapshot, showEntries))
	}

	if visible < len(filtered) {
		panel = panel.Child(makeSecondaryButton("加载更多", EventFileLoadMore))
	}

	panel = panel.Child(buildFileStatusBar(snapshot, len(filtered), visible))
	panel = panel.Child(buildFilePreviewPanel(snapshot))

	return panel
}

func buildFileHeader(snapshot DebugState) *ui.Element {
	header := makePanel().Bg("#10172A").Padding(10)

	pathRow := makeRow().
		AlignCenter().
		Gap(8).
		Child(makeSingleLineInput(snapshot.FileDirInput, EventFileDirInput).FlexGrow(1).MinWidth(0)).
		Child(makePrimaryButton("前往", EventFileGoDir).MinWidth(72))

	navRow := el(ui.ElementTypeGrid, "").
		WidthFull().
		GridTemplateColumns("repeat(2, minmax(0, 1fr))").
		Gap(8).
		MarginTop(8).
		Child(makeSecondaryButton("上级", EventFileParent).WidthFull()).
		Child(makeSecondaryButton("刷新", EventFileRefresh).WidthFull())

	searchBox := makeRow().
		WidthFull().
		AlignCenter().
		Gap(6).
		MarginTop(8).
		Bg("#0E1424").
		Border(1, "#2F3853").
		Radius(8).
		Padding(6).
		Child(makeSVGIcon(IconSVGSearch)).
		Child(makeSingleLineInput(snapshot.FileSearchQuery, EventFileSearchInput).
			Bg("transparent").
			Border(0, "transparent").
			Padding(0).
			FlexGrow(1).
			MinWidth(0))

	viewRow := makeRow().Gap(4).
		Child(makeViewSwitchBtn(IconSVGGrid, EventFileViewGrid, snapshot.FileViewMode == FileViewGrid)).
		Child(makeViewSwitchBtn(IconSVGList, EventFileViewList, snapshot.FileViewMode == FileViewList))

	sortMenu := el(ui.ElementTypeDropdownMenuRoot, "")
	sortTrigger := el(ui.ElementTypeDropdownMenuTrigger, "").
		Child(
			el(ui.ElementTypeButton, "排序").
				Bg("#252D44").
				TextColor("#D4DAEE").
				Border(1, "#394360").
				Radius(8).
				MinHeight(36).
				Padding(8).
				Child(makeSVGIcon(IconSVGSort).MarginLeft(4)),
		)
	sortTrigger = applyButtonMotion(sortTrigger)
	sortContent := el(ui.ElementTypeDropdownMenuContent, "").
		Child(el(ui.ElementTypeDropdownMenuItem, "按名称排序").On(ui.EventClick, EventFileSortName)).
		Child(el(ui.ElementTypeDropdownMenuItem, "按大小排序").On(ui.EventClick, EventFileSortSize)).
		Child(el(ui.ElementTypeDropdownMenuItem, "按日期排序").On(ui.EventClick, EventFileSortDate))
	sortMenu = sortMenu.Child(sortTrigger).Child(sortContent)

	controlRow := makeRow().
		AlignCenter().
		Gap(8).
		MarginTop(8).
		Child(viewRow).
		Child(sortMenu)

	quickActions := el(ui.ElementTypeGrid, "").
		WidthFull().
		GridTemplateColumns("repeat(auto-fit, minmax(96px, 1fr))").
		Gap(6).
		MarginTop(8).
		Child(makeActionButton("新建文件", EventFileNewFile).WidthFull()).
		Child(makeActionButton("新建目录", EventFileNewDir).WidthFull()).
		Child(makeActionButton("上传", EventFileUpload).WidthFull())

	header = header.
		Child(pathRow).
		Child(navRow).
		Child(searchBox).
		Child(controlRow).
		Child(quickActions)

	if len(snapshot.FileSelectedPaths) > 0 {
		selectedActions := el(ui.ElementTypeGrid, "").
			WidthFull().
			GridTemplateColumns("repeat(auto-fit, minmax(86px, 1fr))").
			Gap(6).
			Child(makeActionButton("复制", EventFileCopy).WidthFull()).
			Child(makeActionButton("移动", EventFileMove).WidthFull()).
			Child(makeActionButton("重命名", EventFileRename).WidthFull()).
			Child(makeActionButton("下载", EventFileDownload).WidthFull()).
			Child(makeDangerActionButton("删除", EventFileDelete).WidthFull())

		selectedBar := makeColumn().
			Gap(6).
			MarginTop(8).
			Child(makeBadge(fmt.Sprintf("已选中 %d 项", len(snapshot.FileSelectedPaths)))).
			Child(selectedActions)

		header = header.Child(selectedBar)
	}

	return header
}

func makeViewSwitchBtn(iconSVG string, eventID string, active bool) *ui.Element {
	btn := el(ui.ElementTypeButton, "").
		Padding(6).
		Radius(8).
		MinWidth(42).
		MinHeight(36).
		Child(makeSVGIcon(iconSVG))
	if active {
		btn = btn.Bg("#3A4D8F").Border(1, "#5F77C6").TextColor("#F2F6FF")
	} else {
		btn = btn.Bg("#222A42").Border(1, "#384360").TextColor("#AAB5D6")
	}
	return applyButtonMotion(btn.On(ui.EventClick, eventID))
}

func buildFileGridView(snapshot DebugState, entries []FileEntry) *ui.Element {
	container := makePanel().Bg("#0E1424").Padding(10)
	if len(entries) == 0 {
		return container.Child(makeMutedText("当前目录为空").Padding(8))
	}

	grid := el(ui.ElementTypeGrid, "").
		WidthFull().
		GridTemplateColumns("repeat(auto-fill, minmax(150px, 1fr))").
		Gap(8)

	for _, it := range entries {
		grid = grid.Child(buildFileGridItem(snapshot, it))
	}
	return container.Child(grid)
}

func buildFileGridItem(snapshot DebugState, it FileEntry) *ui.Element {
	ctxRoot := el(ui.ElementTypeContextMenuRoot, "")
	trigger := el(ui.ElementTypeContextMenuTrigger, "")

	card := el(ui.ElementTypeCard, "").
		Bg("#151D33").
		Border(1, "#2A3654").
		Radius(10).
		Padding(10).
		Flex().
		FlexDirection(ui.FlexDirectionColumn).
		AlignCenter().
		Gap(6).
		On(ui.EventClick, fileOpenEventID(it.Path))

	if isPathSelected(snapshot, it.Path) {
		card = card.Bg("#243769").Border(1, "#5A79C8")
	}

	if it.IsDir {
		card = card.Child(makeSVGIcon(IconSVGFolder).TextColor("#F6C85E"))
	} else {
		card = card.Child(makeSVGIcon(IconSVGFile).TextColor("#8AA5FF"))
	}
	card = card.
		Child(makeText(it.Name).Size(13).TextColor("#E8EEFF")).
		Child(makeMutedText(fileMetaText(it)))

	trigger = trigger.Child(card)
	ctxRoot = ctxRoot.Child(trigger).Child(buildFileContextMenu(it))
	return ctxRoot
}

func buildFileListView(snapshot DebugState, entries []FileEntry) *ui.Element {
	container := makePanel().Bg("#0E1424").Padding(8)
	if len(entries) == 0 {
		return container.Child(makeMutedText("当前目录为空").Padding(8))
	}

	list := makeColumn().Gap(6)
	for _, it := range entries {
		list = list.Child(buildFileListItem(snapshot, it))
	}
	return container.Child(list)
}

func buildFileListItem(snapshot DebugState, it FileEntry) *ui.Element {
	ctxRoot := el(ui.ElementTypeContextMenuRoot, "")
	trigger := el(ui.ElementTypeContextMenuTrigger, "")

	row := makeRow().
		AlignCenter().
		Gap(8).
		Padding(8).
		Bg("#141B2F").
		Border(1, "#293452").
		Radius(8).
		On(ui.EventClick, fileOpenEventID(it.Path))

	if isPathSelected(snapshot, it.Path) {
		row = row.Bg("#243769").Border(1, "#5A79C8")
	}

	icon := makeSVGIcon(IconSVGFile).TextColor("#8AA5FF")
	typeText := "文件"
	if it.IsDir {
		icon = makeSVGIcon(IconSVGFolder).TextColor("#F6C85E")
		typeText = "目录"
	}

	content := makeColumn().
		FlexGrow(1).
		MinWidth(0).
		Child(makeText(it.Name).TextColor("#E8EEFF")).
		Child(makeMutedText(typeText + " · " + fileMetaText(it)).MarginTop(2))

	row = row.Child(icon).Child(content)

	trigger = trigger.Child(row)
	ctxRoot = ctxRoot.Child(trigger).Child(buildFileContextMenu(it))
	return ctxRoot
}

func buildFileContextMenu(it FileEntry) *ui.Element {
	menu := el(ui.ElementTypeContextMenuContent, "").
		Child(el(ui.ElementTypeContextMenuItem, "复制").On(ui.EventClick, fileCtxCopyEventID(it.Path))).
		Child(el(ui.ElementTypeContextMenuItem, "移动").On(ui.EventClick, fileCtxMoveEventID(it.Path))).
		Child(el(ui.ElementTypeContextMenuItem, "重命名").On(ui.EventClick, fileCtxRenameEventID(it.Path))).
		Child(el(ui.ElementTypeContextMenuSeparator, "")).
		Child(el(ui.ElementTypeContextMenuItem, "删除").On(ui.EventClick, fileCtxDeleteEventID(it.Path)))
	if !it.IsDir {
		menu = menu.
			Child(el(ui.ElementTypeContextMenuSeparator, "")).
			Child(el(ui.ElementTypeContextMenuItem, "下载").On(ui.EventClick, fileCtxDownloadEventID(it.Path)))
	}
	return menu
}

func buildFileStatusBar(snapshot DebugState, filteredCount int, visibleCount int) *ui.Element {
	modeText := "网格"
	if snapshot.FileViewMode == FileViewList {
		modeText = "列表"
	}
	sortText := "名称"
	switch snapshot.FileSortMode {
	case FileSortBySize:
		sortText = "大小"
	case FileSortByDate:
		sortText = "日期(降级为名称)"
	}

	bar := makeColumn().
		Gap(6).
		Padding(8).
		Bg("#11192E").
		Border(1, "#27324A").
		Radius(10)

	summary := makeRow().
		AlignCenter().
		Gap(8).
		Child(makeSVGIcon(IconSVGHardDrive)).
		Child(makeMutedText(fmt.Sprintf("显示 %d / %d", visibleCount, filteredCount)))

	meta := el(ui.ElementTypeGrid, "").
		WidthFull().
		GridTemplateColumns("repeat(auto-fit, minmax(120px, 1fr))").
		Gap(6).
		Child(makeBadge("视图: " + modeText)).
		Child(makeBadge("排序: " + sortText))

	bar = bar.Child(summary).Child(meta)
	if snapshot.TransferProgress != "" && snapshot.TransferProgress != "idle" {
		bar = bar.Child(makeText("进度: " + snapshot.TransferProgress).TextColor("#87E9C6"))
	}
	return bar
}

func buildFilePreviewPanel(snapshot DebugState) *ui.Element {
	preview := makePanel().Bg("#0E1424").Padding(10)
	if len(snapshot.FileSelectedPaths) == 0 {
		return preview.Child(makeMutedText("未选中文件").Padding(4))
	}

	if len(snapshot.FileSelectedPaths) > 1 {
		return preview.Child(makeText(fmt.Sprintf("已选中 %d 项", len(snapshot.FileSelectedPaths))))
	}

	target := snapshot.FileSelectedPaths[0]
	preview = preview.Child(makeSectionTitle("文件预览")).Child(makeMutedText(target).MarginTop(4))

	if snapshot.FileEditorIsBinary {
		preview = preview.
			Child(makeMutedText("二进制预览(hex)").MarginTop(8)).
			Child(makeCodeBlock(fallback(snapshot.FileEditorHexPreview, "(empty)")).MarginTop(6))
	} else {
		preview = preview.
			Child(makeMutedText("文本编辑").MarginTop(8)).
			Child(makeInputArea(snapshot.FileEditorText, EventFileEditorInput).MinHeight(180).MarginTop(6)).
			Child(makePrimaryButton("保存修改", EventFileSave).MarginTop(8))
	}

	if snapshot.TransferLastLocalPath != "" {
		preview = preview.Child(makeMutedText("最近下载: " + snapshot.TransferLastLocalPath).MarginTop(8))
	}

	return preview
}

func fileMetaText(it FileEntry) string {
	if it.IsDir {
		return "目录"
	}
	if it.Size >= 0 {
		return formatFileSize(it.Size)
	}
	return "-"
}
