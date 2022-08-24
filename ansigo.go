package ansigo

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	tagOpenStart  string = "<ansi"
	tagOpenEnd    string = ">"
	tagCloseStart string = "</ansi"
	tagCloseEnd   string = ">"
	// regex result data indices
	matchPosFull  int = 0
	matchPosTag   int = 1
	matchPosValue int = 2
	// special values to modify 8 bit color codes
	fgToBgIncrement int = 10
	boldIncrement   int = 60

	tagTypeNone  uint8 = 0
	tagTypeOpen  uint8 = 1
	tagTypeClose uint8 = 2
	tagTypeBoth  uint8 = 3
)

var (
	// errors
	errTagsNotFound error = errors.New("ansi tag not found")
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

	var sBuilder strings.Builder
	var nestedDepth int = 0
	var tagStack []*ansiProperties = make([]*ansiProperties, 0, 2)
	var nextTag *tagMatch = nil
	var isCloseTag bool = false
	var closeTagFullLength int = len(tagCloseStart + tagCloseEnd)

	for {
		isCloseTag = false
		nextOpenTag, _ := extractParts(str, tagOpenStart, tagOpenEnd)
		nextCloseTag, _ := extractParts(str, tagCloseStart, tagCloseEnd)

		if nextCloseTag != nil && nextCloseTag.endPos-nextCloseTag.startPos != closeTagFullLength {
			nextCloseTag = nil
		}

		if nextOpenTag != nil {
			if nextCloseTag != nil {

				// Make sure the closing tag doesn't start before the end of the open tag
				// If it does, the open tag is invalid and we must use the close tag
				if nextCloseTag.startPos < nextOpenTag.endPos {

					nextOpenTag = nil
					nextTag = nextCloseTag
					isCloseTag = true

				} else {

					// if the next open tag starts before the next closing tag, use it
					if nextOpenTag.startPos < nextCloseTag.startPos {
						nextTag = nextOpenTag
					}
				}

			} else {
				nextTag = nextOpenTag
			}
		} else if nextCloseTag != nil {
			nextTag = nextCloseTag
			isCloseTag = true
		} else {
			// no tags, only normal string remains
			sBuilder.WriteString(str)
			break
		}

		// if there was a straggler string to start, add that
		if nextTag.startPos != 0 {
			sBuilder.WriteString(str[:nextTag.startPos])
		}

		// open tag, extract it and
		// add it to the end of the stack
		tagStack = append(tagStack, extractProperties(str[nextTag.startPos:nextTag.endPos]))

		// Write the ansi value of the most recent tag extraction
		if isCloseTag {

			// un-nest by 1
			if nestedDepth > 0 {
				nestedDepth--
			}

			if nestedDepth > 0 {
				sBuilder.WriteString(tagStack[nestedDepth-1].AnsiCode())
			} else {
				sBuilder.WriteString(tagStack[nestedDepth].AnsiReset())
			}

		} else {

			sBuilder.WriteString(tagStack[len(tagStack)-1].AnsiCode())
			// Now we are nested
			nestedDepth++

		}

		// text remaining (inside) becomes new str
		str = str[nextTag.endPos:]

	}

	// if an ending tag was forgotten, reset all ansi color
	if nestedDepth > 0 {
		sBuilder.WriteString(tagStack[nestedDepth-1].AnsiReset())
	}

	return sBuilder.String()
}

