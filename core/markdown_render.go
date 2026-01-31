package core

import (
	"regexp"
	"strings"

	com "github.com/amterp/rad/core/common"
)

// Color scheme: "Sunset Terminal" — matching the docs website theme
var (
	colorCoral = com.NewRgb(232, 93, 76)   // H1 headers — primary brand color
	colorAmber = com.NewRgb(244, 162, 97)  // H2+ headers — warning/accent
	colorTeal  = com.NewRgb(45, 212, 191)  // Code — accent color
)

// styledSpan represents a segment of text with a particular style
type styledSpan struct {
	start int
	end   int
	text  string
	style string // "bold", "italic", "code", "url"
}

// Regex patterns for markdown parsing
var (
	// Headers: match # at start of line followed by space and text
	headerPattern = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)

	// Bold: **text** (non-greedy)
	boldPattern = regexp.MustCompile(`\*\*(.+?)\*\*`)

	// Inline code: `code` (non-greedy)
	inlineCodePattern = regexp.MustCompile("`([^`]+)`")

	// URLs: http:// or https:// followed by non-whitespace
	urlPattern = regexp.MustCompile(`https?://[^\s]+`)

	// Code fence: ``` with optional language (may be indented)
	codeFencePattern = regexp.MustCompile("^\\s*```")
)

// RenderMarkdownForTerminal converts markdown text to a styled RadString
// for terminal display. It's a simple line-by-line renderer designed for
// the well-structured error documentation format.
func RenderMarkdownForTerminal(markdown string) RadString {
	lines := strings.Split(markdown, "\n")
	var result RadString

	inCodeBlock := false
	codeBlockBaseIndent := "" // The indentation of the opening fence
	const codeBlockIndent = "    "

	for i, line := range lines {
		var renderedLine RadString

		if codeFencePattern.MatchString(line) {
			if !inCodeBlock {
				// Entering code block — capture the fence's indentation
				codeBlockBaseIndent = extractLeadingWhitespace(line)
			}
			// Toggle code block state, skip the fence line itself
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			// Inside code block: strip the base indentation, then add our standard indent
			strippedLine := strings.TrimPrefix(line, codeBlockBaseIndent)
			renderedLine = renderCodeLine(codeBlockIndent + strippedLine)
		} else {
			// Outside code block: process markdown syntax
			renderedLine = renderMarkdownLine(line)
		}

		// Append newline between lines (but not after the last line)
		if i > 0 || result.Len() > 0 {
			result = result.ConcatStr("\n")
		}
		result = result.Concat(renderedLine)
	}

	return result
}

// extractLeadingWhitespace returns the leading whitespace from a string
func extractLeadingWhitespace(s string) string {
	for i, r := range s {
		if r != ' ' && r != '\t' {
			return s[:i]
		}
	}
	return s
}

// renderMarkdownLine processes a single line of markdown outside of code blocks
func renderMarkdownLine(line string) RadString {
	// Check for headers first
	if match := headerPattern.FindStringSubmatch(line); match != nil {
		level := len(match[1]) // Number of # characters
		text := match[2]
		return renderHeader(text, level)
	}

	// Process inline formatting
	return renderInlineFormatting(line)
}

// renderHeader creates a styled header
func renderHeader(text string, level int) RadString {
	// Process any inline formatting within the header text first
	rendered := renderInlineFormatting(text)

	// Apply header styling
	rendered.SetAttr(BOLD)
	if level == 1 {
		rendered.SetRgb(colorCoral.R, colorCoral.G, colorCoral.B)
	} else {
		rendered.SetRgb(colorAmber.R, colorAmber.G, colorAmber.B)
	}

	return rendered
}

// renderCodeLine creates a teal-colored code line
func renderCodeLine(line string) RadString {
	rs := NewRadString(line)
	rs.SetRgb(colorTeal.R, colorTeal.G, colorTeal.B)
	return rs
}

