package main

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_splitPerfData(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want map[string]PerfData
	}{
		{
			name: "single",
			args: args{"test=1;2;3;4;5"},
			want: map[string]PerfData{
				"test": {Units: "", Value: 1, Min: 4, Max: 5},
			},
		},
		{
			name: "multiple",
			args: args{"test1=1;2;3;4;5 test2=11%;22;33;44;55"},
			want: map[string]PerfData{
				"test1": {Units: "", Value: 1, Min: 4, Max: 5},
				"test2": {Units: "%", Value: 11, Min: 44, Max: 55},
			},
		},
		{
			name: "spaces in label",
			args: args{"'test 1'=1;2;3;4;5 'test 2'=11%;22;33;44;55"},
			want: map[string]PerfData{
				"test 1": {Units: "", Value: 1, Min: 4, Max: 5},
				"test 2": {Units: "%", Value: 11, Min: 44, Max: 55},
			},
		},
		{
			name: "missing value/min/max",
			args: args{"'test'=;2;3;;"},
			want: map[string]PerfData{
				"test": {Units: "", Value: math.NaN(), Min: math.NaN(), Max: math.NaN()},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitPerfData(tt.args.line); !cmp.Equal(got, tt.want, cmpopts.EquateNaNs()) {
				t.Error(cmp.Diff(got, tt.want, cmpopts.EquateNaNs()))
			}
		})
	}
}

func Test_splitQuoted(t *testing.T) {
	type args struct {
		line string
		sep  rune
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "single item unquoted",
			args: args{line: "one", sep: ' '},
			want: []string{"one"},
		},
		{
			name: "single item quoted",
			args: args{line: "'one'", sep: ' '},
			want: []string{"'one'"},
		},
		{
			name: "no quotes",
			args: args{line: "one two", sep: ' '},
			want: []string{"one", "two"},
		},
		{
			name: "with quotes no spaces",
			args: args{line: "'one' 'two'", sep: ' '},
			want: []string{"'one'", "'two'"},
		},
		{
			name: "with quotes and spaces",
			args: args{line: "'o ne' 'tw o'", sep: ' '},
			want: []string{"'o ne'", "'tw o'"},
		},
		{
			name: "partial quotes and spaces",
			args: args{line: "'o ne'=111 'tw o'=222", sep: ' '},
			want: []string{"'o ne'=111", "'tw o'=222"},
		},
		{
			name: "leading separator",
			args: args{line: " one", sep: ' '},
			want: []string{"one"},
		},
		{
			name: "trailing separator",
			args: args{line: "one ", sep: ' '},
			want: []string{"one"},
		},
		{
			name: "repeated separator",
			args: args{line: "one   two", sep: ' '},
			want: []string{"one", "two"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitQuoted(tt.args.line, tt.args.sep); !cmp.Equal(got, tt.want) {
				t.Error(cmp.Diff(got, tt.want))
			}
		})
	}
}
