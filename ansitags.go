package ansitags

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type parseMode uint8
type ParseBehavior uint8

const (
	parseModeNone     parseMode = 0
	parseModeMatching parseMode = 1

	StripTags  ParseBehavior = iota // remove all valid ansitags
	Monochrome                      // ignore any color changing properties
	HTML                            // produce HTML instead of ansi tags

	// maxTagSize is the maximum byte length of a tag we will accumulate.
	// Tags longer than this cannot be valid, so we flush and reset.
	maxTagSize = 256
)

var (
	tagStart byte = '<'
	tagEnd   byte = '>'

	tagOpen  string = "ansi"  // will be wrapped in tagStart and tagEnd
	tagClose string = "/ansi" // will be wrapped in tagStart and tagEnd
)

func Parse(str string, behaviors ...ParseBehavior) string {

	var outputBuffer bytes.Buffer
	outputBuffer.Grow(len(str))
	parseString(str, &outputBuffer, behaviors...)
	return outputBuffer.String()
}

// parseString is the core implementation for string input, avoiding bufio overhead.
func parseString(str string, out *bytes.Buffer, behaviors ...ParseBehavior) {

	rwLock.RLock()
	defer rwLock.RUnlock()

	var stripAllTags bool
	var stripAllColor bool
	var writeHTML bool

	for _, b := range behaviors {
		switch b {
		case StripTags:
			stripAllTags = true
		case Monochrome:
			stripAllColor = true
		case HTML:
			writeHTML = true
		}
	}

	var tagStack []*ansiProperties = make([]*ansiProperties, 0, 5)

	// Fixed-size tag accumulation buffer — avoids heap allocation for the common case.
	var tagBuf [maxTagSize]byte
	var tagLen int

	openMatcher := NewTagMatcher(tagStart, []byte(tagOpen), tagEnd, true)
	closeMatcher := NewTagMatcher(tagStart, []byte(tagClose), tagEnd, false)

	var mode parseMode = parseModeNone

	for i := 0; i < len(str); i++ {
		input := str[i]

		if mode == parseModeNone {
			if input != tagStart {
				out.WriteByte(input)
				continue
			}
			mode = parseModeMatching
		}

		if mode == parseModeMatching {

			openMatch, openMatchDone := openMatcher.MatchNext(input)
			closeMatch, closeMatchDone := closeMatcher.MatchNext(input)

			if openMatch {

				if tagLen < maxTagSize {
					tagBuf[tagLen] = input
					tagLen++
				}

				if !openMatchDone {
					continue
				}

				newTag := extractProperties(string(tagBuf[:tagLen]))

				if stripAllColor {
					newTag.fg = defaultFg256
					newTag.bg = defaultBg256
				}

				if writeHTML {
					newTag.htmlOnly = true
				}

				tagLen = 0

				if !stripAllTags {
					stackLen := len(tagStack)
					if stackLen > 0 {
						out.WriteString(newTag.PropagateAnsiCode(tagStack[stackLen-1]))
					} else {
						out.WriteString(newTag.PropagateAnsiCode(nil))
					}
					tagStack = append(tagStack, newTag)
				} else {
					releaseProperties(newTag)
				}

				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue
			}
			openMatcher.Reset()

			if closeMatch {

				if tagLen < maxTagSize {
					tagBuf[tagLen] = input
					tagLen++
				}

				if !closeMatchDone {
					continue
				}

				tagLen = 0
				if !stripAllTags {
					stackLen := len(tagStack)

					if stackLen > 2 {
						out.WriteString(tagStack[stackLen-2].PropagateAnsiCode(tagStack[stackLen-3]))
					} else if stackLen > 1 {
						out.WriteString(tagStack[stackLen-2].PropagateAnsiCode(nil))
					} else {
						if writeHTML {
							out.WriteString(htmlResetAll)
						} else {
							out.WriteString(ansiResetAll)
						}
					}

					if stackLen > 0 {
						releaseProperties(tagStack[stackLen-1])
						tagStack[stackLen-1] = nil
						tagStack = tagStack[:stackLen-1]
					}
				}

				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue
			}
			closeMatcher.Reset()

			if closeMatchDone && openMatchDone {
				if tagLen < maxTagSize {
					tagBuf[tagLen] = input
					tagLen++
				}
			}

			mode = parseModeNone

			if !stripAllTags {
				out.Write(tagBuf[:tagLen])
			}
			tagLen = 0
			continue
		}
	}

	if !stripAllTags {
		if tagLen > 0 {
			out.Write(tagBuf[:tagLen])
			tagLen = 0
		}

		if len(tagStack) > 0 {
			if writeHTML {
				out.WriteString(htmlResetAll)
			} else {
				out.WriteString(ansiResetAll)
			}
		}
	}

	// Release any remaining pooled properties
	for _, p := range tagStack {
		if p != nil {
			releaseProperties(p)
		}
	}
}

