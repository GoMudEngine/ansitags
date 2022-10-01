package ansigo

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type parseMode uint8

const (
	tagStart byte = '<'
	tagEnd   byte = '>'

	tagOpen  string = "ansi"  // will be wrapped in tagStart and tagEnd
	tagClose string = "/ansi" // will be wrapped in tagStart and tagEnd

	// regex result data indices
	matchPosFull  int = 0
	matchPosTag   int = 1
	matchPosValue int = 2

	// special values to modify 8 bit color codes
	fgToBgIncrement int = 10
	boldIncrement   int = 60

	defaultFg int = 39
	defaultBg int = 49

	parseModeNone     parseMode = 0
	parseModeMatching parseMode = 1
)

var (
	// regular expressions
	propertyRegex, _ = regexp.Compile(" (bg|fg|bold)=[\"']?([a-z0-9]+)[\"']?")

	// map of strings to 4 bit color codes
	colorMap map[string]int = map[string]int{
		"black":   30,
		"red":     31,
		"green":   32,
		"yellow":  33,
		"blue":    34,
		"magenta": 35,
		"cyan":    36,
		"white":   37,
	}
)

type tagMatch struct {
	startPos int
	endPos   int
}

type ansiProperties struct {
	fg   int
	bg   int
	bold bool
}

func Parse(str string) string {

	input := bufio.NewReader(strings.NewReader(str))

	var outputBuffer bytes.Buffer
	output := bufio.NewWriter(&outputBuffer)
	ParseStreaming(input, output)

	return outputBuffer.String()
}

func ParseStreaming(inbound *bufio.Reader, outbound *bufio.Writer) {

	var tagStack []*ansiProperties = make([]*ansiProperties, 0, 5)

	var sBuilder strings.Builder
	var currentTagBuilder strings.Builder

	openMatcher := NewTagMatcher(tagStart, []byte(tagOpen), tagEnd, true)
	closeMatcher := NewTagMatcher(tagStart, []byte(tagClose), tagEnd, false)

	var mode parseMode = parseModeNone

	for {
		input, err := inbound.ReadByte()
		if err != nil && err == io.EOF {
			break
		}

		if sBuilder.Len() >= 256 {
			outbound.WriteString(sBuilder.String())
			sBuilder.Reset()
			outbound.Flush()
		}

		// If not currently in any modes, look for any tags
		if mode == parseModeNone {

			if input != tagStart {
				// If it's not an opening tag and we're looking for it (zero position)
				// Write it to the output string and go to next
				sBuilder.WriteByte(input)
				continue
			}

			// Since the input is a starting tag, switch to a matching mode
			mode = parseModeMatching
		}

		// if attempting to match a tag
		if mode == parseModeMatching {

			openMatch, openMatchDone := openMatcher.MatchNext(input)
			closeMatch, closeMatchDone := closeMatcher.MatchNext(input)

			if openMatch {
				currentTagBuilder.WriteByte(input)

				if !openMatchDone {
					continue
				}
				// If this is the final closing string of the open tag

				newTag := extractProperties(currentTagBuilder.String())
				currentTagBuilder.Reset()

				stackLen := len(tagStack)

				if stackLen > 0 {
					sBuilder.WriteString(newTag.PropagateAnsiCode(tagStack[stackLen-1]))
				} else {
					sBuilder.WriteString(newTag.PropagateAnsiCode(nil))
				}

				tagStack = append(tagStack, newTag)

				// reset matchers
				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue

			}
			// No open match was found. Reset the matcher
			openMatcher.Reset()

			if closeMatch {

				currentTagBuilder.WriteByte(input)

				if !closeMatchDone {
					continue
				}

				currentTagBuilder.Reset()
				// we're already at the end, we can parse it.
				stackLen := len(tagStack)

				if stackLen > 2 {
					sBuilder.WriteString(tagStack[stackLen-2].PropagateAnsiCode(tagStack[stackLen-3]))
				} else if stackLen > 1 {
					sBuilder.WriteString(tagStack[stackLen-2].PropagateAnsiCode(nil))
				} else {
					sBuilder.WriteString(AnsiResetAll())
				}

				if stackLen > 0 {
					tagStack[len(tagStack)-1] = nil
					tagStack = tagStack[0 : len(tagStack)-1]
				}

				// reset matchers
				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue

			}
			// No close match was found. Reset the matcher
			closeMatcher.Reset()

			// open and close both failed to match. Reset everything
			mode = parseModeNone
			sBuilder.WriteString(currentTagBuilder.String())
			currentTagBuilder.Reset()
			continue
		}

	}

	if currentTagBuilder.Len() > 0 {
		sBuilder.WriteString(currentTagBuilder.String())
		currentTagBuilder.Reset()
	}

	// if there were any unclosed tags in the stream
	if len(tagStack) > 0 {
		sBuilder.WriteString(AnsiResetAll())
	}

	if sBuilder.Len() > 0 {
		outbound.WriteString(sBuilder.String())
	}

	outbound.Flush()
}

