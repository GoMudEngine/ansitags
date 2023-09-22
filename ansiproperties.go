package ansitags

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	// regex result data indices
	matchPosFull  int = 0
	matchPosTag   int = 1
	matchPosValue int = 2

	// special values to modify 8 bit color codes
	fgToBgIncrement int = 10
	boldIncrement   int = 60

	defaultFg int = 39
	defaultBg int = 49

	posMax int = 16000
)

var (

	// regular expressions
	propertyRegex, _ = regexp.Compile(" (bg|fg|bold|position|clear)=[\"']?([a-z0-9,_-]+)[\"']?")

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

	positionMap map[string][]string = map[string][]string{
		"topleft": []string{"1", "1"},
	}

	// \033[xJ
	// 0 = clear from cursor and beyond
	// 1 = clear from cursor and before
	// 2 = clear screen but it's still in scrollback
	// 3 = just delete everything in the scrollback buffer
	//
	clearMap map[string]int = map[string]int{
		"aftercursor":  0,
		"beforecursor": 1,
		"all":          2,
		"scrollback":   3,
	}
)

type ansiProperties struct {
	fg       int
	bg       int
	bold     bool
	clear    int
	position []uint16
}

func (p *ansiProperties) AnsiReset() string {
	return "\033[39;49m"
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

	var clearCode string = ""
	if p.clear > -1 {
		clearCode = "\033[" + strconv.Itoa(p.clear) + "J"
	}

	var positionCode string = ""
	if len(p.position) == 2 {
		positionCode = "\033[" + strconv.Itoa(int(p.position[1])) + ";" + strconv.Itoa(int(p.position[0])) + "H"
	}

	var colorCode string = ""
	if p.fg > -1 || p.bg > -1 {
		colorCode = "\033["
		if p.fg > -1 {
			colorCode += strconv.Itoa(p.fg)
			if p.bg > -1 {
				colorCode += ";" + strconv.Itoa(p.bg)
			}
			colorCode += "m"
		} else {
			colorCode += strconv.Itoa(p.bg) + "m"
		}
	}

	return clearCode + positionCode + colorCode
}

func AnsiResetAll() string {
	return "\033[0m"
}

func extractProperties(tagStr string) *ansiProperties {
	ret := &ansiProperties{fg: defaultFg, bg: defaultBg, clear: -1}

	result := propertyRegex.FindAllStringSubmatch(tagStr, -1)
	var err error
	for _, match := range result {

		switch match[matchPosTag] {
		case "fg":
			if ret.fg, err = strconv.Atoi(match[matchPosValue]); err != nil {
				// if couldn't find a number, check for a mapped string
				if val, ok := colorMap[match[matchPosValue]]; ok {
					ret.fg = val
				} else {
					ret.fg = defaultFg
				}
			}
		case "bg":
			if ret.bg, err = strconv.Atoi(match[matchPosValue]); err != nil {
				// if couldn't find a number, check for a mapped string
				if val, ok := colorMap[match[matchPosValue]]; ok {
					// increment value to make it a bg value
					ret.bg = val + fgToBgIncrement
				} else {
					ret.bg = defaultBg
				}
			}
		case "bold":
			if ret.bold, err = strconv.ParseBool(match[matchPosValue]); err != nil {
				ret.bold = false
			}
		case "position":

			var posArr []string

			if val, ok := positionMap[match[matchPosValue]]; ok {
				posArr = val
			} else {
				posArr = strings.Split(match[matchPosValue], ",")
			}

			if len(posArr) == 2 {
				yPos := -1
				xPos := -1
				if xPos, err = strconv.Atoi(posArr[0]); err != nil {
					continue
				}
				if yPos, err = strconv.Atoi(posArr[1]); err != nil {
					continue
				}

				if xPos > -1 && yPos > -1 && xPos <= posMax && yPos <= posMax {
					ret.position = []uint16{uint16(xPos), uint16(yPos)}
				}
			}
		case "clear":
			if val, ok := clearMap[match[matchPosValue]]; ok {
				ret.clear = val
			}
		}
		//fmt.Printf("%#v = %#v\n", val[matchPosTag], val[matchPosValue])

	}

	return ret
}
