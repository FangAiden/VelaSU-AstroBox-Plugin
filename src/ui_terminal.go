package plugin

import (
	"fmt"
	"strings"

	ui "astroboxplugin/bindings/astrobox_psys_host_ui_v3"
)

const terminalScrollBottom uint32 = 1000000

func terminalOutputText(lines []string) string {
	text := strings.Join(lines, "\n")
	if strings.TrimSpace(text) == "" {
		return "AstroBox Terminal Ready."
	}
	return text
}

func buildTerminalPanel(snapshot DebugState) *ui.Element {
	panel := makeColumn().WidthFull().Gap(10)

	allLines := strings.Split(strings.Join(snapshot.TerminalBuffer, "\n"), "\n")
	if len(snapshot.TerminalBuffer) == 0 {
		allLines = []string{"AstroBox Terminal Ready."}
	}

	token := strings.TrimSpace(snapshot.TokenInput)
	if token == "" {
		allLines = append([]string{
			"[提示] 尚未配置 Token，终端命令不会发送。",
			"请先进入“核心配置”填写 4 位 Token。",
			"",
		}, allLines...)
	} else if !tokenPattern.MatchString(token) {
		allLines = append([]string{
			"[提示] 当前 Token 格式无效，终端命令不会发送。",
			"请在“核心配置”中填写 4 位数字 Token。",
			"",
		}, allLines...)
	}

	header := makePanel().
		Bg("#10172A").
		Padding(10).
		Child(
			makeRow().
				AlignCenter().
				Gap(8).
				Child(makeSVGIcon(IconSVGTerminal)).
				Child(makeSectionTitle("终端")).
				Child(makeBadge(fmt.Sprintf("%d 行", len(allLines)))).
				Child(makeSpacer()).
				Child(makeSecondaryButton("导出", EventTerminalExportText)).
				Child(makeDangerButton("清空", EventTerminalClear)),
		)

	screen := makePanel().
		Bg("#050A12").
		Border(1, "#1F2A3F").
		Padding(10)

	scroll := el(ui.ElementTypeScrollArea, "").
		WidthFull().
		Height(320).
		Padding(4).
		ScrollBehavior("auto").
		ScrollTop(terminalScrollBottom)

	outputText := terminalOutputText(allLines)
	output := el(ui.ElementTypeCode, outputText).
		TextColor("#ABEBD4").
		Bg("transparent").
		Size(12)

	scroll = scroll.Child(output)
	screen = screen.Child(scroll)

	cmdInput := el(ui.ElementTypeInput, snapshot.CurrentCommand).
		WithoutDefaultStyles().
		FlexGrow(1).
		Autofocus().
		TabIndex(0).
		Bg("#0E1424").
		TextColor("#ECF2FF").
		Border(1, "#2E3852").
		Radius(8).
		Padding(8).
		On(ui.EventInput, EventCommandInput).
		On(ui.EventChange, EventCommandInput).
		On(ui.EventKeyDown, EventTerminalKeyDown)

	inputRow := makeRow().
		AlignCenter().
		Gap(8).
		Child(cmdInput).
		Child(makePrimaryButton("执行", EventExecCommand))

	historyPanel := makePanel().
		Bg("#0E1423").
		Padding(10).
		Child(makeSectionTitle("最近命令").MarginRight(4))
	if len(snapshot.TerminalHistory) == 0 {
		historyPanel = historyPanel.Child(makeMutedText("暂无历史记录").MarginTop(8))
	} else {
		startIdx := len(snapshot.TerminalHistory) - 5
		if startIdx < 0 {
			startIdx = 0
		}
		for idx := len(snapshot.TerminalHistory) - 1; idx >= startIdx; idx-- {
			item := snapshot.TerminalHistory[idx]
			row := makeRow().
				AlignCenter().
				Gap(8).
				MarginTop(8).
				Child(makeMutedText(item.Timestamp + "  " + item.ExitCode).Width(100)).
				Child(el(ui.ElementTypeCode, item.Command).TextColor("#D8E4FF").FlexGrow(1)).
				Child(makeActionButton("重跑", historyRunEventID(idx)))
			historyPanel = historyPanel.Child(row)
		}
	}

	panel = panel.Child(header).Child(screen).Child(inputRow).Child(historyPanel)
	return panel
}
