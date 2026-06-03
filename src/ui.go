package plugin

import (
	"fmt"
	"strings"
	"sync"

	ui "astroboxplugin/bindings/astrobox_psys_host_ui_v3"
)

var (
	uiRootMu            sync.Mutex
	uiRootElementID     string
	uiLastRenderedRoute string
)

func RenderMainUI(elementID string) {
	uiRootMu.Lock()
	uiRootElementID = elementID
	uiRootMu.Unlock()

	snapshot := readStateSnapshot()
	ui.Render(elementID, buildMainUI(snapshot))
}

func RerenderMainUI() {
	uiRootMu.Lock()
	elementID := uiRootElementID
	uiRootMu.Unlock()
	if elementID == "" {
		return
	}
	snapshot := readStateSnapshot()
	ui.Render(elementID, buildMainUI(snapshot))
}

func buildMainUI(snapshot DebugState) *ui.Element {
	main := makeColumn().
		WidthFull().
		Padding(10).
		Gap(10)

	animateRoute := shouldAnimateRoute(snapshot.CurrentAppRoute)

	if snapshot.CurrentAppRoute == RouteDashboard {
		dashboard := buildHomeDashboard(snapshot)
		if animateRoute {
			dashboard = applyPageMotion(RouteDashboard, dashboard)
		}
		main = main.Child(dashboard)
		return main
	}

	main = main.Child(buildPageHeader(snapshot))

	var routeBody *ui.Element
	switch snapshot.CurrentAppRoute {
	case RouteTerminal:
		routeBody = buildTerminalPanel(snapshot)
	case RouteFileMgr:
		routeBody = buildFileManagerPanel(snapshot)
	case RouteSettings:
		routeBody = buildSettingsOnlyRoot(snapshot)
	case RouteLogs:
		routeBody = buildLogsOnlyRoot(snapshot)
	default:
		routeBody = buildHomeDashboard(snapshot)
	}
	if animateRoute {
		main = main.Child(applyPageMotion(snapshot.CurrentAppRoute, routeBody))
	} else {
		main = main.Child(routeBody)
	}

	if snapshot.CurrentAppRoute != RouteLogs {
		resultPanel := buildResultPanel(snapshot)
		if animateRoute {
			resultPanel = applySectionMotion(resultPanel, sectionAnimationDelayMs)
		}
		main = main.Child(resultPanel)
	}

	return main
}

func shouldAnimateRoute(currentRoute string) bool {
	uiRootMu.Lock()
	defer uiRootMu.Unlock()
	if uiLastRenderedRoute != currentRoute {
		uiLastRenderedRoute = currentRoute
		return true
	}
	return false
}

func buildPageHeader(snapshot DebugState) *ui.Element {
	title := map[string]string{
		RouteTerminal: "终端",
		RouteFileMgr:  "文件管理",
		RouteSettings: "核心设置",
		RouteLogs:     "系统日志",
	}[snapshot.CurrentAppRoute]
	if title == "" {
		title = "VelaSU"
	}

	row := makeRow().
		WidthFull().
		AlignCenter().
		Gap(8).
		Padding(8).
		Bg("#11182C").
		Border(1, "#27324A").
		Radius(10)

	row = row.
		Child(buildOptionPill(IconSVGBack, "返回", EventRouteDashboard)).
		Child(makeSectionTitle(title).Size(18)).
		Child(makeSpacer())

	if snapshot.Pending != nil {
		row = row.Child(makeBadge("请求中"))
	}
	if snapshot.SelectedDeviceAddr != "" {
		row = row.Child(makeBadge(snapshot.SelectedDeviceAddr))
	}
	return row
}