func NewTagMatcher(start byte, mid []byte, end byte, unknownLength bool) *tagMatcher {
	t := tagMatcher{
		startByte:  start,
		midBytes:   mid,
		endByte:    end,
		exactMatch: !unknownLength,
		totalSize:  uint8(len(mid) + 1), // total size without the end byte
		position:   0,
	}
	return &t
}

type tagMatcher struct {
	exactMatch bool // lets unknown bytes keep matching before the end character is found
	totalSize  uint8
	position   uint8
	startByte  byte
	endByte    byte
	midBytes   []byte
}

func (t *tagMatcher) Seek(pos uint8) {
	t.position = pos
}

func (t *tagMatcher) MatchNext(char byte) (matched bool, complete bool) {

	if t.position == 0 {

		// Look for starting byte
		if char == t.startByte {
			t.position++
			// Still matching
			return true, false
		}
		// Failed to match.
		t.position = 0
		return false, true
	}

	if t.position >= t.totalSize {

		// Look for ending byte
		if char == t.endByte {
			t.position++
			// Matched and done.
			return true, true
		}

		if t.exactMatch {
			// Failed to match and required exact
			t.position = 0
			return false, true
		}

		// allows more characters before finishing.
		return true, false
	}

	// Look for mid bytes match
	if char == t.midBytes[t.position-1] {
		t.position++
		// Still matching
		return true, false
	}

	// Failed to match
	t.position = 0
	return false, true
}

func (t *tagMatcher) Reset() {
	t.Seek(0)
}

func (p *ansiProperties) AnsiReset() string {
	return "\033[39;49m"
}

func (p ansiProperties) AnsiCode() string {

	if p.bold {
		if p.fg < 90 && p.fg != defaultFg {
			p.fg += boldIncrement
		}
		if p.bg < 90 && p.fg != defaultBg {
			p.bg += boldIncrement
		}
	}

	return "\033[" + strconv.Itoa(p.fg) + ";" + strconv.Itoa(p.bg) + "m"

}

func (p ansiProperties) PropagateAnsiCode(previous *ansiProperties) string {

	if previous != nil {
		if p.fg == defaultFg {
			p.fg = previous.fg
		}
		if p.bg == defaultBg {
			p.bg = previous.bg
		}
		if !p.bold {
			p.bold = previous.bold
		}
	}

	if p.bold {
		if p.fg < 90 && p.fg != defaultFg {
			p.fg += boldIncrement
		}
		if p.bg < 90 && p.fg != defaultBg {
			p.bg += boldIncrement
		}
	}

	return "\033[" + strconv.Itoa(p.fg) + ";" + strconv.Itoa(p.bg) + "m"

}

func AnsiResetAll() string {
	return "\033[0m"
}

func extractProperties(tagStr string) *ansiProperties {
	ret := &ansiProperties{fg: defaultFg, bg: defaultBg}

	result := propertyRegex.FindAllStringSubmatch(tagStr, -1)
	var err error
	for _, val := range result {

		switch val[matchPosTag] {
		case "fg":
			if ret.fg, err = strconv.Atoi(val[matchPosValue]); err != nil {
				// if couldn't find a number, check for a mapped string
				if val, ok := colorMap[val[matchPosValue]]; ok {
					ret.fg = val
				} else {
					ret.fg = defaultFg //-1
				}
			}
		case "bg":
			if ret.bg, err = strconv.Atoi(val[matchPosValue]); err != nil {
				// if couldn't find a number, check for a mapped string
				if val, ok := colorMap[val[matchPosValue]]; ok {
					// increment value to make it a bg value
					ret.bg = val + fgToBgIncrement
				} else {
					ret.bg = defaultBg //-1
				}
			}
		case "bold":
			if ret.bold, err = strconv.ParseBool(val[matchPosValue]); err != nil {
				ret.bold = false
			}
		}
		//fmt.Printf("%#v = %#v\n", val[matchPosTag], val[matchPosValue])

	}

	return ret
}
