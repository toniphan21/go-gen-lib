package util

import (
	"encoding/json"
	"strings"
)

func ParseOptionsInMarkdownFile(content string) map[string]any {
	var result map[string]any
	var rawOptions string
	lines := strings.Split(content, "\n")
	for _, v := range lines {
		line := strings.TrimSpace(v)
		if strings.HasPrefix(line, "[//]: # (Options:") && strings.HasSuffix(line, ")") {
			line = strings.TrimPrefix(line, "[//]: # (Options:")
			line = strings.TrimSuffix(line, ")")
			rawOptions = line
			continue
		}
	}

	if rawOptions != "" {
		_ = json.Unmarshal([]byte(rawOptions), &result)
	}
	return result
}
