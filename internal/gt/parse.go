package gt

import (
	"strings"
	"unicode/utf8"
)

// Branch represents a single branch in the Graphite stack tree.
type Branch struct {
	Name       string
	IsCurrent  bool
	Annotation string // e.g. "needs restack", "merging", "" if none
	Children   []*Branch
}

// parsedLine holds the extracted data from a single line of gt log short output.
type parsedLine struct {
	name       string
	depth      int
	isCurrent  bool
	annotation string
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
	root := &Branch{Name: parsed[0].name, IsCurrent: parsed[0].isCurrent, Annotation: parsed[0].annotation}
	roots := []*Branch{root}

	// parentAtDepth tracks the "tip" branch at each depth level.
	// A new branch at depth d chains onto parentAtDepth[d] (same-depth continuation)
	// or becomes a child of parentAtDepth[d-1] (new deeper level).
	parentAtDepth := map[int]*Branch{parsed[0].depth: root}
	prevDepth := parsed[0].depth

	for i := 1; i < len(parsed); i++ {
		p := parsed[i]
		b := &Branch{Name: p.name, IsCurrent: p.isCurrent, Annotation: p.annotation}

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
	// connector chars (─, ┘) and whitespace. Also strip any trailing
	// parenthesized annotations like "(merging)" or "(needs restack)"
	// that gt may append to branch names.
	rest := line[byteOffset:]
	name := stripConnectors(rest)
	name = strings.TrimSpace(name)
	name, annotation := extractAnnotation(name)

	if name == "" {
		return parsedLine{}, false
	}

	return parsedLine{
		name:       name,
		depth:      depth,
		isCurrent:  isCurrent,
		annotation: annotation,
	}, true
}

// extractAnnotation splits a trailing parenthesized annotation from a branch
// name, e.g. "my-branch (merging)" → ("my-branch", "merging").
// Returns an empty annotation if none is present.
func extractAnnotation(name string) (string, string) {
	if idx := strings.LastIndex(name, " ("); idx != -1 && strings.HasSuffix(name, ")") {
		annotation := name[idx+2 : len(name)-1]
		return name[:idx], annotation
	}
	return name, ""
}

// FindParent searches the branch tree for the parent of the named branch.
// Returns the parent name and true if found, or ("", false) if the branch
// is a root or not present in the tree.
func FindParent(branches []*Branch, name string) (string, bool) {
	for _, root := range branches {
		if found, parent := findParentRecursive(root, name); found {
			return parent, true
		}
	}
	return "", false
}

// findParentRecursive walks the tree rooted at node, returning (true, parentName)
// if name is found among its descendants.
func findParentRecursive(node *Branch, name string) (bool, string) {
	for _, child := range node.Children {
		if child.Name == name {
			return true, node.Name
		}
		if found, parent := findParentRecursive(child, name); found {
			return true, parent
		}
	}
	return false, ""
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
