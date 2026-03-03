package plugin

import (
	"fmt"
	"strings"
)

func formatSelectedDevice(snapshot DebugState) string {
	if snapshot.SelectedDeviceAddr == "" {
		return "(未选择)"
	}
	if snapshot.SelectedDeviceName != "" {
		return fmt.Sprintf("%s (%s)", snapshot.SelectedDeviceName, snapshot.SelectedDeviceAddr)
	}
	return snapshot.SelectedDeviceAddr
}

func fallback(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
}

func deviceSelectEventID(addr string) string {
	return EventDeviceSelectPrefix + addr
}

func historyRunEventID(index int) string {
	return fmt.Sprintf("%s%d", EventTerminalHistoryRunPrefix, index)
}

func favoriteRunEventID(index int) string {
	return fmt.Sprintf("%s%d", EventTerminalFavoriteRunPrefix, index)
}

func favoriteDelEventID(index int) string {
	return fmt.Sprintf("%s%d", EventTerminalFavoriteDelPrefix, index)
}

func fileOpenEventID(path string) string {
	return EventFileEntryOpenPrefix + path
}

func fileCtxCopyEventID(path string) string {
	return EventFileCtxCopyPrefix + path
}

func fileCtxMoveEventID(path string) string {
	return EventFileCtxMovePrefix + path
}

func fileCtxRenameEventID(path string) string {
	return EventFileCtxRenamePrefix + path
}

func fileCtxDeleteEventID(path string) string {
	return EventFileCtxDeletePrefix + path
}

func fileCtxDownloadEventID(path string) string {
	return EventFileCtxDownloadPrefix + path
}

func isPathSelected(snapshot DebugState, path string) bool {
	for _, p := range snapshot.FileSelectedPaths {
		if p == path {
			return true
		}
	}
	return false
}

func filterEntries(entries []FileEntry, query string) []FileEntry {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return entries
	}
	filtered := make([]FileEntry, 0, len(entries))
	for _, it := range entries {
		if strings.Contains(strings.ToLower(it.Name), query) || strings.Contains(strings.ToLower(it.Path), query) {
			filtered = append(filtered, it)
		}
	}
	return filtered
}