func buildHomeDashboard(snapshot DebugState) *ui.Element {
	root := makeColumn().WidthFull().HeightFull().AlignCenter().JustifyCenter().Gap(40).PaddingTop(60).PaddingBottom(60)
	ready := dependencyReady(snapshot)

	statusText := "设备未连接"
	statusColor := "#F87171"
	statusBg := "#2A1616"
	statusIcon := IconSVGDevice
	if snapshot.SelectedDeviceAddr != "" {
		statusText = "已连接设备: " + formatSelectedDevice(snapshot)
		statusColor = "#4ADE80"
		statusBg = "#142A1E"
	} else if len(snapshot.ConnectedDevices) > 0 {
		statusText = fmt.Sprintf("发现可用设备: %d 台", len(snapshot.ConnectedDevices))
		statusColor = "#FACC15"
		statusBg = "#2A2612"
	}

	depText := fallback(snapshot.DependencyMessage, "未检查依赖状态")
	depColor := "#FACC15"
	depBg := "#2A2612"
	depIcon := IconSVGSettings
	if ready {
		depColor = "#4ADE80"
		depBg = "#142A1E"
		depText = "依赖正常"
	}

	badges := makeRow().AlignCenter().JustifyCenter().Gap(16).
		Child(makeStatusBadge(statusIcon, statusText, statusColor, statusBg)).
		Child(makeStatusBadge(depIcon, depText, depColor, depBg))

	terminalBox := makeAppIcon(IconSVGTerminal, "终端", "#1F2937", "#4ADE80", EventRouteTerminal)
	fileBox := makeAppIcon(IconSVGFolder, "文件管理", "#3B82F6", "#FFFFFF", EventRouteFileMgr)
	settingsBox := makeAppIcon(IconSVGSettings, "核心设置", "#4B5563", "#FFFFFF", EventRouteSettings)
	logsBox := makeAppIcon(IconSVGLogs, "系统日志", "#111827", "#FACC15", EventRouteLogs)

	if !ready {
		terminalBox = terminalBox.Disabled().Opacity(0.45)
		fileBox = fileBox.Disabled().Opacity(0.45)
	}

	apps := makeRow().AlignCenter().JustifyCenter().Gap(48).MarginTop(40).MarginBottom(40).
		Child(terminalBox).Child(fileBox).Child(settingsBox).Child(logsBox)

	refreshIcon := `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>`

	
	btnRefreshDevice := makeOutlinedButtonWithIcon(refreshIcon, "刷新设备", EventDeviceRefresh)
	btnRefreshDep := makeOutlinedButtonWithIcon(IconSVGHardDrive, "手动刷新依赖", EventDependencyRefresh)
	btnLaunch := makePrimaryButtonWithIcon(IconSVGPlay, "一键启动", EventLaunchQA)
	if !ready {
		btnLaunch = btnLaunch.Disabled().Opacity(0.6)
	}

	buttons := makeRow().AlignCenter().JustifyCenter().Gap(16).
		Child(btnRefreshDevice).Child(btnRefreshDep).Child(btnLaunch)

	bottomText := makeRow().AlignCenter().JustifyCenter().Gap(12).MarginTop(16).
		Child(makeMutedText("目标表盘 " + TargetWatchfaceName + ": " + mapInstallStatus(snapshot.TargetWatchfaceFound))).
		Child(makeMutedText("|")).
		Child(makeMutedText("目标快应用 " + TargetPackageName + ": " + mapInstallStatus(snapshot.TargetQuickAppFound)))

	root = root.Child(badges).Child(apps).Child(buttons).Child(bottomText)
	return root
}

func makeStatusBadge(icon, text, color, bgColor string) *ui.Element {
	return makeRow().
		AlignCenter().Gap(6).
		Bg(bgColor).
		Radius(20).
		PaddingTop(6).PaddingBottom(6).PaddingLeft(16).PaddingRight(16).
		Border(1, color).
		Child(el(ui.ElementTypeSvg, strings.Replace(icon, `width="16" height="16"`, `width="14" height="14"`, 1)).TextColor(color)).
		Child(makeText(text).Size(13).TextColor(color))
}

