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
		Padding(12).
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
	root := makeColumn().WidthFull().Gap(16)

	header := makePanel().
		Bg("#10172A").
		Padding(14).
		Child(makeTitle("VelaSU 控制台").Size(26).MarginRight(4)).
		Child(makeMutedText("Interconnect 远程终端与文件管理").MarginTop(4))

	statusText := "未连接设备"
	statusColor := "#93A0BE"
	if snapshot.SelectedDeviceAddr != "" {
		statusText = "已连接设备: " + formatSelectedDevice(snapshot)
		statusColor = "#87E9C6"
	} else if len(snapshot.ConnectedDevices) > 0 {
		statusText = fmt.Sprintf("发现可用设备: %d 台", len(snapshot.ConnectedDevices))
		statusColor = "#F9D17E"
	}

	header = header.Child(
		makeRow().
			AlignCenter().
			MarginTop(10).
			Child(makeSVGIcon(IconSVGDevice)).
			Child(makeText(statusText).TextColor(statusColor).MarginLeft(6)),
	)

	grid := el(ui.ElementTypeGrid, "").
		WidthFull().
		GridTemplateColumns("repeat(auto-fill, minmax(180px, 1fr))").
		Gap(10).
		Child(makeDashboardCard(IconSVGTerminal, "终端", "命令执行与输入输出", EventRouteTerminal).WidthFull().MinHeight(130).Margin(0)).
		Child(makeDashboardCard(IconSVGFolderOpen, "文件管理", "浏览与编辑远端文件", EventRouteFileMgr).WidthFull().MinHeight(130).Margin(0)).
		Child(makeDashboardCard(IconSVGSettings, "核心设置", "设备连接与配置项", EventRouteSettings).WidthFull().MinHeight(130).Margin(0)).
		Child(makeDashboardCard(IconSVGLogs, "系统日志", "调试日志与导出", EventRouteLogs).WidthFull().MinHeight(130).Margin(0))

	quick := makePanel().
		Bg("#10172A").
		Padding(12).
		Child(makeSectionTitle("快速操作")).
		Child(
			makeRow().
				Gap(8).
				MarginTop(8).
				Child(makeSecondaryButton("刷新设备", EventDeviceRefresh)).
				Child(makeSecondaryButton("连接测试", EventHello)).
				Child(makeSecondaryButton("启动快应用", EventLaunchQA)),
		)

	root = root.Child(header).Child(grid).Child(quick)
	return root
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
