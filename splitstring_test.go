package ansitags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitStringNestedTags(t *testing.T) {
	input := `<ansi fg="yellow">This is some <ansi fg="black">long as heck</ansi> text</ansi>`
	result := SplitString(input, 17, false)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, `<ansi fg="yellow">This is some <ansi fg="black">long</ansi></ansi>`, result[0])
	assert.Equal(t, `<ansi fg="yellow"><ansi fg="black"> as heck</ansi> text</ansi>`, result[1])
}

func TestSplitStringNoTags(t *testing.T) {
	result := SplitString("Hello World", 5, false)
	assert.Equal(t, []string{"Hello", " Worl", "d"}, result)
}

func TestSplitStringNoSplitNeeded(t *testing.T) {
	input := `<ansi fg="red">Hi</ansi>`
	result := SplitString(input, 10, false)
	assert.Equal(t, []string{input}, result)
}

func TestSplitStringExactLength(t *testing.T) {
	input := `<ansi fg="red">Hello</ansi>`
	result := SplitString(input, 5, false)
	assert.Equal(t, []string{input}, result)
}

func TestSplitStringEmptyInput(t *testing.T) {
	result := SplitString("", 10)
	assert.Equal(t, []string{""}, result)
}

func TestSplitStringZeroMaxLen(t *testing.T) {
	result := SplitString("hello", 0)
	assert.Equal(t, []string{"hello"}, result)
}

func TestSplitStringMultipleSplits(t *testing.T) {
	input := `<ansi fg="red">abcdefghijklmno</ansi>`
	result := SplitString(input, 5, false)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red">abcde</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red">fghij</ansi>`, result[1])
	assert.Equal(t, `<ansi fg="red">klmno</ansi>`, result[2])
}

func TestSplitStringTagAtSplitBoundary(t *testing.T) {
	input := `AB<ansi fg="red">CD</ansi>EF`
	result := SplitString(input, 3, false)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, `AB<ansi fg="red">C</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red">D</ansi>EF`, result[1])
}

