package ansitags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- SplitString tests ---

func TestSplitStringBasic(t *testing.T) {
	result := SplitString("Hello World", 5, false)
	assert.Equal(t, []string{"Hello", " Worl", "d"}, result)
}

func TestSplitStringNoSplitNeeded(t *testing.T) {
	input := `<ansi fg="red">Hi</ansi>`
	assert.Equal(t, []string{input}, SplitString(input, 10, false))
}

func TestSplitStringExactLength(t *testing.T) {
	input := `<ansi fg="red">Hello</ansi>`
	assert.Equal(t, []string{input}, SplitString(input, 5, false))
}

func TestSplitStringEdgeCases(t *testing.T) {
	assert.Equal(t, []string{""}, SplitString("", 10))
	assert.Equal(t, []string{"hello"}, SplitString("hello", 0))
}

func TestSplitStringNestedTags(t *testing.T) {
	input := `<ansi fg="yellow">This is some <ansi fg="black">long as heck</ansi> text</ansi>`
	result := SplitString(input, 17, false)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, `<ansi fg="yellow">This is some <ansi fg="black">long</ansi></ansi>`, result[0])
	assert.Equal(t, `<ansi fg="yellow"><ansi fg="black"> as heck</ansi> text</ansi>`, result[1])
}

func TestSplitStringMultipleSplits(t *testing.T) {
	input := `<ansi fg="red">abcdefghijklmno</ansi>`
	result := SplitString(input, 5, false)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red">abcde</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red">fghij</ansi>`, result[1])
	assert.Equal(t, `<ansi fg="red">klmno</ansi>`, result[2])
}

func TestSplitStringTagAtBoundary(t *testing.T) {
	result := SplitString(`AB<ansi fg="red">CD</ansi>EF`, 3, false)
	assert.Equal(t, []string{`AB<ansi fg="red">C</ansi>`, `<ansi fg="red">D</ansi>EF`}, result)
}

func TestSplitStringMaxLen1(t *testing.T) {
	input := `<ansi fg="red">ABCDE</ansi>`
	result := SplitString(input, 1, false)

	assert.Equal(t, 5, len(result))
	for i, ch := range "ABCDE" {
		assert.Equal(t, `<ansi fg="red">`+string(ch)+`</ansi>`, result[i])
	}
}

func TestSplitStringPreservesAllChars(t *testing.T) {
	input := `<ansi fg="red">Hello <ansi fg="green">World, this is a <ansi fg="blue">deeply nested and very long</ansi> string that</ansi> continues here</ansi>`
	result := SplitString(input, 15, false)

	var all string
	for _, seg := range result {
		all += Parse(seg, StripTags)
	}
	assert.Equal(t, Parse(input, StripTags), all)

	for i, seg := range result[:len(result)-1] {
		assert.Equal(t, 15, visibleLen(seg), "segment %d", i)
	}
}

func TestSplitStringTagClosesExactlyAtSplit(t *testing.T) {
	result := SplitString(`<ansi fg="red">12345</ansi>67890ABCDE`, 5, false)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red">12345</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red"></ansi>67890`, result[1])
	assert.Equal(t, `ABCDE`, result[2])
}

func TestSplitStringAllSegmentsParseable(t *testing.T) {
	inputs := []string{
		`<ansi fg="red">Simple</ansi>`,
		`<ansi fg="red"><ansi fg="green"><ansi fg="blue">Triple nested long text here</ansi></ansi></ansi>`,
		`Before<ansi fg="red">middle</ansi>after`,
		`<ansi fg="red" bg="blue">Mixed <ansi fg="green">attributes <ansi fg="yellow" bg="white">everywhere</ansi> in this</ansi> string</ansi>`,
	}

	for _, input := range inputs {
		for maxLen := 1; maxLen <= 10; maxLen++ {
			result := SplitString(input, maxLen, false)
			for i, seg := range result {
				assert.LessOrEqual(t, visibleLen(seg), maxLen, "input=%q maxLen=%d seg=%d", input, maxLen, i)
			}
			var all string
			for _, seg := range result {
				all += Parse(seg, StripTags)
			}
			assert.Equal(t, Parse(input, StripTags), all, "input=%q maxLen=%d", input, maxLen)
		}
	}
}

// --- SplitString trimSpace tests (default true) ---

func TestSplitStringTrimDefault(t *testing.T) {
	input := `<ansi fg="yellow">This is some <ansi fg="black">long as heck</ansi> text</ansi>`
	result := SplitString(input, 17)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, `<ansi fg="yellow">This is some <ansi fg="black">long</ansi></ansi>`, result[0])
	assert.Equal(t, `<ansi fg="yellow"><ansi fg="black">as heck</ansi> text</ansi>`, result[1])
}

func TestSplitStringTrimLeadingAndTrailing(t *testing.T) {
	assert.Equal(t, []string{"Hello", "Worl", "d"}, SplitString("Hello World", 5))

	input := `<ansi fg="red">Hello World</ansi>`
	result := SplitString(input, 5)
	assert.Equal(t, `<ansi fg="red">Hello</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red">Worl</ansi>`, result[1])
}

