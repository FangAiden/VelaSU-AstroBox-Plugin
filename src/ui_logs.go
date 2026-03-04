package plugin

import (
	"fmt"

	ui "astroboxplugin/bindings/astrobox_psys_host_ui_v3"
)

func buildLogsOnlyRoot(snapshot DebugState) *ui.Element {
	root := makeColumn().WidthFull().Gap(10)

	header := makePanel().
		Bg("#10172A").
		Padding(12).
		Child(makeSectionTitle("系统日志").MarginRight(4)).
		Child(makeMutedText(fmt.Sprintf("日志容量: %d / %d", len(snapshot.Logs), MaxLogEntries)).MarginTop(6)).
		Child(makeRow().Gap(8).MarginTop(8).
			Child(makeSecondaryButton("导出全部", EventLogExportText).FlexGrow(1)).
			Child(makeDangerButton("清空日志", EventLogClear).FlexGrow(1)))
	root = root.Child(header)

	logPanel := makePanel().Bg("#0D1425").Padding(10)
	logs := snapshot.Logs
	if len(logs) == 0 {
		logPanel = logPanel.Child(makeMutedText("暂无日志输出"))
		return root.Child(logPanel)
	}

	totalPages := (len(logs) + LogPageSize - 1) / LogPageSize
	currentPage := snapshot.LogPage
	if currentPage >= totalPages {
		currentPage = totalPages - 1
	}
	if currentPage < 0 {
		currentPage = 0
	}
	start := currentPage * LogPageSize
	end := start + LogPageSize
	if end > len(logs) {
		end = len(logs)
	}

	scroll := el(ui.ElementTypeScrollArea, "").
		WidthFull().
		MinHeight(280).
		MaxHeight(420).
		Padding(4)
	list := makeColumn().Gap(6)
	for _, item := range logs[start:end] {
		line := fmt.Sprintf("[%s] [%s] %s", item.Timestamp, item.Level, item.Message)
		list = list.Child(makeCodeBlock(line).Bg("#0A111E").Border(1, "#23324A"))
	}
	scroll = scroll.Child(list)
	logPanel = logPanel.Child(scroll)

	if totalPages > 1 {
		nav := makeRow().AlignCenter().Gap(8).MarginTop(8)
		prev := makeSecondaryButton("上一页", EventLogPagePrev)
		if currentPage == 0 {
			prev = prev.Disabled().Opacity(0.4)
		}
		next := makeSecondaryButton("下一页", EventLogPageNext)
		if currentPage >= totalPages-1 {
			next = next.Disabled().Opacity(0.4)
		}
		nav = nav.
			Child(prev).
			Child(makeMutedText(fmt.Sprintf("%d / %d", currentPage+1, totalPages))).
			Child(next)
		logPanel = logPanel.Child(nav)
	}

	return root.Child(logPanel)
}
