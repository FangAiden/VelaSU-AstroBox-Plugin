package plugin

const (
	TargetPackageName = "com.vela.su.aigik"

	DefaultRpcTimeoutMs   uint64 = 16000
	DefaultShellTimeoutMs        = 15000

	MaxLogEntries     = 300
	MaxLogViewEntries = 120
	LogPageSize       = 15
	TerminalPageSize  = 15

	DefaultCommand = "pwd"

	RouteDashboard = "dashboard"
	RouteTerminal  = "terminal"
	RouteFileMgr   = "filemgr"
	RouteSettings  = "settings"
	RouteLogs      = "logs"

	MaxTerminalHistory   = 80
	MaxTerminalFavorites = 20

	DefaultFileDir       = "/data"
	DefaultDirPageSize   = 100
	DefaultFsChunkSize   = 4096
	DefaultUploadChunk   = 8 * 1024
	MaxEditorPreviewSize = 1 * 1024 * 1024
	HexPreviewBytes      = 1024
	LocalDownloadDir     = "downloads"

	timerPayloadPrefix = "rpc_timeout:"
)

var CommandPresets = []string{
	"pwd",
	"ls",
	"ls /data",
}