func TestSplitStringTrimPreservesInternalSpaces(t *testing.T) {
	input := `<ansi fg="red">a b c d e f g</ansi>`
	result := SplitString(input, 5)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, `<ansi fg="red">a b c</ansi>`, result[0])
	assert.Equal(t, `<ansi fg="red">d e</ansi>`, result[1])
	assert.Equal(t, `<ansi fg="red">f g</ansi>`, result[2])
}

func TestSplitStringVisibleLength(t *testing.T) {
	assert.Equal(t, 5, visibleLen("Hello"))
	assert.Equal(t, 5, visibleLen(`<ansi fg="red">Hello</ansi>`))
	assert.Equal(t, 11, visibleLen(`<ansi fg="red">Hello</ansi> World`))
	assert.Equal(t, 0, visibleLen(`<ansi fg="red"></ansi>`))
}

// --- SplitStringOnSpaces tests ---

func TestSplitStringOnSpacesBasic(t *testing.T) {
	// Splits at the space between words.
	result := SplitStringOnSpaces("Hello World", 8)
	assert.Equal(t, []string{"Hello", "World"}, result)
}

func TestSplitStringOnSpacesFallbackNoSpace(t *testing.T) {
	// No space available — falls back to character split.
	result := SplitStringOnSpaces("ABCDEFGHIJ", 4, false)
	assert.Equal(t, []string{"ABCD", "EFGH", "IJ"}, result)
}

func TestSplitStringOnSpacesNoSplitNeeded(t *testing.T) {
	input := `<ansi fg="red">Hello</ansi>`
	assert.Equal(t, []string{input}, SplitStringOnSpaces(input, 10))
}

func TestSplitStringOnSpacesEdgeCases(t *testing.T) {
	assert.Equal(t, []string{""}, SplitStringOnSpaces("", 10))
	assert.Equal(t, []string{"hello"}, SplitStringOnSpaces("hello", 0))
}

func TestSplitStringOnSpacesWithTags(t *testing.T) {
	input := `<ansi fg="yellow">Hello <ansi fg="red">World foo</ansi> bar</ansi>`
	// visible: "Hello World foo bar" = 19 chars, maxLen=10
	// space at 6, 12, 16. nextTarget=10. At consumed=6, lastSpaceAt=6.
	// At consumed=10, split at 6. nextTarget=16. At consumed=12, lastSpaceAt=12.
	// At consumed=16, split at 12. nextTarget=22. Remainder "bar".
	result := SplitStringOnSpaces(input, 10)

	assert.Equal(t, 3, len(result))
	assert.Equal(t, "Hello", Parse(result[0], StripTags))
	assert.Equal(t, "World foo", Parse(result[1], StripTags))
	assert.Equal(t, "bar", Parse(result[2], StripTags))
}

func TestSplitStringOnSpacesSpaceAtLimit(t *testing.T) {
	// Space falls just after the maxLen boundary — char split fires first,
	// then the space splits the next window immediately.
	// "Hello World" maxLen=5, trimSpace=false:
	//   split at 5 (char), then split at 6 (space) → ["Hello", " ", "World"]
	result := SplitStringOnSpaces("Hello World", 5, false)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "Hello", result[0])
	assert.Equal(t, " ", result[1])
	assert.Equal(t, "World", result[2])
}

func TestSplitStringOnSpacesPreservesAllChars(t *testing.T) {
	input := `<ansi fg="white">The quick <ansi fg="green">brown fox</ansi> jumps over the lazy dog</ansi>`
	result := SplitStringOnSpaces(input, 12, false)

	var all string
	for _, seg := range result {
		all += Parse(seg, StripTags)
	}
	assert.Equal(t, Parse(input, StripTags), all)

	for _, seg := range result {
		assert.LessOrEqual(t, visibleLen(seg), 12)
	}
}

func TestSplitStringOnSpacesMixedFallback(t *testing.T) {
	// "word toolongword next" — "toolongword" exceeds maxLen=8, falls back to char split.
	result := SplitStringOnSpaces("word toolongword next", 8, false)

	var all string
	for _, seg := range result {
		all += Parse(seg, StripTags)
	}
	assert.Equal(t, "word toolongword next", all)

	for _, seg := range result {
		assert.LessOrEqual(t, visibleLen(seg), 8)
	}
}

// --- Benchmarks ---

func BenchmarkSplitString(b *testing.B) {
	input := `<ansi fg="red">This is some <ansi fg="blue">long text that needs to be split</ansi> across multiple lines</ansi>`
	for n := 0; n < b.N; n++ {
		SplitString(input, 10, false)
	}
}

func BenchmarkSplitStringOnSpaces(b *testing.B) {
	input := `<ansi fg="red">This is some <ansi fg="blue">long text that needs to be split</ansi> across multiple lines</ansi>`
	for n := 0; n < b.N; n++ {
		SplitStringOnSpaces(input, 10)
	}
}