func ParseStreaming(input *os.File, output *os.File) error {

	var stackPosition int = 0
	var tagStack []*ansiProperties = make([]*ansiProperties, 0, 10)

	inbound := bufio.NewReader(input)
	outbound := bufio.NewWriter(output)
	var sBuilder strings.Builder
	var currentTagBuilder strings.Builder

	/*
		tagOpenStart  string = "<ansi"
		tagOpenEnd    string = ">"
		tagCloseStart string = "</ansi"
		tagCloseEnd   string = ">"
	*/

	var tagPosition int = 0
	var tagType uint8 = tagTypeNone

	var tagOpenParts []rune = []rune(tagOpenStart)
	var tagCloseParts []rune = []rune(tagCloseStart + tagCloseEnd)

	for {
		input, _, err := inbound.ReadRune()
		if err != nil && err == io.EOF {
			break
		}

		// if NOT in a sequence:
		// 1.) Check whether the input is the next char in a sequence

		// if IN a sequence:
		// 1.) If current tag type is both, try and narrow down which one we are looking at
		// 2.) Check wehther next char is in the sequence for the current tag type
		// 3.) if not, reset everything.

		// If not in the middle of a tag, or the tag hasn't been narrowed yet
		if tagType == tagTypeNone || tagType == tagTypeBoth {

			if input == tagOpenParts[tagPosition] && input == tagCloseParts[tagPosition] {
				tagType = tagTypeBoth
				currentTagBuilder.WriteRune(input)
				tagPosition++
				continue
			}

			if input == tagOpenParts[tagPosition] {
				tagType = tagTypeOpen
				currentTagBuilder.WriteRune(input)
				tagPosition++
				continue
			}

			if input == tagCloseParts[tagPosition] {
				tagType = tagTypeClose
				currentTagBuilder.WriteRune(input)
				tagPosition++
				continue
			}

			if tagType != tagTypeNone {
				// We've failed to find a match, reset everything, write what we've got to the output
				tagType = tagTypeNone
				sBuilder.WriteString(currentTagBuilder.String())
				currentTagBuilder.Reset()
				tagPosition = 0
			}

		} else {

			if tagType == tagTypeOpen {

				// if still trying to reach the end...
				if tagPosition < len(tagOpenParts) {

					// if still tracking...
					if input == tagOpenParts[tagPosition] {
						currentTagBuilder.WriteRune(input)
						tagPosition++
						continue
					} else {

						// fell off. Reset it all.
						tagType = tagTypeNone
						sBuilder.WriteString(currentTagBuilder.String())
						currentTagBuilder.Reset()
						tagPosition = 0
						continue
					}

				} else {

					// If this is the final closing string of the open tag
					if string(input) == tagOpenEnd {

						currentTagBuilder.WriteRune(input)
						tagStack = append(tagStack, extractProperties(currentTagBuilder.String()))
						stackPosition++

						fmt.Println(currentTagBuilder.String())
						for _, v := range tagStack {
							fmt.Println(*v)
						}

						currentTagBuilder.Reset()

						sBuilder.WriteString(tagStack[stackPosition-1].AnsiCode())

						tagType = tagTypeNone
						tagPosition = 0

						continue

					} else {
						currentTagBuilder.WriteRune(input)
						continue
					}

				}

			}

			if tagType == tagTypeClose {

				// if still trying to reach the end...
				if tagPosition < len(tagCloseParts)-1 {

					// if still tracking...
					if input == tagCloseParts[tagPosition] {
						currentTagBuilder.WriteRune(input)
						tagPosition++
						continue
					} else {
						// fell off. Reset it all.
						tagType = tagTypeNone
						sBuilder.WriteString(currentTagBuilder.String())
						currentTagBuilder.Reset()
						tagPosition = 0
						continue
					}

				} else {
					currentTagBuilder.Reset()

					// we're already at the end, we can parse it.
					if stackPosition > 1 {
						sBuilder.WriteString(tagStack[stackPosition-2].AnsiCode())
					} else {
						sBuilder.WriteString(tagStack[0].AnsiReset())
					}

					stackPosition--
					tagType = tagTypeNone
					tagPosition = 0

					continue
				}

			}

		}

		if tagType == tagTypeNone {
			sBuilder.WriteRune(input)

			if sBuilder.Len() >= 512 {

				outbound.WriteString(sBuilder.String())
				sBuilder.Reset()
				outbound.Flush()
			}
		}
	}

	if sBuilder.Len() > 0 {
		bytes, _ := json.Marshal(sBuilder.String())
		fmt.Println(string(bytes))
		outbound.WriteString(sBuilder.String())
	}

	outbound.Flush()
	return nil
}

func extractParts(str string, strStart string, strEnd string) (matchData *tagMatch, err error) {

	lPosStart := strings.Index(str, strStart)
	strStartLen := len(strStart)

	if lPosStart == -1 {
		return nil, errTagsNotFound
	}

	lPosEnd := strings.Index(str[lPosStart+strStartLen:], strEnd)

	if lPosEnd == -1 {
		return nil, errTagsNotFound
	}

	return &tagMatch{
			startPos: lPosStart,
			endPos:   lPosEnd + lPosStart + strStartLen + 1,
		},
		nil
}

func (p *ansiProperties) AnsiReset() string {
	return "\033[0m"
}

func (p *ansiProperties) AnsiCode() string {
	if p.fg == -1 && p.bg == -1 {
		return ""
	}

	fgBoldMod := 0
	bgBoldMod := 0

	if p.bold {
		if p.fg < 90 {
			fgBoldMod = boldIncrement
		}
		if p.bg < 90 {
			bgBoldMod = boldIncrement
		}
	}

	if p.fg != -1 {
		if p.bg != -1 {
			// fg and bg
			return "\033[" + strconv.Itoa(p.fg+fgBoldMod) + ";" + strconv.Itoa(p.bg+bgBoldMod) + "m"
		}
		// only fg
		return "\033[" + strconv.Itoa(p.fg+fgBoldMod) + "m"
	}

	// only bg
	return "\033[" + strconv.Itoa(p.bg+bgBoldMod) + "m"
}

func extractProperties(tagStr string) *ansiProperties {
	ret := &ansiProperties{fg: -1, bg: -1}

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
					ret.fg = -1
				}
			}
		case "bg":
			if ret.bg, err = strconv.Atoi(val[matchPosValue]); err != nil {
				// if couldn't find a number, check for a mapped string
				if val, ok := colorMap[val[matchPosValue]]; ok {
					// increment value to make it a bg value
					ret.bg = val + fgToBgIncrement
				} else {
					ret.bg = -1
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
