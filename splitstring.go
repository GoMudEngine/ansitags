package ansitags

import "strings"

func SplitString(input string, maxLen int, trimSpace ...bool) []string {
	doTrim := true
	if len(trimSpace) > 0 {
		doTrim = trimSpace[0]
	}

	if maxLen <= 0 || len(input) == 0 {
		return []string{input}
	}

	totalVisible := visibleLen(input)
	if totalVisible <= maxLen {
		return []string{input}
	}

	var result []string
	var tagStack []string
	var current strings.Builder
	visibleCount := 0
	totalConsumed := 0

	openMatcher := NewTagMatcher(tagStart, []byte(tagOpen), tagEnd, true)
	closeMatcher := NewTagMatcher(tagStart, []byte(tagClose), tagEnd, false)

	var tagBuf [maxTagSize]byte
	var tagLen int
	var mode parseMode = parseModeNone

	split := func() {
		for j := len(tagStack) - 1; j >= 0; j-- {
			current.WriteString("</ansi>")
		}
		result = append(result, current.String())
		current.Reset()
		for _, tag := range tagStack {
			current.WriteString(tag)
		}
		visibleCount = 0
	}

	writeVisible := func(b byte) {
		current.WriteByte(b)
		visibleCount++
		totalConsumed++
		if visibleCount >= maxLen && totalConsumed < totalVisible {
			split()
		}
	}

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if mode == parseModeNone {
			if ch != tagStart {
				writeVisible(ch)
				continue
			}
			mode = parseModeMatching
		}

		if mode == parseModeMatching {
			openMatch, openMatchDone := openMatcher.MatchNext(ch)
			closeMatch, closeMatchDone := closeMatcher.MatchNext(ch)

			if openMatch {
				if tagLen < maxTagSize {
					tagBuf[tagLen] = ch
					tagLen++
				}
				if !openMatchDone {
					continue
				}
				tagStr := string(tagBuf[:tagLen])
				tagStack = append(tagStack, tagStr)
				current.WriteString(tagStr)
				tagLen = 0
				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue
			}
			openMatcher.Reset()

			if closeMatch {
				if tagLen < maxTagSize {
					tagBuf[tagLen] = ch
					tagLen++
				}
				if !closeMatchDone {
					continue
				}
				tagLen = 0
				if len(tagStack) > 0 {
					tagStack = tagStack[:len(tagStack)-1]
				}
				current.WriteString("</ansi>")
				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue
			}
			closeMatcher.Reset()

			if closeMatchDone && openMatchDone {
				if tagLen < maxTagSize {
					tagBuf[tagLen] = ch
					tagLen++
				}
			}

			mode = parseModeNone

			for j := 0; j < tagLen; j++ {
				writeVisible(tagBuf[j])
			}
			tagLen = 0
			continue
		}
	}

	if tagLen > 0 {
		for j := 0; j < tagLen; j++ {
			writeVisible(tagBuf[j])
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	if len(result) == 0 {
		return []string{input}
	}

	if doTrim {
		for i, seg := range result {
			result[i] = trimTagAwareSpaces(seg)
		}
	}

	return result
}

// splitPoints computes the 1-based visible-character counts at which to split,
// preferring a space at or before each maxLen boundary. The split occurs after
// the character at that count (inclusive). If no space is found, falls back to
// a character-based split at the maxLen boundary.
func splitPoints(input string, maxLen int) []int {
	openMatcher := NewTagMatcher(tagStart, []byte(tagOpen), tagEnd, true)
	closeMatcher := NewTagMatcher(tagStart, []byte(tagClose), tagEnd, false)
	var tagBuf [maxTagSize]byte
	var tagLen int
	var mode parseMode = parseModeNone

	// consumed is 1-based: the count of visible chars seen so far.
	consumed := 0
	totalVisible := visibleLen(input)

	var points []int
	nextTarget := maxLen
	// lastSpaceAt is the 1-based consumed value just after the last seen space.
	lastSpaceAt := -1

	recordVisible := func(ch byte) {
		consumed++
		if ch == ' ' {
			lastSpaceAt = consumed
		}
		for consumed >= nextTarget && consumed < totalVisible {
			var splitAt int
			if lastSpaceAt > 0 {
				splitAt = lastSpaceAt
				nextTarget = lastSpaceAt + maxLen
			} else {
				splitAt = nextTarget
				nextTarget = nextTarget + maxLen
			}
			lastSpaceAt = -1
			points = append(points, splitAt)
		}
	}

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if mode == parseModeNone {
			if ch != tagStart {
				recordVisible(ch)
				continue
			}
			mode = parseModeMatching
		}

		if mode == parseModeMatching {
			openMatch, openMatchDone := openMatcher.MatchNext(ch)
			closeMatch, closeMatchDone := closeMatcher.MatchNext(ch)

			if openMatch {
				if tagLen < maxTagSize {
					tagBuf[tagLen] = ch
					tagLen++
				}
				if !openMatchDone {
					continue
				}
				tagLen = 0
				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue
			}
			openMatcher.Reset()

			if closeMatch {
				if tagLen < maxTagSize {
					tagBuf[tagLen] = ch
					tagLen++
				}
				if !closeMatchDone {
					continue
				}
				tagLen = 0
				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue
			}
			closeMatcher.Reset()

			if closeMatchDone && openMatchDone {
				if tagLen < maxTagSize {
					tagBuf[tagLen] = ch
					tagLen++
				}
			}
			mode = parseModeNone
			for j := 0; j < tagLen; j++ {
				recordVisible(tagBuf[j])
			}
			tagLen = 0
			continue
		}
	}
	return points
}

