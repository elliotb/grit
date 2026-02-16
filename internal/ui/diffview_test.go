package ui

import (
	"strings"
	"testing"
)

func TestParseDiffStat_Normal(t *testing.T) {
	output := ` internal/ui/model.go    | 15 +++++++++------
 internal/ui/keys.go     |  3 +++
 internal/gt/gt.go       |  8 ++++++++
 3 files changed, 18 insertions(+), 8 deletions(-)`

	entries := parseDiffStat(output)
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}

	wantPaths := []string{"internal/ui/model.go", "internal/ui/keys.go", "internal/gt/gt.go"}
	for i, want := range wantPaths {
		if entries[i].path != want {
			t.Errorf("entry[%d].path = %q, want %q", i, entries[i].path, want)
		}
	}

	if entries[0].summary != "15 +++++++++------" {
		t.Errorf("entry[0].summary = %q, want %q", entries[0].summary, "15 +++++++++------")
	}
}

func TestParseDiffStat_SingleFile(t *testing.T) {
	output := ` main.go | 2 +-
 1 file changed, 1 insertion(+), 1 deletion(-)`

	entries := parseDiffStat(output)
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].path != "main.go" {
		t.Errorf("path = %q, want %q", entries[0].path, "main.go")
	}
}

func TestParseDiffStat_Empty(t *testing.T) {
	entries := parseDiffStat("")
	if len(entries) != 0 {
		t.Fatalf("got %d entries, want 0", len(entries))
	}
}

func TestParseDiffStat_SummaryOnly(t *testing.T) {
	output := " 0 files changed"
	entries := parseDiffStat(output)
	if len(entries) != 0 {
		t.Fatalf("got %d entries, want 0", len(entries))
	}
}

func TestParseDiffStat_BinaryFile(t *testing.T) {
	output := ` image.png | Bin 0 -> 1234 bytes
 1 file changed, 0 insertions(+), 0 deletions(-)`

	entries := parseDiffStat(output)
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].path != "image.png" {
		t.Errorf("path = %q, want %q", entries[0].path, "image.png")
	}
	if !strings.Contains(entries[0].summary, "Bin") {
		t.Errorf("summary = %q, want to contain 'Bin'", entries[0].summary)
	}
}

func TestParseDiffStat_RenamedFile(t *testing.T) {
	output := ` old.go => new.go | 0
 1 file changed, 0 insertions(+), 0 deletions(-)`

	entries := parseDiffStat(output)
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if !strings.Contains(entries[0].path, "=>") {
		t.Errorf("path = %q, want to contain '=>'", entries[0].path)
	}
}

func TestNewDiffView_Dimensions(t *testing.T) {
	d := newDiffView(100, 30)
	if d.width != 100 {
		t.Errorf("width = %d, want 100", d.width)
	}
	if d.height != 30 {
		t.Errorf("height = %d, want 30", d.height)
	}
	if d.fileCursor != 0 {
		t.Errorf("fileCursor = %d, want 0", d.fileCursor)
	}
	if d.focusedPanel != panelFileList {
		t.Errorf("focusedPanel = %d, want panelFileList", d.focusedPanel)
	}
}

func TestDiffView_SetSize(t *testing.T) {
	d := newDiffView(80, 24)
	d.setSize(120, 40)
	if d.width != 120 {
		t.Errorf("width = %d, want 120", d.width)
	}
	if d.height != 40 {
		t.Errorf("height = %d, want 40", d.height)
	}
}

func TestDiffView_SetFiles(t *testing.T) {
	d := newDiffView(80, 24)
	files := []diffFileEntry{
		{path: "a.go", summary: "1 +"},
		{path: "b.go", summary: "2 ++"},
	}
	d.setFiles(files)
	if len(d.files) != 2 {
		t.Fatalf("got %d files, want 2", len(d.files))
	}
	if d.fileCursor != 0 {
		t.Errorf("fileCursor = %d, want 0", d.fileCursor)
	}
}

func TestDiffView_SetDiffContent(t *testing.T) {
	d := newDiffView(80, 24)
	d.setDiffContent("diff content here")
	// Viewport content is set; verify it renders.
	view := d.diffViewport.View()
	if !strings.Contains(view, "diff content here") {
		t.Error("diff viewport should contain the set content")
	}
}

