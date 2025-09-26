package rpm

import (
	"cmp"
	"strconv"
	"strings"

	"github.com/rmohr/bazeldnf/pkg/api"
)

func ComparePackage(a *api.Package, b *api.Package) int {
	if a.Repository.Priority == b.Repository.Priority {
		return Compare(a.Version, b.Version)
	}

	return b.Repository.Priority - a.Repository.Priority
}

func ComparePackageKey(a, b api.PackageKey) int {
	return cmp.Or(
		cmp.Compare(a.Name, b.Name), Compare(a.Version, b.Version))
}

func Compare(a api.Version, b api.Version) int {
	return cmp.Or(
		compare(a.Epoch, b.Epoch),
		compare(a.Ver, b.Ver),
		compare(a.Rel, b.Rel),
	)
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

	if len(a) == 0 && len(b) == 0 {
		return 0
	}

	// if a starts with a ~ and b does not, a is older
	if a[0] == '~' && b[0] != '~' {
		return -1
	}
	// if b starts with a ~ and a does not, a is newer
	if a[0] != '~' && b[0] == '~' {
		return 1
	}

	if a[0] == '~' && b[0] == '~' {
		a = a[1:]
		b = b[1:]
	}

	toA := Tokenizer{text: a}
	toB := Tokenizer{text: b}
	for {
		ta := toA.NextToken()
		tb := toB.NextToken()
		res := ta.Compare(tb)

		if res != 0 || ta.Type == "" || tb.Type == "" {
			return res
		}
	}
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
			end = idx - 1
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
	if a.Type == "" && b.Type == "" {
		return 0
	}

	if a.Type != "" && b.Type == "" {
		return 1
	} else if a.Type == "" && b.Type != "" {
		return -1
	}

	if a.Type != SepToken && b.Type == SepToken {
		return 1
	} else if a.Type == SepToken && b.Type != SepToken {
		return -1
	}

	if a.Type == NumToken && b.Type == AlphaToken {
		return 1
	} else if a.Type == AlphaToken && b.Type == NumToken {
		return -1
	}

	if a.Type == NumToken && b.Type == NumToken {
		aInt, _ := strconv.Atoi(a.Text)
		bInt, _ := strconv.Atoi(b.Text)
		if aInt == bInt {
			return 0
		} else if aInt > bInt {
			return 1
		} else {
			return -1
		}
	} else {
		return strings.Compare(a.Text, b.Text)
	}
}

type TokenType string

const (
	NumToken   TokenType = "num"
	AlphaToken TokenType = "alpha"
	SepToken   TokenType = "sep"
)
