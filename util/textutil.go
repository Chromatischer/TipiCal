package util

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

var urlRegex = regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)

var ansiRegex = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]|\x1b\\][^\x07]*\x07")

type URLMatch struct {
	URL   string
	Start int
	End   int
}

func DetectURLs(text string) []URLMatch {
	matches := urlRegex.FindAllStringIndex(text, -1)
	if matches == nil {
		return nil
	}

	result := make([]URLMatch, len(matches))
	for i, m := range matches {
		result[i] = URLMatch{
			URL:   text[m[0]:m[1]],
			Start: m[0],
			End:   m[1],
		}
	}
	return result
}

func FormatWithHyperlinks(text string, maxDisplayLen int) string {
	urls := DetectURLs(text)
	if len(urls) == 0 {
		return text
	}

	var result strings.Builder
	lastEnd := 0

	for _, match := range urls {
		result.WriteString(text[lastEnd:match.Start])

		displayURL := shortenURL(match.URL, maxDisplayLen)
		hyperlink := formatHyperlinkDirect(match.URL, displayURL)
		result.WriteString(hyperlink)
		lastEnd = match.End
	}

	result.WriteString(text[lastEnd:])
	return result.String()
}

func FormatWithHyperlinksStyled(text string, maxDisplayLen int, linkPrefix, linkSuffix string) string {
	urls := DetectURLs(text)
	if len(urls) == 0 {
		return text
	}

	var result strings.Builder
	lastEnd := 0

	for _, match := range urls {
		result.WriteString(text[lastEnd:match.Start])

		displayURL := shortenURL(match.URL, maxDisplayLen)
		hyperlink := formatHyperlinkStyled(match.URL, displayURL, linkPrefix, linkSuffix)
		result.WriteString(hyperlink)
		lastEnd = match.End
	}

	result.WriteString(text[lastEnd:])
	return result.String()
}

func shortenURL(rawURL string, maxLen int) string {
	if maxLen <= 0 || DisplayWidth(rawURL) <= maxLen {
		return rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return TruncateText(rawURL, maxLen)
	}

	domain := parsed.Host
	shortened := domain + "/…"

	if DisplayWidth(shortened) <= maxLen {
		return shortened
	}

	return TruncateText(domain, maxLen)
}

func formatHyperlinkDirect(url, display string) string {
	return "\x1b]8;;" + url + "\x07" + display + "\x1b]8;;\x07"
}

func formatHyperlinkStyled(url, display, prefix, suffix string) string {
	return "\x1b]8;;" + url + "\x07" + prefix + display + suffix + "\x1b]8;;\x07"
}

func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func DisplayWidthANSI(s string) int {
	return runewidth.StringWidth(StripANSI(s))
}

func WrapTextANSI(s string, width int) []string {
	if width <= 0 {
		return []string{""}
	}

	var result []string
	lines := strings.Split(s, "\n")

	for _, line := range lines {
		if DisplayWidthANSI(line) <= width {
			result = append(result, line)
			continue
		}
		wrapped := wrapLineANSI(line, width)
		result = append(result, wrapped...)
	}

	return result
}

func wrapLineANSI(line string, width int) []string {
	segments := parseANSI(line)
	var result []string
	var currentLine strings.Builder
	currentWidth := 0
	activeHyperlink := ""

	for _, seg := range segments {
		if seg.isCode {
			if strings.HasPrefix(seg.text, "\x1b]8;;") {
				if strings.HasSuffix(seg.text, "\x07") && len(seg.text) > 6 {
					content := seg.text[5 : len(seg.text)-1]
					if content == "" {
						activeHyperlink = ""
					} else {
						activeHyperlink = content
					}
				}
			}
			currentLine.WriteString(seg.text)
		} else {
			runes := []rune(seg.text)
			wordStart := 0
			for i := 0; i <= len(runes); i++ {
				atWordEnd := i == len(runes) || runes[i] == ' ' || runes[i] == '\t'
				if atWordEnd && i > wordStart {
					word := string(runes[wordStart:i])
					wordWidth := runewidth.StringWidth(word)
					if currentWidth+wordWidth > width {
						if currentLine.Len() > 0 {
							if activeHyperlink != "" {
								currentLine.WriteString("\x1b]8;;\x07")
							}
							result = append(result, currentLine.String())
							currentLine.Reset()
							currentWidth = 0
							if activeHyperlink != "" {
								currentLine.WriteString("\x1b]8;;" + activeHyperlink + "\x07")
							}
						}
						if wordWidth > width {
							wordRunes := []rune(word)
							chunkStart := 0
							chunkWidth := 0
							for j, r := range wordRunes {
								rw := runewidth.RuneWidth(r)
								if chunkWidth+rw > width {
									currentLine.WriteString(string(wordRunes[chunkStart:j]))
									if activeHyperlink != "" {
										currentLine.WriteString("\x1b]8;;\x07")
									}
									result = append(result, currentLine.String())
									currentLine.Reset()
									currentWidth = 0
									if activeHyperlink != "" {
										currentLine.WriteString("\x1b]8;;" + activeHyperlink + "\x07")
									}
									chunkStart = j
									chunkWidth = 0
								}
								chunkWidth += rw
							}
							if chunkStart < len(wordRunes) {
								remaining := string(wordRunes[chunkStart:])
								currentLine.WriteString(remaining)
								currentWidth = runewidth.StringWidth(remaining)
							} else {
								currentWidth = 0
							}
						} else {
							currentLine.WriteString(word)
							currentWidth += wordWidth
						}
					} else {
						currentLine.WriteString(word)
						currentWidth += wordWidth
					}
				}
				if i < len(runes) && (runes[i] == ' ' || runes[i] == '\t') {
					if currentWidth+1 <= width {
						currentLine.WriteRune(runes[i])
						currentWidth++
					}
					wordStart = i + 1
				}
			}
		}
	}

	if currentLine.Len() > 0 {
		result = append(result, currentLine.String())
	}

	if len(result) == 0 {
		return []string{""}
	}
	return result
}

type ansiSegment struct {
	text   string
	isCode bool
}

func parseANSI(s string) []ansiSegment {
	var segments []ansiSegment
	lastEnd := 0

	matches := ansiRegex.FindAllStringIndex(s, -1)
	for _, m := range matches {
		if m[0] > lastEnd {
			segments = append(segments, ansiSegment{text: s[lastEnd:m[0]], isCode: false})
		}
		segments = append(segments, ansiSegment{text: s[m[0]:m[1]], isCode: true})
		lastEnd = m[1]
	}
	if lastEnd < len(s) {
		segments = append(segments, ansiSegment{text: s[lastEnd:], isCode: false})
	}

	if len(segments) == 0 {
		return []ansiSegment{{text: s, isCode: false}}
	}
	return segments
}