// SplitStringOnSpaces splits input into segments of at most maxLen visible
// characters, preferring to split at a space boundary. If no space exists
// at or before the limit, it falls back to a character-based split.
func SplitStringOnSpaces(input string, maxLen int, trimSpace ...bool) []string {
	doTrim := true
	if len(trimSpace) > 0 {
		doTrim = trimSpace[0]
	}

	if maxLen <= 0 || len(input) == 0 {
		return []string{input}
	}

	totalVisible := visibleLen(input)
	if totalVisible <= maxLen {
		return []string{input}
	}

	points := splitPoints(input, maxLen)

	var result []string
	var tagStack []string
	var current strings.Builder
	visibleCount := 0
	totalConsumed := 0
	pointIdx := 0

	openMatcher := NewTagMatcher(tagStart, []byte(tagOpen), tagEnd, true)
	closeMatcher := NewTagMatcher(tagStart, []byte(tagClose), tagEnd, false)

	var tagBuf [maxTagSize]byte
	var tagLen int
	var mode parseMode = parseModeNone

	split := func() {
		for j := len(tagStack) - 1; j >= 0; j-- {
			current.WriteString("</ansi>")
		}
		result = append(result, current.String())
		current.Reset()
		for _, tag := range tagStack {
			current.WriteString(tag)
		}
		visibleCount = 0
	}

	writeVisible := func(b byte) {
		current.WriteByte(b)
		visibleCount++
		totalConsumed++
		if pointIdx < len(points) && totalConsumed == points[pointIdx] && totalConsumed < totalVisible {
			pointIdx++
			split()
		}
	}

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if mode == parseModeNone {
			if ch != tagStart {
				writeVisible(ch)
				continue
			}
			mode = parseModeMatching
		}

		if mode == parseModeMatching {
			openMatch, openMatchDone := openMatcher.MatchNext(ch)
			closeMatch, closeMatchDone := closeMatcher.MatchNext(ch)

			if openMatch {
				if tagLen < maxTagSize {
					tagBuf[tagLen] = ch
					tagLen++
				}
				if !openMatchDone {
					continue
				}
				tagStr := string(tagBuf[:tagLen])
				tagStack = append(tagStack, tagStr)
				current.WriteString(tagStr)
				tagLen = 0
				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue
			}
			openMatcher.Reset()

			if closeMatch {
				if tagLen < maxTagSize {
					tagBuf[tagLen] = ch
					tagLen++
				}
				if !closeMatchDone {
					continue
				}
				tagLen = 0
				if len(tagStack) > 0 {
					tagStack = tagStack[:len(tagStack)-1]
				}
				current.WriteString("</ansi>")
				mode = parseModeNone
				openMatcher.Reset()
				closeMatcher.Reset()
				continue
			}
			closeMatcher.Reset()

			if closeMatchDone && openMatchDone {
				if tagLen < maxTagSize {
					tagBuf[tagLen] = ch
					tagLen++
				}
			}

			mode = parseModeNone

			for j := 0; j < tagLen; j++ {
				writeVisible(tagBuf[j])
			}
			tagLen = 0
			continue
		}
	}

	if tagLen > 0 {
		for j := 0; j < tagLen; j++ {
			writeVisible(tagBuf[j])
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	if len(result) == 0 {
		return []string{input}
	}

	if doTrim {
		for i, seg := range result {
			result[i] = trimTagAwareSpaces(seg)
		}
	}

	return result
}

func trimTagAwareSpaces(input string) string {
	n := len(input)
	if n == 0 {
		return input
	}

	isVisible := make([]bool, n)

	openMatcher := NewTagMatcher(tagStart, []byte(tagOpen), tagEnd, true)
	closeMatcher := NewTagMatcher(tagStart, []byte(tagClose), tagEnd, false)
	var mode parseMode = parseModeNone
	var tagBufStart int

	for i := 0; i < n; i++ {
		ch := input[i]
		if mode == parseModeNone {
			if ch != tagStart {
				isVisible[i] = true
				continue
			}
			mode = parseModeMatching
			tagBufStart = i
		}
		if mode == parseModeMatching {
			openMatch, openDone := openMatcher.MatchNext(ch)
			closeMatch, closeDone := closeMatcher.MatchNext(ch)

			if openMatch {
				if openDone {
					mode = parseModeNone
					openMatcher.Reset()
					closeMatcher.Reset()
				}
				continue
			}
			openMatcher.Reset()

			if closeMatch {
				if closeDone {
					mode = parseModeNone
					openMatcher.Reset()
					closeMatcher.Reset()
				}
				continue
			}
			closeMatcher.Reset()

			mode = parseModeNone
			for j := tagBufStart; j <= i; j++ {
				isVisible[j] = true
			}
			continue
		}
	}

	if mode == parseModeMatching {
		for j := tagBufStart; j < n; j++ {
			isVisible[j] = true
		}
	}

	firstNonSpace := -1
	lastNonSpace := -1
	for i := 0; i < n; i++ {
		if isVisible[i] && input[i] != ' ' {
			if firstNonSpace == -1 {
				firstNonSpace = i
			}
			lastNonSpace = i
		}
	}

	if firstNonSpace == -1 {
		var buf strings.Builder
		for i := 0; i < n; i++ {
			if !isVisible[i] {
				buf.WriteByte(input[i])
			}
		}
		return buf.String()
	}

	var buf strings.Builder
	buf.Grow(n)
	for i := 0; i < n; i++ {
		if isVisible[i] && (i < firstNonSpace || i > lastNonSpace) {
			continue
		}
		buf.WriteByte(input[i])
	}
	return buf.String()
}

func visibleLen(input string) int {
	count := 0
	openMatcher := NewTagMatcher(tagStart, []byte(tagOpen), tagEnd, true)
	closeMatcher := NewTagMatcher(tagStart, []byte(tagClose), tagEnd, false)
	var tagLen int
	var mode parseMode = parseModeNone

	for i := 0; i < len(input); i++ {
		ch := input[i]
		if mode == parseModeNone {
			if ch != tagStart {
				count++
				continue
			}
			mode = parseModeMatching
		}
		if mode == parseModeMatching {
			openMatch, openMatchDone := openMatcher.MatchNext(ch)
			closeMatch, closeMatchDone := closeMatcher.MatchNext(ch)
			if openMatch {
				tagLen++
				if openMatchDone {
					tagLen = 0
					mode = parseModeNone
					openMatcher.Reset()
					closeMatcher.Reset()
				}
				continue
			}
			openMatcher.Reset()
			if closeMatch {
				tagLen++
				if closeMatchDone {
					tagLen = 0
					mode = parseModeNone
					openMatcher.Reset()
					closeMatcher.Reset()
				}
				continue
			}
			closeMatcher.Reset()
			if closeMatchDone && openMatchDone {
				tagLen++
			}
			mode = parseModeNone
			count += tagLen
			tagLen = 0
			continue
		}
	}
	if tagLen > 0 {
		count += tagLen
	}
	return count
}
