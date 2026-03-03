# 开发说明（终端 + 文件管理器）

## 1. 总体架构

- 保持原有 Interconnect 通道，不修改 WIT。
- 传输模型：
  - 外层约束：同一时刻仅 1 个 pending RPC。
  - 内层能力：通过 `task_queue.go` 串行执行批量任务（如目录 `stat` 扫描、上传/下载分块）。

## 2. 主要状态字段（`DebugState`）

- 连接态：`ConnectedDevices`、`SelectedDeviceAddr`、`RegisteredDeviceAddr`
- RPC 态：`Pending`、`LastRequestID`、`LastResponseStatus`、`LastError`
- 终端态：`CurrentCommand`、`CurrentCwdInput`、`TerminalHistory`、`TerminalFavorites`
- 文件管理态：`FileCurrentDir`、`FileEntries`、`FileSelectedPath`、`FileEditorText`
- 传输态：`TransferProgress`、`TransferLastLocalPath`
- 队列态：`TaskQueueBusy`、`TaskQueueLength`

## 3. 核心流程

### 3.1 终端执行

1. UI 触发执行 -> 入队。
2. `RpcShellExec` 发包 -> pending。
3. 回包按 `id` 匹配 -> 更新结果、历史、日志。

### 3.2 目录浏览

1. `shell.exec("ls <dir>")` 获取名称列表。
2. 对每个条目入队执行 `fs.stat` 补全元信息。
3. 目录优先排序并刷新 UI。

### 3.3 上传

1. `dialog.pick-file(read=true)` 读取本地文件。
2. 按块（默认 8KB）base64 编码后写入：
   - 首块 `truncate`
   - 后续 `append`
3. 完成后自动刷新当前目录。

### 3.4 下载

1. 循环 `fs.read(path, offset, length)` 拉取分块。
2. 聚合后写入本地 `downloads/<name>`。
3. UI 显示本地路径与传输状态。

## 4. 错误处理

- token 非 4 位数字：除 `hello` 外阻断发送。
- `REMOTE_DISABLED`：提示手表侧开启远控。
- `AUTH_FAILED`：提示 token 错误。
- 非 JSON / 缺 `id`：标记 `invalid_response`。
- `id` 不匹配：标记 `id_mismatch`，不清理当前 pending。
- 超时：标记 `timeout` 并释放 pending，允许继续下一请求。

## 5. 构建

```bash
python scripts/build_dist.py
```

产物在 `dist/`：

- `shellbridge_interconnect_plugin.wasm`
- `manifest.json`
- `icon.png`
