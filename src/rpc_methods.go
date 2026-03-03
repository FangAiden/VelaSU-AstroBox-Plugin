package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	register "astroboxplugin/bindings/astrobox_psys_host_register"
)

func SendProtectedRpc(method string, params any) error {
	return sendRpcForTarget(method, params, true, nil)
}

func sendProtectedRpcWithCallback(method string, params any, cb rpcDoneFunc) error {
	return sendRpcForTarget(method, params, true, cb)
}

func sendRpcForTarget(method string, params any, requireToken bool, cb rpcDoneFunc) error {
	if err := ensureRegisteredForCurrentDevice(); err != nil {
		return err
	}
	token := ""
	if requireToken {
		var err error
		token, err = currentValidToken()
		if err != nil {
			return err
		}
	}
	return sendRpcRequestWithCallback(method, token, params, cb)
}

func RpcHello() error {
	return sendRpcForTarget("hello", nil, false, nil)
}

func RpcHelloWithCallback(cb func(HelloResult, error)) error {
	return sendRpcForTarget("hello", nil, false, func(resp RpcResponse, err error) {
		if err != nil {
			cb(HelloResult{}, err)
			return
		}
		var out HelloResult
		if err := json.Unmarshal(resp.Result, &out); err != nil {
			cb(HelloResult{}, err)
			return
		}
		cb(out, nil)
	})
}

func RpcShellExec(cmd string, timeoutMs int, cb func(ShellExecResult, error)) error {
	cmd = normalizeInline(cmd)
	if cmd == "" {
		return errors.New("命令不能为空")
	}
	if timeoutMs <= 0 {
		timeoutMs = DefaultShellTimeoutMs
	}
	params := map[string]any{"cmd": cmd, "sync": true, "timeoutMs": timeoutMs}
	return sendProtectedRpcWithCallback("shell.exec", params, func(resp RpcResponse, err error) {
		if err != nil {
			cb(ShellExecResult{}, err)
			return
		}
		var out ShellExecResult
		if err := json.Unmarshal(resp.Result, &out); err != nil {
			cb(ShellExecResult{}, err)
			return
		}
		cb(out, nil)
	})
}

func RpcShellGetCwd(cb func(ShellCwdResult, error)) error {
	return sendProtectedRpcWithCallback("shell.getCwd", map[string]any{}, func(resp RpcResponse, err error) {
		if err != nil {
			cb(ShellCwdResult{}, err)
			return
		}
		var out ShellCwdResult
		if err := json.Unmarshal(resp.Result, &out); err != nil {
			cb(ShellCwdResult{}, err)
			return
		}
		cb(out, nil)
	})
}

func RpcShellSetCwd(path string, cb func(ShellCwdResult, error)) error {
	path = NormalizePath(readState(func(state DebugState) string { return state.CurrentCwdInput }), path)
	return sendProtectedRpcWithCallback("shell.setCwd", map[string]any{"cwd": path}, func(resp RpcResponse, err error) {
		if err != nil {
			cb(ShellCwdResult{}, err)
			return
		}
		var out ShellCwdResult
		if err := json.Unmarshal(resp.Result, &out); err != nil {
			cb(ShellCwdResult{}, err)
			return
		}
		cb(out, nil)
	})
}

func RpcFsStat(path string, cb func(FsStatResult, error)) error {
	path = NormalizePath(readState(func(state DebugState) string { return state.FileCurrentDir }), path)
	return sendProtectedRpcWithCallback("fs.stat", map[string]any{"path": path}, func(resp RpcResponse, err error) {
		if err != nil {
			cb(FsStatResult{}, err)
			return
		}
		var out FsStatResult
		if err := json.Unmarshal(resp.Result, &out); err != nil {
			cb(FsStatResult{}, err)
			return
		}
		cb(out, nil)
	})
}

func RpcFsReadChunk(path string, offset int, length int, cb func(FsReadResult, error)) error {
	if length <= 0 {
		length = DefaultFsChunkSize
	}
	if length > 32*1024 {
		length = 32 * 1024
	}
	path = NormalizePath(readState(func(state DebugState) string { return state.FileCurrentDir }), path)
	params := map[string]any{"path": path, "offset": offset, "length": length, "encoding": "base64"}
	return sendProtectedRpcWithCallback("fs.read", params, func(resp RpcResponse, err error) {
		if err != nil {
			cb(FsReadResult{}, err)
			return
		}
		var out FsReadResult
		if err := json.Unmarshal(resp.Result, &out); err != nil {
			cb(FsReadResult{}, err)
			return
		}
		cb(out, nil)
	})
}

func RpcFsWriteChunk(path string, mode string, b64 string, cb func(FsWriteResult, error)) error {
	mode = strings.TrimSpace(mode)
	if mode != "truncate" && mode != "append" {
		mode = "append"
	}
	path = NormalizePath(readState(func(state DebugState) string { return state.FileCurrentDir }), path)
	params := map[string]any{"path": path, "mode": mode, "encoding": "base64", "data": b64}
	return sendProtectedRpcWithCallback("fs.write", params, func(resp RpcResponse, err error) {
		if err != nil {
			cb(FsWriteResult{}, err)
			return
		}
		var out FsWriteResult
		if err := json.Unmarshal(resp.Result, &out); err != nil {
			cb(FsWriteResult{}, err)
			return
		}
		cb(out, nil)
	})
}

func ensureRegisteredForCurrentDevice() error {
	addr, err := selectedDeviceAddr()
	if err != nil {
		return err
	}
	snapshot := readStateSnapshot()
	if snapshot.RegisteredDeviceAddr == addr {
		return nil
	}
	ret := register.RegisterInterconnectRecv(addr, TargetPackageName).Read()
	if ResultUnitFailed(ret) {
		return fmt.Errorf("register-interconnect-recv failed")
	}
	withState(func(state *DebugState) {
		state.RegisteredDeviceAddr = addr
	})
	appendLogf("INFO", "已注册 interconnect 接收: device=%s pkg=%s", addr, TargetPackageName)
	return nil
}

func selectedDeviceAddr() (string, error) {
	addr := readState(func(state DebugState) string {
		return strings.TrimSpace(state.SelectedDeviceAddr)
	})
	if addr == "" {
		return "", fmt.Errorf("请先选择设备")
	}
	return addr, nil
}

func currentValidToken() (string, error) {
	token := readState(func(state DebugState) string {
		return strings.TrimSpace(state.TokenInput)
	})
	if token == "" {
		return "", fmt.Errorf("token 不能为空")
	}
	if !tokenPattern.MatchString(token) {
		return "", fmt.Errorf("token 必须是 4 位数字")
	}
	return token, nil
}
