package main

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"testing"
)

type mockReader struct {
	reads int
	lines []string
}

func newMockReader(s string) *mockReader {
	var m mockReader
	m.lines = strings.Split(s, "\n")
	return &m
}

func (mr *mockReader) Read(p []byte) (int, error) {
	if mr.reads < len(mr.lines) {
		i := copy(p, []byte(mr.lines[mr.reads]))
		mr.reads++
		return i, nil
	}
	return 0, errors.New("too many reads")
}

func TestPrompt(t *testing.T) {
	tt := []struct {
		prompt string
		want   string
		output string
		input  string
	}{
		{"choose (A) or (B) ", "B", "choose (A) or (B) ", "B\n"},
		{"choose (A) or (B) ", "A", "choose (A) or (B) \"D\" invalid choice, try again\nchoose (A) or (B) ", "D\nA\n"},
	}

	for i, tc := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r := bytes.NewBufferString(tc.input)
			var w bytes.Buffer
			got, _ := prompt(r, &w, tc.prompt)
			if got != tc.want {
				t.Fatal("got", got, "want", tc.want)
			}
			if tc.output != w.String() {
				t.Fatalf("got %q want %q", w.String(), tc.output)
			}

		})
	}
}

func TestInputRowsCols(t *testing.T) {
	tt := []struct {
		input            string
		wantOutput       string
		wantRow, wantCol int
		wantErr          bool
	}{
		{
			input:      "34 6\nC\n",
			wantOutput: "Enter dimensions row count and column count separated by a space. You entered 34 rows and 6 columns.\nContinue (C), Retry (R) Cancel (X)? ",
			wantRow:    34,
			wantCol:    6,
			wantErr:    false,
		},
		{
			input:      "1 2\nR\n4 5\nC\n",
			wantOutput: "Enter dimensions row count and column count separated by a space. You entered 1 rows and 2 columns.\nContinue (C), Retry (R) Cancel (X)? Enter dimensions row count and column count separated by a space. You entered 4 rows and 5 columns.\nContinue (C), Retry (R) Cancel (X)? ",
			wantRow:    4,
			wantCol:    5,
			wantErr:    false,
		},
		{
			input:      "1 2\nX\n",
			wantOutput: "",
			wantRow:    0,
			wantCol:    0,
			wantErr:    true,
		},
	}

	for i, tc := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var w bytes.Buffer
			rows, cols, err := getDimensions(bytes.NewBufferString(tc.input), &w)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected err")
				}
				return
			}
			if !tc.wantErr && err != nil {
				t.Fatal("unexpected error", err)
			}
			if rows != tc.wantRow {
				t.Fatalf("bad rows got %d want %d", rows, tc.wantRow)
			}
			if cols != tc.wantCol {
				t.Fatalf("bad cols got %d want %d", cols, tc.wantCol)
			}
			if tc.wantOutput != w.String() {
				t.Logf("want %q", tc.wantOutput)
				t.Logf("got  %q", w.String())
				t.Fatal()
			}
		})
	}
}

func TestReadRow(t *testing.T) {
	tt := []struct {
		input string
		want  []int
		err   bool
		cols  int
	}{
		{"0 1 0 0\n", []int{0, 1, 0, 0}, false, 4},
		{"bob 0 0 0\n", nil, true, 4},
		{"0 20 0 0\n", nil, true, 4},
		{"0 1 0 0 0\n", nil, true, 4},
		{"0 1 0 0\n", nil, true, 5},
	}
	for i, tc := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			rdr := bytes.NewBufferString(tc.input)
			got, err := readRow(rdr, tc.cols)
			if err != nil {
				t.Log("got error", err)
			}
			if tc.err && err == nil {
				t.Fatalf("expected error")
			}
			if !tc.err && err != nil {
				t.Fatal("unexpected error", err)
			}

			assertEqual(t, got, tc.want)
		})
	}

}

