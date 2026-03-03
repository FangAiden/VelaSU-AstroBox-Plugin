package plugin

import "sync"

var lifecycleOnce sync.Once

func Init() {
	lifecycleOnce.Do(func() {
		initLogger()
		OnLoad()
	})
}

func OnLoad() {
	initDebugState()
	appendLog("INFO", "ShellBridge Interconnect 插件已加载")
}
