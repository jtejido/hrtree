// Copyright 2012 Daniel Connelly.  All rights reserved.
package hrtree

import (
	"fmt"
	"math"
	"strings"
)

type Point [Dim]float64

func (p Point) ToRect(tol float64) *Rect {
	var r Rect
	for i := range p {
		r.lower[i] = p[i] - tol
		r.upper[i] = p[i] + tol
	}
	return &r
}

func (p Point) dist(q Point) float64 {
	sum := 0.0
	for i := range p {
		dx := p[i] - q[i]
		sum += dx * dx
	}
	return math.Sqrt(sum)
}

type Rect struct {
	lower, upper Point // the upper-left and lower-right bounds
}

func NewRect(lower, upper Point) (r Rect, err error) {
	if len(lower) != len(upper) {
		err = fmt.Errorf("lower left and upper right bounds must have the same dimension.")
		return
	}

	r.lower = lower
	r.upper = upper

	return
}

func (r *Rect) String() string {
	var s [Dim]string
	for i, a := range r.lower {
		b := r.upper[i]
		s[i] = fmt.Sprintf("[%.2f, %.2f]", a, b)
	}
	return strings.Join(s[:], "x")
}

// size computes the measure of a rectangle (the product of its side lengths).
func (r *Rect) size() float64 {
	size := 1.0
	for i, a := range r.lower {
		b := r.upper[i]
		size *= b - a
	}
	return size
}

func (r1 *Rect) enlarge(r2 *Rect) {
	for i := 0; i < Dim; i++ {
		if r1.lower[i] > r2.lower[i] {
			r1.lower[i] = r2.lower[i]
		}
		if r1.upper[i] < r2.upper[i] {
			r1.upper[i] = r2.upper[i]
		}
	}
}

// containsRect tests whether r2 is is located inside r1.
func (r1 *Rect) contains(r2 *Rect) bool {
	for i, a1 := range r1.lower {
		b1, a2, b2 := r1.upper[i], r2.lower[i], r2.upper[i]
		if a1 > a2 || b2 > b1 {
			return false
		}
	}

	return true
}

func (r1 *Rect) equal(r2 *Rect) (ok bool) {
	for i, a1 := range r1.lower {
		b1, a2, b2 := r1.upper[i], r2.lower[i], r2.upper[i]
		if a1 != a2 && b2 != b1 {
			return false
		}
	}

	return true
}

// func (r1 *Rect) equal(r2 *Rect) (ok bool) {
// 	xlow1, ylow1 := r1.lower[0], r1.lower[1]
// 	xhigh2, yhigh2 := r2.upper[0], r2.upper[1]

// 	xhigh1, yhigh1 := r1.upper[0], r1.upper[1]
// 	xlow2, ylow2 := r2.lower[0], r2.lower[1]

// 	return xlow1 == xlow2 && xhigh1 == xhigh2 && ylow1 == ylow2 && yhigh1 == yhigh2
// }

func (r *Rect) ToCenter() (center Point) {
	for i := 0; i < Dim; i++ {
		center[i] = (r.lower[i] + r.upper[i]) / 2
	}

	return
}

// We return positive result on rectangles touching
func intersect(r1, r2 *Rect) (ok bool) {
	ok = true
	for i := 0; ok && i < Dim; i++ {
		ok = r1.lower[i] <= r2.upper[i] && r1.upper[i] >= r2.lower[i]
	}
	return
}

// func intersect(r1, r2 *Rect) bool {
// 	for i := 0; i < Dim; i++ {
// 		if r2.upper[i] <= r1.lower[i] || r1.upper[i] <= r2.lower[i] {
// 			return false
// 		}
// 	}
// 	return true
// }

// minDist computes the square of the distance from a point to a rectangle.
// If the point is contained in the rectangle then the distance is zero.
//
// Implemented per Definition 2 of "Nearest Neighbor Queries" by
// N. Roussopoulos, S. Kelley and F. Vincent, ACM SIGMOD, pages 71-79, 1995.
func (p Point) minDist(r *Rect) float64 {
	sum := 0.0
	for i, pi := range p {
		if pi < r.lower[i] {
			d := pi - r.lower[i]
			sum += d * d
		} else if pi > r.upper[i] {
			d := pi - r.upper[i]
			sum += d * d
		} else {
			sum += 0
		}
	}
	return sum
}

// minMaxDist computes the minimum of the maximum distances from p to points
// on r.  If r is the bounding box of some geometric objects, then there is
// at least one object contained in r within minMaxDist(p, r) of p.
//
// Implemented per Definition 4 of "Nearest Neighbor Queries" by
// N. Roussopoulos, S. Kelley and F. Vincent, ACM SIGMOD, pages 71-79, 1995.
func (p Point) minMaxDist(r *Rect) float64 {
	// by definition, MinMaxDist(p, r) =
	// min{1<=k<=n}(|pk - rmk|^2 + sum{1<=i<=n, i != k}(|pi - rMi|^2))
	// where rmk and rMk are defined as follows:

	rm := func(k int) float64 {
		if p[k] <= (r.lower[k]+r.upper[k])/2 {
			return r.lower[k]
		}
		return r.upper[k]
	}

	rM := func(k int) float64 {
		if p[k] >= (r.lower[k]+r.upper[k])/2 {
			return r.lower[k]
		}
		return r.upper[k]
	}

	// This formula can be computed in linear time by precomputing
	// S = sum{1<=i<=n}(|pi - rMi|^2).

	S := 0.0
	for i := range p {
		d := p[i] - rM(i)
		S += d * d
	}

	// Compute MinMaxDist using the precomputed S.
	min := math.MaxFloat64
	for k := range p {
		d1 := p[k] - rM(k)
		d2 := p[k] - rm(k)
		d := S - d1*d1 + d2*d2
		if d < min {
			min = d
		}
	}

	return min
}
