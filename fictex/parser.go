package fictex

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// TODO(kevlar): Make this support UTF-8 (e.g. ReadRune instead of ReadByte)

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

func ParseBytes(b []byte) (Node, os.Error) {
	return Parse(bytes.NewBuffer(b))
}

func ParseString(s string) (Node, os.Error) {
	return Parse(strings.NewReader(s))
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
		m, err = p.top()
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

func (p *parser) top() (Node, os.Error) {
	// Create a new Group node
	n := Node{}

	for {
		c, err := p.ReadByte()
		if err != nil {
			return n, err
		}

		var next Node
		switch c {
			case '\n', '\t', ' ':
				continue // slurp whitespace
			case '<':
				next, err = p.readPreview()
			default:
				p.UnreadByte()
				next, err = p.readParagraph(false)
		}

		n.Child = append(n.Child, next)

		if err != nil {
			return n, err
		}
	}
	panic("unreachable")
}

// readPreview reads a preview like:
//   <Preview text goes here
//   full text goes here
//   >
// The first < must have already been read.
func (p *parser) readPreview() (Node, os.Error) {
	n := Node{Type: Preview}

preview:
	line, err := p.ReadSlice('\n')

	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	n.Text = append(n.Text, line...)

	if err == bufio.ErrBufferFull {
		goto preview
	}
	if err != nil {
		return n, err
	}

more:
	for {
		c, err := p.ReadByte()
		if err != nil {
			return n, err
		}

		var next Node
		switch c {
			case '>':
				break more
			case '\n', '\t', ' ':
				continue // slurp whitespace
			case '<':
				next, err = p.readPreview() // sub preview
			default:
				p.UnreadByte()
				next, err = p.readParagraph(true)
		}

		n.Child = append(n.Child, next)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (p *parser) readParagraph(preview bool) (Node, os.Error) {
	n := Node{Type: Paragraph}

	// Check for dashes
	c, err := p.ReadByte()
	if err != nil {
		return n, err
	}

	if c == '-' {
		node, err := p.readDash()
		if err != nil {
			return node, err
		}
		if node.Type == HLine {
			return node, nil
		}
		n.Child = append(n.Child, node)
	} else {
		p.UnreadByte()
	}

	var last *Node
	for {
		c, err := p.ReadByte()
		if err != nil {
			break
		}

		// End a preview with a > on its own line
		if preview && c == '>' {
			p.UnreadByte()
			break
		}

		// End the paragraph with \n on its own line
		if c == '\n' {
			break
		}

		p.UnreadByte()

	more:
		next, err := p.readText()

		eol := false
		if length := len(next.Text); length > 0 && next.Text[length-1] == '\n' {
			//next.Text = next.Text[:len(next.Text)-1]
			eol = true
		}

		if next.Type != Text || len(next.Text) > 0 {
			if last != nil && next.Type == last.Type {
				last.Text = append(last.Text, next.Text...)
				last.Child = append(last.Child, next.Child...)
			} else {
				n.Child = append(n.Child, next)
				last = &n.Child[len(n.Child)-1]
			}
		}

		if err != nil {
			break
		}
		if eol {
			continue
		}

		goto more
	}

	for i := range n.Child {
		n.Child[i].Text = bytes.TrimRight(n.Child[i].Text, "\n")
		n.Child[i].Text = bytes.Replace(n.Child[i].Text, []byte{'\n'}, []byte{' '}, -1)
	}

	return n, nil
}

// readText reads a "normal" piece of text:
//   - Formatted if it starts with / * or _
//   - As a dash if it starts with -
//   - Up to the next dash, space, or newline otherwise
func (p *parser) readText() (Node, os.Error) {
	n := Node{Type: Text}

	// The first character determines what kind of text this is
	start, err := p.ReadByte()
	if err != nil {
		return n, err
	}

	switch start {
		case '-':
			return p.readDash()
		case '/', '*', '_':
			p.UnreadByte()
			return p.readFormatted()
		default:
			p.UnreadByte()
			start = 0
	}

	for {
		c, err := p.ReadByte()
		if err != nil {
			break
		}
		if c == '-' {
			p.UnreadByte()
			break
		}
		n.Text = append(n.Text, c)
		if c == '\n' || c == ' ' {
			break
		}
	}

	return n, nil
}

func (p *parser) readFormatted() (Node, os.Error) {
	n := Node{Type: Text}

	start, err := p.ReadByte()
	if err != nil {
		return n, err
	}

	switch start {
		case '*': n.Type = Bold
		case '/': n.Type = Slant
		case '_': n.Type = Underline
		default:
			// Shouldn't happen, but...
			start = 0
			p.UnreadByte()
	}

	for {
		pair, err := p.Peek(2)
		if len(pair) == 0 {
			return n, err
		}
		p.ReadByte()
		if pair[0] == start {
			if len(pair) == 1 {
				break
			}
			if r := int(pair[1]); unicode.IsSpace(r) || unicode.IsPunct(r) {
				break
			}
		}
		if pair[0] == '\n' {
			break
		}
		n.Text = append(n.Text, pair[0])
	}

	if len(n.Text) == 0 {
		n.Type = Text
		n.Text = append(n.Text, start)
	}

	return n, nil
}

func (p *parser) readDash() (Node, os.Error) {
	cnt := 1

	for {
		c, err := p.ReadByte()
		if err != nil {
			break
		}

		if c != '-' {
			p.UnreadByte()
			break
		}

		cnt++
	}

	switch cnt {
		case 1:
			return Node{Type: Text, Text: []byte{'-'}}, nil
		case 2:
			return Node{Type: NDash}, nil
		case 3:
			return Node{Type: MDash}, nil
	}
	return Node{Type: HLine}, nil
}

type Unimplemented string
func (e Unimplemented) String() string { return string(e) + " unimplemented" }
