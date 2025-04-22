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

			// fmt.Println(output)
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

			// fmt.Println(output)
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

			// fmt.Println(output)
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

			// fmt.Println(output)
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
	delete(colorMap256, alias)
	// Set alias in default (color256) group
	if err := SetAlias(alias, value); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := colorMap256[alias]; got != value {
		t.Errorf("colorMap256[%q] = %d; want %d", alias, got, value)
	}
}

func TestSetAliasValidColor8Group(t *testing.T) {
	alias := "testAlias8"
	value := 45
	delete(colorMap8, alias)
	// Set alias in color8 group
	if err := SetAlias(alias, value, "color8"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := colorMap8[alias]; got != value {
		t.Errorf("colorMap8[%q] = %d; want %d", alias, got, value)
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
	delete(colorMap256, "multi256_1")
	delete(colorMap256, "multi256_2")
	// Bulk set in default (color256) group
	if err := SetAliases(aliases); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	for alias, expected := range aliases {
		if got := colorMap256[alias]; got != expected {
			t.Errorf("colorMap256[%q] = %d; want %d", alias, got, expected)
		}
	}
}

func TestSetAliasesValidColor8Group(t *testing.T) {
	aliases := map[string]int{
		"multi8_1": 10,
		"multi8_2": 20,
	}
	delete(colorMap8, "multi8_1")
	delete(colorMap8, "multi8_2")
	// Bulk set in color8 group
	if err := SetAliases(aliases, "color8"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	for alias, expected := range aliases {
		if got := colorMap8[alias]; got != expected {
			t.Errorf("colorMap8[%q] = %d; want %d", alias, got, expected)
		}
	}
}

func TestSetAliasesInvalidValue(t *testing.T) {
	aliases := map[string]int{
		"badAlias1": -1,
		"badAlias2": 300,
	}
	delete(colorMap256, "badAlias1")
	delete(colorMap256, "badAlias2")
	// Expect error and no partial application
	if err := SetAliases(aliases); err == nil {
		t.Fatalf("expected error for invalid alias value, got none")
	}
}

//
// Benchmarks
// cpu: Intel(R) Core(TM) i9-9880H CPU @ 2.30GHz
//
/*
//
// BenchmarkSprintf
// BenchmarkSprintf-16               500000               114.7 ns/op             8 B/op          1 allocs/op
//
func BenchmarkSprintf(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = fmt.Sprintf("\033[%d;%dm", 8+60, 7)
	}
}

//
// BenchmarkAtoI
// BenchmarkAtoI-16                  500000                34.44 ns/op            0 B/op          0 allocs/op
//
func BenchmarkAtoI(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = "\033[" + strconv.Itoa(8+60) + ";" + strconv.Itoa(7) + "m"
	}
}

//
// BenchmarkConcat
// BenchmarkConcat-16                500000            168857 ns/op         1403048 B/op          2 allocs/op
//

func BenchmarkConcat(b *testing.B) {
	var sConcat string = ""
	for n := 0; n < b.N; n++ {
		sConcat += strconv.Itoa(n)
	}
}

//
// BenchmarkStringBuilder
// BenchmarkStringBuilder-16         500000                38.46 ns/op           41 B/op          0 allocs/op
//
func BenchmarkStringBuilder(b *testing.B) {
	var sBuilder strings.Builder
	for n := 0; n < b.N; n++ {
		sBuilder.WriteString(strconv.Itoa(n))
	}
}

// BenchmarkParseColorString
// BenchmarkParseColorString-16    	  617618	      1665 ns/op	     808 B/op	      16 allocs/op
func BenchmarkParseColorString(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Parse("This is a prefix <ansi fg=blue bg=red>This is color</ansi> This is a suffix")
	}
}

// BenchmarkParseColorInt
// BenchmarkParseColorInt-16    	  880810	      1393 ns/op	     712 B/op	      14 allocs/op
func BenchmarkParseColorInt(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Parse("This is a prefix <ansi fg='9' bg='2'>This is color</ansi> This is a suffix")
	}
}
*/
// Name                       # Run      Avg Runtime       Bytes Allocated   # of Allocat
// BenchmarkParseStreaming-16 34398	     32720 ns/op	   14958 B/op	     297 allocs/op
func BenchmarkParseStreaming(b *testing.B) {

	testStr := "This is text"
	for i := 0; i < 20; i++ {
		testStr = "<ansi fg=black bg=\"white\">" + testStr + "</ansi>"
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

// BenchmarkParseNew-16    	   33194	     35641 ns/op	   23330 B/op	     304 allocs/op
func BenchmarkParse(b *testing.B) {

	testStr := "This is text"
	for i := 0; i < 20; i++ {
		testStr = "<ansi fg=black bg=\"white\">" + testStr + "</ansi>"
	}
	for n := 0; n < b.N; n++ {
		Parse(testStr)
	}
}
