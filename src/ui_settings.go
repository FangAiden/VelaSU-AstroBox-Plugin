package plugin

import (
	"fmt"

	ui "astroboxplugin/bindings/astrobox_psys_host_ui_v3"
)

func buildSettingsOnlyRoot(snapshot DebugState) *ui.Element {
	root := makeColumn().WidthFull().Gap(10)
	ready := dependencyReady(snapshot)

	statusColor := "#F9D17E"
	if ready {
		statusColor = "#87E9C6"
	}

	devicePanel := makePanel().
		Bg("#10172A").
		Padding(12).
		Child(makeRow().WidthFull().Child(makeSectionTitle("设备状态"))).
		Child(
			makeColumn().
				WidthFull().
				Gap(4).
				MarginTop(6).
				Child(makeMutedText("当前设备: " + formatSelectedDevice(snapshot))).
				Child(makeMutedText("已连接设备数: " + itoa(len(snapshot.ConnectedDevices)))).
				Child(makeText("依赖状态: " + fallback(snapshot.DependencyMessage, "未检查依赖状态")).TextColor(statusColor)),
		)
	root = root.Child(devicePanel)

	depWatchface := "未安装"
	if snapshot.TargetWatchfaceFound {
		depWatchface = "已安装"
	}
	depWatchfaceCurrent := "否"
	if snapshot.TargetWatchfaceOn {
		depWatchfaceCurrent = "是"
	}
	depQuickApp := "未安装"
	if snapshot.TargetQuickAppFound {
		depQuickApp = "已安装"
	}

	depsPanel := makePanel().
		Bg("#10172A").
		Padding(12).
		Child(makeRow().WidthFull().Child(makeSectionTitle("目标依赖"))).
		Child(
			makeColumn().
				WidthFull().
				Gap(4).
				MarginTop(6).
				Child(makeMutedText("目标表盘名称: " + TargetWatchfaceName)).
				Child(makeMutedText("表盘安装状态: " + depWatchface)).
				Child(makeMutedText("是否当前表盘: " + depWatchfaceCurrent)).
				Child(makeMutedText("目标快应用包名: " + TargetPackageName).MarginTop(4)).
				Child(makeMutedText("快应用安装状态: " + depQuickApp)),
		)
	root = root.Child(depsPanel)

	tokenCard := makePanel().
		Bg("#10172A").
		Padding(12).
		Child(makeRow().AlignCenter().Gap(8).
			Child(makeSVGIcon(IconSVGKey)).
			Child(makeSectionTitle("安全授权"))).
		Child(makeMutedText("绑定 4 位 Token 以保护 RPC 调用").MarginTop(6)).
		Child(makeSingleLineInput(snapshot.TokenInput, EventTokenInput).MarginTop(8))
	root = root.Child(tokenCard)

	oneClick := makePrimaryButton("一键启动", EventLaunchQA).Width(0).MinWidth(0).FlexGrow(1)
	if !snapshot.TargetWatchfaceFound || !snapshot.TargetQuickAppFound {
		oneClick = oneClick.Disabled().Opacity(0.6)
	}
	helloBtn := makePrimaryButton("测试连接", EventHello).Width(0).MinWidth(0).FlexGrow(1)
	registerBtn := makeSecondaryButton("注册服务", EventRegisterInterconnect).Width(0).MinWidth(0).FlexGrow(1)
	if !ready {
		helloBtn = helloBtn.Disabled().Opacity(0.6)
		registerBtn = registerBtn.Disabled().Opacity(0.6)
	}

	actions := makePanel().
		Bg("#10172A").
		Padding(12).
		Child(makeSectionTitle("连接操作")).
		Child(
			el(ui.ElementTypeGrid, "").
				WidthFull().
				GridTemplateColumns("repeat(auto-fit, minmax(160px, 1fr))").
				Gap(8).
				MarginTop(8).
				Child(makeSecondaryButton("刷新设备", EventDeviceRefresh).WidthFull().MinWidth(0)).
				Child(makeSecondaryButton("手动刷新依赖", EventDependencyRefresh).WidthFull().MinWidth(0)),
		).
		Child(
			el(ui.ElementTypeGrid, "").
				WidthFull().
				GridTemplateColumns("repeat(auto-fit, minmax(160px, 1fr))").
				Gap(8).
				MarginTop(8).
				Child(oneClick.WidthFull().MinWidth(0)).
				Child(registerBtn.WidthFull().MinWidth(0)),
		).
		Child(
			el(ui.ElementTypeGrid, "").
				WidthFull().
				GridTemplateColumns("repeat(auto-fit, minmax(160px, 1fr))").
				Gap(8).
				MarginTop(8).
				Child(helloBtn.WidthFull().MinWidth(0)),
		)
	if !ready {
		actions = actions.Child(makeMutedText("依赖未满足，核心功能已禁用。").MarginTop(8))
	}
	root = root.Child(actions)

	if len(snapshot.ConnectedDevices) > 0 {
		deviceList := makePanel().
			Bg("#10172A").
			Padding(12).
			Child(makeSectionTitle("设备切换"))
		for _, d := range snapshot.ConnectedDevices {
			label := d.Name
			if label == "" {
				label = d.Addr
			}
			subtitle := d.Addr
			if d.Addr == snapshot.SelectedDeviceAddr {
				subtitle += " (当前)"
			}
			deviceList = deviceList.Child(
				makeSettingCard(IconSVGDevice, label, subtitle, makeSecondaryButton("切换", deviceSelectEventID(d.Addr))),
			)
		}
		root = root.Child(deviceList)
	}

	return root
}

func itoa(v int) string {
	return fmt.Sprintf("%d", v)
}
