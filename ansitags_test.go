package ansitags

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type TestCase struct {
	Input    string `yaml:"input"`
	Expected string `yaml:"expected"`
}

func TestParseTagOpenerMismatches(t *testing.T) {

	testTable := loadTestFile("testdata/ansitags_test_tag_openers.yaml")

	for name, testCase := range testTable {

		t.Run(name, func(t *testing.T) {

			output := Parse(testCase.Input)
			assert.Equal(t, testCase.Expected, output)

			//fmt.Println(output)
			//bytes, _ := json.Marshal(output)
			//fmt.Println(string(bytes))
		})
	}

}

func TestParseAliases(t *testing.T) {

	testTable := loadTestFile("testdata/ansitags_test_aliases.yaml")

	LoadAliases("aliases.yaml")

	for name, testCase := range testTable {

		t.Run(name, func(t *testing.T) {

			output := Parse(testCase.Input)
			assert.Equal(t, testCase.Expected, output)

			//fmt.Println(output)
			//bytes, _ := json.Marshal(output)
			//fmt.Println(string(bytes))
		})
	}

}

func TestParseClear(t *testing.T) {

	testTable := loadTestFile("testdata/ansitags_test_clear.yaml")

	for name, testCase := range testTable {

		t.Run(name, func(t *testing.T) {

			output := Parse(testCase.Input)
			assert.Equal(t, testCase.Expected, output)

			//fmt.Println(output)
			//bytes, _ := json.Marshal(output)
			//fmt.Println(string(bytes))
		})
	}

}

func TestParsePosition(t *testing.T) {

	testTable := loadTestFile("testdata/ansitags_test_position.yaml")

	for name, testCase := range testTable {

		t.Run(name, func(t *testing.T) {

			output := Parse(testCase.Input)
			assert.Equal(t, testCase.Expected, output)

			//fmt.Println(output)
			//bytes, _ := json.Marshal(output)
			//fmt.Println(string(bytes))
		})
	}

}

func TestParseColor(t *testing.T) {

	testTable := loadTestFile("testdata/ansitags_test_color.yaml")

	for name, testCase := range testTable {

		t.Run(name, func(t *testing.T) {

			output := Parse(testCase.Input)
			assert.Equal(t, testCase.Expected, output)

			//fmt.Println(output)
			//bytes, _ := json.Marshal(output)
			//fmt.Println(string(bytes))
		})
	}

}

func TestHtmlMode(t *testing.T) {

	testTable := loadTestFile("testdata/ansitags_test_html.yaml")

	for name, testCase := range testTable {

		t.Run(name, func(t *testing.T) {

			output := Parse(testCase.Input, HTML)
			assert.Equal(t, testCase.Expected, output)

			//fmt.Println(output)
			//bytes, _ := json.Marshal(output)
			//fmt.Println(string(bytes))
		})
	}

}

func TestParseStripped(t *testing.T) {

	testTable := loadTestFile("testdata/ansitags_test_strip.yaml")

	for name, testCase := range testTable {

		t.Run(name, func(t *testing.T) {

			output := Parse(testCase.Input, StripTags)
			assert.Equal(t, testCase.Expected, output)

			//fmt.Println(output)
			//bytes, _ := json.Marshal(output)
			//fmt.Println(string(bytes))
		})
	}

}

func TestParseMono(t *testing.T) {

	testTable := loadTestFile("testdata/ansitags_test_monochrome.yaml")

	for name, testCase := range testTable {

		t.Run(name, func(t *testing.T) {

			output := Parse(testCase.Input, Monochrome)
			assert.Equal(t, testCase.Expected, output)

			//fmt.Println(output)
			//bytes, _ := json.Marshal(output)
			//fmt.Println(string(bytes))
		})
	}

}

func TestParseLarge(t *testing.T) {

	testString := loadRawFile("testdata/ansitags_test_streaming.yaml")

	_ = Parse(testString)

}

func loadTestFile(filename string) map[string]TestCase {

	data := make(map[string]TestCase, 100)

	if yfile, err := ioutil.ReadFile(filename); err != nil {
		panic(err)
	} else {
		if err := yaml.Unmarshal(yfile, &data); err != nil {
			panic(err)
		}
	}

	return data

}

func loadRawFile(filename string) string {

	yfile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	return string(yfile)
}

func TestSetAliasValidDefaultGroup(t *testing.T) {
	alias := "testAlias256"
	value := 123
	// Ensure clean state
	delete(colorAliases, alias)
	// Set alias in default (color256) group
	if err := SetAlias(alias, value); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := colorAliases[alias]; got != value {
		t.Errorf("colorAliases[%q] = %d; want %d", alias, got, value)
	}
}

func TestSetAliasInvalidValue(t *testing.T) {
	alias := "invalidAlias"
	invalidValue := -5
	// Expect error for out-of-range value
	if err := SetAlias(alias, invalidValue); err == nil {
		t.Errorf("expected error for invalid value %d, got none", invalidValue)
	}
}

func TestSetAliasesValidDefaultGroup(t *testing.T) {
	aliases := map[string]int{
		"multi256_1": 200,
		"multi256_2": 201,
	}
	delete(colorAliases, "multi256_1")
	delete(colorAliases, "multi256_2")
	// Bulk set in default (color256) group
	if err := SetAliases(aliases); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	for alias, expected := range aliases {
		if got := colorAliases[alias]; got != expected {
			t.Errorf("colorAliases[%q] = %d; want %d", alias, got, expected)
		}
	}
}

