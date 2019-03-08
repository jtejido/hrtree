package hrtree

import (
	"fmt"
	"strings"
)

type Rectangle interface {
	UpperRight() Point
	LowerLeft() Point
}

type Point [Dim]uint64

type rectangle struct {
	lowerLeft, upperRight Point // the upper-left and lower-right bounds
}

func newRect(lowerLeft, upperRight Point) (r rectangle, err error) {
	if len(lowerLeft) != len(upperRight) {
		err = fmt.Errorf("lower left and upper right bounds must have the same dimension.")
		return
	}

	r.lowerLeft = lowerLeft
	r.upperRight = upperRight

	return
}

func (r *rectangle) String() string {
	var s [Dim]string
	for i, a := range r.lowerLeft {
		b := r.upperRight[i]
		s[i] = fmt.Sprintf("[%v, %v]", a, b)
	}
	return strings.Join(s[:], "x")
}

func (r *rectangle) size() float64 {
	size := 1.0
	for i, a := range r.lowerLeft {
		b := r.upperRight[i]
		size *= float64(b - a)
	}
	return size
}

func (r1 *rectangle) enlarge(r2 *rectangle) {
	for i := 0; i < Dim; i++ {
		if r1.lowerLeft[i] > r2.lowerLeft[i] {
			r1.lowerLeft[i] = r2.lowerLeft[i]
		}
		if r1.upperRight[i] < r2.upperRight[i] {
			r1.upperRight[i] = r2.upperRight[i]
		}
	}
}

func (r1 *rectangle) contains(r2 Rectangle) bool {
	for i, a1 := range r1.lowerLeft {
		b1, a2, b2 := r1.upperRight[i], r2.LowerLeft()[i], r2.UpperRight()[i]
		if a1 > a2 || b2 > b1 {
			return false
		}
	}

	return true
}

func equal(r1, r2 Rectangle) (ok bool) {
	for i, a1 := range r1.LowerLeft() {
		b1, a2, b2 := r1.UpperRight()[i], r2.LowerLeft()[i], r2.UpperRight()[i]
		if a1 != a2 && b2 != b1 {
			return false
		}
	}

	return true
}

func getCenter(r Rectangle) []uint64 {
	center := make([]uint64, Dim)
	for i := 0; i < Dim; i++ {
		center[i] = (r.LowerLeft()[i] + r.UpperRight()[i]) / 2
	}

	return center
}

func intersect(r1 *rectangle, r2 Rectangle) (ok bool) {
	ok = true
	for i := 0; ok && i < Dim; i++ {
		ok = r1.lowerLeft[i] <= r2.UpperRight()[i] && r1.upperRight[i] >= r2.LowerLeft()[i]
	}
	return
}