func ParseStreaming(inbound *bufio.Reader, outbound *bufio.Writer, behaviors ...ParseBehavior) {

	rwLock.RLock()
	defer rwLock.RUnlock()

	var stripAllTags bool
	var stripAllColor bool
	var writeHTML bool

	for _, b := range behaviors {
		switch b {
		case StripTags:
			stripAllTags = true
		case Monochrome:
			stripAllColor = true
		case HTML:
			writeHTML = true
		}
	}

	var tagStack []*ansiProperties = make([]*ansiProperties, 0, 5)

	var tagBuf [maxTagSize]byte
	var tagLen int

	openMatcher := NewTagMatcher(tagStart, []byte(tagOpen), tagEnd, true)
	closeMatcher := NewTagMatcher(tagStart, []byte(tagClose), tagEnd, false)

	var mode parseMode = parseModeNone

	for {
		input, err := inbound.ReadByte()

		if err != nil && err == io.EOF {
			break
		}

		if mode == parseModeNone {
			if input != tagStart {
				outbound.WriteByte(input)
				continue
			}
			mode = parseModeMatching
		}

		if mode == parseModeMatching {

			openMatch, openMatchDone := openMatcher.MatchNext(input)
			closeMatch, closeMatchDone := closeMatcher.MatchNext(input)

			if openMatch {

				if tagLen < maxTagSize {
					tagBuf[tagLen] = input
					tagLen++
				}

				if !openMatchDone {
					continue
				}

				newTag := extractProperties(string(tagBuf[:tagLen]))

				if stripAllColor {
					newTag.fg = defaultFg256
					newTag.bg = defaultBg256
				}

				if writeHTML {
					newTag.htmlOnly = true
				}

				tagLen = 0

				if !stripAllTags {
					stackLen := len(tagStack)
					if stackLen > 0 {
						outbound.WriteString(newTag.PropagateAnsiCode(tagStack[stackLen-1]))
					} else {
						outbound.WriteString(newTag.PropagateAnsiCode(nil))
					}
					tagStack = append(tagStack, newTag)
				} else {
					releaseProperties(newTag)
				}

				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue
			}
			openMatcher.Reset()

			if closeMatch {

				if tagLen < maxTagSize {
					tagBuf[tagLen] = input
					tagLen++
				}

				if !closeMatchDone {
					continue
				}

				tagLen = 0
				if !stripAllTags {
					stackLen := len(tagStack)

					if stackLen > 2 {
						outbound.WriteString(tagStack[stackLen-2].PropagateAnsiCode(tagStack[stackLen-3]))
					} else if stackLen > 1 {
						outbound.WriteString(tagStack[stackLen-2].PropagateAnsiCode(nil))
					} else {
						if writeHTML {
							outbound.WriteString(htmlResetAll)
						} else {
							outbound.WriteString(ansiResetAll)
						}
					}

					if stackLen > 0 {
						releaseProperties(tagStack[stackLen-1])
						tagStack[stackLen-1] = nil
						tagStack = tagStack[:stackLen-1]
					}
				}

				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue
			}
			closeMatcher.Reset()

			if closeMatchDone && openMatchDone {
				if tagLen < maxTagSize {
					tagBuf[tagLen] = input
					tagLen++
				}
			}

			mode = parseModeNone

			if !stripAllTags {
				outbound.Write(tagBuf[:tagLen])
			}
			tagLen = 0
			continue
		}
	}

	if !stripAllTags {
		if tagLen > 0 {
			outbound.Write(tagBuf[:tagLen])
			tagLen = 0
		}

		if len(tagStack) > 0 {
			if writeHTML {
				outbound.WriteString(htmlResetAll)
			} else {
				outbound.WriteString(ansiResetAll)
			}
		}
	}

	for _, p := range tagStack {
		if p != nil {
			releaseProperties(p)
		}
	}

	outbound.Flush()
}

func SetAlias(alias string, value int) error {

	rwLock.Lock()
	defer rwLock.Unlock()

	if value < 0 || value > 255 {
		return fmt.Errorf(`value "%d" out of allowable range for alias "%s"`, value, alias)
	}

	newMap := make(map[string]int, len(colorAliases)+1)
	for k, v := range colorAliases {
		newMap[k] = v
	}
	newMap[alias] = value
	colorAliases = newMap
	storeAliasSnapshot(colorAliases)

	return nil
}

func SetAliases(aliases map[string]int) error {

	rwLock.Lock()
	defer rwLock.Unlock()

	for alias, value := range aliases {
		if value < 0 || value > 255 {
			return fmt.Errorf(`value "%d" out of allowable range for alias "%s"`, value, alias)
		}
	}

	newMap := make(map[string]int, len(colorAliases)+len(aliases))
	for k, v := range colorAliases {
		newMap[k] = v
	}
	for alias, value := range aliases {
		newMap[alias] = value
	}
	colorAliases = newMap
	storeAliasSnapshot(colorAliases)

	return nil
}

func LoadAliases(yamlFilePaths ...string) error {

	rwLock.Lock()
	defer rwLock.Unlock()

	data := make(map[string]map[string]string, 100)

	newMap := make(map[string]int, len(colorAliases))
	for k, v := range colorAliases {
		newMap[k] = v
	}

	for _, yamlFilePath := range yamlFilePaths {

		if yfile, err := os.ReadFile(yamlFilePath); err != nil {
			return err
		} else {
			if err := yaml.Unmarshal(yfile, &data); err != nil {
				return err
			}
		}

		for aliasGroup, aliases := range data {

			if aliasGroup == "colors" || aliasGroup == "color256" {

				aliasToAlias := map[string]string{}

				for alias, real := range aliases {

					if numVal, err := strconv.Atoi(real); err == nil {

						newMap[alias] = numVal

					} else {

						aliasToAlias[alias] = real
					}

				}

				for alias, otherAlias := range aliasToAlias {
					if val, ok := newMap[otherAlias]; ok {
						newMap[alias] = val
					}
				}
			}

			if aliasGroup == "position" {
				for alias, real := range aliases {
					posArr := strings.Split(real, ",")
					if len(posArr) == 2 {
						positionMap[alias] = posArr
					}
				}
			}

		}
	}

	colorAliases = newMap
	storeAliasSnapshot(colorAliases)

	return nil
}
