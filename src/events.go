package plugin

import (
	ui "astroboxplugin/bindings/astrobox_psys_host_ui_v3"
	pluginEvent "astroboxplugin/bindings/astrobox_psys_plugin_event_v3"
	"fmt"
	"strings"
)

func OnEvent(eventType pluginEvent.EventType, eventPayload string) string {
	switch eventType {
	case pluginEvent.EventTypeInterconnectMessage:
		if handleInterconnectEventPayload(eventPayload) {
			RerenderMainUI()
		}
	case pluginEvent.EventTypeTimer:
		payloadText, err := ExtractPayloadText(eventPayload)
		if err != nil {
			break
		}
		if handleQueueDrainPayloadText(payloadText) {
			break
		}
		if handleRpcTimeoutPayloadText(payloadText) {
			RerenderMainUI()
		}
	case pluginEvent.EventTypePluginMessage:
		appendLogf("INFO", "plugin-message: %s", truncateText(eventPayload, 160))
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

func OnUiEventV3(eventID string, event ui.Event, eventPayload string) string {
	appendLogf("INFO", "OnUiEventV3 called! id=%s, event=%d", eventID, event)
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
	if event == ui.EventInput {
		return eventID == EventFileSearchInput
	}
	if event == ui.EventChange {
		_ = eventPayload
		return false
	}
	if event == ui.EventKeyDown {
		payload, ok := parseUIEventPayload(eventPayload)
		if ok && strings.EqualFold(payload.Key, "Enter") {
			return true
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