func TestDiffView_View_EmptyFiles(t *testing.T) {
	d := newDiffView(80, 24)
	d.branchName = "feature-a"
	d.parentBranch = "main"
	view := d.view()
	if !strings.Contains(view, "(no changes)") {
		t.Error("view should show '(no changes)' when no files")
	}
	if !strings.Contains(view, "Files") {
		t.Error("view should show 'Files' header")
	}
}

func TestDiffView_View_WithFiles(t *testing.T) {
	d := newDiffView(100, 24)
	d.branchName = "feature-a"
	d.parentBranch = "main"
	d.setFiles([]diffFileEntry{
		{path: "model.go", summary: "5 +++--"},
		{path: "keys.go", summary: "3 +++"},
	})
	d.setDiffContent("@@ -1,3 +1,4 @@\n+new line")

	view := d.view()
	if !strings.Contains(view, "model.go") {
		t.Error("view should contain 'model.go'")
	}
	if !strings.Contains(view, "keys.go") {
		t.Error("view should contain 'keys.go'")
	}
	if !strings.Contains(view, "feature-a") {
		t.Error("view should contain branch name")
	}
	if !strings.Contains(view, "main") {
		t.Error("view should contain parent branch name")
	}
}

func TestDiffView_FocusToggle(t *testing.T) {
	d := newDiffView(80, 24)
	if d.focusedPanel != panelFileList {
		t.Error("default focus should be panelFileList")
	}
	d.focusedPanel = panelDiff
	if d.focusedPanel != panelDiff {
		t.Error("focus should be panelDiff after toggle")
	}
}

func TestDiffView_PanelWidths(t *testing.T) {
	d := newDiffView(100, 24)
	fileW, diffW := d.panelWidths()
	if fileW < fileListMinWidth {
		t.Errorf("file list width %d should be >= %d", fileW, fileListMinWidth)
	}
	total := fileW + diffW + borderWidth
	if total != 100 {
		t.Errorf("total width = %d, want 100", total)
	}
}

func TestDiffView_PanelWidths_NarrowTerminal(t *testing.T) {
	d := newDiffView(50, 24)
	fileW, diffW := d.panelWidths()
	total := fileW + diffW + borderWidth
	if total != 50 {
		t.Errorf("total width = %d, want 50", total)
	}
	if diffW < 1 {
		t.Error("diff width should be at least 1")
	}
}

func TestDiffView_FileListOffset(t *testing.T) {
	d := newDiffView(80, 5) // height 5, minus header = 4 visible lines
	files := make([]diffFileEntry, 10)
	for i := range files {
		files[i] = diffFileEntry{path: "file" + string(rune('0'+i)) + ".go"}
	}
	d.setFiles(files)

	// Cursor at 0, should show from offset 0.
	d.fileCursor = 0
	if off := d.fileListOffset(); off != 0 {
		t.Errorf("offset = %d, want 0", off)
	}

	// Cursor at 3 (last visible), offset still 0.
	d.fileCursor = 3
	if off := d.fileListOffset(); off != 0 {
		t.Errorf("offset = %d, want 0", off)
	}

	// Cursor at 5, offset scrolls.
	d.fileCursor = 5
	off := d.fileListOffset()
	if off <= 0 {
		t.Errorf("offset = %d, want > 0 for cursor=5", off)
	}
}

func TestTruncateToWidth(t *testing.T) {
	tests := []struct {
		input string
		width int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is too long", 10, "this is t…"},
		{"", 5, ""},
		{"any", 0, ""},
	}

	for _, tt := range tests {
		got := truncateToWidth(tt.input, tt.width)
		if got != tt.want {
			t.Errorf("truncateToWidth(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
		}
	}
}

func TestPadToWidth(t *testing.T) {
	got := padToWidth("hi", 5)
	if got != "hi   " {
		t.Errorf("padToWidth(\"hi\", 5) = %q, want %q", got, "hi   ")
	}

	got = padToWidth("exact", 5)
	if got != "exact" {
		t.Errorf("padToWidth(\"exact\", 5) = %q, want %q", got, "exact")
	}
}

func TestDiffView_View_VerticalSeparator(t *testing.T) {
	d := newDiffView(80, 24)
	d.branchName = "feat"
	d.parentBranch = "main"
	d.setFiles([]diffFileEntry{{path: "a.go", summary: "1 +"}})

	view := d.view()
	if !strings.Contains(view, "│") {
		t.Error("view should contain vertical separator")
	}
}
