package plugin

import (
	"fmt"
	"strings"

	device "astroboxplugin/bindings/astrobox_psys_host_device"
	thirdpartyapp "astroboxplugin/bindings/astrobox_psys_host_thirdpartyapp"
	watchface "astroboxplugin/bindings/astrobox_psys_host_watchface"
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

	snapshot = readStateSnapshot()
	if snapshot.SelectedDeviceAddr != "" {
		refreshDependencyStatus(snapshot.SelectedDeviceAddr)
	}
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
		state.FileSortAsc = true
		state.FileEntries = nil
		state.FileVisibleCount = DefaultDirPageSize
		state.FileSelectedPaths = nil
		resetDependencyStateLocked(state, "未检查依赖状态")
	})
	appendLogf("INFO", "已选择设备: %s (%s)", picked.Name, picked.Addr)
	refreshDependencyStatus(picked.Addr)
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

func actionDependencyRefresh() {
	snapshot := readStateSnapshot()
	if snapshot.SelectedDeviceAddr == "" {
		appendLog("WARN", "请先在设置中选择设备")
		alertDialog("无法刷新", "请先在设置中选择设备")
		return
	}
	if refreshDependencyStatus(snapshot.SelectedDeviceAddr) {
		appendLog("INFO", "依赖检查通过，可使用 VelaSU")
		return
	}
	latest := readStateSnapshot()
	appendLogf("WARN", "依赖检查未通过: %s", latest.DependencyMessage)
}

func actionLaunchQA() {
	snapshot := readStateSnapshot()
	if snapshot.SelectedDeviceAddr == "" {
		appendLog("WARN", "请先在设置中选择设备后再启动")
		alertDialog("无法启动", "请先在设置中选择设备")
		return
	}

	ready := refreshDependencyStatus(snapshot.SelectedDeviceAddr)
	latest := readStateSnapshot()
	if !ready {
		appendLogf("ERROR", "启动前检查失败: %s", latest.DependencyMessage)
		alertDialog("无法启动", latest.DependencyMessage)
		return
	}

	if !latest.TargetWatchfaceOn {
		setResult := watchface.SetCurrentWatchface(snapshot.SelectedDeviceAddr, latest.TargetWatchfaceID).Read()
		if setResult.IsErr() {
			appendLog("ERROR", "设置目标表盘失败")
			alertDialog("启动失败", "设置目标表盘失败")
			return
		}
		appendLogf("INFO", "已设置目标表盘: %s", TargetWatchfaceName)
	}

	listResult := thirdpartyapp.GetThirdpartyAppList(snapshot.SelectedDeviceAddr).Read()
	if listResult.IsErr() {
		appendLog("ERROR", "获取应用列表失败")
		alertDialog("启动失败", "获取应用列表失败")
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
		appendLogf("ERROR", "设备上未找到应用 %s", TargetPackageName)
		alertDialog("启动失败", fmt.Sprintf("设备上未找到应用\n%s", TargetPackageName))
		_ = refreshDependencyStatus(snapshot.SelectedDeviceAddr)
		return
	}

	launchResult := thirdpartyapp.LaunchQa(snapshot.SelectedDeviceAddr, *targetApp, "").Read()
	if launchResult.IsOk() {
		appendLogf("INFO", "已启动快应用: %s", targetApp.AppName)
		alertDialog("启动成功", fmt.Sprintf("已设置表盘并启动应用\n%s", targetApp.AppName))
		_ = refreshDependencyStatus(snapshot.SelectedDeviceAddr)
		return
	}

	appendLog("ERROR", "快应用启动失败")
	alertDialog("启动失败", "快应用启动失败，请检查设备状态")
}

