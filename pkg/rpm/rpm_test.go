package rpm

import (
	"reflect"
	"testing"
)

func TestTokenizer_NextToken(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []Token
	}{
		{name: "tokenizing an empty string", text: "", want: []Token{{}}},
		{name: "tokenizing an empty string", text: ".", want: []Token{{Type: SepToken}, {}}},
		{name: "tokenizing a normal semver", text: "1.2.3", want: []Token{{Text: "1", Type: NumToken}, {Type: SepToken}, {Text: "2", Type: NumToken}, {Type: SepToken}, {Text: "3", Type: NumToken}, {}}},
		{name: "tokenizing a semver with no major number", text: ".22.3", want: []Token{{Type: SepToken}, {Text: "22", Type: NumToken}, {Type: SepToken}, {Text: "3", Type: NumToken}, {}}},
		{name: "tokenizing a semver with no minor number", text: "2..3", want: []Token{{Text: "2", Type: NumToken}, {Type: SepToken}, {Type: SepToken}, {Text: "3", Type: NumToken}, {}}},
		{name: "tokenizing a semver with leading alpha byte", text: "a2.2.3", want: []Token{{Text: "a", Type: AlphaToken}, {Text: "2", Type: NumToken}, {Type: SepToken}, {Text: "2", Type: NumToken}, {Type: SepToken}, {Text: "3", Type: NumToken}, {}}},
		{name: "tokenizing a semver with not leading alpha byte", text: "2a.2.3", want: []Token{{Text: "2", Type: NumToken}, {Text: "a", Type: AlphaToken}, {Type: SepToken}, {Text: "2", Type: NumToken}, {Type: SepToken}, {Text: "3", Type: NumToken}, {}}},
		{name: "tokenizing a semver with leading alpha byte and alphas in minor and patch", text: "a2.eb.r", want: []Token{{Text: "a", Type: AlphaToken}, {Text: "2", Type: NumToken}, {Type: SepToken}, {Text: "eb", Type: AlphaToken}, {Type: SepToken}, {Text: "r", Type: AlphaToken}, {}}},
		{name: "tokenizing a normal semver with leading zeroes", text: "a001.02.0", want: []Token{{Text: "a", Type: AlphaToken}, {Text: "1", Type: NumToken}, {Type: SepToken}, {Text: "2", Type: NumToken}, {Type: SepToken}, {Text: "0", Type: NumToken}, {}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := &Tokenizer{
				text: tt.text,
			}
			got := []Token{}
			for {
				t := tk.NextToken()
				if t == nil {
					break
				}
				got = append(got, *t)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NextToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
