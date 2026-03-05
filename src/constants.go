package plugin

const (
	TargetPackageName   = "com.vela.su.aigik"
	TargetWatchfaceName = "VelaShellBridge"

	DefaultRpcTimeoutMs   uint64 = 16000
	DefaultShellTimeoutMs        = 15000

	MaxLogEntries     = 300
	MaxLogViewEntries = 120
	LogPageSize       = 15
	TerminalPageSize  = 15

	DefaultCommand = ""

	RouteDashboard = "dashboard"
	RouteTerminal  = "terminal"
	RouteFileMgr   = "filemgr"
	RouteSettings  = "settings"
	RouteLogs      = "logs"

	MaxTerminalHistory = 80

	DefaultFileDir       = "/data"
	DefaultDirPageSize   = 100
	DefaultFsChunkSize   = 1024
	DefaultUploadChunk   = 8 * 1024
	MaxEditorPreviewSize = 256 * 1024
	HexPreviewBytes      = 1024
	LocalDownloadDir     = "downloads"

	timerPayloadPrefix = "rpc_timeout:"
	queueDrainPayload  = "__queue_drain__"
)
