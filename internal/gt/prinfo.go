package gt

import (
	"encoding/json"
	"strings"
)

// prInfoJSON matches the JSON output of `gt branch pr-info`.
type prInfoJSON struct {
	PRNumber int    `json:"prNumber"`
	State    string `json:"state"`
}

// ParsePRInfo parses the JSON output of `gt branch pr-info` into a PRInfo.
// Returns a zero-value PRInfo if the output is empty or unparseable.
func ParsePRInfo(output string) PRInfo {
	output = strings.TrimSpace(output)
	if output == "" {
		return PRInfo{}
	}

	var raw prInfoJSON
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		return PRInfo{}
	}

	return PRInfo{
		Number: raw.PRNumber,
		State:  raw.State,
	}
}