func refreshDependencyStatus(addr string) bool {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		withState(func(state *DebugState) {
			resetDependencyStateLocked(state, "未选择设备")
		})
		return false
	}

	watchfaceResult := watchface.GetWatchfaceList(addr).Read()
	if watchfaceResult.IsErr() {
		withState(func(state *DebugState) {
			resetDependencyStateLocked(state, "获取表盘列表失败")
		})
		return false
	}
	watchfaceList := watchfaceResult.Ok()
	var targetWatchface *watchface.WatchfaceInfo
	for i := range watchfaceList {
		if strings.EqualFold(strings.TrimSpace(watchfaceList[i].Name), TargetWatchfaceName) {
			targetWatchface = &watchfaceList[i]
			break
		}
	}

	appResult := thirdpartyapp.GetThirdpartyAppList(addr).Read()
	if appResult.IsErr() {
		withState(func(state *DebugState) {
			resetDependencyStateLocked(state, "获取快应用列表失败")
			if targetWatchface != nil {
				state.TargetWatchfaceID = targetWatchface.Id
				state.TargetWatchfaceFound = true
				state.TargetWatchfaceOn = targetWatchface.IsCurrent
			}
		})
		return false
	}
	appList := appResult.Ok()
	targetAppFound := false
	for i := range appList {
		if appList[i].PackageName == TargetPackageName {
			targetAppFound = true
			break
		}
	}

	targetWatchfaceFound := targetWatchface != nil
	targetWatchfaceOn := targetWatchface != nil && targetWatchface.IsCurrent
	targetWatchfaceID := ""
	if targetWatchface != nil {
		targetWatchfaceID = targetWatchface.Id
	}

	message := "依赖已就绪"
	switch {
	case !targetWatchfaceFound && !targetAppFound:
		message = fmt.Sprintf("缺少目标表盘 %s 与快应用 %s", TargetWatchfaceName, TargetPackageName)
	case !targetWatchfaceFound:
		message = fmt.Sprintf("缺少目标表盘 %s", TargetWatchfaceName)
	case !targetAppFound:
		message = fmt.Sprintf("缺少目标快应用 %s", TargetPackageName)
	}

	withState(func(state *DebugState) {
		state.DependencyChecked = true
		state.DependencyMessage = message
		state.TargetWatchfaceID = targetWatchfaceID
		state.TargetWatchfaceFound = targetWatchfaceFound
		state.TargetWatchfaceOn = targetWatchfaceOn
		state.TargetQuickAppFound = targetAppFound
	})

	appendLogf(
		"INFO",
		"依赖状态: watchface(%s)=%t current=%t, quickapp(%s)=%t",
		TargetWatchfaceName,
		targetWatchfaceFound,
		targetWatchfaceOn,
		TargetPackageName,
		targetAppFound,
	)

	return targetWatchfaceFound && targetAppFound
}

func dependencyReady(snapshot DebugState) bool {
	if strings.TrimSpace(snapshot.SelectedDeviceAddr) == "" {
		return false
	}
	if !snapshot.DependencyChecked {
		return false
	}
	return snapshot.TargetWatchfaceFound && snapshot.TargetQuickAppFound
}

func dependencyBlockedReason(snapshot DebugState) string {
	if strings.TrimSpace(snapshot.SelectedDeviceAddr) == "" {
		return "请先选择设备"
	}
	if !snapshot.DependencyChecked {
		return "请先刷新依赖状态"
	}
	if snapshot.DependencyMessage != "" {
		return snapshot.DependencyMessage
	}
	return "依赖状态未知"
}

func ensureDependencyReadyForUsage() bool {
	snapshot := readStateSnapshot()
	if dependencyReady(snapshot) {
		return true
	}
	reason := dependencyBlockedReason(snapshot)
	appendLogf("WARN", "依赖未满足，拒绝继续操作: %s", reason)
	alertDialog("无法使用 VelaSU", reason)
	return false
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

func resetDependencyStateLocked(state *DebugState, message string) {
	state.DependencyChecked = false
	state.DependencyMessage = message
	state.TargetWatchfaceID = ""
	state.TargetWatchfaceFound = false
	state.TargetWatchfaceOn = false
	state.TargetQuickAppFound = false
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
	state.FileSortAsc = true
	state.FileEntries = nil
	state.FileVisibleCount = 0
	state.FileSelectedPaths = nil
	state.FileEditorText = ""
	state.FileEditorHexPreview = ""
	state.FileEditorIsBinary = false
	resetDependencyStateLocked(state, "未选择设备")
}
