package search

import (
	"bytes"
	"strconv"
	"testing"
)

func TestWrap(t *testing.T) {
	tt := []struct {
		i, dim, want int
	}{
		{0, 3, 0},
		{2, 3, 2},
		{3, 3, 0},
		{7, 3, 1},
		{-1, 3, 2},
		{-5, 3, 1},
		{-1, 11, 10},
		{-3, 3, 0},
	}
	for i, tc := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := wrap(tc.i, tc.dim)
			if got != tc.want {
				t.Fatalf("want %d got %d", tc.want, got)
			}
		})
	}
}

func TestWasSeen(t *testing.T) {
	tt := []struct {
		visited, test point
		mark, want    bool
	}{
		{getPoint(3, 3), getPoint(0, 0), false, false},
		{getPoint(1, 2), getPoint(1, 2), true, true},
		{getPoint(1, 2), getPoint(1, 3), true, false},
		{getPoint(1, 2), getPoint(2, 2), true, false},
		{getPoint(1, 2), getPoint(2, 3), true, false},
		{getPoint(0, 1), getPoint(0, 11), true, true},
		{getPoint(0, 0), getPoint(11, 0), true, true},
		{getPoint(10, 9), getPoint(-1, 9), true, true},
		{getPoint(9, 9), getPoint(9, -1), true, true},
	}

	for i, tc := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			grid := [][]int{
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			}
			s := state{
				grid: grid,
				rows: len(grid),
				cols: len(grid[0]),
			}
			if tc.mark {
				s.visit(tc.visited)
			}
			got := s.visited(tc.test)
			if got != tc.want {
				t.Fatal("unexpected")
			}
		})
	}
}

func TestShapesMatch(t *testing.T) {
	p1 := []point{
		{0, 0},
		{1, 0},
		{1, 1},
		{1, 2},
		{2, 2},
		{2, 1},
	}
	p2 := []point{
		{4, 1},
		{5, 1},
		{5, 2},
		{5, 3},
		{6, 3},
		{6, 2},
	}
	p3 := []point{
		{5, 0},
		{5, 1},
		{5, 2},
		{5, 3},
		{6, 3},
		{6, 2},
	}
	p4 := []point{
		{4, 1},
		{5, 1},
		{5, 2},
		{5, 3},
		{6, 3},
	}
	p5 := []point{
		{1, 1},
	}
	p6 := []point{
		{7, 3},
	}
	p7 := []point{
		{0, 8},
		{1, 8},
		{1, 9},
		{1, 10},
		{2, 10},
		{2, 9},
	}
	p8 := []point{
		{-1, 8},
		{0, 8},
		{0, 9},
		{0, 10},
		{1, 10},
		{1, 9},
	}

	tt := []struct {
		u     []point
		v     []point
		match bool
	}{
		{p1, p2, true},
		{p1, p3, false},
		{p1, p4, false},
		{p5, p6, true},
		{p1, p7, true},
		{p1, p8, true},
	}

	for i, tc := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			shp := shape{tc.u}
			comp := shape{tc.v}
			if shp.match(comp) != tc.match {
				t.Logf("-> %q", shp)
				t.Logf("-> %q", comp)
				t.Fatal("did not conform to expectations")
			}
		})
	}
}

func TestGetDirection(t *testing.T) {
	tt := []struct {
		u, v point
		want direction
	}{
		{point{0, 0}, point{1, 0}, right},
		{point{1, 0}, point{0, 0}, left},
		{point{0, 1}, point{0, 0}, up},
		{point{0, 0}, point{0, 1}, down},
		// handle wraps
		{point{0, 0}, point{-1, 0}, left},
		{point{1, 0}, point{2, 0}, right},
		{point{0, -1}, point{0, 0}, down},
		{point{0, 0}, point{0, -1}, up},
	}

	for i, tc := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			got := getDirection(tc.u, tc.v)
			if got != tc.want {
				t.Fatalf("wanted %q got %q", tc.want, got)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	tt := []struct {
		grid [][]int
		want string
	}{
		{
			grid: [][]int{
				{0, 0, 0},
				{0, 1, 0},
				{0, 0, 0},
			},
			want: "    X\n-----\n",
		},
		{
			grid: [][]int{
				{0, 1, 1, 0, 0},
				{0, 1, 0, 0, 0},
				{0, 1, 0, 0, 0},
				{0, 0, 0, 0, 0},
				{0, 1, 1, 0, 0},
				{0, 1, 0, 0, 0},
				{0, 1, 0, 0, 0},
				{0, 0, 0, 0, 0},
			},
			want: "    XX\n    X \n    X \n------\n",
		},
		{
			grid: [][]int{
				{0, 1, 1, 1, 0},
				{0, 1, 0, 1, 0},
				{0, 1, 1, 1, 0},
				{0, 0, 0, 0, 0},
				{0, 1, 1, 0, 0},
			},
			want: "    XXX\n    X X\n    XXX\n       \n    XX \n-------\n",
		},
		{
			grid: [][]int{
				{0, 1, 1, 0, 0},
				{0, 1, 0, 0, 0},
				{0, 1, 0, 0, 0},
				{0, 0, 0, 0, 0},
				{0, 1, 1, 0, 0},
				{0, 1, 0, 0, 0},
				{0, 1, 0, 0, 0},
			},
			want: "    XX\n    X \n    X \n      \n    XX\n    X \n    X \n------\n",
		},
		{
			grid: [][]int{
				{0, 0, 0, 1},
				{1, 1, 1, 1},
				{0, 0, 0, 1},
			},
			want: "       X\n    XXXX\n       X\n--------\n",
		},
	}

	for i, tc := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var w bytes.Buffer
			s, err := New(tc.grid)
			if err != nil {
				t.Fatal("unexpected error", err)
			}
			s.Print(&w)
			if w.String() != tc.want {
				t.Logf("want %q", tc.want)
				t.Logf("got  %q", w.String())
				t.Fatal()
			}
		})
	}
}
