package plugin

import (
	ui "astroboxplugin/bindings/astrobox_psys_host_ui"
	pluginEvent "astroboxplugin/bindings/astrobox_psys_plugin_event"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

func OnEvent(eventType pluginEvent.EventType, eventPayload string) string {
	switch eventType {
	case pluginEvent.EventTypeInterconnectMessage:
		handleInterconnectEventPayload(eventPayload)
		RerenderMainUI()
	case pluginEvent.EventTypeTimer:
		handleRpcTimeoutEventPayload(eventPayload)
		RerenderMainUI()
	case pluginEvent.EventTypePluginMessage:
		if !tryHandleUIEventV3Message(eventPayload) {
			appendLogf("INFO", "plugin-message: %s", truncateText(eventPayload, 160))
		}
	case pluginEvent.EventTypeDeviceAction:
		appendLogf("INFO", "device-action: %s", truncateText(eventPayload, 160))
	case pluginEvent.EventTypeProviderAction:
		appendLogf("INFO", "provider-action: %s", truncateText(eventPayload, 160))
	case pluginEvent.EventTypeDeeplinkAction:
		appendLogf("INFO", "deeplink-action: %s", truncateText(eventPayload, 160))
	case pluginEvent.EventTypeTransportPacket:
		appendLogf("INFO", "transport-packet: %s", truncateText(eventPayload, 160))
	default:
		appendLogf("WARN", "unknown event type=%d payload=%s", eventType, truncateText(eventPayload, 160))
	}
	return ""
}

func OnUiEvent(eventID string, event ui.Event, eventPayload string) string {
	HandleUIEvent(eventID, event, eventPayload)
	if shouldRerenderAfterUIEvent(event, eventID, eventPayload) {
		RerenderMainUI()
	}
	return ""
}

func OnUiRender(elementID string) {
	RenderMainUI(elementID)
}

func OnCardRender(cardID string) {
	appendLogf("INFO", "card-render: %s", cardID)
}

func shouldRerenderAfterUIEvent(event ui.Event, eventID string, eventPayload string) bool {
	if event == ui.EventInput || event == ui.EventChange {
		if eventID == EventTerminalKeyDown {
			payload, ok := parseUIEventPayload(eventPayload)
			if !ok {
				return false
			}
			return strings.EqualFold(payload.Key, "Enter")
		}
		return false
	}
	return true
}

func truncateText(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return fmt.Sprintf("%s...", value[:max])
}

type uiEventV3Message struct {
	Kind         string `json:"kind"`
	EventID      string `json:"event_id"`
	EventName    string `json:"event"`
	EventPayload string `json:"event_payload"`
}

func tryHandleUIEventV3Message(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}

	var msg uiEventV3Message
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		return false
	}
	if msg.Kind != "ui-event-v3" {
		return false
	}

	event, ok := mapUIEventV3NameToLegacy(msg.EventName)
	if !ok {
		appendLogf("WARN", "unsupported ui-event-v3 event: %s", msg.EventName)
		return true
	}

	HandleUIEvent(msg.EventID, event, msg.EventPayload)
	if shouldRerenderAfterUIEvent(event, msg.EventID, msg.EventPayload) {
		RerenderMainUI()
	}
	return true
}

func mapUIEventV3NameToLegacy(eventName string) (ui.Event, bool) {
	var normalized strings.Builder
	normalized.Grow(len(eventName))
	for _, ch := range strings.TrimSpace(eventName) {
		if ch == '-' || ch == '_' {
			continue
		}
		normalized.WriteRune(unicode.ToUpper(ch))
	}

	switch normalized.String() {
	case "CLICK":
		return ui.EventClick, true
	case "HOVER":
		return ui.EventHover, true
	case "CHANGE":
		return ui.EventChange, true
	case "INPUT":
		return ui.EventInput, true
	case "FOCUS":
		return ui.EventFocus, true
	case "BLUR":
		return ui.EventBlur, true
	case "MOUSEENTER":
		return ui.EventMouseEnter, true
	case "MOUSELEAVE":
		return ui.EventMouseLeave, true
	case "POINTERDOWN":
		return ui.EventPointerDown, true
	case "POINTERUP":
		return ui.EventPointerUp, true
	case "POINTERMOVE":
		return ui.EventPointerMove, true
	case "KEYDOWN":
		return ui.EventInput, true
	case "KEYUP":
		return ui.EventInput, true
	case "LONGPRESS":
		return ui.EventClick, true
	default:
		return 0, false
	}
}
