// actions.go - UI event dispatcher, input handling, and event payload helpers.
package plugin

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	ui "astroboxplugin/bindings/astrobox_psys_host_ui"

	"github.com/bytecodealliance/wit-bindgen/wit_types"
)

var tokenPattern = regexp.MustCompile(`^\d{4}$`)

func HandleUIEvent(eventID string, event ui.Event, eventPayload string) {
	if event == ui.EventInput || event == ui.EventChange {
		handleInputChange(eventID, eventPayload)
		if eventID == EventTerminalKeyDown {
			handleKeyDown(eventID, eventPayload)
		}
		return
	}

	if event != ui.EventClick {
		return
	}

	if requiresDependencyReady(eventID) && !ensureDependencyReadyForUsage() {
		return
	}

	switch {
	// Device & Connection
	case eventID == EventDeviceRefresh:
		actionDeviceRefresh()
	case strings.HasPrefix(eventID, EventDeviceSelectPrefix):
		actionDeviceSelect(strings.TrimPrefix(eventID, EventDeviceSelectPrefix))
	case eventID == EventRegisterInterconnect:
		actionRegisterInterconnect()
	case eventID == EventHello:
		actionHello()
	case eventID == EventLaunchQA:
		actionLaunchQA()
	case eventID == EventDependencyRefresh:
		actionDependencyRefresh()

	// Routing
	case eventID == EventRouteDashboard:
		withState(func(state *DebugState) { state.CurrentAppRoute = RouteDashboard })
	case eventID == EventRouteTerminal:
		withState(func(state *DebugState) { state.CurrentAppRoute = RouteTerminal })
	case eventID == EventRouteFileMgr:
		withState(func(state *DebugState) { state.CurrentAppRoute = RouteFileMgr })
	case eventID == EventRouteSettings:
		withState(func(state *DebugState) { state.CurrentAppRoute = RouteSettings })
	case eventID == EventRouteLogs:
		withState(func(state *DebugState) { state.CurrentAppRoute = RouteLogs })

	// Terminal
	case eventID == EventExecCommand:
		actionExecCommand()
	case eventID == EventTerminalClear:
		actionTerminalClear()
	case strings.HasPrefix(eventID, EventTerminalHistoryRunPrefix):
		actionTerminalRunHistory(parseSuffixIndex(eventID, EventTerminalHistoryRunPrefix))

	// File Manager
	case eventID == EventFileRefresh:
		actionFileRefresh()
	case eventID == EventFileGoDir:
		actionFileGoDir()
	case eventID == EventFileParent:
		actionFileParent()
	case eventID == EventFileLoadMore:
		actionFileLoadMore()
	case eventID == EventFileViewGrid:
		withState(func(state *DebugState) { state.FileViewMode = FileViewGrid })
	case eventID == EventFileViewList:
		withState(func(state *DebugState) { state.FileViewMode = FileViewList })
	case eventID == EventFileSortName:
		actionFileSort(FileSortByName)
	case eventID == EventFileSortSize:
		actionFileSort(FileSortBySize)
	case eventID == EventFileSortDate:
		appendLog("INFO", "日期排序已移除")
	case strings.HasPrefix(eventID, EventFileEntryOpenPrefix):
		actionFileOpenEntry(strings.TrimPrefix(eventID, EventFileEntryOpenPrefix))
	case eventID == EventFileNewDir:
		actionFileNewDir()
	case eventID == EventFileDelete:
		actionFileDelete()
	case eventID == EventFileRename:
		actionFileRename()
	case eventID == EventFileCopy:
		actionFileCopy()
	case eventID == EventFileMove:
		actionFileMove()
	case eventID == EventFilePaste:
		actionFilePaste()
	case eventID == EventFileClipboardClear:
		actionFileClipboardClear()
	case eventID == EventFileUpload:
		actionFileUpload()
	case eventID == EventFileDownload:
		actionFileDownload()

	// File Manager Context menu shortcuts
	case strings.HasPrefix(eventID, EventFileCtxCopyPrefix):
		selectSinglePath(strings.TrimPrefix(eventID, EventFileCtxCopyPrefix))
		actionFileCopy()
	case strings.HasPrefix(eventID, EventFileCtxMovePrefix):
		selectSinglePath(strings.TrimPrefix(eventID, EventFileCtxMovePrefix))
		actionFileMove()
	case strings.HasPrefix(eventID, EventFileCtxRenamePrefix):
		selectSinglePath(strings.TrimPrefix(eventID, EventFileCtxRenamePrefix))
		actionFileRename()
	case strings.HasPrefix(eventID, EventFileCtxDeletePrefix):
		selectSinglePath(strings.TrimPrefix(eventID, EventFileCtxDeletePrefix))
		actionFileDelete()
	case strings.HasPrefix(eventID, EventFileCtxDownloadPrefix):
		selectSinglePath(strings.TrimPrefix(eventID, EventFileCtxDownloadPrefix))
		actionFileDownload()

	// Logs
	case eventID == EventLogClear:
		actionLogClear()

	// Pagination
	case eventID == EventLogPagePrev:
		withState(func(s *DebugState) {
			if s.LogPage > 0 {
				s.LogPage--
			}
		})
	case eventID == EventLogPageNext:
		withState(func(s *DebugState) { s.LogPage++ })
	case eventID == EventTerminalPagePrev:
		withState(func(s *DebugState) {
			if s.TerminalPage > 0 {
				s.TerminalPage--
			}
		})
	case eventID == EventTerminalPageNext:
		withState(func(s *DebugState) { s.TerminalPage++ })

	default:
		appendLogf("WARN", "未知 UI 事件: %s", eventID)
	}
}

