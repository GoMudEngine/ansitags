package ansitags

import (
	"strconv"
	"sync"
	"sync/atomic"
	"unsafe"
)

type ColorMode uint8

const (
	defaultFg256 int = -2
	defaultBg256 int = -2

	posMax int = 16000

	ansiResetAll = "\033[0m"
	htmlResetAll = "</span>"
)

const (
	// 256 bit color mode
	Color8Bit ColorMode = iota
	Color24Bit
)

var (
	// map of strings to 8 bit color codes — accessed via atomic pointer for lock-free reads
	defaultColorAliases = map[string]int{
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

	// atomicAliases holds a *aliasSnapshot; readers load it without any lock.
	atomicAliases unsafe.Pointer

	positionMap map[string][]string = map[string][]string{
		"topleft": {"1", "1"},
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

	ansiFgSeq [256]string
	ansiBgSeq [256]string

	// Pre-computed HTML color style fragments, e.g. "color:#ff0000;"
	htmlFgStyle [256]string
	htmlBgStyle [256]string

	rwLock = sync.RWMutex{}
)

// aliasSnapshot is an immutable copy of colorAliases used for lock-free reads.
type aliasSnapshot struct {
	m map[string]int
}

// loadAliasSnapshot returns the current alias map without acquiring any lock.
func loadAliasSnapshot() map[string]int {
	p := atomic.LoadPointer(&atomicAliases)
	return (*aliasSnapshot)(p).m
}

// storeAliasSnapshot replaces the alias map atomically. Must be called under rwLock.
func storeAliasSnapshot(m map[string]int) {
	snap := &aliasSnapshot{m: m}
	atomic.StorePointer(&atomicAliases, unsafe.Pointer(snap))
}

// colorAliases is the mutable backing map, written only under rwLock.Write.
// Readers always go through loadAliasSnapshot().
var colorAliases map[string]int

type ansiProperties struct {
	fg       int
	bg       int
	clear    int
	position []uint16
	htmlOnly bool
}

// propertiesPool recycles ansiProperties to reduce heap allocations.
var propertiesPool = sync.Pool{
	New: func() any {
		return &ansiProperties{}
	},
}

func acquireProperties() *ansiProperties {
	p := propertiesPool.Get().(*ansiProperties)
	p.fg = defaultFg256
	p.bg = defaultBg256
	p.clear = -1
	p.position = p.position[:0]
	p.htmlOnly = false
	return p
}

func releaseProperties(p *ansiProperties) {
	propertiesPool.Put(p)
}

func (p *ansiProperties) AnsiReset() string {
	return ansiResetAll
}

func (p ansiProperties) PropagateAnsiCode(previous *ansiProperties) string {

	origFg := p.fg
	origBg := p.bg

	if previous != nil {

		if p.fg == defaultFg256 {
			p.fg = previous.fg
		}
		if p.bg == defaultBg256 {
			p.bg = previous.bg
		}
	}

	if p.htmlOnly {

		if previous != nil {

			if p.fg == previous.fg && p.bg == previous.bg {
				return `<span>`
			}
		}

		if p.fg == defaultFg256 && p.bg == defaultBg256 {
			return `<span>`
		}

		if p.fg > -1 && p.bg > -1 {
			return `<span style="` + htmlFgStyle[p.fg] + htmlBgStyle[p.bg] + `">`
		}
		if p.fg > -1 {
			return `<span style="` + htmlFgStyle[p.fg] + `">`
		}
		if p.bg > -1 {
			return `<span style="` + htmlBgStyle[p.bg] + `">`
		}
		return `<span>`
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
			colorCode += ansiFgSeq[p.fg]
		} else if origFg == defaultFg256 {
			colorCode += "\033[39m"
		}

		if p.bg > -1 {
			colorCode += ansiBgSeq[p.bg]
		} else if origBg == defaultBg256 {
			colorCode += "\033[49m"
		}
	}

	return clearCode + positionCode + colorCode
}

func SetColorMode(mode ColorMode) {
	// This is a NOOP now, left for backwards compatibility
}

// extractProperties parses an open tag string like `<ansi fg=red bg="0" >` and
// returns a populated ansiProperties. The caller is responsible for releasing
// the returned pointer via releaseProperties when it is no longer needed.
func extractProperties(tagStr string) *ansiProperties {

	ret := acquireProperties()

	aliases := loadAliasSnapshot()

	i := 0
	n := len(tagStr)

	for i < n {
		// Skip until we find a space (attribute separator)
		if tagStr[i] != ' ' {
			i++
			continue
		}
		i++ // consume the space

		// Read the key (runs until '=')
		keyStart := i
		for i < n && tagStr[i] != '=' {
			i++
		}
		if i >= n {
			break
		}
		key := tagStr[keyStart:i]
		i++ // consume '='

		// Read the value, optionally quoted with ' or "
		if i >= n {
			break
		}
		var quote byte
		if tagStr[i] == '\'' || tagStr[i] == '"' {
			quote = tagStr[i]
			i++
		}
		valStart := i
		if quote != 0 {
			for i < n && tagStr[i] != quote {
				i++
			}
		} else {
			for i < n && tagStr[i] != ' ' && tagStr[i] != '>' {
				i++
			}
		}
		val := tagStr[valStart:i]
		if quote != 0 && i < n {
			i++ // consume closing quote
		}

		if len(val) == 0 {
			continue
		}

		switch key {
		case "fg":
			if num, err := strconv.Atoi(val); err == nil {
				ret.fg = num
			} else if colorVal, ok := aliases[val]; ok {
				ret.fg = colorVal
			} else {
				ret.fg = defaultFg256
			}
		case "bg":
			if num, err := strconv.Atoi(val); err == nil {
				ret.bg = num
			} else if colorVal, ok := aliases[val]; ok {
				ret.bg = colorVal
			} else {
				ret.bg = defaultBg256
			}
		case "position":
			var posArr []string
			if mapped, ok := positionMap[val]; ok {
				posArr = mapped
			} else {
				// split on comma inline to avoid allocating a slice via strings.Split
				comma := -1
				for ci := 0; ci < len(val); ci++ {
					if val[ci] == ',' {
						comma = ci
						break
					}
				}
				if comma > 0 {
					posArr = []string{val[:comma], val[comma+1:]}
				}
			}
			if len(posArr) == 2 {
				xPos, xErr := strconv.Atoi(posArr[0])
				yPos, yErr := strconv.Atoi(posArr[1])
				if xErr == nil && yErr == nil && xPos > -1 && yPos > -1 && xPos <= posMax && yPos <= posMax {
					ret.position = append(ret.position[:0], uint16(xPos), uint16(yPos))
				}
			}
		case "clear":
			if val, ok := clearMap[val]; ok {
				ret.clear = val
			}
		}
	}

	return ret
}

// Speed up by pre-computing these values
func init() {
	// Seed colorAliases and the atomic snapshot
	colorAliases = make(map[string]int, len(defaultColorAliases))
	for k, v := range defaultColorAliases {
		colorAliases[k] = v
	}
	storeAliasSnapshot(colorAliases)

	for i := 0; i < 256; i++ {
		ansiFgSeq[i] = "\033[38;5;" + strconv.Itoa(i) + "m"
		ansiBgSeq[i] = "\033[48;5;" + strconv.Itoa(i) + "m"
	}
}