// renderInlineFormatting processes bold, italic, and inline code in a line
func renderInlineFormatting(line string) RadString {
	// We'll process the line by finding all formatting markers and building
	// segments. This is a simplified approach that handles non-overlapping
	// formatting patterns.

	var spans []styledSpan

	// Find all bold spans: **text**
	for _, match := range boldPattern.FindAllStringSubmatchIndex(line, -1) {
		spans = append(spans, styledSpan{
			start: match[0],
			end:   match[1],
			text:  line[match[2]:match[3]], // The captured group (text without **)
			style: "bold",
		})
	}

	// Find all inline code spans: `code`
	for _, match := range inlineCodePattern.FindAllStringSubmatchIndex(line, -1) {
		spans = append(spans, styledSpan{
			start: match[0],
			end:   match[1],
			text:  line[match[2]:match[3]], // The captured group (text without backticks)
			style: "code",
		})
	}

	// Find URLs: https://...
	for _, match := range urlPattern.FindAllStringIndex(line, -1) {
		spans = append(spans, styledSpan{
			start: match[0],
			end:   match[1],
			text:  line[match[0]:match[1]],
			style: "url",
		})
	}

	// Find italic spans: *text* (but not **)
	// We need a more careful approach here to avoid matching inside **
	italicMatches := findItalicSpans(line, spans)
	spans = append(spans, italicMatches...)

	// If no formatting found, return plain string
	if len(spans) == 0 {
		return NewRadString(line)
	}

	// Sort spans by start position
	sortSpans(spans)

	// Remove overlapping spans (keep the first one)
	spans = removeOverlappingSpans(spans)

	// Build the result by iterating through the line
	var result RadString
	pos := 0

	for _, span := range spans {
		// Add any plain text before this span
		if span.start > pos {
			result = result.Concat(NewRadString(line[pos:span.start]))
		}

		// Add the styled span
		styledText := NewRadString(span.text)
		switch span.style {
		case "bold":
			styledText.SetAttr(BOLD)
		case "italic":
			styledText.SetAttr(ITALIC)
		case "code":
			styledText.SetRgb(colorTeal.R, colorTeal.G, colorTeal.B)
		case "url":
			styledText.SetRgb(colorTeal.R, colorTeal.G, colorTeal.B)
			styledText.SetAttr(UNDERLINE)
			styledText.SetSegmentsHyperlink(NewRadString(span.text))
		}
		result = result.Concat(styledText)

		pos = span.end
	}

	// Add any remaining plain text
	if pos < len(line) {
		result = result.Concat(NewRadString(line[pos:]))
	}

	return result
}

// findItalicSpans finds *text* patterns that aren't part of **text**
func findItalicSpans(line string, existingSpans []styledSpan) []styledSpan {
	var result []styledSpan

	// Simple state machine to find single * pairs
	i := 0
	for i < len(line) {
		// Skip if we're at a ** (bold marker)
		if i+1 < len(line) && line[i] == '*' && line[i+1] == '*' {
			i += 2
			continue
		}

		// Found a single *
		if line[i] == '*' {
			start := i
			i++

			// Find the closing *
			textStart := i
			for i < len(line) {
				// Skip **
				if i+1 < len(line) && line[i] == '*' && line[i+1] == '*' {
					i += 2
					continue
				}
				if line[i] == '*' {
					// Found closing *
					text := line[textStart:i]
					if len(text) > 0 {
						// Check this doesn't overlap with existing spans
						span := styledSpan{
							start: start,
							end:   i + 1,
							text:  text,
							style: "italic",
						}
						if !overlapsAny(span, existingSpans) {
							result = append(result, span)
						}
					}
					i++
					break
				}
				i++
			}
			continue
		}
		i++
	}

	return result
}

// overlapsAny checks if a span overlaps with any existing spans
func overlapsAny(span styledSpan, existing []styledSpan) bool {
	for _, e := range existing {
		if span.start < e.end && span.end > e.start {
			return true
		}
	}
	return false
}

// sortSpans sorts spans by start position using simple insertion sort
func sortSpans(spans []styledSpan) {
	for i := 1; i < len(spans); i++ {
		j := i
		for j > 0 && spans[j].start < spans[j-1].start {
			spans[j], spans[j-1] = spans[j-1], spans[j]
			j--
		}
	}
}

// removeOverlappingSpans removes spans that overlap with earlier spans
func removeOverlappingSpans(spans []styledSpan) []styledSpan {
	if len(spans) == 0 {
		return spans
	}

	result := []styledSpan{spans[0]}
	for i := 1; i < len(spans); i++ {
		lastEnd := result[len(result)-1].end
		if spans[i].start >= lastEnd {
			result = append(result, spans[i])
		}
	}
	return result
}
