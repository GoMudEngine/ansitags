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
	propertyRegex, _ = regexp.Compile(" (bg|fg|bold|position)=[\"']?([a-z0-9]+)[\"']?")

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

type ansiProperties struct {
	fg       int
	bg       int
	bold     bool
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

	if p.fg == 0 && p.bg == 0 {
		return ""
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
		case "position":

			posArr := strings.Split(val[matchPosValue], ",")
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
		}
		//fmt.Printf("%#v = %#v\n", val[matchPosTag], val[matchPosValue])

	}

	return ret
}
