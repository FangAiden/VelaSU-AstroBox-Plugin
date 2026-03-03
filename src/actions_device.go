package plugin

import (
	"fmt"
	"strings"

	device "astroboxplugin/bindings/astrobox_psys_host_device"
	thirdpartyapp "astroboxplugin/bindings/astrobox_psys_host_thirdpartyapp"
)

func actionDeviceRefresh() {
	connected := device.GetConnectedDeviceList().Read()
	snapshot := readStateSnapshot()
	if !deviceExists(connected, snapshot.SelectedDeviceAddr) {
		clearPendingRequest("设备状态变化，已取消当前请求")
		resetTaskQueue()
	}
	withState(func(state *DebugState) {
		state.ConnectedDevices = connected
		if !deviceExistsLocked(state, state.SelectedDeviceAddr) {
			clearDeviceSelectionLocked(state)
		}
	})
	appendLogf("INFO", "设备刷新完成，数量=%d", len(connected))
}

func actionDeviceSelect(addr string) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		appendLog("ERROR", "设备地址为空")
		return
	}

	snapshot := readStateSnapshot()
	var picked device.DeviceInfo
	found := false
	for _, d := range snapshot.ConnectedDevices {
		if d.Addr == addr {
			picked = d
			found = true
			break
		}
	}
	if !found {
		appendLog("ERROR", "设备不存在，请先刷新设备列表")
		return
	}

	clearPendingRequest("切换设备，已取消当前请求")
	resetTaskQueue()

	withState(func(state *DebugState) {
		state.SelectedDeviceAddr = picked.Addr
		state.SelectedDeviceName = picked.Name
		state.RegisteredDeviceAddr = ""
		state.LastResponseStatus = "idle"
		state.LastResponseRaw = ""
		state.LastResponsePretty = ""
		state.LastRequestID = ""
		state.LastRequestMethod = ""
		state.LastLatencyMs = 0
		state.LastError = ""
		state.FileCurrentDir = DefaultFileDir
		state.FileDirInput = DefaultFileDir
		state.FileSearchQuery = ""
		state.FileViewMode = FileViewGrid
		state.FileSortMode = FileSortByName
		state.FileEntries = nil
		state.FileVisibleCount = DefaultDirPageSize
		state.FileSelectedPaths = nil
	})
	appendLogf("INFO", "已选择设备: %s (%s)", picked.Name, picked.Addr)
}

func actionRegisterInterconnect() {
	if err := ensureRegisteredForCurrentDevice(); err != nil {
		appendLogf("ERROR", "注册接收失败: %v", err)
		alertDialog("注册失败", fmt.Sprintf("服务注册失败: %v", err))
		return
	}
	appendLog("INFO", "已完成接收注册")
	alertDialog("注册成功", "服务注册完成，可以开始通信")
}

func actionHello() {
	EnqueueRpcTask("hello", func() error {
		return RpcHelloWithCallback(func(result HelloResult, err error) {
			if err != nil {
				appendLogf("ERROR", "hello 失败: %v", err)
				alertDialog("连接失败", fmt.Sprintf("测试连接失败: %v", err))
				return
			}
			appendLogf("INFO", "连接成功，remoteEnabled=%t hasToken=%t protocol=%d", result.RemoteEnabled, result.HasToken, result.Protocol)
			alertDialog("连接成功", fmt.Sprintf("服务端已响应\nprotocol=%d\nremoteEnabled=%t\nhasToken=%t", result.Protocol, result.RemoteEnabled, result.HasToken))
		})
	})
}

func actionLaunchQA() {
	snapshot := readStateSnapshot()
	if snapshot.SelectedDeviceAddr == "" {
		appendLog("WARN", "请先在“核心配置”中选择设备后再拉起快应用")
		alertDialog("无法启动", "请先在核心配置中选择设备")
		return
	}

	appendLog("INFO", "正在获取已安装应用列表...")
	listResult := thirdpartyapp.GetThirdpartyAppList(snapshot.SelectedDeviceAddr).Read()
	if listResult.IsErr() {
		appendLog("ERROR", "获取应用列表失败")
		alertDialog("启动失败", "获取设备应用列表失败")
		return
	}

	appList := listResult.Ok()
	var targetApp *thirdpartyapp.AppInfo
	for i := range appList {
		if appList[i].PackageName == TargetPackageName {
			targetApp = &appList[i]
			break
		}
	}
	if targetApp == nil {
		appendLogf("ERROR", "设备上未找到应用 %s，请确认已安装", TargetPackageName)
		alertDialog("启动失败", fmt.Sprintf("设备上未找到应用\n%s", TargetPackageName))
		return
	}

	appendLogf("INFO", "找到应用 %s (v%d)，正在拉起...", targetApp.AppName, targetApp.VersionCode)
	launchResult := thirdpartyapp.LaunchQa(snapshot.SelectedDeviceAddr, *targetApp, "").Read()
	if launchResult.IsOk() {
		appendLog("INFO", "快应用启动成功")
		alertDialog("启动成功", fmt.Sprintf("已启动 %s 快应用", targetApp.AppName))
	} else {
		appendLog("ERROR", "快应用启动失败")
		alertDialog("启动失败", "快应用启动失败，请检查设备状态")
	}
}

func deviceExistsLocked(state *DebugState, addr string) bool {
	if strings.TrimSpace(addr) == "" {
		return false
	}
	for _, d := range state.ConnectedDevices {
		if d.Addr == addr {
			return true
		}
	}
	return false
}

func deviceExists(devices []device.DeviceInfo, addr string) bool {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return false
	}
	for _, d := range devices {
		if d.Addr == addr {
			return true
		}
	}
	return false
}

func clearDeviceSelectionLocked(state *DebugState) {
	state.SelectedDeviceAddr = ""
	state.SelectedDeviceName = ""
	state.RegisteredDeviceAddr = ""
	state.LastRequestID = ""
	state.LastRequestMethod = ""
	state.LastResponseStatus = "idle"
	state.LastResponseRaw = ""
	state.LastResponsePretty = ""
	state.LastLatencyMs = 0
	state.LastError = ""
	state.FileCurrentDir = DefaultFileDir
	state.FileDirInput = DefaultFileDir
	state.FileSearchQuery = ""
	state.FileViewMode = FileViewGrid
	state.FileSortMode = FileSortByName
	state.FileEntries = nil
	state.FileVisibleCount = 0
	state.FileSelectedPaths = nil
	state.FileEditorText = ""
	state.FileEditorHexPreview = ""
	state.FileEditorIsBinary = false
}
