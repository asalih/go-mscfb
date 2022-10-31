package mscfb

import (
	"reflect"
	"testing"
)

func TestNameChainFromPath(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty",
			args: args{s: ""},
			want: []string{"."},
		},
		{
			name: "valid abs",
			args: args{s: "/foo/bar/baz/"},
			want: []string{"foo", "bar", "baz"},
		},
		{
			name: "valid rel",
			args: args{s: "foo/bar/baz"},
			want: []string{"foo", "bar", "baz"},
		},
		{
			name: "valid up",
			args: args{s: "foo/bar/../baz"},
			want: []string{"foo", "baz"},
		},
		{
			name: "invalid up",
			args: args{s: "foo/../../baz"},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NameChainFromPath(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NameChainFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathFromNameChain(t *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{names: []string{}},
			want: "/",
		},
		{
			name: "valid",
			args: args{names: []string{"foo", "bar", "baz"}},
			want: "/foo/bar/baz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PathFromNameChain(tt.args.names); got != tt.want {
				t.Errorf("PathFromNameChain() = %v, want %v", got, tt.want)
			}
		})
	}
}
