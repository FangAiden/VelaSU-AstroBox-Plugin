package plugin

import (
	"fmt"
	"strings"

	ui "astroboxplugin/bindings/astrobox_psys_host_ui_v3"
)

func actionLogClear() {
	clearLogs()
	appendLog("INFO", "日志已清空")
}

func actionLogExportText() {
	snapshot := readStateSnapshot()
	var builder strings.Builder
	for _, item := range snapshot.Logs {
		builder.WriteString(fmt.Sprintf("[%s] [%s] %s\n", item.Timestamp, item.Level, item.Message))
	}
	ui.RenderToTextCard("插件日志记录", builder.String())
	appendLog("INFO", "已导出日志内容供复制")
}
