package plugin

import (
	"encoding/json"
	"errors"
	"strings"
)

var errEmptyPayload = errors.New("event payload is empty")

// ExtractPayloadText unwraps AstroBox interconnect event payloads.
// The host may wrap actual data in payloadText/payload fields.
func ExtractPayloadText(eventPayload string) (string, error) {
	raw := strings.TrimSpace(eventPayload)
	if raw == "" {
		return "", errEmptyPayload
	}

	var wrapper map[string]any
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return raw, nil
	}

	if v, ok := wrapper["payloadText"]; ok {
		switch text := v.(type) {
		case string:
			text = strings.TrimSpace(text)
			if text != "" {
				return text, nil
			}
		default:
			buf, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(buf), nil
		}
	}

	if v, ok := wrapper["payload"]; ok {
		switch text := v.(type) {
		case string:
			text = strings.TrimSpace(text)
			if text != "" {
				return text, nil
			}
		default:
			buf, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(buf), nil
		}
	}

	return raw, nil
}