func TestGetRow(t *testing.T) {
	tt := []struct {
		input      string
		wantOutput string
		wantRow    []int
		cols       int
		err        error
	}{
		{
			input:      "1 0 0 1\nC\n",
			wantRow:    []int{1, 0, 0, 1},
			wantOutput: "Enter a 4 element row containing space separated ones or zeros\nYou entered [1 0 0 1]\nContinue (C) Retry (R) Cancel (X)? ",
			cols:       4,
		},
		{
			input:      "1 0 0\nX\n",
			wantOutput: "Enter a 4 element row containing space separated ones or zeros\nError: \"unexpected newline\" Retry (R) Cancel (X)? ",
			wantRow:    nil,
			cols:       4,
			err:        errorUserTerminated,
		},
		{
			input: "1 0 0\nR\n1 1 1 1\nC\n",
			wantOutput: "Enter a 4 element row containing space separated ones or zeros\nError: \"unexpected newline\"" +
				" Retry (R) Cancel (X)? Enter a 4 element row containing space separated ones or zeros\nYou entered [1 1 " +
				"1 1]\nContinue (C) Retry (R) Cancel (X)? ",
			wantRow: []int{1, 1, 1, 1},
			cols:    4,
			err:     nil,
		},
		{
			input: "1 0 0 1\nR\n1 1 1 1\nC\n",
			wantOutput: "Enter a 4 element row containing space separated ones or zeros\nYou entered [1 0 0 1]\nContinue " +
				"(C) Retry (R) Cancel (X)? Enter a 4 element row containing space separated ones or zeros\nYou entered " +
				"[1 1 1 1]\nContinue (C) Retry (R) Cancel (X)? ",
			wantRow: []int{1, 1, 1, 1},
			cols:    4,
		},
		{
			input: "1 0 0 1\nR\n1 1 1 1\nX\n",
			wantOutput: "Enter a 4 element row containing space separated ones or zeros\nYou entered [1 0 0 1]\nContinue " +
				"(C) Retry (R) Cancel (X)? Enter a 4 element row containing space separated ones or zeros\nYou entered " +
				"[1 1 1 1]\nContinue (C) Retry (R) Cancel (X)? ",
			wantRow: nil,
			cols: 4,
			err: errorUserTerminated,
		},
	}
	for i, tc := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var w bytes.Buffer
			got, err := getRow(bytes.NewBufferString(tc.input), &w, tc.cols)
			if err != tc.err {
				t.Fatalf("unexpected error value %v", err)
			}
			if tc.wantOutput != w.String() {
				t.Logf("want %q", tc.wantOutput)
				t.Logf("got  %q", w.String())
				t.Fatal()
			}
			assertEqual(t, got, tc.wantRow)
		})
	}
}

func TestReadGrid(t *testing.T) {
	want := [][]int{
		{1,0,1,0,1,0},
		{0,1,0,1,0,1},
		{1,1,1,1,1,1},
		{1,1,1,0,0,0},
	}
	input := "4 6\nC\n1 0 1 0 1 0\nC\n0 1 0 1 0 1\nC\n1 1 1 1 1 1\nC\n1 1 1 0 0 0\nC\n"
	var w bytes.Buffer
	got, err := readGrid(bytes.NewBufferString(input), &w)
	if err != nil {
		t.Fatal("unexpected", err )
	}

	if len(got) != len(want) {
		t.Fatalf("wanted rows %d got rows %d", len(want), len(got))
	}
	for i := 0; i < len(want); i++ {
		t.Logf("row %d", i)
		assertEqual(t, got[i], want[i])
	}
}

func assertEqual(t *testing.T, got, want []int) {
	if len(got) == len(want) {
		for i := 0; i < len(got); i++ {
			if got[i] != want[i] {
				t.Logf("got  %v", got)
				t.Logf("want %v", want)
				t.Fatal()
			}
		}
		return
	}
}


