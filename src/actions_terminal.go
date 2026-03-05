package plugin

import (
	"fmt"
	"strings"

	ui "astroboxplugin/bindings/astrobox_psys_host_ui_v3"
)

func actionGetCwd() {
	EnqueueRpcTask("shell.getCwd", func() error {
		return RpcShellGetCwd(func(result ShellCwdResult, err error) {
			if err != nil {
				appendLogf("ERROR", "获取 cwd 失败: %v", err)
				return
			}
			withState(func(state *DebugState) {
				state.CurrentCwdInput = result.Cwd
			})
		})
	})
}

func actionSetCwd() {
	cwd := readState(func(state DebugState) string {
		return strings.TrimSpace(state.CurrentCwdInput)
	})
	if cwd == "" {
		appendLog("ERROR", "cwd 不能为空")
		return
	}
	EnqueueRpcTask("shell.setCwd", func() error {
		return RpcShellSetCwd(cwd, func(result ShellCwdResult, err error) {
			if err != nil {
				appendLogf("ERROR", "设置 cwd 失败: %v", err)
				return
			}
			withState(func(state *DebugState) {
				state.CurrentCwdInput = result.Cwd
				state.FileCurrentDir = result.Cwd
				state.FileDirInput = result.Cwd
			})
		})
	})
}

func actionExecPreset(cmd string) {
	withState(func(state *DebugState) {
		state.CurrentCommand = cmd
	})
	actionExecCommand()
}

func actionExecCommand() {
	cmd := readState(func(state DebugState) string {
		return strings.TrimSpace(state.CurrentCommand)
	})
	if cmd == "" {
		appendLog("ERROR", "命令不能为空")
		return
	}

	token := readState(func(state DebugState) string {
		return strings.TrimSpace(state.TokenInput)
	})
	if token == "" {
		withState(func(state *DebugState) {
			state.TerminalBuffer = append(state.TerminalBuffer,
				"[提示] 尚未配置 Token，命令未发送。",
				"请先进入“核心配置”填写 4 位 Token。",
			)
		})
		appendLog("WARN", "终端命令未发送：Token 为空")
		return
	}
	if !tokenPattern.MatchString(token) {
		withState(func(state *DebugState) {
			state.TerminalBuffer = append(state.TerminalBuffer,
				"[提示] Token 格式无效，命令未发送。",
				"请填写 4 位数字 Token。",
			)
		})
		appendLog("WARN", "终端命令未发送：Token 格式无效")
		return
	}

	withState(func(state *DebugState) {
		state.TerminalBuffer = append(state.TerminalBuffer, "$ "+cmd)
		allLines := strings.Join(state.TerminalBuffer, "\n")
		lineCount := len(strings.Split(allLines, "\n"))
		lastPage := (lineCount+TerminalPageSize-1)/TerminalPageSize - 1
		if lastPage < 0 {
			lastPage = 0
		}
		state.TerminalPage = lastPage
		state.CurrentCommand = ""
	})
	EnqueueRpcTask("terminal.exec", func() error {
		return RpcShellExec(cmd, DefaultShellTimeoutMs, func(result ShellExecResult, err error) {
			if err != nil {
				appendLogf("ERROR", "执行失败: %v", err)
				withState(func(state *DebugState) {
					state.TerminalBuffer = append(state.TerminalBuffer, fmt.Sprintf("Error: %v", err))
				})
				return
			}
		})
	})
}

func actionTerminalClear() {
	withState(func(state *DebugState) {
		state.TerminalBuffer = nil
		state.TerminalPage = 0
	})
	appendLog("INFO", "终端输出已清空")
}

func actionTerminalRunHistory(index int) {
	if index < 0 {
		return
	}
	snapshot := readStateSnapshot()
	if index >= len(snapshot.TerminalHistory) {
		return
	}
	cmd := snapshot.TerminalHistory[index].Command
	withState(func(state *DebugState) {
		state.CurrentCommand = cmd
	})
	actionExecCommand()
}

func actionTerminalExportText() {
	snapshot := readStateSnapshot()
	text := strings.Join(snapshot.TerminalBuffer, "\n")
	ui.RenderToTextCard("终端输出记录", text)
	appendLog("INFO", "已导出终端内容供复制")
}

func enqueueShellCommandTask(name string, cmd string, onSuccess func(ShellExecResult)) {
	EnqueueRpcTask(name, func() error {
		return RpcShellExec(cmd, DefaultShellTimeoutMs, func(result ShellExecResult, err error) {
			if err != nil {
				appendLogf("ERROR", "%s 执行失败: %v", name, err)
				return
			}
			if execErr := shellExecNonZeroErr(result); execErr != nil {
				appendLogf("ERROR", "%s 执行失败: cmd=%s %v", name, cmd, execErr)
				return
			}
			if onSuccess != nil {
				onSuccess(result)
			}
		})
	})
}

func shellExecNonZeroErr(result ShellExecResult) error {
	if result.ExitCode == nil || *result.ExitCode == 0 {
		return nil
	}
	output := normalizeInline(result.Output)
	if output == "" {
		return fmt.Errorf("exitCode=%d", *result.ExitCode)
	}
	const maxOutput = 240
	if len(output) > maxOutput {
		output = output[:maxOutput] + "..."
	}
	return fmt.Errorf("exitCode=%d output=%s", *result.ExitCode, output)
}
