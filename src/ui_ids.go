package plugin

const (
	EventDeviceRefresh      = "device.refresh"
	EventDeviceSelectPrefix = "device.select:"

	EventTokenInput           = "conn.token.input"
	EventRegisterInterconnect = "conn.register"
	EventHello                = "rpc.hello"

	EventRouteDashboard = "route.dashboard"
	EventRouteTerminal  = "route.terminal"
	EventRouteFileMgr   = "route.filemgr"
	EventRouteSettings  = "route.settings"
	EventRouteLogs      = "route.logs"
	EventLaunchQA       = "action.launch.qa"

	EventGetCwd   = "rpc.cwd.get"
	EventCwdInput = "rpc.cwd.input"
	EventSetCwd   = "rpc.cwd.set"

	EventCommandInput             = "terminal.command.input"
	EventTerminalKeyDown          = "terminal.keydown"
	EventExecCommand              = "terminal.exec.run"
	EventTerminalClear            = "terminal.clear"
	EventTerminalHistoryRunPrefix = "terminal.history.run:"

	EventFileRefresh         = "file.refresh"
	EventFileDirInput        = "file.dir.input"
	EventFileSearchInput     = "file.search.input"
	EventFileViewGrid        = "file.view.grid"
	EventFileViewList        = "file.view.list"
	EventFileSortName        = "file.sort.name"
	EventFileSortSize        = "file.sort.size"
	EventFileSortDate        = "file.sort.date"
	EventFileGoDir           = "file.go.dir"
	EventFileParent          = "file.go.parent"
	EventFileLoadMore        = "file.load.more"
	EventFileEntryOpenPrefix = "file.entry.open:"
	EventFileNewFile         = "file.new.file"
	EventFileNewDir          = "file.new.dir"
	EventFileDelete          = "file.delete"
	EventFileRename          = "file.rename"
	EventFileCopy            = "file.copy"
	EventFileMove            = "file.move"
	EventFileUpload          = "file.upload"
	EventFileDownload        = "file.download"
	EventFileSave            = "file.save"

	EventLogClear      = "log.clear"
	EventLogExportText = "log.export"
	EventLogPagePrev   = "log.page.prev"
	EventLogPageNext   = "log.page.next"

	EventTerminalPagePrev = "terminal.page.prev"
	EventTerminalPageNext = "terminal.page.next"

	EventSetCwdToDevice = "action.file.setcwd_device"
	EventSetCwdToApp    = "action.file.setcwd_app"
	EventSetCwdToData   = "action.file.setcwd_data"

	EventFileManagerClickDirPrefix  = "filemgr.click.dir:"
	EventFileManagerClickFilePrefix = "filemgr.click.file:"

	EventFileEditorInput    = "file.editor.input"
	EventTerminalExportText = "terminal.export"

	EventFileCtxCopyPrefix     = "file.ctx.copy:"
	EventFileCtxMovePrefix     = "file.ctx.move:"
	EventFileCtxRenamePrefix   = "file.ctx.rename:"
	EventFileCtxDeletePrefix   = "file.ctx.delete:"
	EventFileCtxDownloadPrefix = "file.ctx.download:"
)
