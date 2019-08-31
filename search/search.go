package search

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type stateError string

func (se stateError) Error() string { return string(se) }

const errorNoRows = stateError("no rows in grid")
const errorNoCols = stateError("no columns in grid")

const (
	up direction = iota
	down
	left
	right

	set     int = 1
	unset   int = 0
	visited int = 2

	leftPadding = "    "
)

func New(g [][]int) (*state, error) {
	rows := len(g)
	if rows == 0 {
		return nil, errorNoRows
	}
	cols := len(g[0])
	if cols == 0 {
		return nil, errorNoCols
	}
	st := state{
		grid:   g,
		shapes: nil,
		rows:   rows,
		cols:   cols,
	}
	st.findShapes()
	return &st, nil
}

type state struct {
	grid       [][]int
	shapes     []shape
	rows, cols int
}

func (s state) Print(w io.Writer) {
	for _, shp := range s.shapes {
		shp.print(w, s.rows, s.cols)
	}
}

func (s *state) findShape(p point) *shape {
	var result *shape
	if !s.visited(p) {
		if s.visit(p) == set {
			result = new(shape)
			result.points = append(result.points, p)
			for _, dir := range []direction{up, right, down, left} {
				children := s.findShape(nextPoint(p, dir))
				if children != nil {
					result.points = append(result.points, children.points...)
				}
			}
		}
	}

	return result
}

func (s *state) findShapes() {
	for row := 0; row < s.rows; row++ {
		for col := 0; col < s.cols; col++ {
			shape := s.findShape(getPoint(col, row))
			if shape != nil {
				if !s.hasShape(*shape) {
					s.shapes = append(s.shapes, *shape)
				}
			}
		}
	}
}

func (s state) hasShape(shp shape) bool {
	for _, comp := range s.shapes {
		if comp.match(shp) {
			return true
		}
	}
	return false
}

func (s state) isShapePart(p point) bool {
	trns := p.transform(s.rows, s.cols)
	if s.grid[trns.x][trns.y] == set {
		return true
	}
	return false
}

func (s *state) visit(p point) int {
	t := p.transform(s.rows, s.cols)
	old := s.grid[t.y][t.x]
	if old == visited {
		panic(fmt.Sprint("original", p, "transformed", t))
	}
	s.grid[t.y][t.x] = visited
	return old
}

func (s state) visited(p point) bool {
	t := p.transform(s.rows, s.cols)
	if s.grid[t.y][t.x] == visited {
		return true
	}
	return false
}

type direction int

func (d direction) String() string {
	dirs := []string{
		"up",
		"down",
		"left",
		"right",
	}
	return dirs[d]
}

type point struct {
	x, y int
}

func (p point) index(rows, cols int) int {
	pp := p.transform(rows, cols)
	return pp.y*cols + pp.x
}

func (p point) match(v point) bool {
	if p.x != v.x {
		return false
	}
	if p.y != v.y {
		return false
	}
	return true
}

func (p point) transform(rows, cols int) point {
	x := wrap(p.x, cols)
	y := wrap(p.y, rows)
	return point{x, y}
}

func (p point) String() string {
	return fmt.Sprintf("point(x:%d, y:%d)", p.x, p.y)
}

type sorter struct {
	points     []point
	rows, cols int
}

func (s *sorter) Len() int      { return len(s.points) }
func (s *sorter) Swap(i, j int) { s.points[i], s.points[j] = s.points[j], s.points[i] }
func (s *sorter) Less(i, j int) bool {
	return s.points[i].index(s.rows, s.cols) < s.points[j].index(s.rows, s.cols)
}

type rowState []int

func (r rowState) empty() bool {
	for _, p := range r {
		if p == set {
			return false
		}
	}
	return true
}

func (r rowState) print(w io.Writer) {
	fmt.Fprint(w, leftPadding)
	for _, p := range r {
		switch p {
		case set:
			fmt.Fprint(w, "X")
		case unset:
			fmt.Fprint(w, " ")
		}
	}
	fmt.Fprintln(w)
}

type shape struct {
	points []point
}

func (s shape) String() string {
	var resp string
	for i := 0; i < len(s.points)-1; i++ {
		if resp != "" {
			resp += ", "
		}
		resp += getDirection(s.points[i], s.points[i+1]).String()
	}
	return resp
}

func (s shape) print(w io.Writer, rows, cols int) {
	transformedPoints, newRows, newCols := transform(s.points, rows, cols)

	byIndex := sorter{
		points: transformedPoints,
		rows:   newRows,
		cols:   newCols,
	}
	sort.Sort(&byIndex)

	for row := 0; row < newRows; row++ {
		rowPoints := make(rowState, newCols)
		for _, p := range byIndex.points {
			if p.y == row {
				rowPoints[p.x] = set
			}
		}
		rowPoints.print(w)
	}
	fmt.Fprintln(w, strings.Repeat("-", newCols+len(leftPadding)))
}

func getDirection(b, e point) direction {
	if b.y > e.y {
		return up
	}
	if b.y < e.y {
		return down
	}
	if b.x > e.x {
		return left
	}
	if b.x < e.x {
		return right
	}
	panic("should never get direction on same point")
}

func getPoint(x, y int) point {
	return point{x, y}
}

func (s shape) match(v shape) bool {
	if len(s.points) != len(v.points) {
		return false
	}
	if len(v.points) == 1 {
		return true
	}
	for i := 0; i < len(s.points)-1; i++ {
		d1 := getDirection(s.points[i], s.points[i+1])
		d2 := getDirection(v.points[i], v.points[i+1])
		if d1 != d2 {
			return false
		}
	}
	return true
}

func nextPoint(curr point, d direction) point {
	switch d {
	case up:
		return getPoint(curr.x, curr.y-1)
	case down:
		return getPoint(curr.x, curr.y+1)
	case left:
		return getPoint(curr.x-1, curr.y)
	case right:
		return getPoint(curr.x+1, curr.y)
	}
	panic("nextPoint is f'd")
}

func shapeDimensions(ps []point, rows, cols int) (lx, ux, ly, uy int) {
	lx, ux, ly, uy = cols, 0, rows, 0
	for _, p := range ps {
		if p.x < lx {
			lx = p.x
		}
		if p.x > ux {
			ux = p.x
		}
		if p.y < ly {
			ly = p.y
		}
		if p.y > uy {
			uy = p.y
		}
	}
	return
}

func transform(ps []point, rows, cols int) ([]point, int, int) {
	var transformed []point

	for _, p := range ps {
		tp := p.transform(rows, cols)
		transformed = append(transformed, tp)
	}
	lowX, highX, lowY, highY := shapeDimensions(transformed, rows, cols)

	for i := 0; i < len(transformed); i++ {
		transformed[i].x -= lowX
		transformed[i].y -= lowY
	}
	return transformed, highY - lowY + 1, highX - lowX + 1
}

func wrap(i, dim int) int {
	if i >= 0 {
		return i % dim
	}
	return (dim + i%dim) % dim
}
