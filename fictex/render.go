package fictex

import (
	"fmt"
	"html"
	"io"
	"os"
)

type StringPair [2]string

type Renderer struct {
	// The Escape function is used to escape plaintext
	Escape func(string) string

	// The following are used to bracket the appropriate text
	Bold StringPair
	Slant StringPair
	Underline StringPair
	Paragraph StringPair

	// The following are used in place of the corresponding node
	NDash string
	MDash string
	HLine string

	// The first of the pair will be formatted with Sprintf(fmt, preview)
	Preview StringPair
}

var TextRenderer = Renderer{
	Bold: StringPair{"*", "*"},
	Slant: StringPair{"/", "/"},
	Underline: StringPair{"_", "_"},
	Paragraph: StringPair{"\n    ", "\n"},

	NDash: "--",
	MDash: "---",
	HLine: "\n-----\n",

	Preview: StringPair{"\n<<%s", ">>\n"},
}

var HTMLRenderer = Renderer{
	Escape: html.EscapeString,

	Bold: StringPair{"<b>", "</b>"},
	Slant: StringPair{"<i>", "</i>"},
	Underline: StringPair{"<u>", "</u>"},
	Paragraph: StringPair{"<p>\n", "\n</p>\n"},

	NDash: "&#8211;", //"&ndash;",
	MDash: "&#8212;", //"&mdash;",
	HLine: "<hr />\n",

	Preview: StringPair{"<!-- Fold: %q -->\n", "<!-- /Fold -->\n"},
}

func (r Renderer) Render(w io.Writer, n Node) os.Error {
	esc := func(b []byte) string {
		if r.Escape == nil {
			return string(b)
		}
		return r.Escape(string(b))
	}

	var render func(Node) os.Error
	render = func(n Node) (err os.Error) {
		switch n.Type {
			case Group:
				for _, n := range n.Child {
					if err := render(n); err != nil {
						return err
					}
				}
			case Text:
				_, err = io.WriteString(w, esc(n.Text))
			case Bold:
				_, err = fmt.Fprintf(w, "%s%s%s", r.Bold[0], esc(n.Text), r.Bold[1])
			case Slant:
				_, err = fmt.Fprintf(w, "%s%s%s", r.Slant[0], esc(n.Text), r.Slant[1])
			case Underline:
				_, err = fmt.Fprintf(w, "%s%s%s", r.Underline[0], esc(n.Text), r.Underline[1])
			case Paragraph:
				if _, err := io.WriteString(w, r.Paragraph[0]); err != nil {
					return err
				}
				for _, n := range n.Child {
					render(n)
				}
				_, err = io.WriteString(w, r.Paragraph[1])
			case NDash:
				_, err = io.WriteString(w, r.NDash)
			case MDash:
				_, err = io.WriteString(w, r.MDash)
			case HLine:
				_, err = io.WriteString(w, r.HLine)
			case Preview:
				if _, err := fmt.Fprintf(w, r.Preview[0], esc(n.Text)); err != nil {
					return err
				}
				for _, n := range n.Child {
					if err := render(n); err != nil {
						return err
					}
				}
				_, err = io.WriteString(w, r.Preview[1])
			default:
				_, err = fmt.Fprintf(w, "Unhandled %T\n", n)
		}
		return err
	}
	return render(n)
}
