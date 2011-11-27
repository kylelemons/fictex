package fictex

import (
	"testing"
	"strings"
	"reflect"
)

var parseTests = []struct {
	Desc   string
	Input  string
	Output Node
}{
	{
		Desc:   "Zero Test",
		Input:  "",
		Output: Node{},
	},
	{
		Desc:   "Whitespace Test",
		Input:  "  \t\n",
		Output: Node{},
	},
	{
		Desc:  "Basic Test",
		Input: "text",
		Output: Node{
			Type: Group,
			Child: []Node{{
				Type: Paragraph,
				Child: []Node{{
					Type: Text,
					Text: []byte("text"),
				}},
			}},
		},
	},
	{
		Desc:  "Formatted Test",
		Input: "/slant/ *bold* _underline_",
		Output: Node{
			Type: Group,
			Child: []Node{{
				Type: Paragraph,
				Child: []Node{{
					Type: Slant,
					Text: []byte("slant"),
				}, {
					Type: Text,
					Text: []byte(" "),
				}, {
					Type: Bold,
					Text: []byte("bold"),
				}, {
					Type: Text,
					Text: []byte(" "),
				}, {
					Type: Underline,
					Text: []byte("underline"),
				}},
			}},
		},
	},
	{
		Desc:  "Unformatted Test",
		Input: "/slant/*bold*_underline_",
		Output: Node{
			Type: Group,
			Child: []Node{{
				Type: Paragraph,
				Child: []Node{{
					Type: Text,
					Text: []byte("/slant/*bold*_underline_"),
				}},
			}},
		},
	},
	{
		Desc:  "Inline Slashes",
		Input: "If a/b, /a a/b /b c/ d",
		Output: Node{
			Type: Group,
			Child: []Node{{
				Type: Paragraph,
				Child: []Node{{
					Type: Text,
					Text: []byte("If a/b, "),
				}, {
					Type: Slant,
					Text: []byte("a a/b /b c"),
				}, {
					Type: Text,
					Text: []byte(" d"),
				}},
			}},
		},
	},
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		desc := test.Desc
		in := strings.NewReader(test.Input)
		out, err := Parse(in)
		if err != nil {
			t.Fatalf("%s: parse: %s", desc, err)
		}
		if !reflect.DeepEqual(out, test.Output) {
			t.Errorf("%s: Parse tree mismatch:", desc)
			t.Logf("Got:\n%s", out)
			t.Logf("Want:\n%s", test.Output)
		}
	}
}
