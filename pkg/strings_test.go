package pkg

import (
	"reflect"
	"testing"
)

func TestChunk(t *testing.T) {
	type args struct {
		input  string
		length int
		sep    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty input returns empty result",
			args: args{input: "", length: 42, sep: ""},
			want: "",
		},
		{
			name: "chunk length of 0 returns untouched input",
			args: args{input: "foo", length: 0, sep: ""},
			want: "foo",
		},
		{
			name: "negative chunk length returns untouched input",
			args: args{input: "foo", length: -42, sep: ""},
			want: "foo",
		},
		{
			name: "empty set returns untouched input",
			args: args{input: "foobar", length: 3, sep: ""},
			want: "foobar",
		},
		{
			name: "input is returned untouched when chunk length is greater than the input",
			args: args{input: "foo", length: 5, sep: ""},
			want: "foo",
		},
		{
			name: "input is returned untouched when chunk length is equal the input",
			args: args{input: "foo", length: 3, sep: ""},
			want: "foo",
		},
		{
			name: "input is returned untouched if sep length equals chunk length",
			args: args{input: "foo", length: 1, sep: "-"},
			want: "foo",
		},
		{
			name: "input is returned untouched if sep length is greater than chunk length",
			args: args{input: "foo", length: 2, sep: "---"},
			want: "foo",
		},
		{
			name: "input is chunked with 2 chars sep",
			args: args{input: "foo bar baz", length: 4, sep: " -"},
			want: "foo  -bar  -baz",
		},
		{
			name: "input is uft-8",
			args: args{input: "形声字", length: 6, sep: "$"},
			want: "形声$字",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Chunk(tt.args.input, tt.args.length, tt.args.sep); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Chunk() = %#v, want %#v", string(got), string(tt.want))
			}
		})
	}
}
