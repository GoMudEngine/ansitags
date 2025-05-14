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
)

var (
	tagStart byte = '<'
	tagEnd   byte = '>'

	tagOpen  string = "ansi"  // will be wrapped in tagStart and tagEnd
	tagClose string = "/ansi" // will be wrapped in tagStart and tagEnd
)

func Parse(str string, behaviors ...ParseBehavior) string {

	input := bufio.NewReader(strings.NewReader(str))

	var outputBuffer bytes.Buffer
	output := bufio.NewWriter(&outputBuffer)
	ParseStreaming(input, output, behaviors...)

	return outputBuffer.String()
}

func ParseStreaming(inbound *bufio.Reader, outbound *bufio.Writer, behaviors ...ParseBehavior) {

	rwLock.RLock()
	defer rwLock.RUnlock()

	var stripAllTags bool = false
	var stripAllColor bool = false
	var writeHTML bool = false

	for _, b := range behaviors {
		if b == StripTags {
			stripAllTags = true
		} else if b == Monochrome {
			stripAllColor = true
		} else if b == HTML {
			writeHTML = true
		}
	}

	var tagStack []*ansiProperties = make([]*ansiProperties, 0, 5)

	var currentTagBuilder strings.Builder

	openMatcher := NewTagMatcher(tagStart, []byte(tagOpen), tagEnd, true)
	closeMatcher := NewTagMatcher(tagStart, []byte(tagClose), tagEnd, false)

	var mode parseMode = parseModeNone

	for {
		input, err := inbound.ReadByte()

		if err != nil && err == io.EOF {
			break
		}

		// If not currently in any modes, look for any tags
		if mode == parseModeNone {
			if input != tagStart {
				// If it's not an opening tag and we're looking for it (zero position)
				// Write it to the output string and go to next
				outbound.WriteByte(input)
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

				if stripAllColor {
					newTag.fg = defaultFg256
					newTag.bg = defaultBg256
				}

				if writeHTML {
					newTag.htmlOnly = true
				}

				currentTagBuilder.Reset()

				if !stripAllTags {
					stackLen := len(tagStack)
					if stackLen > 0 {
						outbound.WriteString(newTag.PropagateAnsiCode(tagStack[stackLen-1]))
					} else {
						outbound.WriteString(newTag.PropagateAnsiCode(nil))
					}
					tagStack = append(tagStack, newTag)
				}

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
				if !stripAllTags {
					// we're already at the end, we can parse it.
					stackLen := len(tagStack)

					if stackLen > 2 {
						outbound.WriteString(tagStack[stackLen-2].PropagateAnsiCode(tagStack[stackLen-3]))
					} else if stackLen > 1 && !writeHTML {
						outbound.WriteString(tagStack[stackLen-2].PropagateAnsiCode(nil))
					} else {
						if writeHTML {
							outbound.WriteString(htmlResetAll)
						} else {
							outbound.WriteString(ansiResetAll)
						}
					}

					if stackLen > 0 {
						tagStack[len(tagStack)-1] = nil
						tagStack = tagStack[0 : len(tagStack)-1]
					}
				}

				// reset matchers
				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue

			}
			// No close match was found. Reset the matcher
			closeMatcher.Reset()

			if closeMatchDone && openMatchDone {
				currentTagBuilder.WriteByte(input)
			}

			// open and close both failed to match. Reset everything
			mode = parseModeNone

			if !stripAllTags {
				outbound.WriteString(currentTagBuilder.String())
			}
			currentTagBuilder.Reset()
			continue
		}

	}

	if !stripAllTags {
		if currentTagBuilder.Len() > 0 {
			outbound.WriteString(currentTagBuilder.String())
			currentTagBuilder.Reset()
		}

		// if there were any unclosed tags in the stream
		if len(tagStack) > 0 {
			if writeHTML {
				outbound.WriteString(htmlResetAll)
			} else {
				outbound.WriteString(ansiResetAll)
			}
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

	colorAliases[alias] = value

	return nil
}

func SetAliases(aliases map[string]int) error {

	rwLock.Lock()
	defer rwLock.Unlock()

	for alias, value := range aliases {
		if value < 0 || value > 255 {
			return fmt.Errorf(`value "%d" out of allowable range for alias "%s"`, value, alias)
		}

		colorAliases[alias] = value
	}

	return nil
}

func LoadAliases(yamlFilePaths ...string) error {

	rwLock.Lock()
	defer rwLock.Unlock()

	data := make(map[string]map[string]string, 100)

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

					// If a number value supplied, map it
					if numVal, err := strconv.Atoi(real); err == nil {

						colorAliases[alias] = numVal

					} else {

						// If it's a string value, this is allowed if the string is already defined as an alias
						// Save these for a second pass after all numeric aliases have been assigned.
						// example:
						// red: 9
						// bloody: red
						aliasToAlias[alias] = real
					}

				}

				// Second loop to process alias-to-alias maps
				for alias, otherAlias := range aliasToAlias {
					// Only accept if the otherAlias exists
					// If so, map it
					if val, ok := colorAliases[otherAlias]; ok {
						colorAliases[alias] = val
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

	return nil
}