func TestGetAliases(t *testing.T) {
	// Default aliases should all be present as int values
	aliases := GetAliases()
	for name, num := range defaultColorAliases {
		got, ok := aliases[name]
		if !ok {
			t.Errorf("GetAliases() missing default alias %q", name)
			continue
		}
		if got != num {
			t.Errorf("GetAliases()[%q] = %v; want %d", name, got, num)
		}
	}

	// Adding an alias via SetAlias should be reflected
	if err := SetAlias("getaliases-test", 42); err != nil {
		t.Fatalf("SetAlias returned unexpected error: %v", err)
	}
	aliases = GetAliases()
	if got, ok := aliases["getaliases-test"]; !ok || got != 42 {
		t.Errorf("GetAliases()[%q] = %v, ok=%v; want %d, true", "getaliases-test", got, ok, 42)
	}

	// GetAliases must return a copy — mutations must not affect internal state
	aliases["getaliases-test"] = 999
	if got := GetAliases()["getaliases-test"]; got != 42 {
		t.Errorf("mutating returned map affected internal state: got %v, want %d", got, 42)
	}
}

func TestSetAliasesInvalidValue(t *testing.T) {
	aliases := map[string]int{
		"badAlias1": -1,
		"badAlias2": 300,
	}
	delete(colorAliases, "badAlias1")
	delete(colorAliases, "badAlias2")
	// Expect error and no partial application
	if err := SetAliases(aliases); err == nil {
		t.Fatalf("expected error for invalid alias value, got none")
	}
}

// cpu: Apple M3 Max
// Name       		                   # Run 	     Avg Runtime       Bytes Allocated   # of Allocat
// BenchmarkParseStreaming-14    	   39277	     29422 ns/op	   27360 B/op	     491 allocs/o
func BenchmarkParseStreaming(b *testing.B) {

	testStr := "This is text"
	for i := 0; i < 5; i++ { // heavily nested tags
		testStr = testStr + "<ansi fg=black bg=\"white\">" + testStr + "</ansi>"
	}

	reader := strings.NewReader(testStr)
	input := bufio.NewReader(reader)
	var outputBuffer bytes.Buffer
	output := bufio.NewWriter(&outputBuffer)

	for n := 0; n < b.N; n++ {
		ParseStreaming(input, output)
		reader.Reset(testStr)
	}
}

// cpu: Apple M3 Max
// Name 	                   # Run   		 Avg Runtime       Bytes Allocated   # of Allocat
// BenchmarkParse-14    	   37892	     30006 ns/op	   33892 B/op	     497 allocs/op
func BenchmarkParse(b *testing.B) {

	testStr := "This is text"
	for i := 0; i < 5; i++ { // heavily nested tags
		testStr = testStr + "<ansi fg=black bg=\"white\">" + testStr + "</ansi>"
	}
	for n := 0; n < b.N; n++ {
		Parse(testStr)
	}
}

// cpu: Apple M3 Max
// Name 	                       # Run     	 Avg Runtime       Bytes Allocated   # of Allocat
// BenchmarkParseHTML-14    	   40748	     28125 ns/op	   32295 B/op	     455 allocs/op
func BenchmarkParseHTML(b *testing.B) {

	testStr := "This is text"
	for i := 0; i < 5; i++ { // heavily nested tags
		testStr = testStr + "<ansi fg=black bg=\"white\">" + testStr + "</ansi>"
	}
	for n := 0; n < b.N; n++ {
		Parse(testStr, HTML)
	}
}

// A numeric color past 255 has no entry in the color tables. It falls back
// to the default instead of indexing out of range.
func TestParseColorOutOfRange(t *testing.T) {

	testCases := map[string]struct {
		input string
		same  string
	}{
		"fg past the end":  {`<ansi fg="999">x</ansi>`, `<ansi>x</ansi>`},
		"fg one past":      {`<ansi fg="256">x</ansi>`, `<ansi>x</ansi>`},
		"bg past the end":  {`<ansi fg="9" bg="300">x</ansi>`, `<ansi fg="9">x</ansi>`},
		"last valid color": {`<ansi fg="255">x</ansi>`, `<ansi fg="255">x</ansi>`},
		"nested":           {`<ansi fg="9">a<ansi fg="999">b</ansi>c</ansi>`, `<ansi fg="9">a<ansi>b</ansi>c</ansi>`},
	}

	for name, testCase := range testCases {

		t.Run(name, func(t *testing.T) {

			for _, behaviors := range [][]ParseBehavior{nil, {HTML}, {Monochrome}} {
				assert.NotPanics(t, func() { Parse(testCase.input, behaviors...) })
				assert.Equal(t, Parse(testCase.same, behaviors...), Parse(testCase.input, behaviors...))

				var streamed bytes.Buffer
				assert.NotPanics(t, func() {
					ParseStreaming(bufio.NewReader(strings.NewReader(testCase.input)), bufio.NewWriter(&streamed), behaviors...)
				})
				assert.Equal(t, Parse(testCase.same, behaviors...), streamed.String())
			}
		})
	}

	assert.Contains(t, Parse(`<ansi fg="255">x</ansi>`, HTML), `color:`)
}

// An alias file entry outside 0-255 is skipped, so it can never reach the
// color tables.
func TestLoadAliasesSkipsOutOfRange(t *testing.T) {

	path := t.TempDir() + "/aliases.yaml"
	err := ioutil.WriteFile(path, []byte("colors:\n  rangetest-good: 9\n  rangetest-bad: 999\n  rangetest-negative: -4\n"), 0644)
	assert.NoError(t, err)
	assert.NoError(t, LoadAliases(path))

	aliases := GetAliases()
	assert.Equal(t, 9, aliases["rangetest-good"])
	assert.NotContains(t, aliases, "rangetest-bad")
	assert.NotContains(t, aliases, "rangetest-negative")

	assert.NotPanics(t, func() { Parse(`<ansi fg="rangetest-bad">x</ansi>`, HTML) })
}
