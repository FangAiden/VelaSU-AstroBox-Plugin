package plugin

import (
	interconnect "astroboxplugin/bindings/astrobox_psys_host_interconnect"
	timer "astroboxplugin/bindings/astrobox_psys_host_timer"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type rpcDoneFunc func(resp RpcResponse, err error)

var (
	callbackMu       sync.Mutex
	requestCallbacks = map[string]rpcDoneFunc{}
)

func BuildRpcRequest(method string, token string, params any) (RpcRequest, string, error) {
	method = strings.TrimSpace(method)
	if method == "" {
		return RpcRequest{}, "", errors.New("rpc method is empty")
	}

	req := RpcRequest{ID: makeRequestID(), Method: method}
	if token != "" {
		req.Token = token
	}
	if params != nil {
		req.Params = params
	}

	buf, err := json.Marshal(req)
	if err != nil {
		return RpcRequest{}, "", err
	}
	return req, string(buf), nil
}

func ParseRpcResponse(raw string) (RpcResponse, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return RpcResponse{}, errors.New("rpc response is empty")
	}
	var resp RpcResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return RpcResponse{}, err
	}
	if strings.TrimSpace(resp.ID) == "" {
		return RpcResponse{}, errors.New("rpc response missing id")
	}
	return resp, nil
}

func sendRpcRequest(method string, token string, params any) error {
	return sendRpcRequestWithCallback(method, token, params, nil)
}

func sendRpcRequestWithCallback(method string, token string, params any, cb rpcDoneFunc) error {
	snapshot := readStateSnapshot()
	if strings.TrimSpace(snapshot.SelectedDeviceAddr) == "" {
		return errors.New("请先选择设备")
	}
	if snapshot.Pending != nil {
		return errors.New("请等待当前请求完成")
	}

	req, payload, err := BuildRpcRequest(method, token, params)
	if err != nil {
		return err
	}

	if cb != nil {
		registerRequestCallback(req.ID, cb)
	}

	ret := interconnect.SendQaicMessage(snapshot.SelectedDeviceAddr, TargetPackageName, payload).Read()
	if ret.IsErr() {
		if cb != nil {
			invokeRequestCallback(req.ID, RpcResponse{}, errors.New("send-qaic-message failed"))
		}
		return errors.New("send-qaic-message failed")
	}

	timeoutPayload := buildTimeoutPayload(req.ID)
	timeoutTimerID := timer.SetTimeout(DefaultRpcTimeoutMs, timeoutPayload).Read()
	nowMs := time.Now().UnixMilli()

	withState(func(state *DebugState) {
		state.Pending = &PendingRequest{ID: req.ID, Method: method, TimeoutTimerID: timeoutTimerID, SentAtMs: nowMs}
		state.LastRequestID = req.ID
		state.LastRequestMethod = method
		state.LastResponseStatus = "waiting"
		state.LastResponseRaw = ""
		state.LastResponsePretty = ""
		state.LastLatencyMs = 0
		state.LastError = ""
	})

	appendLogf("INFO", "已发送 RPC: method=%s id=%s", method, req.ID)
	return nil
}

func registerRequestCallback(id string, cb rpcDoneFunc) {
	callbackMu.Lock()
	requestCallbacks[id] = cb
	callbackMu.Unlock()
}

func popRequestCallback(id string) rpcDoneFunc {
	callbackMu.Lock()
	defer callbackMu.Unlock()
	cb := requestCallbacks[id]
	delete(requestCallbacks, id)
	return cb
}

func invokeRequestCallback(id string, resp RpcResponse, err error) {
	cb := popRequestCallback(id)
	if cb == nil {
		return
	}
	cb(resp, err)
}

func handleInterconnectEventPayload(eventPayload string) bool {
	hasPending := readState(func(state DebugState) bool {
		return state.Pending != nil
	})
	if !hasPending {
		return false
	}

	text, err := ExtractPayloadText(eventPayload)
	if err != nil {
		appendLogf("ERROR", "解析 interconnect 事件失败: %v", err)
		return false
	}
	text = strings.TrimSpace(text)
	if text == "" {
		appendLog("WARN", "收到空 interconnect 消息")
		return false
	}

	pendingMethod := readState(func(state DebugState) string {
		if state.Pending == nil {
			return ""
		}
		return state.Pending.Method
	})

	if shouldRecordResponsePreview(pendingMethod) {
		withState(func(state *DebugState) {
			rawPreview, prettyPreview := buildResponsePreview(text)
			state.LastResponseRaw = rawPreview
			state.LastResponsePretty = prettyPreview
		})
	}

	resp, err := ParseRpcResponse(text)
	if err != nil {
		appendLogf("ERROR", "回包 JSON 解析失败: %v", err)
		snapshot := readStateSnapshot()
		if snapshot.Pending != nil {
			clearPendingTimer(snapshot.Pending.TimeoutTimerID)
			invokeRequestCallback(snapshot.Pending.ID, RpcResponse{}, fmt.Errorf("invalid rpc response: %w", err))
		}
		withState(func(state *DebugState) {
			state.Pending = nil
			state.LastResponseStatus = "invalid_response"
			state.LastError = "回包不是有效的 RPC JSON"
		})
		appendLog("WARN", "无效回包导致 pending 请求终止")
		scheduleQueueDrain()
		return true
	}

	return handleParsedRpcResponse(resp)
}

