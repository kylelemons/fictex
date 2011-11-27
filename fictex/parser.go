package fictex

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

type nodeType int
const (
	Group nodeType = iota  // A group of Child nodes (no Text)
	Text                   // A single node with Text filled in
	Paragraph              // A sequence of Child nodes (no Text) in a paragraph
	Bold
	Slant
	Underline
	MDash                  // No Text or Child
	NDash                  // No Text or Child
	HLine                  // No Text or Child
	Preview                // Text is the preview, Children are the full
)
var typeString = [...]string{
	"Group", "Text", "Paragraph", "Bold", "Slant",
	"Underline", "M-Dash", "N-Dash", "Separator", "Preview",
}

func (t nodeType) String() string {
	return typeString[t]
}

type Node struct {
	Type  nodeType
	Text  []byte
	Child []Node
}

func (n Node) String() string {
	var b bytes.Buffer
	n.str(&b, 0)
	return b.String()
}

func (n Node) str(w io.Writer, depth int) {
	indent := strings.Repeat("| ", depth)
	fmt.Fprintf(w, "%s+ %s:\n", indent, n.Type)
	if len(n.Text) > 0 {
		fmt.Fprintf(w, "%s| + %q\n", indent, n.Text)
	}
	for _, c := range n.Child {
		c.str(w, depth+1)
	}
}

func Parse(r io.Reader) (Node, os.Error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	p := parser{br}

	var (
		n, m Node
		err  os.Error
	)

	for err == nil {
		m, err = p.next()
		n.Child = append(n.Child, m.Child...)
	}
	if err == os.EOF {
		err = nil
	}
	return n, err
}

type parser struct {
	*bufio.Reader
}

func (p *parser) next() (Node, os.Error) {
	// Create a new Group node
	n := Node{}

	for {
		c, err := p.ReadByte()
		if err != nil {
			return n, err
		}

		var next Node
		switch c {
			case '-':
				// The first - will have aready been read
				next, err = p.readDash()
			case '\n', '\t', ' ':
				continue // slurp whitespace
			case '<':
				// The first < will have already been read
				next, err = p.readPreview()
			default:
				p.UnreadByte()
				next, err = p.readParagraph()
		}

		switch next.Type {
			case Group: n.Child = append(n.Child, next.Child...)
			default:    n.Child = append(n.Child, next)
		}

		if err != nil {
			return n, err
		}
	}
	panic("unreachable")
}

func (p *parser) readDash() (Node, os.Error) {
	return Node{}, Unimplemented("readDash")
}

func (p *parser) readPreview() (Node, os.Error) {
	return Node{}, Unimplemented("readPreview")
}

func (p *parser) readParagraph() (Node, os.Error) {
	n := Node{Type: Paragraph}

	bow := true // beginning-of-word

more:
	for {
		c, err := p.ReadByte()
		if err != nil {
			return n, err
		}

		var next Node
		switch c {
			case '\n':
				break more
			case '-':
				next, err = p.readDash()
			case '/':
				if !bow {
					goto plain
				}
				next, err = p.readText(Slant, c)
			case '*':
				if !bow {
					goto plain
				}
				next, err = p.readText(Bold, c)
			case '_':
				if !bow {
					goto plain
				}
				next, err = p.readText(Underline, c)
			default:
				goto plain
		}
		goto push

	plain:
		p.UnreadByte()
		next, err = p.readText(Text, '\n')

	push:
		if next.Type == Text {
			l := len(next.Text)
			bow = l > 0 && next.Text[l-1] == ' '

			l = len(n.Child)
			if l > 0 && n.Child[l-1].Type == Text {
				n.Child[l-1].Text = append(n.Child[l-1].Text, next.Text...)
			} else {
				n.Child = append(n.Child, next)
			}
		} else {
			bow = false
			n.Child = append(n.Child, next)
		}

		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (p *parser) readText(t nodeType, end byte) (Node, os.Error) {
	n := Node{Type: Text}

more:
	for {
		c, err := p.ReadByte()
		if err != nil {
			return n, err
		}

		if len(n.Text) == 0 {
			n.Text = append(n.Text, c)
			continue
		}

		switch c {
			case end:
				next, err := p.Peek(1)
				if err != nil || next[0] == ' ' {
					n.Type = t
					break more
				}
				n.Text = append(n.Text, c)
			case '/', '_', '*':
				p.UnreadByte()
				if t != Text {
					// TODO(kevlar): reuse Text if it has capacity
					n.Text = append([]byte{end}, n.Text...)
				}
				break more
			case '\n':
				break more
			default:
				n.Text = append(n.Text, c)
		}
	}

	return n, nil
}

type Unimplemented string
func (e Unimplemented) String() string { return string(e) + " unimplemented" }
