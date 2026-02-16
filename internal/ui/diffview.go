package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// viewMode distinguishes the active screen.
type viewMode int

const (
	modeTree viewMode = iota
	modeDiff
	modeHelp
)

// diffPanel tracks which panel has focus in the diff view.
type diffPanel int

const (
	panelFileList diffPanel = iota
	panelDiff
)

// diffFileEntry represents a single file from git diff --stat output.
type diffFileEntry struct {
	path    string
	summary string // e.g. "5 +++--"
}

// diffView holds all state for the diff view.
type diffView struct {
	branchName   string
	parentBranch string
	files        []diffFileEntry
	fileCursor   int
	diffViewport viewport.Model
	focusedPanel diffPanel
	width        int
	height       int
}

const (
	// fileListWidthFraction is the fraction of width for the file list panel.
	fileListMinWidth = 30
	fileListMaxFrac  = 0.35
	// borderWidth is the width of the vertical separator between panels.
	borderWidth = 1
)

var (
	diffHeaderStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	diffFileStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	diffFileSelectedStyle = lipgloss.NewStyle().Bold(true).Reverse(true)
	diffBorderStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	diffPanelHeaderStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7"))
	diffPanelFocusedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
)

func newDiffView(width, height int) diffView {
	d := diffView{
		width:  width,
		height: height,
	}
	_, diffWidth := d.panelWidths()
	vpHeight := height - 1 // minus header line
	if vpHeight < 1 {
		vpHeight = 1
	}
	d.diffViewport = viewport.New(diffWidth, vpHeight)
	d.diffViewport.KeyMap = viewport.KeyMap{}
	return d
}

func (d *diffView) setSize(width, height int) {
	d.width = width
	d.height = height
	_, diffWidth := d.panelWidths()
	vpHeight := height - 1
	if vpHeight < 1 {
		vpHeight = 1
	}
	d.diffViewport.Width = diffWidth
	d.diffViewport.Height = vpHeight
}

// panelWidths returns the widths for the file list and diff panels.
func (d diffView) panelWidths() (fileListWidth, diffWidth int) {
	fileListWidth = int(float64(d.width) * fileListMaxFrac)
	if fileListWidth < fileListMinWidth && d.width > fileListMinWidth+10 {
		fileListWidth = fileListMinWidth
	}
	if fileListWidth > d.width-10 {
		fileListWidth = d.width / 3
	}
	diffWidth = d.width - fileListWidth - borderWidth
	if diffWidth < 1 {
		diffWidth = 1
	}
	return
}

func (d *diffView) setFiles(files []diffFileEntry) {
	d.files = files
	d.fileCursor = 0
}

func (d *diffView) setDiffContent(content string) {
	d.diffViewport.SetContent(content)
	d.diffViewport.SetYOffset(0)
}

// ensureFileCursorVisible returns the offset for the file list so the cursor is visible.
func (d diffView) fileListOffset() int {
	listHeight := d.height - 1 // minus header
	if listHeight < 1 {
		listHeight = 1
	}
	offset := 0
	if d.fileCursor >= listHeight {
		offset = d.fileCursor - listHeight + 1
	}
	return offset
}

func (d diffView) view() string {
	fileListWidth, diffWidth := d.panelWidths()

	// Header for file list panel.
	fileHeaderStyle := diffPanelHeaderStyle
	diffHeaderSt := diffPanelHeaderStyle
	if d.focusedPanel == panelFileList {
		fileHeaderStyle = diffPanelFocusedStyle
	} else {
		diffHeaderSt = diffPanelFocusedStyle
	}

	fileHeader := fileHeaderStyle.Render(truncateToWidth("Files", fileListWidth))
	diffHeader := diffHeaderSt.Render(truncateToWidth(
		"Diff: "+d.branchName+" (vs "+d.parentBranch+")", diffWidth))

	// Render file list.
	listHeight := d.height - 1
	if listHeight < 1 {
		listHeight = 1
	}

	var fileLines []string
	if len(d.files) == 0 {
		fileLines = append(fileLines, diffFileStyle.Render("(no changes)"))
	} else {
		offset := d.fileListOffset()
		end := offset + listHeight
		if end > len(d.files) {
			end = len(d.files)
		}
		for i := offset; i < end; i++ {
			name := d.files[i].path
			displayName := truncateToWidth(name, fileListWidth)
			if i == d.fileCursor {
				fileLines = append(fileLines, diffFileSelectedStyle.Render(padToWidth(displayName, fileListWidth)))
			} else {
				fileLines = append(fileLines, diffFileStyle.Render(padToWidth(displayName, fileListWidth)))
			}
		}
	}

	// Pad file list to full height.
	for len(fileLines) < listHeight {
		fileLines = append(fileLines, strings.Repeat(" ", fileListWidth))
	}

	fileListContent := strings.Join(fileLines, "\n")

	// Build separator column.
	separator := diffBorderStyle.Render("│")
	var sepLines []string
	sepLines = append(sepLines, separator) // header line separator
	for i := 0; i < listHeight; i++ {
		sepLines = append(sepLines, separator)
	}
	sepContent := strings.Join(sepLines, "\n")

	// Build diff panel (header + viewport).
	diffContent := diffHeader + "\n" + d.diffViewport.View()

	// File list panel (header + content).
	filePanel := fileHeader + "\n" + fileListContent

	return lipgloss.JoinHorizontal(lipgloss.Top, filePanel, sepContent, diffContent)
}

// truncateToWidth truncates a string to fit within the given width,
// accounting for ANSI escape sequences.
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := ansi.StringWidth(s)
	if w <= width {
		return s
	}
	return ansi.Truncate(s, width, "…")
}

// padToWidth pads a string with spaces to reach the given width.
func padToWidth(s string, width int) string {
	w := ansi.StringWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// parseDiffStat parses the output of `git diff --stat` into file entries.
func parseDiffStat(output string) []diffFileEntry {
	lines := strings.Split(output, "\n")
	var entries []diffFileEntry

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip the summary line (e.g. "3 files changed, 18 insertions(+), 8 deletions(-)")
		if strings.Contains(line, "file changed") || strings.Contains(line, "files changed") {
			continue
		}
		// Lines are formatted as: " path | stat"
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		path := strings.TrimSpace(parts[0])
		summary := strings.TrimSpace(parts[1])
		if path != "" {
			entries = append(entries, diffFileEntry{path: path, summary: summary})
		}
	}

	return entries
}
