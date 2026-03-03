package plugin

import (
	device "astroboxplugin/bindings/astrobox_psys_host_device"
	"fmt"
	"sync"
	"time"
)

type PendingRequest struct {
	ID             string
	Method         string
	TimeoutTimerID uint64
	SentAtMs       int64
}

type LogEntry struct {
	Timestamp string
	Level     string
	Message   string
}

type DebugState struct {
	ConnectedDevices     []device.DeviceInfo
	SelectedDeviceAddr   string
	SelectedDeviceName   string
	RegisteredDeviceAddr string

	TokenInput      string
	CurrentCommand  string
	CurrentCwdInput string

	Pending *PendingRequest

	LastRequestID      string
	LastRequestMethod  string
	LastResponseStatus string
	LastResponseRaw    string
	LastResponsePretty string
	LastLatencyMs      int64
	LastError          string

	CurrentAppRoute string

	TerminalHistory   []TerminalHistoryEntry
	TerminalFavorites []TerminalFavorite
	TerminalBuffer    []string

	FileCurrentDir       string
	FileDirInput         string
	FileSearchQuery      string
	FileViewMode         FileViewMode
	FileSortMode         FileSortMode
	FileEntries          []FileEntry
	FileVisibleCount     int
	FileSelectedPaths    []string // Supports multi-select
	FileEditorText       string
	FileEditorIsBinary   bool
	FileEditorHexPreview string

	TransferLastLocalPath string
	TransferProgress      string

	TaskQueueBusy   bool
	TaskQueueLength int

	Logs    []LogEntry
	LogPage int

	TerminalPage int
}

var (
	debugStateMu sync.Mutex
	debugState   DebugState
)

func withState(fn func(*DebugState)) {
	debugStateMu.Lock()
	defer debugStateMu.Unlock()
	fn(&debugState)
}

func readState[T any](fn func(DebugState) T) T {
	debugStateMu.Lock()
	defer debugStateMu.Unlock()
	return fn(debugState)
}

func initDebugState() {
	favorites := make([]TerminalFavorite, 0, len(CommandPresets))
	for _, cmd := range CommandPresets {
		favorites = append(favorites, TerminalFavorite{Name: cmd, Command: cmd})
	}

	withState(func(state *DebugState) {
		*state = DebugState{
			CurrentCommand:     DefaultCommand,
			CurrentCwdInput:    DefaultFileDir,
			LastResponseStatus: "idle",
			CurrentAppRoute:    RouteDashboard,
			TerminalHistory:    make([]TerminalHistoryEntry, 0, 32),
			TerminalFavorites:  favorites,
			TerminalBuffer:     make([]string, 0, 100),
			FileCurrentDir:     DefaultFileDir,
			FileDirInput:       DefaultFileDir,
			FileViewMode:       FileViewGrid,
			FileSortMode:       FileSortByName,
			FileEntries:        make([]FileEntry, 0, 64),
			FileVisibleCount:   DefaultDirPageSize,
			TransferProgress:   "idle",
			Logs:               make([]LogEntry, 0, 64),
		}
	})
	appendLog("INFO", "插件已初始化")
}

func readStateSnapshot() DebugState {
	return readState(func(state DebugState) DebugState {
		copyState := state
		copyState.ConnectedDevices = append([]device.DeviceInfo(nil), state.ConnectedDevices...)
		copyState.TerminalHistory = append([]TerminalHistoryEntry(nil), state.TerminalHistory...)
		copyState.TerminalFavorites = append([]TerminalFavorite(nil), state.TerminalFavorites...)
		copyState.TerminalBuffer = append([]string(nil), state.TerminalBuffer...)
		copyState.FileEntries = append([]FileEntry(nil), state.FileEntries...)
		copyState.FileSelectedPaths = append([]string(nil), state.FileSelectedPaths...)
		copyState.Logs = append([]LogEntry(nil), state.Logs...)
		if state.Pending != nil {
			pendingCopy := *state.Pending
			copyState.Pending = &pendingCopy
		}
		return copyState
	})
}

func clearLogs() {
	withState(func(state *DebugState) {
		state.Logs = nil
		state.LastError = ""
	})
}

func appendLog(level string, message string) {
	entry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     level,
		Message:   message,
	}
	withState(func(state *DebugState) {
		state.Logs = append(state.Logs, entry)
		if len(state.Logs) > MaxLogEntries {
			overflow := len(state.Logs) - MaxLogEntries
			copy(state.Logs, state.Logs[overflow:])
			state.Logs = state.Logs[:MaxLogEntries]
		}
		if level == "ERROR" {
			state.LastError = message
		}
	})
	Logf("[%s] %s", level, message)
}

func appendLogf(level string, format string, args ...any) {
	appendLog(level, fmt.Sprintf(format, args...))
}

func addTerminalHistory(command string, output string, exitCode string) {
	withState(func(state *DebugState) {
		entry := TerminalHistoryEntry{
			Command:   command,
			Output:    output,
			ExitCode:  exitCode,
			Timestamp: time.Now().Format("15:04:05"),
		}
		state.TerminalHistory = append(state.TerminalHistory, entry)
		if len(state.TerminalHistory) > MaxTerminalHistory {
			overflow := len(state.TerminalHistory) - MaxTerminalHistory
			copy(state.TerminalHistory, state.TerminalHistory[overflow:])
			state.TerminalHistory = state.TerminalHistory[:MaxTerminalHistory]
		}
	})
}

func addTerminalFavorite(command string) {
	command = normalizeInline(command)
	if command == "" {
		return
	}
	withState(func(state *DebugState) {
		for _, it := range state.TerminalFavorites {
			if it.Command == command {
				return
			}
		}
		state.TerminalFavorites = append(state.TerminalFavorites, TerminalFavorite{Name: command, Command: command})
		if len(state.TerminalFavorites) > MaxTerminalFavorites {
			overflow := len(state.TerminalFavorites) - MaxTerminalFavorites
			copy(state.TerminalFavorites, state.TerminalFavorites[overflow:])
			state.TerminalFavorites = state.TerminalFavorites[:MaxTerminalFavorites]
		}
	})
}

func removeTerminalFavorite(index int) {
	withState(func(state *DebugState) {
		if index < 0 || index >= len(state.TerminalFavorites) {
			return
		}
		state.TerminalFavorites = append(state.TerminalFavorites[:index], state.TerminalFavorites[index+1:]...)
	})
}