func TestSplitStringDeeplyNested(t *testing.T) {
	input := `<ansi fg="red"><ansi fg="green"><ansi fg="blue">Hello World</ansi></ansi></ansi>`
	result := SplitString(input, 5, false)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red"><ansi fg="green"><ansi fg="blue">Hello</ansi></ansi></ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red"><ansi fg="green"><ansi fg="blue"> Worl</ansi></ansi></ansi>`, result[1])
	assert.Equal(t, `<ansi fg="red"><ansi fg="green"><ansi fg="blue">d</ansi></ansi></ansi>`, result[2])
}

func TestSplitStringTagClosesBeforeSplit(t *testing.T) {
	input := `<ansi fg="red">AB</ansi>CDEF`
	result := SplitString(input, 3, false)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, `<ansi fg="red">AB</ansi>C`, result[0])
	assert.Equal(t, `DEF`, result[1])
}

func TestSplitStringOnlyTags(t *testing.T) {
	input := `<ansi fg="red"></ansi>`
	result := SplitString(input, 5, false)
	assert.Equal(t, []string{input}, result)
}

func TestSplitStringVisibleLength(t *testing.T) {
	assert.Equal(t, 5, visibleLen("Hello"))
	assert.Equal(t, 5, visibleLen(`<ansi fg="red">Hello</ansi>`))
	assert.Equal(t, 11, visibleLen(`<ansi fg="red">Hello</ansi> World`))
	assert.Equal(t, 0, visibleLen(`<ansi fg="red"></ansi>`))
	assert.Equal(t, 5, visibleLen(`<ansi fg="red"><ansi fg="blue">Hello</ansi></ansi>`))
}

func TestSplitStringSplitPreservesAllChars(t *testing.T) {
	input := `<ansi fg="yellow">This is some <ansi fg="black">long as heck</ansi> text</ansi>`
	result := SplitString(input, 17, false)

	stripped1 := Parse(result[0], StripTags)
	stripped2 := Parse(result[1], StripTags)
	strippedAll := Parse(input, StripTags)

	assert.Equal(t, strippedAll, stripped1+stripped2)
}

func TestSplitStringLongMultiColorParagraph(t *testing.T) {
	input := `<ansi fg="red">The quick </ansi><ansi fg="green">brown fox </ansi><ansi fg="blue">jumps over </ansi><ansi fg="yellow">the lazy </ansi><ansi fg="magenta">dog and then runs away</ansi>`
	result := SplitString(input, 20, false)

	// visible: "The quick brown fox jumps over the lazy dog and then runs away" = 62 chars
	// 62 / 20 => 4 segments (20, 20, 20, 2)
	assert.Equal(t, 4, len(result))

	// Verify no visible characters are lost
	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, Parse(input, StripTags), allStripped)

	// Verify segment visible lengths
	assert.Equal(t, 20, visibleLen(result[0]))
	assert.Equal(t, 20, visibleLen(result[1]))
	assert.Equal(t, 20, visibleLen(result[2]))
	assert.Equal(t, 2, visibleLen(result[3]))

	// Verify each segment parses without panic
	for i, seg := range result {
		parsed := Parse(seg)
		assert.NotEmpty(t, parsed, "segment %d should produce output", i)
	}
}

func TestSplitStringNestedTagsOpenAndCloseAcrossMultipleSplits(t *testing.T) {
	// Outer tag spans entire string, inner tags open and close at various points
	input := `<ansi fg="red">Hello <ansi fg="green">World, this is a <ansi fg="blue">deeply nested and very long</ansi> string that</ansi> continues outside the inner tags with more text here</ansi>`
	result := SplitString(input, 15, false)

	// visible: "Hello World, this is a deeply nested and very long string that continues outside the inner tags with more text here" = 115 chars
	// 115 / 15 => 8 segments (7 full + 1 partial)
	expectedVisible := "Hello World, this is a deeply nested and very long string that continues outside the inner tags with more text here"
	assert.Equal(t, len(expectedVisible), visibleLen(input))

	expectedSegments := (len(expectedVisible) + 14) / 15 // ceil division
	assert.Equal(t, expectedSegments, len(result))

	// Every segment except the last must have exactly 15 visible chars
	for i := 0; i < len(result)-1; i++ {
		assert.Equal(t, 15, visibleLen(result[i]), "segment %d visible length", i)
	}
	assert.LessOrEqual(t, visibleLen(result[len(result)-1]), 15)

	// Concatenation of stripped segments must equal stripped original
	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, expectedVisible, allStripped)
}

func TestSplitStringManySequentialTags(t *testing.T) {
	// Many short sequential tags, no nesting
	input := `<ansi fg="red">AB</ansi><ansi fg="green">CD</ansi><ansi fg="blue">EF</ansi><ansi fg="yellow">GH</ansi><ansi fg="magenta">IJ</ansi><ansi fg="cyan">KL</ansi><ansi fg="white">MN</ansi><ansi fg="red">OP</ansi>`
	result := SplitString(input, 3, false)

	// visible: "ABCDEFGHIJKLMNOP" = 16 chars
	// 16 / 3 => 6 segments (5 full + 1 with 1 char)
	assert.Equal(t, 16, visibleLen(input))

	expectedSegments := 6
	assert.Equal(t, expectedSegments, len(result))

	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, "ABCDEFGHIJKLMNOP", allStripped)

	// First segment: "ABC" — 'A','B' from red, 'C' from green
	assert.Equal(t, `<ansi fg="red">AB</ansi><ansi fg="green">C</ansi>`, result[0])
	// Second segment: "DEF" — 'D' continues green, 'E','F' from blue
	assert.Equal(t, `<ansi fg="green">D</ansi><ansi fg="blue">EF</ansi>`, result[1])
}

func TestSplitStringAlternatingTaggedAndUntagged(t *testing.T) {
	input := `plain1<ansi fg="red">RED</ansi>plain2<ansi fg="blue">BLUE</ansi>plain3<ansi fg="green">GREEN</ansi>plain4`
	result := SplitString(input, 10, false)

	// visible: "plain1REDplain2BLUEplain3GREENplain4" = 36 chars
	assert.Equal(t, 36, visibleLen(input))

	// 36 / 10 => 4 segments (3 full + 1 with 6)
	assert.Equal(t, 4, len(result))

	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, "plain1REDplain2BLUEplain3GREENplain4", allStripped)

	for i := 0; i < len(result)-1; i++ {
		assert.Equal(t, 10, visibleLen(result[i]), "segment %d", i)
	}
}

func TestSplitStringDeep5LevelNesting(t *testing.T) {
	input := `<ansi fg="red"><ansi fg="green"><ansi fg="blue"><ansi fg="yellow"><ansi fg="magenta">This text is five levels deep and should be split properly across segments</ansi></ansi></ansi></ansi></ansi>`
	result := SplitString(input, 12, false)

	// visible = "This text is five levels deep and should be split properly across segments" = 74 chars
	expectedVisible := "This text is five levels deep and should be split properly across segments"
	assert.Equal(t, len(expectedVisible), visibleLen(input))

	// ceil(74/12) = 7 segments
	assert.Equal(t, 7, len(result))

	// Each segment must reopen all 5 tags and close all 5
	for i := 0; i < len(result)-1; i++ {
		assert.Equal(t, 12, visibleLen(result[i]), "segment %d visible length", i)
	}

	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, expectedVisible, allStripped)

	// Verify the first segment has all 5 opening tags and 5 closing tags
	assert.Contains(t, result[0], `<ansi fg="red">`)
	assert.Contains(t, result[0], `<ansi fg="green">`)
	assert.Contains(t, result[0], `<ansi fg="blue">`)
	assert.Contains(t, result[0], `<ansi fg="yellow">`)
	assert.Contains(t, result[0], `<ansi fg="magenta">`)

	// Middle segments must reopen all 5 levels
	for i := 1; i < len(result)-1; i++ {
		seg := result[i]
		assert.Contains(t, seg, `<ansi fg="red">`, "segment %d missing red reopen", i)
		assert.Contains(t, seg, `<ansi fg="magenta">`, "segment %d missing magenta reopen", i)
	}
}

func TestSplitStringNestingChangesAcrossSplits(t *testing.T) {
	// Nesting depth changes across the string: 1 level -> 2 levels -> 1 level -> 0 levels
	input := `<ansi fg="red">Level one <ansi fg="green">level two here</ansi> back to one</ansi> and now plain`
	result := SplitString(input, 10, false)

	// visible: "Level one level two here back to one and now plain" = 50 chars
	expectedVisible := "Level one level two here back to one and now plain"
	assert.Equal(t, len(expectedVisible), visibleLen(input))
	assert.Equal(t, 5, len(result))

	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, expectedVisible, allStripped)

	// Parse each segment to verify they produce valid ANSI output
	for i, seg := range result {
		parsed := Parse(seg)
		assert.NotEmpty(t, parsed, "segment %d should parse", i)
	}
}

func TestSplitStringWithBgAttributes(t *testing.T) {
	input := `<ansi fg="red" bg="white">Warning: <ansi fg="yellow" bg="black">critical error in module</ansi> please check logs immediately</ansi>`
	result := SplitString(input, 15, false)

	// visible: "Warning: critical error in module please check logs immediately" = 63 chars
	expectedVisible := "Warning: critical error in module please check logs immediately"
	assert.Equal(t, len(expectedVisible), visibleLen(input))

	expectedSegments := (len(expectedVisible) + 14) / 15
	assert.Equal(t, expectedSegments, len(result))

	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, expectedVisible, allStripped)

	// Verify bg attributes are preserved in reopened tags
	assert.Contains(t, result[0], `bg="white"`)
	if len(result) > 1 {
		assert.Contains(t, result[1], `bg="white"`, "second segment should reopen outer tag with bg")
	}
}

func TestSplitStringMaxLen1(t *testing.T) {
	input := `<ansi fg="red">ABCDE</ansi>`
	result := SplitString(input, 1, false)

	assert.Equal(t, 5, len(result))
	assert.Equal(t, `<ansi fg="red">A</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red">B</ansi>`, result[1])
	assert.Equal(t, `<ansi fg="red">C</ansi>`, result[2])
	assert.Equal(t, `<ansi fg="red">D</ansi>`, result[3])
	assert.Equal(t, `<ansi fg="red">E</ansi>`, result[4])
}

