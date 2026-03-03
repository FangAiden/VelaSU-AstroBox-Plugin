package plugin

import (
	"fmt"

	ui "astroboxplugin/bindings/astrobox_psys_host_ui_v3"
)

func buildSettingsOnlyRoot(snapshot DebugState) *ui.Element {
	root := makeColumn().WidthFull().Gap(10)

	root = root.Child(makePanel().
		Bg("#10172A").
		Padding(12).
		Child(makeSectionTitle("设备状态")).
		Child(makeMutedText("当前设备: " + formatSelectedDevice(snapshot)).MarginTop(6)).
		Child(makeMutedText("已连接设备数: " + itoa(len(snapshot.ConnectedDevices))).MarginTop(4)))

	tokenCard := makePanel().
		Bg("#10172A").
		Padding(12).
		Child(makeRow().AlignCenter().Gap(8).
			Child(makeSVGIcon(IconSVGKey)).
			Child(makeSectionTitle("安全授权"))).
		Child(makeMutedText("绑定 4 位 Token 以保护 RPC 调用").MarginTop(6)).
		Child(makeSingleLineInput(snapshot.TokenInput, EventTokenInput).MarginTop(8))
	root = root.Child(tokenCard)

	actions := makePanel().
		Bg("#10172A").
		Padding(12).
		Child(makeSectionTitle("连接操作")).
		Child(makeRow().Gap(8).MarginTop(8).
			Child(makeSecondaryButton("刷新设备", EventDeviceRefresh).FlexGrow(1)).
			Child(makeSecondaryButton("启动快应用", EventLaunchQA).FlexGrow(1))).
		Child(makeRow().Gap(8).MarginTop(8).
			Child(makeSecondaryButton("注册服务", EventRegisterInterconnect).FlexGrow(1)).
			Child(makePrimaryButton("测试连接", EventHello).FlexGrow(1)))
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