func makeAppIcon(iconStr, title, bgColor, iconColor, eventID string) *ui.Element {
	bigIconStr := strings.Replace(iconStr, `width="16" height="16"`, `width="48" height="48"`, 1)
	bigIconStr = strings.Replace(bigIconStr, `width="20" height="20"`, `width="48" height="48"`, 1)

	iconBox := makeColumn().
		AlignCenter().JustifyCenter().
		Width(100).Height(100).
		Radius(24).
		Bg(bgColor).
		Border(1, "#374151").
		Child(el(ui.ElementTypeSvg, bigIconStr).TextColor(iconColor))

	content := makeColumn().
		AlignCenter().JustifyCenter().Gap(12).
		Child(iconBox).
		Child(makeText(title).Size(15).TextColor("#E8ECF8"))

	return applyButtonMotion(
		el(ui.ElementTypeButton, "").
			WithoutDefaultStyles().
			On(ui.EventClick, eventID).
			Child(content),
	)
}

func makeOutlinedButtonWithIcon(iconStr, label, eventID string) *ui.Element {
	content := makeRow().AlignCenter().Gap(8).
		Child(el(ui.ElementTypeSvg, iconStr).TextColor("#D4DAEE")).
		Child(makeText(label).Size(14).TextColor("#D4DAEE"))

	return applyButtonMotion(
		el(ui.ElementTypeButton, "").
			WithoutDefaultStyles().
			Bg("#252D44").
			Border(1, "#394360").
			Radius(8).
			PaddingTop(8).PaddingBottom(8).PaddingLeft(16).PaddingRight(16).
			On(ui.EventClick, eventID).
			Child(content),
	)
}

func makePrimaryButtonWithIcon(iconStr, label, eventID string) *ui.Element {
	content := makeRow().AlignCenter().Gap(8).
		Child(el(ui.ElementTypeSvg, iconStr).TextColor("#FFFFFF")).
		Child(makeText(label).Size(14).TextColor("#FFFFFF"))

	return applyButtonMotion(
		el(ui.ElementTypeButton, "").
			WithoutDefaultStyles().
			Bg("#4A6CF7").
			Border(1, "#5B7AF8").
			Radius(8).
			PaddingTop(8).PaddingBottom(8).PaddingLeft(16).PaddingRight(16).
			On(ui.EventClick, eventID).
			Child(content),
	)
}

func buildResultPanel(snapshot DebugState) *ui.Element {
	panel := makePanel().
		WidthFull().
		MinWidth(0).
		Flex().
		FlexDirection(ui.FlexDirectionColumn).
		Gap(6).
		Bg("#0F1424").
		Padding(12).
		Child(makeSectionTitle("请求结果").MarginRight(4)).
		Child(makeText(fmt.Sprintf("请求: id=%s method=%s", fallback(snapshot.LastRequestID, "-"), fallback(snapshot.LastRequestMethod, "-"))).MarginTop(8).MarginRight(4)).
		Child(makeText("状态: " + fallback(snapshot.LastResponseStatus, "idle")).MarginTop(4).MarginRight(4)).
		Child(makeText(fmt.Sprintf("耗时: %d ms", snapshot.LastLatencyMs)).MarginTop(4).MarginRight(4)).
		Child(makeText("错误: " + fallback(snapshot.LastError, "(none)")).MarginTop(4))

	scroll := el(ui.ElementTypeScrollArea, "").
		WidthFull().
		MinWidth(0).
		MaxHeight(280).
		PaddingTop(8)

	content := makeColumn().
		WidthFull().
		MinWidth(0).
		Gap(8).
		Child(makeMutedText("原始响应")).
		Child(makeCodeBlock(clipPanelText(snapshot.LastResponseRaw, 1800)).MinWidth(0)).
		Child(makeMutedText("格式化响应")).
		Child(makeCodeBlock(clipPanelText(snapshot.LastResponsePretty, 1800)).MinWidth(0))

	scroll = scroll.Child(content)
	panel = panel.Child(scroll)
	return panel
}

func clipPanelText(value string, max int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "(empty)"
	}
	if len(value) <= max {
		return value
	}
	return value[:max] + "\n...(truncated)"
}

func mapInstallStatus(installed bool) string {
	if installed {
		return "已安装"
	}
	return "未安装"
}
