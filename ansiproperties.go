package ansitags

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type ColorMode uint8

const (
	// regex result data indices
	matchPosFull  int = 0
	matchPosTag   int = 1
	matchPosValue int = 2

	defaultFg256 int = -2
	defaultBg256 int = -2

	posMax int = 16000
)

const (
	// 256 bit color mode
	Color8Bit ColorMode = iota
	Color24Bit
)

var (

	// regular expressions
	propertyRegex, _ = regexp.Compile(" (bg|fg|bold|position|clear)=[\"']?([a-z0-9,_-]+)[\"']?")

	// map of strings to 8 bit color codes
	colorAliases map[string]int = map[string]int{
		"black":        0,
		"red":          1,
		"green":        2,
		"yellow":       3,
		"blue":         4,
		"magenta":      5,
		"cyan":         6,
		"white":        7,
		"black-bold":   8,
		"red-bold":     9,
		"green-bold":   10,
		"yellow-bold":  11,
		"blue-bold":    12,
		"magenta-bold": 13,
		"cyan-bold":    14,
		"white-bold":   15,
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

	colorMode ColorMode = Color8Bit

	rwLock = sync.RWMutex{}
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
		if p.fg == defaultFg256 {
			p.fg = previous.fg
		}
		if p.bg == defaultBg256 {
			p.bg = previous.bg
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

	if p.fg == defaultFg256 && p.bg == defaultBg256 {
		colorCode = "\033[0m"
	} else {
		if p.fg > -1 {
			colorCode += "\033[38;5;" + strconv.Itoa(p.fg) + `m`
		} else if p.fg == defaultFg256 {
			colorCode += "\033[39m"
		}

		if p.bg > -1 {
			colorCode += "\033[48;5;" + strconv.Itoa(p.bg) + `m`
		} else if p.bg == defaultBg256 {
			colorCode += "\033[49m"
		}
	}

	return clearCode + positionCode + colorCode
}

func SetColorMode(mode ColorMode) {
	// This is a NOOP now, left for backwards compatibility
}

func AnsiResetAll() string {
	return "\033[0m"
}

func extractProperties(tagStr string) *ansiProperties {

	var ret = &ansiProperties{fg: defaultFg256, bg: defaultBg256, clear: -1}

	result := propertyRegex.FindAllStringSubmatch(tagStr, -1)
	var err error
	var colorVal int
	var ok bool
	for _, match := range result {

		switch match[matchPosTag] {
		case "fg":
			if ret.fg, err = strconv.Atoi(match[matchPosValue]); err != nil {

				if colorVal, ok = colorAliases[match[matchPosValue]]; ok {
					ret.fg = colorVal
				} else {
					ret.fg = defaultFg256
				}
			}
		case "bg":
			if ret.bg, err = strconv.Atoi(match[matchPosValue]); err != nil {

				if colorVal, ok = colorAliases[match[matchPosValue]]; ok {
					ret.bg = colorVal
				} else {
					ret.bg = defaultBg256
				}
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
