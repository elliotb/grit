package gt

import (
	"strings"
	"unicode/utf8"
)

// Branch represents a single branch in the Graphite stack tree.
type Branch struct {
	Name      string
	IsCurrent bool
	Children  []*Branch
}

// parsedLine holds the extracted data from a single line of gt log short output.
type parsedLine struct {
	name      string
	depth     int
	isCurrent bool
}

// ParseLogShort parses the output of `gt log short` into a tree of branches.
// Returns a slice of root branches (typically one trunk like main/master).
func ParseLogShort(output string) ([]*Branch, error) {
	lines := strings.Split(output, "\n")

	var parsed []parsedLine
	for _, line := range lines {
		pl, ok := parseLine(line)
		if ok {
			parsed = append(parsed, pl)
		}
	}

	if len(parsed) == 0 {
		return nil, nil
	}

	// Reverse: gt log short lists top-of-stack first, trunk last.
	// We want trunk first so we can build parent→child relationships.
	for i, j := 0, len(parsed)-1; i < j; i, j = i+1, j-1 {
		parsed[i], parsed[j] = parsed[j], parsed[i]
	}

	// Build tree. The first entry (after reversal) is the trunk/root.
	root := &Branch{Name: parsed[0].name, IsCurrent: parsed[0].isCurrent}
	roots := []*Branch{root}

	// parentAtDepth tracks the "tip" branch at each depth level.
	// A new branch at depth d chains onto parentAtDepth[d] (same-depth continuation)
	// or becomes a child of parentAtDepth[d-1] (new deeper level).
	parentAtDepth := map[int]*Branch{parsed[0].depth: root}
	prevDepth := parsed[0].depth

	for i := 1; i < len(parsed); i++ {
		p := parsed[i]
		b := &Branch{Name: p.name, IsCurrent: p.isCurrent}

		switch {
		case p.depth == 0:
			// Another trunk-level branch: child of the root.
			root.Children = append(root.Children, b)
			parentAtDepth[0] = b

		case p.depth == prevDepth:
			// Same depth as previous: chain within the stack.
			// This branch is a child of the current tip at this depth.
			if parent, ok := parentAtDepth[p.depth]; ok {
				parent.Children = append(parent.Children, b)
			}
			parentAtDepth[p.depth] = b

		case p.depth > prevDepth:
			// Going deeper: start of a new stack level.
			// Parent is the tip at the depth above.
			if parent, ok := parentAtDepth[p.depth-1]; ok {
				parent.Children = append(parent.Children, b)
			}
			parentAtDepth[p.depth] = b

		default:
			// Going shallower: finished a deeper stack.
			// Clear stale deeper entries.
			for d := p.depth + 1; d <= prevDepth; d++ {
				delete(parentAtDepth, d)
			}
			// Attach to parent at depth-1, or root if depth 0.
			if p.depth > 0 {
				if parent, ok := parentAtDepth[p.depth-1]; ok {
					parent.Children = append(parent.Children, b)
				}
			}
			parentAtDepth[p.depth] = b
		}

		prevDepth = p.depth
	}

	return roots, nil
}

const (
	currentMarker = '◉'
	otherMarker   = '◯'
)

// parseLine extracts branch info from a single line of gt log short output.
// Returns false if the line doesn't contain a branch marker.
func parseLine(line string) (parsedLine, bool) {
	// Find the first ◉ or ◯ in the line by scanning runes.
	runePos := 0
	markerFound := false
	var isCurrent bool
	byteOffset := 0

	for byteOffset < len(line) {
		r, size := utf8.DecodeRuneInString(line[byteOffset:])
		if r == currentMarker {
			isCurrent = true
			markerFound = true
			byteOffset += size
			break
		}
		if r == otherMarker {
			isCurrent = false
			markerFound = true
			byteOffset += size
			break
		}
		runePos++
		byteOffset += size
	}

	if !markerFound {
		return parsedLine{}, false
	}

	// Depth is determined by the rune column of the marker.
	// Column 0 = depth 0, column 2 = depth 1, etc.
	depth := runePos / 2

	// Extract the branch name: everything after the marker, stripped of
	// connector chars (─, ┘) and whitespace.
	rest := line[byteOffset:]
	name := stripConnectors(rest)
	name = strings.TrimSpace(name)

	if name == "" {
		return parsedLine{}, false
	}

	return parsedLine{
		name:      name,
		depth:     depth,
		isCurrent: isCurrent,
	}, true
}

// stripConnectors removes tree-drawing characters (─, ┘) from the string.
func stripConnectors(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '─', '┘', '│':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
