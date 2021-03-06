package fictex

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

var renderTests = []struct{
	Desc  string
	Input Node
	Text  string
	HTML  string
}{
	{
		Desc: "Basic test",
		Input: Node{
			Type: Text,
			Text: []byte("test"),
		},
		Text: "test",
		HTML: "test",
	},
	{
		Desc: "Escape test",
		Input: Node{
			Type: Text,
			Text: []byte("<"),
		},
		Text: "<",
		HTML: "&lt;",
	},
	{
		Desc: "Group",
		Input: Node{
			Type: Group,
			Child: []Node{{
				Type: Text,
				Text: []byte("a"),
			}, {
				Type: Text,
				Text: []byte("b"),
			}},
		},
		Text: "ab",
		HTML: "ab",
	},
	{
		Desc: "Wrapping",
		Input: Node{
			Type: Paragraph,
			Child: []Node{{
				Type: Bold,
				Text: []byte("a"),
			}, {
				Type: Slant,
				Text: []byte("b"),
			}, {
				Type: Underline,
				Text: []byte("c"),
			}},
		},
		Text: "\n    *a*/b/_c_\n",
		HTML: "<p>\n<b>a</b><i>b</i><u>c</u>\n</p>\n",
	},
	{
		Desc: "Dashes",
		Input: Node{
			Type: Group,
			Child: []Node{{
				Type: NDash,
			}, {
				Type: Text,
				Text: []byte(" "),
			}, {
				Type: MDash,
			}, {
				Type: HLine,
			}},
		},
		Text: "-- ---\n-----\n",
		HTML: "&#8211; &#8212;<hr />\n",
	},
	{
		Desc: "Preview",
		Input: Node{
			Type: Preview,
			Text: []byte("short"),
			Child: []Node{{
				Type: Paragraph,
				Child: []Node{{
					Type: Text,
					Text: []byte("long"),
				}},
			}},
		},
		Text: "\n<<short\n    long\n>>\n",
		HTML: "<!-- Fold: \"short\" -->\n<p>\nlong\n</p>\n<!-- /Fold -->\n",
	},
}

func TestRender(t *testing.T) {
	for _, test := range renderTests {
		desc := test.Desc
		b := new(bytes.Buffer)

		if err := TextRenderer.Render(b, test.Input); err != nil {
			t.Fatalf("%s: rendertext: %s", desc, err)
		}
		if got, want := b.String(), test.Text; got != want {
			t.Errorf("%s: rendertext = %q, want %q", desc, got, want)
		}

		b.Truncate(0)
		if err := HTMLRenderer.Render(b, test.Input); err != nil {
			t.Fatalf("%s: renderhtml: %s", desc, err)
		}
		if got, want := b.String(), test.HTML; got != want {
			t.Errorf("%s: renderhtml = %q, want %q", desc, got, want)
		}
	}
}

func BenchmarkRender(b *testing.B) {
	file, err := os.Open("testdata/lipsum.txt")
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	fic, err := Parse(bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	buf := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		buf.Truncate(0)
		HTMLRenderer.Render(buf, fic)
	}
}
