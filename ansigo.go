package ansigo

import (
	"bufio"
	"bytes"
	"errors"
	"io"
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

	defaultFg int = 39
	defaultBg int = 49
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

	input := bufio.NewReader(strings.NewReader(str))

	var outputBuffer bytes.Buffer
	output := bufio.NewWriter(&outputBuffer)
	if err := ParseStreaming(input, output); err != nil {
		panic(err)
	}

	return outputBuffer.String()
}

func ParseStreaming(inbound *bufio.Reader, outbound *bufio.Writer) error {

	var tagStack []*ansiProperties = make([]*ansiProperties, 0, 5)

	var sBuilder strings.Builder
	var currentTagBuilder strings.Builder
	var tagPosition int = 0
	var tagType uint8 = tagTypeNone

	var tagOpenParts []byte = []byte(tagOpenStart)
	var tagCloseParts []byte = []byte(tagCloseStart + tagCloseEnd)
	var tagStartChar byte = tagOpenParts[0]

	for {
		input, err := inbound.ReadByte()
		if err != nil && err == io.EOF {
			break
		}

		if sBuilder.Len() >= 128 {
			outbound.WriteString(sBuilder.String())
			sBuilder.Reset()
			outbound.Flush()
		}

		if tagType == tagTypeNone && tagPosition == 0 && input != tagStartChar {
			sBuilder.WriteByte(input)
			continue
		}

		// if NOT in a sequence:
		// 1.) Check whether the input is the next char in a sequence

		// if IN a sequence:
		// 1.) If current tag type is both, try and narrow down which one we are looking at
		// 2.) Check wehther next char is in the sequence for the current tag type
		// 3.) if not, reset everything.
		//fmt.Print(string(input))
		// If not in the middle of a tag, or the tag hasn't been narrowed yet
		if tagType == tagTypeNone || tagType == tagTypeBoth {

			if input == tagOpenParts[tagPosition] {

				if input == tagCloseParts[tagPosition] {
					tagType = tagTypeBoth
				} else {
					tagType = tagTypeOpen
				}

				currentTagBuilder.WriteByte(input)
				tagPosition++
				continue
			}

			if input == tagCloseParts[tagPosition] {
				tagType = tagTypeClose
				currentTagBuilder.WriteByte(input)
				tagPosition++
				continue
			}

			if tagType != tagTypeNone {
				// We've failed to find a match, reset everything, write what we've got to the output
				tagType = tagTypeNone
				currentTagBuilder.WriteByte(input)
				sBuilder.WriteString(currentTagBuilder.String())
				currentTagBuilder.Reset()
				tagPosition = 0
				continue
			}

			sBuilder.WriteByte(input)

		} else {

			if tagType == tagTypeOpen {

				// if still trying to reach the end...
				if tagPosition < len(tagOpenParts) {

					// if still tracking...
					if input == tagOpenParts[tagPosition] {
						currentTagBuilder.WriteByte(input)
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

						currentTagBuilder.WriteByte(input)

						newTag := extractProperties(currentTagBuilder.String())

						stackLen := len(tagStack)

						currentTagBuilder.Reset()

						if stackLen > 0 {
							sBuilder.WriteString(newTag.PropagateAnsiCode(tagStack[stackLen-1]))
						} else {
							sBuilder.WriteString(newTag.PropagateAnsiCode(nil))
						}

						tagStack = append(tagStack, newTag)

						tagType = tagTypeNone
						tagPosition = 0

						continue

					} else {
						currentTagBuilder.WriteByte(input)
						continue
					}

				}

			}

			if tagType == tagTypeClose {

				// if still trying to reach the end...
				if tagPosition < len(tagCloseParts)-1 {

					// if still tracking...
					if input == tagCloseParts[tagPosition] {
						currentTagBuilder.WriteByte(input)
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

					tagType = tagTypeNone
					tagPosition = 0

					continue
				}

			}

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
	return nil
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
