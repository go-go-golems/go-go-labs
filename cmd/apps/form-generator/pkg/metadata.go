package pkg

import (
	"encoding/base64"
	"strings"
)

const metadataPrefix = "uhoh::"

// encodeMetadata stores step and field identifiers in a compact string that can be
// round-tripped even if the IDs contain spaces or special characters.
func encodeMetadata(stepID, fieldKey string) string {
	stepID = strings.TrimSpace(stepID)
	fieldKey = strings.TrimSpace(fieldKey)
	if stepID == "" && fieldKey == "" {
		return ""
	}

	parts := []string{}
	if stepID != "" {
		parts = append(parts, base64.RawURLEncoding.EncodeToString([]byte(stepID)))
	} else {
		parts = append(parts, "")
	}
	if fieldKey != "" {
		parts = append(parts, base64.RawURLEncoding.EncodeToString([]byte(fieldKey)))
	}

	return metadataPrefix + strings.Join(parts, ":")
}

// decodeMetadata reverses encodeMetadata.
func decodeMetadata(value string) (stepID string, fieldKey string, ok bool) {
	if !strings.HasPrefix(value, metadataPrefix) {
		return "", "", false
	}
	payload := strings.TrimPrefix(value, metadataPrefix)
	if payload == "" {
		return "", "", false
	}
	parts := strings.Split(payload, ":")
	decode := func(in string) string {
		if in == "" {
			return ""
		}
		b, err := base64.RawURLEncoding.DecodeString(in)
		if err != nil {
			return ""
		}
		return string(b)
	}
	if len(parts) > 0 {
		stepID = decode(parts[0])
	}
	if len(parts) > 1 {
		fieldKey = decode(parts[1])
	}
	return stepID, fieldKey, true
}
