package gt

import "testing"

func TestParsePRInfo_ValidJSON(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantNumber int
		wantState  string
	}{
		{
			name:       "open PR",
			input:      `{"prNumber": 142, "state": "OPEN", "title": "Add auth"}`,
			wantNumber: 142,
			wantState:  "OPEN",
		},
		{
			name:       "draft PR",
			input:      `{"prNumber": 143, "state": "DRAFT", "title": "WIP: tests"}`,
			wantNumber: 143,
			wantState:  "DRAFT",
		},
		{
			name:       "merged PR",
			input:      `{"prNumber": 100, "state": "MERGED", "title": "Fix bug"}`,
			wantNumber: 100,
			wantState:  "MERGED",
		},
		{
			name:       "closed PR",
			input:      `{"prNumber": 50, "state": "CLOSED", "title": "Abandoned"}`,
			wantNumber: 50,
			wantState:  "CLOSED",
		},
		{
			name:       "no title field",
			input:      `{"prNumber": 200, "state": "OPEN"}`,
			wantNumber: 200,
			wantState:  "OPEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ParsePRInfo(tt.input)
			if info.Number != tt.wantNumber {
				t.Errorf("Number = %d, want %d", info.Number, tt.wantNumber)
			}
			if info.State != tt.wantState {
				t.Errorf("State = %q, want %q", info.State, tt.wantState)
			}
		})
	}
}

func TestParsePRInfo_EmptyInput(t *testing.T) {
	info := ParsePRInfo("")
	if info.Number != 0 {
		t.Errorf("Number = %d, want 0", info.Number)
	}
	if info.State != "" {
		t.Errorf("State = %q, want empty", info.State)
	}
}

func TestParsePRInfo_WhitespaceOnly(t *testing.T) {
	info := ParsePRInfo("  \n  ")
	if info.Number != 0 {
		t.Errorf("Number = %d, want 0", info.Number)
	}
}

func TestParsePRInfo_MalformedJSON(t *testing.T) {
	info := ParsePRInfo("not json at all")
	if info.Number != 0 {
		t.Errorf("Number = %d, want 0", info.Number)
	}
	if info.State != "" {
		t.Errorf("State = %q, want empty", info.State)
	}
}

func TestParsePRInfo_EmptyJSON(t *testing.T) {
	info := ParsePRInfo("{}")
	if info.Number != 0 {
		t.Errorf("Number = %d, want 0", info.Number)
	}
	if info.State != "" {
		t.Errorf("State = %q, want empty", info.State)
	}
}

func TestParsePRInfo_WithWhitespace(t *testing.T) {
	input := `  {"prNumber": 142, "state": "OPEN"}  `
	info := ParsePRInfo(input)
	if info.Number != 142 {
		t.Errorf("Number = %d, want 142", info.Number)
	}
}