func handleInputChange(eventID string, eventPayload string) {
	payload, ok := parseUIEventPayload(eventPayload)
	if !ok {
		return
	}
	value := payload.Value
	switch eventID {
	case EventTokenInput:
		withState(func(state *DebugState) {
			state.TokenInput = strings.TrimSpace(value)
		})
	case EventCwdInput:
		withState(func(state *DebugState) {
			state.CurrentCwdInput = value
		})
	case EventCommandInput:
		withState(func(state *DebugState) {
			state.CurrentCommand = value
		})
	case EventFileDirInput:
		withState(func(state *DebugState) {
			state.FileDirInput = value
		})
	case EventFileSearchInput:
		withState(func(state *DebugState) {
			state.FileSearchQuery = value
		})
	case EventFileRenameInput:
		withState(func(state *DebugState) {
			state.FileRenameInput = strings.TrimSpace(value)
		})
	}
}

func handleKeyDown(eventID string, eventPayload string) {
	payload, ok := parseUIEventPayload(eventPayload)
	if !ok {
		return
	}
	if eventID == EventTerminalKeyDown && strings.EqualFold(payload.Key, "Enter") {
		actionExecCommand()
	}
}

func selectSinglePath(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	withState(func(state *DebugState) {
		state.FileSelectedPaths = []string{path}
		state.FileRenameInput = BaseName(path)
		state.FilePreviewSeq++
	})
}

func parseSuffixIndex(eventID string, prefix string) int {
	if !strings.HasPrefix(eventID, prefix) {
		return -1
	}
	raw := strings.TrimPrefix(eventID, prefix)
	idx, err := strconv.Atoi(raw)
	if err != nil {
		return -1
	}
	return idx
}

type uiEventPayload struct {
	Type     string `json:"type"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	ClientX  int    `json:"clientX"`
	ClientY  int    `json:"clientY"`
	Modifier string `json:"modifier"`
}

func parseUIEventPayload(eventPayload string) (uiEventPayload, bool) {
	eventPayload = strings.TrimSpace(eventPayload)
	if eventPayload == "" {
		return uiEventPayload{}, false
	}
	var payload uiEventPayload
	if err := json.Unmarshal([]byte(eventPayload), &payload); err == nil {
		return payload, true
	}
	payload.Value = eventPayload
	return payload, true
}

func ResultUnitFailed(ret wit_types.Result[wit_types.Unit, wit_types.Unit]) bool {
	return ret.IsErr()
}

func requiresDependencyReady(eventID string) bool {
	switch {
	case eventID == EventDeviceRefresh:
		return false
	case strings.HasPrefix(eventID, EventDeviceSelectPrefix):
		return false
	case eventID == EventDependencyRefresh:
		return false
	case eventID == EventLaunchQA:
		return false
	case eventID == EventRouteDashboard:
		return false
	case eventID == EventRouteSettings:
		return false
	case eventID == EventRouteLogs:
		return false
	case eventID == EventLogClear:
		return false
	case eventID == EventLogExportText:
		return false
	case eventID == EventLogPagePrev:
		return false
	case eventID == EventLogPageNext:
		return false
	default:
		return true
	}
}
