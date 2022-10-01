package ansigo

import (
	"bufio"
	"bytes"
	"io"
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
