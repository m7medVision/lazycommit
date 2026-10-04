package domain

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ErrEmptyDiff signals that there are no changes to describe. Callers use it
// to short-circuit before any backend is invoked.
var ErrEmptyDiff = errors.New("diff is empty")

// DefaultMaxDiffChars bounds the diff sent to a backend (roughly 10k
// tokens), so large change sets do not exceed model context or rate limits.
const DefaultMaxDiffChars = 40000

// MinDiffBudget is the smallest budget Compact honours; smaller values are
// raised to it so the summary header always fits.
const MinDiffBudget = 1000

const truncatedMarker = "[truncated]\n"

// noiseFiles are generated files whose diffs carry no meaning for a commit
// message; only their names and line counts are kept when compacting.
var noiseFiles = map[string]bool{
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"go.sum":            true,
	"Cargo.lock":        true,
	"poetry.lock":       true,
	"uv.lock":           true,
	"composer.lock":     true,
	"Gemfile.lock":      true,
	"Pipfile.lock":      true,
}

var noiseSuffixes = []string{".min.js", ".min.css", ".js.map", ".css.map"}

// Diff is a non-empty git diff.
type Diff struct {
	text string
}

func NewDiff(text string) (Diff, error) {
	if strings.TrimSpace(text) == "" {
		return Diff{}, ErrEmptyDiff
	}
	return Diff{text: text}, nil
}

func (d Diff) String() string {
	return d.text
}

// Compact returns a diff of at most budget bytes; a positive budget below
// MinDiffBudget is raised to it. A diff that already fits is returned
// unchanged. Otherwise the result starts with a summary of the changed
// files (cut short with a count when the list alone would not fit),
// followed by every file diff that fits in full; leftover space goes
// to the first file that did not fit, cut at a line boundary. Lock files,
// minified files, and binary files appear only in the summary.
func (d Diff) Compact(budget int) Diff {
	if budget <= 0 {
		return d
	}
	budget = max(budget, MinDiffBudget)
	if len(d.text) <= budget {
		return d
	}

	files := splitFileDiffs(d.text)
	if len(files) == 0 {
		return Diff{text: truncate(d.text, budget)}
	}

	var out strings.Builder
	writeSummary(&out, files, budget)
	var skipped *fileDiff
	for i := range files {
		f := &files[i]
		if f.binary || f.noise {
			continue
		}
		if out.Len()+len(f.text) <= budget {
			out.WriteString(f.text)
		} else if skipped == nil {
			skipped = f
		}
	}
	if skipped != nil {
		out.WriteString(truncate(skipped.text, budget-out.Len()))
	}
	return Diff{text: out.String()}
}

type fileDiff struct {
	text    string
	path    string
	added   int
	removed int
	binary  bool
	noise   bool
}

// splitFileDiffs splits a unified git diff into per-file sections. Text
// before the first "diff --git" header is dropped.
func splitFileDiffs(text string) []fileDiff {
	const header = "diff --git "
	var starts []int
	if strings.HasPrefix(text, header) {
		starts = append(starts, 0)
	}
	for i := 0; ; {
		j := strings.Index(text[i:], "\n"+header)
		if j < 0 {
			break
		}
		i += j + 1
		starts = append(starts, i)
	}

	files := make([]fileDiff, len(starts))
	for k, start := range starts {
		end := len(text)
		if k+1 < len(starts) {
			end = starts[k+1]
		}
		files[k] = newFileDiff(text[start:end])
	}
	return files
}

func newFileDiff(text string) fileDiff {
	f := fileDiff{text: text}
	var header, oldPath, newPath, renamePath string
	inHunk := false
	for line := range strings.Lines(text) {
		line = strings.TrimSuffix(line, "\n")
		if header == "" {
			header = line
			continue
		}
		if inHunk {
			switch {
			case strings.HasPrefix(line, "+"):
				f.added++
			case strings.HasPrefix(line, "-"):
				f.removed++
			}
			continue
		}
		switch {
		case strings.HasPrefix(line, "@@"):
			inHunk = true
		case strings.HasPrefix(line, "Binary files "), strings.HasPrefix(line, "GIT binary patch"):
			f.binary = true
		case strings.HasPrefix(line, "--- "):
			oldPath = markerPath(line[len("--- "):])
		case strings.HasPrefix(line, "+++ "):
			newPath = markerPath(line[len("+++ "):])
		case strings.HasPrefix(line, "rename to "):
			renamePath = unquotePath(line[len("rename to "):])
		}
	}

	switch {
	case newPath != "":
		f.path = newPath
	case oldPath != "":
		f.path = oldPath
	case renamePath != "":
		f.path = renamePath
	default:
		f.path = headerPath(header)
	}

	base := path.Base(f.path)
	f.noise = noiseFiles[base]
	for _, suffix := range noiseSuffixes {
		if strings.HasSuffix(base, suffix) {
			f.noise = true
		}
	}
	return f
}