func buildResponsePreview(raw string) (string, string) {
	const maxRawChars = 4096
	const maxPrettyInputChars = 12288
	const maxPrettyChars = 4096

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}

	rawPreview := raw
	if len(rawPreview) > maxRawChars {
		rawPreview = rawPreview[:maxRawChars] + "\n...(truncated)"
	}

	if len(raw) > maxPrettyInputChars {
		return rawPreview, "(omitted: payload too large to pretty-print)"
	}

	pretty := prettyJSON(raw)
	if len(pretty) > maxPrettyChars {
		pretty = pretty[:maxPrettyChars] + "\n...(truncated)"
	}
	return rawPreview, pretty
}

func shouldRecordResponsePreview(method string) bool {
	switch strings.TrimSpace(method) {
	case "fs.read":
		return false
	default:
		return true
	}
}

func handleParsedRpcResponse(resp RpcResponse) bool {
	snapshot := readStateSnapshot()
	pending := snapshot.Pending
	if pending == nil {
		withState(func(state *DebugState) {
			state.LastResponseStatus = "unmatched"
		})
		appendLogf("WARN", "收到未匹配回包 id=%s", resp.ID)
		return false
	}
	if resp.ID != pending.ID {
		withState(func(state *DebugState) {
			state.LastResponseStatus = "id_mismatch"
		})
		appendLogf("WARN", "回包 id 不匹配: pending=%s resp=%s", pending.ID, resp.ID)
		return false
	}

	clearPendingTimer(pending.TimeoutTimerID)

	latency := time.Now().UnixMilli() - pending.SentAtMs
	status := "ok"
	lastErr := ""
	if !resp.OK {
		status = "error"
		lastErr = rpcErrorMessage(resp)
	}

	withState(func(state *DebugState) {
		state.Pending = nil
		state.LastLatencyMs = latency
		state.LastResponseStatus = status
		state.LastError = lastErr
	})

	if !resp.OK {
		appendLogf("ERROR", "RPC 失败: %s", lastErr)
		invokeRequestCallback(resp.ID, resp, errors.New(lastErr))
		scheduleQueueDrain()
		return true
	}

	appendLogf("INFO", "RPC 成功: method=%s latency=%dms", pending.Method, latency)
	handleSuccessResult(pending.Method, resp.Result)
	invokeRequestCallback(resp.ID, resp, nil)
	scheduleQueueDrain()

	if strings.TrimSpace(pending.Method) == "fs.read" {
		progress := readState(func(state DebugState) string {
			return strings.TrimSpace(state.TransferProgress)
		})
		if strings.Contains(progress, "done") || strings.Contains(progress, "failed") {
			return true
		}
		return false
	}
	return true
}

