package rpm

import (
	"reflect"
	"testing"

	"github.com/rmohr/bazel-dnf/pkg/api"
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

func Test_compare(t *testing.T) {
	type args struct {
		a string
		b string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "", args: args{"", ""}, want: 0},
		{name: "", args: args{".", "."}, want: 0},
		{name: "", args: args{"..", ".."}, want: 0},
		{name: "", args: args{"..", "..."}, want: -1},
		{name: "", args: args{"...", ".."}, want: 1},
		{name: "", args: args{"..1", ".1."}, want: -1},
		{name: "", args: args{".1.", "..1"}, want: 1},
		{name: "", args: args{"001.1", "1.1"}, want: 0},
		{name: "", args: args{"001.01", "000001.1"}, want: 0},
		{name: "", args: args{"001.02", "000001.1"}, want: 1},
		{name: "", args: args{"001.01", "000001.2"}, want: -1},
		{name: "", args: args{"~1", "~1"}, want: 0},
		{name: "", args: args{"1", "~1"}, want: 1},
		{name: "", args: args{"~1", "1"}, want: -1},
		{name: "", args: args{"1", "1"}, want: 0},
		{name: "", args: args{"1", "2"}, want: -1},
		{name: "", args: args{"2", "1"}, want: 1},
		{name: "", args: args{"a1", "a1"}, want: 0},
		{name: "", args: args{"a1", "1"}, want: -1},
		{name: "", args: args{"1", "a1"}, want: 1},
		{name: "", args: args{"1.2", "1.2"}, want: 0},
		{name: "", args: args{"1.2", "1.3"}, want: -1},
		{name: "", args: args{"1.3", "1.2"}, want: 1},
		{name: "", args: args{"1.2.3", "1.2.3"}, want: 0},
		{name: "", args: args{"1.2.2", "1.2.3"}, want: -1},
		{name: "", args: args{"1.2.3", "1.2.2"}, want: 1},
		{name: "", args: args{"1.a1", "1.a1"}, want: 0},
		{name: "", args: args{"1.a1", "1.1"}, want: -1},
		{name: "", args: args{"1.1", "1.a1"}, want: 1},
		{name: "", args: args{"1.1", "1.1.2"}, want: -1},
		{name: "", args: args{"1.1.2", "1.1"}, want: 1},
		{name: "", args: args{"1.a", "1.a1"}, want: -1},
		{name: "", args: args{"1.a1", "1.a"}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compare(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	type args struct {
		a api.Version
		b api.Version
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Epoch equal",
			args: args{
				a: api.Version{Epoch: "1"},
				b: api.Version{Epoch: "1"},
			},
			want: 0,
		},
		{
			name: "Epoch less",
			args: args{
				a: api.Version{Epoch: "1"},
				b: api.Version{Epoch: "2"},
			},
			want: -1,
		},
		{
			name: "Epoch more",
			args: args{
				a: api.Version{Epoch: "2"},
				b: api.Version{Epoch: "1"},
			},
			want: 1,
		},
		{
			name: "Version equal",
			args: args{
				a: api.Version{Epoch: "1", Ver: "1"},
				b: api.Version{Epoch: "1", Ver: "1"},
			},
			want: 0,
		},
		{
			name: "Version less",
			args: args{
				a: api.Version{Epoch: "1", Ver: "1"},
				b: api.Version{Epoch: "1", Ver: "2"},
			},
			want: -1,
		},
		{
			name: "Version more",
			args: args{
				a: api.Version{Epoch: "1", Ver: "2"},
				b: api.Version{Epoch: "1", Ver: "1"},
			},
			want: 1,
		},
		{
			name: "Release equal",
			args: args{
				a: api.Version{Epoch: "1", Ver: "1", Rel: "1"},
				b: api.Version{Epoch: "1", Ver: "1", Rel: "1"},
			},
			want: 0,
		},
		{
			name: "Release less",
			args: args{
				a: api.Version{Epoch: "1", Ver: "1", Rel: "1"},
				b: api.Version{Epoch: "1", Ver: "1", Rel: "2"},
			},
			want: -1,
		},
		{
			name: "Release more",
			args: args{
				a: api.Version{Epoch: "1", Ver: "1", Rel: "2"},
				b: api.Version{Epoch: "1", Ver: "1", Rel: "1"},
			},
			want: 1,
		},
		{
			name: "Version more in subversion",
			args: args{
				a: api.Version{Epoch: "0", Ver: "3.14", Rel: "2.fc32"},
				b: api.Version{Epoch: "0", Ver: "3", Rel: ""},
			},
			want: 1,
		},
		{
			name: "Release more",
			args: args{
				a: api.Version{Epoch: "0", Ver: "4.16.0", Rel: ""},
				b: api.Version{Epoch: "0", Ver: "4.3", Rel: ""},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Compare(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}