// markerPath extracts the path from the value of a "--- " or "+++ " line.
// It returns "" for /dev/null, the side of an added or deleted file.
func markerPath(value string) string {
	value = strings.TrimSuffix(value, "\t") // git appends a tab to names with spaces
	if value == "/dev/null" {
		return ""
	}
	return stripSidePrefix(unquotePath(value))
}

// headerPath reads the path from a "diff --git" header. It is the fallback
// for sections without "---", "+++", or "rename to" lines: binary and
// mode-only changes, where both sides name the same path.
func headerPath(header string) string {
	rest := strings.TrimPrefix(header, "diff --git ")
	// "a/X b/X" splits exactly in the middle, even when X contains " b/".
	if mid := len(rest) / 2; len(rest)%2 == 1 && rest[mid] == ' ' &&
		strings.HasPrefix(rest, "a/") && rest[mid+1:mid+3] == "b/" && rest[2:mid] == rest[mid+3:] {
		return rest[mid+3:]
	}
	if i := strings.LastIndex(rest, ` "b/`); i >= 0 {
		return stripSidePrefix(unquotePath(rest[i+1:]))
	}
	if i := strings.LastIndex(rest, " b/"); i >= 0 {
		return rest[i+len(" b/"):]
	}
	return rest
}

// unquotePath decodes git's C-style quoting, used for paths with special
// or non-ASCII characters.
func unquotePath(p string) string {
	if strings.HasPrefix(p, `"`) {
		if u, err := strconv.Unquote(p); err == nil {
			return u
		}
	}
	return p
}

func stripSidePrefix(p string) string {
	if strings.HasPrefix(p, "a/") || strings.HasPrefix(p, "b/") {
		return p[2:]
	}
	return p
}

// writeSummary lists every file with its line counts. When the list would
// not fit in the budget, the remaining files are replaced by a count.
func writeSummary(out *strings.Builder, files []fileDiff, budget int) {
	fmt.Fprintf(out, "Diff shortened to fit the size limit. Changed files (%d):\n", len(files))
	for i, f := range files {
		line := fmt.Sprintf("%s +%d -%d\n", f.path, f.added, f.removed)
		switch {
		case f.binary:
			line = fmt.Sprintf("%s (binary, diff omitted)\n", f.path)
		case f.noise:
			line = fmt.Sprintf("%s +%d -%d (generated, diff omitted)\n", f.path, f.added, f.removed)
		}
		tail := len("\n")
		if rest := len(files) - i - 1; rest > 0 {
			tail += len(moreFilesLine(rest))
		}
		if out.Len()+len(line)+tail > budget {
			out.WriteString(moreFilesLine(len(files) - i))
			break
		}
		out.WriteString(line)
	}
	out.WriteString("\n")
}

func moreFilesLine(n int) string {
	if n == 1 {
		return "... and 1 more file\n"
	}
	return fmt.Sprintf("... and %d more files\n", n)
}

// truncate cuts text to at most n bytes and marks the cut; text that
// already fits is returned unchanged. It cuts at a line boundary when one
// fits, otherwise at a rune boundary. It returns "" when n leaves no room
// for content.
func truncate(text string, n int) string {
	if len(text) <= n {
		return text
	}
	limit := n - len(truncatedMarker)
	if limit <= 0 {
		return ""
	}
	if cut := strings.LastIndexByte(text[:limit], '\n'); cut >= 0 {
		return text[:cut+1] + truncatedMarker
	}
	cut := limit - len("\n")
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	if cut == 0 {
		return ""
	}
	return text[:cut] + "\n" + truncatedMarker
}