func TestSplitStringTagOpensExactlyAtSplit(t *testing.T) {
	// A new tag opens right where the split happens
	// "1234567890" = 10 chars, split at 5
	// The tag opens around char 5
	input := `12345<ansi fg="red">67890</ansi>`
	result := SplitString(input, 5, false)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, `12345`, result[0])
	assert.Equal(t, `<ansi fg="red">67890</ansi>`, result[1])
}

func TestSplitStringTagClosesExactlyAtSplit(t *testing.T) {
	// Tag closes right at the split boundary — the tag is still open at
	// the split point, so it gets reopened and then immediately closed
	// by the original </ansi> in the next segment.
	input := `<ansi fg="red">12345</ansi>67890ABCDE`
	result := SplitString(input, 5, false)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red">12345</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red"></ansi>67890`, result[1])
	assert.Equal(t, `ABCDE`, result[2])
}

func TestSplitStringLongParagraphMultipleTagStyles(t *testing.T) {
	input := `<ansi fg="white">In the beginning, the universe was created. </ansi>` +
		`<ansi fg="yellow">This has made a lot of people very angry </ansi>` +
		`<ansi fg="red">and has been widely regarded as a <ansi fg="white" bg="red">bad move</ansi>. </ansi>` +
		`<ansi fg="green">The ships hung in the sky in much the same way that <ansi fg="cyan">bricks don't</ansi>.</ansi>`
	result := SplitString(input, 25, false)

	expectedVisible := "In the beginning, the universe was created. This has made a lot of people very angry and has been widely regarded as a bad move. The ships hung in the sky in much the same way that bricks don't."
	assert.Equal(t, len(expectedVisible), visibleLen(input))

	// Verify all text preserved
	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, expectedVisible, allStripped)

	// Verify each segment is at most 25 visible chars
	for i, seg := range result {
		vl := visibleLen(seg)
		assert.LessOrEqual(t, vl, 25, "segment %d has %d visible chars", i, vl)
	}

	// All but last should be exactly 25
	for i := 0; i < len(result)-1; i++ {
		assert.Equal(t, 25, visibleLen(result[i]), "segment %d", i)
	}
}

