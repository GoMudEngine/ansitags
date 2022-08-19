package ansigo

import (
	"errors"
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

	//var tabs string = ""
	var sBuilder strings.Builder
	var nestedDepth int = 0
	var tagStack []*ansiProperties = make([]*ansiProperties, 0, 2)
	var nextTag *tagMatch = nil
	var isCloseTag bool = false

	for {
		isCloseTag = false
		nextOpenTag, _ := extractParts(str, tagOpenStart, tagOpenEnd)
		nextCloseTag, _ := extractParts(str, tagCloseStart, tagCloseEnd)

		if nextOpenTag != nil {
			if nextCloseTag != nil {
				if nextOpenTag.startPos < nextCloseTag.startPos {
					nextTag = nextOpenTag
				} else {
					nextTag = nextCloseTag
					isCloseTag = true
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
			nestedDepth--

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
		return "\033[0m"
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
