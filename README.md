# ShellBridge AstroBox 插件（终端 + 文件管理器）

基于 `tmp-shellbridge-plugin` 扩展实现，固定通过 Interconnect 对接：

- 目标 QuickApp：`com.vela.su.aigik`
- 协议方法：`hello`、`shell.exec`、`shell.getCwd`、`shell.setCwd`、`fs.stat`、`fs.read`、`fs.write`

## 已实现能力

- 终端：
  - 命令输入、预设命令、历史回填、收藏命令
  - 同步执行（`sync=true`）
  - `cwd` 获取/设置
  - pending 状态、耗时、原始/格式化回包展示
- 文件管理器：
  - 目录浏览（刷新、进入目录、返回上级、分页加载）
  - 文件/目录操作（新建、删除、重命名、移动、复制，均带确认弹窗）
  - 文件预览/编辑（UTF-8 文本优先，二进制 hex 回退）
  - 上传（`pick-file` + `fs.write` 分块）
  - 下载（`fs.read` 分块 + 落地 `downloads/`）
- 协议调度：
  - 保持“单 pending 请求”语义
  - 通过内部任务队列串行化批量动作（目录扫描、上传下载分块）

## 关键目录

```text
src/
  actions.go              # UI 动作分发
  debug_state.go          # 全局状态
  task_queue.go           # 串行任务队列
  rpc_client.go           # RPC 发包/回包/超时
  rpc_methods.go          # 高层 RPC 封装
  file_listing.go         # 目录列表（ls + fs.stat）
  file_transfer.go        # 上传/下载/预览
  path_utils.go           # 路径与 shell quote 工具
  ui.go                   # 页签式 UI（终端、文件管理）
```

## 构建

```bash
python scripts/build_dist.py
```

产物：

- `dist/shellbridge_interconnect_plugin.wasm`
- `dist/manifest.json`
- `dist/icon.png`