func TestSplitStringRepeatedSplitAndRejoin(t *testing.T) {
	// Split a string, then join and strip — should match the original stripped text
	input := `<ansi fg="red">AAA<ansi fg="green">BBB<ansi fg="blue">CCC</ansi>DDD</ansi>EEE</ansi>` +
		`<ansi fg="yellow">FFF<ansi fg="magenta">GGG</ansi>HHH</ansi>` +
		`<ansi fg="cyan">III</ansi>JJJ`
	// visible: "AAABBBCCCDDDEEEFFFGGGHHH IIIJJJ" = 30 chars

	expectedVisible := Parse(input, StripTags)

	for _, maxLen := range []int{1, 2, 3, 4, 5, 7, 10, 13, 15, 29, 30, 100} {
		t.Run("maxLen="+string(rune('0'+maxLen/10))+string(rune('0'+maxLen%10)), func(t *testing.T) {
			result := SplitString(input, maxLen, false)

			var allStripped string
			for _, seg := range result {
				allStripped += Parse(seg, StripTags)
			}
			assert.Equal(t, expectedVisible, allStripped, "maxLen=%d", maxLen)

			for i, seg := range result {
				vl := visibleLen(seg)
				assert.LessOrEqual(t, vl, maxLen, "segment %d has %d visible chars (maxLen=%d)", i, vl, maxLen)
			}

			if visibleLen(input) > maxLen {
				for i := 0; i < len(result)-1; i++ {
					assert.Equal(t, maxLen, visibleLen(result[i]), "segment %d should be full (maxLen=%d)", i, maxLen)
				}
			}
		})
	}
}

func TestSplitStringEmptyTagsBetweenText(t *testing.T) {
	input := `AA<ansi fg="red"></ansi>BB<ansi fg="green"></ansi>CC`
	result := SplitString(input, 3, false)

	// visible: "AABBCC" = 6 chars
	assert.Equal(t, 6, visibleLen(input))
	assert.Equal(t, 2, len(result))

	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, "AABBCC", allStripped)
}