func handleSuccessResult(method string, resultRaw json.RawMessage) {
	if len(resultRaw) == 0 {
		return
	}

	switch method {
	case "hello":
		var hello HelloResult
		if err := json.Unmarshal(resultRaw, &hello); err != nil {
			appendLogf("WARN", "hello 结果解析失败: %v", err)
			return
		}
		appendLogf("INFO", "hello: server=%s protocol=%d remoteEnabled=%t hasToken=%t", hello.Server, hello.Protocol, hello.RemoteEnabled, hello.HasToken)
	case "shell.exec":
		var result ShellExecResult
		if err := json.Unmarshal(resultRaw, &result); err != nil {
			appendLogf("WARN", "shell.exec 结果解析失败: %v", err)
			return
		}
		exitCode := "null"
		if result.ExitCode != nil {
			exitCode = fmt.Sprintf("%d", *result.ExitCode)
		}
		withState(func(state *DebugState) {
			if strings.TrimSpace(result.Output) != "" && state.CurrentAppRoute == RouteTerminal {
				state.TerminalBuffer = append(state.TerminalBuffer, result.Output)
			}
			if len(state.TerminalBuffer) > 20 {
				state.TerminalBuffer = state.TerminalBuffer[len(state.TerminalBuffer)-20:]
			}
			// Auto-scroll to last page
			allLines := strings.Join(state.TerminalBuffer, "\n")
			lineCount := len(strings.Split(allLines, "\n"))
			lastPage := (lineCount+TerminalPageSize-1)/TerminalPageSize - 1
			if lastPage < 0 {
				lastPage = 0
			}
			state.TerminalPage = lastPage
		})
		addTerminalHistory(result.Cmd, truncateOutput(result.Output, 600), exitCode)
		appendLogf("INFO", "shell.exec: mode=%s exitCode=%s cwd=%s outputBytes=%d", result.Mode, exitCode, result.Cwd, len(result.Output))
	case "shell.getCwd", "shell.setCwd":
		var result ShellCwdResult
		if err := json.Unmarshal(resultRaw, &result); err != nil {
			appendLogf("WARN", "%s 结果解析失败: %v", method, err)
			return
		}
		withState(func(state *DebugState) {
			state.CurrentCwdInput = result.Cwd
			if state.CurrentAppRoute == RouteFileMgr && state.FileCurrentDir == "" {
				state.FileCurrentDir = result.Cwd
				state.FileDirInput = result.Cwd
			}
		})
		appendLogf("INFO", "%s: cwd=%s", method, result.Cwd)
	case "fs.stat", "fs.write":
		appendLogf("INFO", "%s 完成", method)
	}
}

func clearPendingRequest(reason string) {
	snapshot := readStateSnapshot()
	if snapshot.Pending == nil {
		return
	}

	clearPendingTimer(snapshot.Pending.TimeoutTimerID)
	invokeRequestCallback(snapshot.Pending.ID, RpcResponse{}, errors.New(reason))
	withState(func(state *DebugState) {
		state.Pending = nil
		state.LastResponseStatus = "canceled"
		if reason != "" {
			state.LastError = reason
		}
	})

	if reason != "" {
		appendLogf("WARN", "取消 pending 请求: %s", reason)
	}
	scheduleQueueDrain()
}

func handleRpcTimeoutPayloadText(text string) bool {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, timerPayloadPrefix) {
		return false
	}

	timeoutID := strings.TrimPrefix(text, timerPayloadPrefix)
	snapshot := readStateSnapshot()
	if snapshot.Pending == nil || snapshot.Pending.ID != timeoutID {
		return false
	}

	invokeRequestCallback(timeoutID, RpcResponse{}, errors.New("rpc timeout"))
	withState(func(state *DebugState) {
		state.Pending = nil
		state.LastResponseStatus = "timeout"
		state.LastError = "请求超时，请重试"
	})
	appendLogf("ERROR", "RPC 超时: id=%s", timeoutID)
	scheduleQueueDrain()
	return true
}

func handleRpcTimeoutEventPayload(eventPayload string) bool {
	text, err := ExtractPayloadText(eventPayload)
	if err != nil {
		return false
	}
	return handleRpcTimeoutPayloadText(text)
}

func rpcErrorMessage(resp RpcResponse) string {
	code := ""
	msg := strings.TrimSpace(resp.Message)
	if resp.Error != nil {
		code = strings.TrimSpace(resp.Error.Code)
		if strings.TrimSpace(resp.Error.Message) != "" {
			msg = strings.TrimSpace(resp.Error.Message)
		}
	}
	if msg == "" {
		msg = "unknown rpc error"
	}

	switch code {
	case "REMOTE_DISABLED":
		msg += "（请在手表端开启远程控制）"
	case "AUTH_FAILED":
		msg += "（请检查 4 位 token）"
	}

	if code == "" {
		return msg
	}
	return code + ": " + msg
}

func makeRequestID() string {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("req_%d_fallback", time.Now().UnixMilli())
	}
	return fmt.Sprintf("req_%d_%s", time.Now().UnixMilli(), hex.EncodeToString(buf))
}

func buildTimeoutPayload(requestID string) string {
	return timerPayloadPrefix + requestID
}

func clearPendingTimer(timerID uint64) {
	if timerID == 0 {
		return
	}
	_ = timer.ClearTimer(timerID).Read()
}

func prettyJSON(raw string) string {
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(raw), "", "  "); err != nil {
		return raw
	}
	return out.String()
}

func truncateOutput(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
