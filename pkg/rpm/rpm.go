package rpm

import (
	"regexp"
	"strings"

	"github.com/rmohr/bazel-dnf/pkg/api"
)

var (
	alpa = regexp.MustCompile("[a-zA-Z]*")
	num  = regexp.MustCompile("[0.9]*")
)

func Compare(a api.Version, b api.Version) int {

	return 0
}

func compare(a string, b string) int {

	// if a is empty and b is not, a is older
	if len(a) == 0 && len(b) != 0 {
		return -1
	}
	// if b is empty and a is not, a is newer
	if len(b) == 0 && len(a) != 0 {
		return 1
	}

	// if a starts with a ~ and b does not, a is older
	if a[0] == '~' && b[0] != '~' {
		return -1
	}
	// if b starts with a ~ and a does not, a is newer
	if a[0] != '~' && b[0] == '~' {
		return 1
	}

	for {
		var ca, cb int
		// take digits
		if a[ca] <= 9 {
		}

		if ca == len(a) {
			break
		}
		if cb == len(b) {
			break
		}
	}

	// the longest string wins from that point on
	if len(a) > len(b) {
		return 1
	} else if len(a) < len(b) {
		return -1
	}
	// the strings are completely equal
	return 0
}

type Tokenizer struct {
	text string
	end  bool
}

func (tk *Tokenizer) NextToken() *Token {
	if tk.end {
		return nil
	}

	if tk.text == "" {
		tk.end = true
		return &Token{}
	}

	if tk.text[0] == '.' {
		tk.text = tk.text[1:]
		return &Token{Type: SepToken}
	}

	token := &Token{}
	end := 0
	for idx, t := range tk.text {
		if token.Type == "" {
			if isNum(t) {
				token.Type = NumToken
			} else {
				token.Type = AlphaToken
			}
		}
		if t == '.' {
			end = idx-1
			break
		} else if isNum(t) && token.Type == NumToken {
			token.Text = token.Text + string(t)
		} else if !isNum(t) && token.Type == AlphaToken {
			token.Text = token.Text + string(t)
		} else {
			break
		}
		end = idx
	}
	tk.text = tk.text[end+1:]

	// remove leading zeros
	start := 0
	for idx, t := range token.Text {
		start = idx
		if t != '0' {
			break
		}
	}
	token.Text = token.Text[start:]
	return token
}

func isNum(b int32) bool {
	if '0' <= b && b <= '9' {
		return true
	}
	return false
}

type Token struct {
	Text string
	Type TokenType
}

func (a *Token) Compare(b *Token) int {
	if a != nil && b == nil {
		return 1
	} else if a == nil && b != nil {
		return -1
	} else if a == nil && b == nil {
		return 0
	}

	if a.Type == "" && b.Type == "" {
		return 0
	}

	if a.Type != SepToken && b.Type == SepToken {
		return 1
	} else if a.Type == SepToken && b.Type != SepToken {
		return -1
	}

	if a.Type != "" && b.Type == "" {
		return 1
	} else if a.Type == "" && b.Type != "" {
		return -1
	}

	if a.Type == NumToken && b.Type == AlphaToken {
		return 1
	} else if a.Type == AlphaToken && b.Type == NumToken {
		return -1
	}
	return strings.Compare(a.Text, b.Text)
}

type TokenType string

const (
	NumToken   TokenType = "num"
	AlphaToken TokenType = "alpha"
	SepToken TokenType = "sep"
)