func TestSplitStringConsecutiveTagsNoGap(t *testing.T) {
	input := `<ansi fg="red">Hello</ansi><ansi fg="blue">World</ansi><ansi fg="green">Again</ansi>`
	result := SplitString(input, 7, false)

	// visible: "HelloWorldAgain" = 15 chars
	assert.Equal(t, 15, visibleLen(input))
	assert.Equal(t, 3, len(result))

	assert.Equal(t, 7, visibleLen(result[0]))
	assert.Equal(t, 7, visibleLen(result[1]))
	assert.Equal(t, 1, visibleLen(result[2]))

	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, "HelloWorldAgain", allStripped)
}

func TestSplitStringUnquotedAttributes(t *testing.T) {
	input := `<ansi fg=red bg=blue>Some colored text that is fairly long</ansi>`
	result := SplitString(input, 10, false)

	expectedVisible := "Some colored text that is fairly long"
	assert.Equal(t, len(expectedVisible), visibleLen(input))

	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, expectedVisible, allStripped)

	// Verify unquoted attributes are preserved
	assert.Contains(t, result[0], `<ansi fg=red bg=blue>`)
	if len(result) > 1 {
		assert.Contains(t, result[1], `<ansi fg=red bg=blue>`)
	}
}

func TestSplitStringAllSegmentsParseable(t *testing.T) {
	inputs := []string{
		`<ansi fg="red">Simple</ansi>`,
		`<ansi fg="red"><ansi fg="green"><ansi fg="blue">Triple nested long text here</ansi></ansi></ansi>`,
		`Before<ansi fg="red">middle</ansi>after`,
		`<ansi fg="red">A</ansi><ansi fg="green">B</ansi><ansi fg="red">C</ansi><ansi fg="green">D</ansi>`,
		`<ansi fg="red" bg="blue">Mixed <ansi fg="green">attributes <ansi fg="yellow" bg="white">everywhere</ansi> in this</ansi> string</ansi>`,
	}

	for _, input := range inputs {
		for maxLen := 1; maxLen <= 10; maxLen++ {
			result := SplitString(input, maxLen, false)
			for i, seg := range result {
				parsed := Parse(seg)
				_ = parsed
				parsedHTML := Parse(seg, HTML)
				_ = parsedHTML
				stripped := Parse(seg, StripTags)
				_ = stripped
				assert.LessOrEqual(t, visibleLen(seg), maxLen, "input=%q maxLen=%d seg=%d", input, maxLen, i)
			}
		}
	}
}

func TestSplitStringLongRealWorldExample(t *testing.T) {
	input := `<ansi fg="white">You are standing in a <ansi fg="green">lush forest</ansi>. ` +
		`The trees tower above you, their <ansi fg="green">leaves rustling</ansi> in the wind. ` +
		`A <ansi fg="yellow">narrow path</ansi> leads <ansi fg="red">north</ansi> toward ` +
		`a <ansi fg="magenta">dark cave</ansi>, while another trail winds ` +
		`<ansi fg="cyan">east</ansi> through the <ansi fg="green">underbrush</ansi>. ` +
		`<ansi fg="yellow">A small <ansi fg="red">treasure chest</ansi> sits near the base of an old oak tree</ansi>.</ansi>`
	result := SplitString(input, 40, false)

	expectedVisible := Parse(input, StripTags)
	assert.Equal(t, len(expectedVisible), visibleLen(input))

	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, expectedVisible, allStripped)

	for i, seg := range result {
		vl := visibleLen(seg)
		assert.LessOrEqual(t, vl, 40, "segment %d has %d visible chars", i, vl)
	}
	for i := 0; i < len(result)-1; i++ {
		assert.Equal(t, 40, visibleLen(result[i]), "segment %d should be exactly 40", i)
	}
}

func TestSplitStringSingleTagSpansManySegments(t *testing.T) {
	input := `<ansi fg="red">` +
		`abcdefghijklmnopqrstuvwxyz` +
		`ABCDEFGHIJKLMNOPQRSTUVWXYZ` +
		`0123456789` +
		`</ansi>`
	result := SplitString(input, 8, false)

	// visible: 26+26+10 = 62 chars, ceil(62/8) = 8 segments
	assert.Equal(t, 62, visibleLen(input))
	assert.Equal(t, 8, len(result))

	// Every segment must open and close the red tag
	for i, seg := range result {
		assert.Contains(t, seg, `<ansi fg="red">`, "segment %d missing open tag", i)
		assert.Contains(t, seg, `</ansi>`, "segment %d missing close tag", i)
	}

	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", allStripped)
}

func TestSplitStringNestingDepthChangesEveryFewChars(t *testing.T) {
	// Rapidly changing nesting depth
	input := `A<ansi fg="red">B<ansi fg="green">C</ansi>D</ansi>E` +
		`<ansi fg="blue">F<ansi fg="yellow">G<ansi fg="magenta">H</ansi>I</ansi>J</ansi>` +
		`K<ansi fg="cyan">L</ansi>M`
	result := SplitString(input, 4, false)

	// visible: "ABCDEFGHIJKLM" = 13 chars
	assert.Equal(t, 13, visibleLen(input))

	var allStripped string
	for _, seg := range result {
		allStripped += Parse(seg, StripTags)
	}
	assert.Equal(t, "ABCDEFGHIJKLM", allStripped)

	for i, seg := range result {
		assert.LessOrEqual(t, visibleLen(seg), 4, "segment %d exceeds maxLen", i)
	}
}

func BenchmarkSplitString(b *testing.B) {
	input := `<ansi fg="red">This is some <ansi fg="blue">long text that needs to be split</ansi> across multiple lines</ansi>`
	for n := 0; n < b.N; n++ {
		SplitString(input, 10, false)
	}
}

func BenchmarkSplitStringDeepNesting(b *testing.B) {
	input := `<ansi fg="red"><ansi fg="green"><ansi fg="blue"><ansi fg="yellow"><ansi fg="magenta">` +
		`This is a deeply nested string that should be split into many segments for benchmarking purposes` +
		`</ansi></ansi></ansi></ansi></ansi>`
	for n := 0; n < b.N; n++ {
		SplitString(input, 10, false)
	}
}

func BenchmarkSplitStringManyTags(b *testing.B) {
	input := ""
	for i := 0; i < 20; i++ {
		input += `<ansi fg="red">AB</ansi><ansi fg="green">CD</ansi>`
	}
	for n := 0; n < b.N; n++ {
		SplitString(input, 15, false)
	}
}

// --- trimSpace tests (default true) ---

func TestSplitStringTrimDefault(t *testing.T) {
	input := `<ansi fg="yellow">This is some <ansi fg="black">long as heck</ansi> text</ansi>`
	result := SplitString(input, 17)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, `<ansi fg="yellow">This is some <ansi fg="black">long</ansi></ansi>`, result[0])
	assert.Equal(t, `<ansi fg="yellow"><ansi fg="black">as heck</ansi> text</ansi>`, result[1])
}

func TestSplitStringTrimLeadingSpaceInsideTag(t *testing.T) {
	input := `<ansi fg="red">Hello World</ansi>`
	result := SplitString(input, 5)

	// "Hello" (5) | " Worl" (5) → trim → "Worl" | "d" (1)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red">Hello</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red">Worl</ansi>`, result[1])
	assert.Equal(t, `<ansi fg="red">d</ansi>`, result[2])
}

func TestSplitStringTrimTrailingSpaceInsideTag(t *testing.T) {
	input := `<ansi fg="red">test </ansi><ansi fg="blue">more</ansi>`
	result := SplitString(input, 5)

	// "test " splits at 5. Red tag is still open at split, so it reopens
	// then immediately closes when the original </ansi> is hit.
	assert.Equal(t, 2, len(result))
	assert.Equal(t, `<ansi fg="red">test</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red"></ansi><ansi fg="blue">more</ansi>`, result[1])
}

func TestSplitStringTrimPlainText(t *testing.T) {
	result := SplitString("Hello World", 5)
	assert.Equal(t, []string{"Hello", "Worl", "d"}, result)
}

func TestSplitStringTrimNestedTags(t *testing.T) {
	input := `<ansi fg="red"><ansi fg="green"><ansi fg="blue">Hello World</ansi></ansi></ansi>`
	result := SplitString(input, 5)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red"><ansi fg="green"><ansi fg="blue">Hello</ansi></ansi></ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red"><ansi fg="green"><ansi fg="blue">Worl</ansi></ansi></ansi>`, result[1])
	assert.Equal(t, `<ansi fg="red"><ansi fg="green"><ansi fg="blue">d</ansi></ansi></ansi>`, result[2])
}

func TestSplitStringTrimPreservesInternalSpaces(t *testing.T) {
	input := `<ansi fg="red">a b c d e f g</ansi>`
	result := SplitString(input, 5)

	// "a b c" (5) | " d e " (5) → trim → "d e" | "f g" (3)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red">a b c</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red">d e</ansi>`, result[1])
	assert.Equal(t, `<ansi fg="red">f g</ansi>`, result[2])
}

func TestSplitStringTrimFalsePreservesSpaces(t *testing.T) {
	input := `<ansi fg="red">Hello World</ansi>`
	result := SplitString(input, 5, false)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red">Hello</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red"> Worl</ansi>`, result[1])
	assert.Equal(t, `<ansi fg="red">d</ansi>`, result[2])
}

func TestSplitStringTrimMultipleLeadingSpaces(t *testing.T) {
	input := `<ansi fg="red">abc   def</ansi>`
	result := SplitString(input, 4)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red">abc</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red">de</ansi>`, result[1])
	assert.Equal(t, `<ansi fg="red">f</ansi>`, result[2])
}

func TestSplitStringTrimMultipleTrailingSpaces(t *testing.T) {
	input := `abc   def`
	result := SplitString(input, 5)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, `abc`, result[0])
	assert.Equal(t, `def`, result[1])
}

func TestSplitStringTrimAllSpaceSegment(t *testing.T) {
	input := `<ansi fg="red">     hello</ansi>`
	result := SplitString(input, 3)

	// Without trim: ["   ", "  h", "ell", "o"]
	// With trim:    ["", "h", "ell", "o"] — first two segments have spaces trimmed
	assert.Equal(t, 4, len(result))
	assert.Equal(t, `<ansi fg="red"></ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red">h</ansi>`, result[1])
	assert.Equal(t, `<ansi fg="red">ell</ansi>`, result[2])
	assert.Equal(t, `<ansi fg="red">o</ansi>`, result[3])
}

func TestSplitStringTrimTagCloseThenSpace(t *testing.T) {
	input := `<ansi fg="red">Hello</ansi> <ansi fg="blue">World</ansi>`
	result := SplitString(input, 6)

	// Without trim: ["Hello ", "<ansi fg=\"blue\">World</ansi>"]
	// With trim:    ["Hello", "<ansi fg=\"blue\">World</ansi>"]
	assert.Equal(t, 2, len(result))
	assert.Equal(t, `<ansi fg="red">Hello</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="blue">World</ansi>`, result[1])
}

func TestSplitStringTrimLongRealWorld(t *testing.T) {
	input := `<ansi fg="white">You enter the <ansi fg="green">forest clearing</ansi>. ` +
		`A <ansi fg="yellow">golden light</ansi> shines from above.</ansi>`
	result := SplitString(input, 20)

	// Each segment should have no leading/trailing spaces
	for i, seg := range result {
		stripped := Parse(seg, StripTags)
		if len(stripped) > 0 {
			assert.NotEqual(t, ' ', rune(stripped[0]), "segment %d has leading space", i)
			assert.NotEqual(t, ' ', rune(stripped[len(stripped)-1]), "segment %d has trailing space", i)
		}
	}

	// All segments should be parseable
	for i, seg := range result {
		parsed := Parse(seg)
		assert.NotEmpty(t, parsed, "segment %d should parse", i)
	}
}

func TestSplitStringTrimSpaceBetweenTags(t *testing.T) {
	// Space sits between two tags at the split boundary.
	// Red is still open at the split point, so it reopens then immediately closes.
	input := `<ansi fg="red">AAAA</ansi> <ansi fg="blue">BBBB</ansi>`
	result := SplitString(input, 4)

	// " BBB" (4) → trim → "BBB" with reopened/closed red prefix
	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red">AAAA</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red"></ansi><ansi fg="blue">BBB</ansi>`, result[1])
	assert.Equal(t, `<ansi fg="blue">B</ansi>`, result[2])
}
