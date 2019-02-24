// Copyright 2012 Daniel Connelly.  All rights reserved.
package hrtree

import (
	"math"
	"testing"
)

const EPS = 0.000000001

func TestDist(t *testing.T) {
	p := Point{1, 2}
	q := Point{4, 5}
	dist := math.Sqrt(18)

	if d := p.dist(q); d != dist {
		t.Errorf("dist(%v, %v) = %v; expected %v", p, q, d, dist)
	}
}

func TestNewRect(t *testing.T) {

	lengths := [Dim]float64{2.5, 8.0}

	rect, err := NewRect(Point{1.0, -2.5}, Point{3.5, 5.5})
	if err != nil {
		t.Errorf("Error on NewRect(%v, %v): %v", rect.lower, lengths, err)
	}
	if d := rect.lower.dist(rect.lower); d > EPS {
		t.Errorf("Expected p == rect.p")
	}
	if d := rect.upper.dist(rect.upper); d > EPS {
		t.Errorf("Expected q == rect.q")
	}
}

func TestRectSize(t *testing.T) {
	p := Point{1.0, -2.5}
	q := Point{3.5, 5.5}
	lengths := [Dim]float64{2.5, 8.0}
	rect, _ := NewRect(p, q)
	size := lengths[0] * lengths[1]
	actual := rect.size()
	if size != actual {
		t.Errorf("Expected %v.size() == %v, got %v", rect, size, actual)
	}
}

func TestContains(t *testing.T) {
	rect1, _ := NewRect(Point{3.7, -2.4}, Point{9.9, -1.3})
	rect2, _ := NewRect(Point{4.1, -1.9}, Point{7.3, -1.3})

	if yes := rect1.contains(&rect2); !yes {
		t.Errorf("Expected %v.containsRect(%v", rect1, rect2)
	}
}

func TestDoesNotContainRectOverlaps(t *testing.T) {

	rect1, _ := NewRect(Point{3.7, -2.4}, Point{9.9, -1.3})
	rect2, _ := NewRect(Point{4.1, -1.9}, Point{7.2, -0.5})

	if yes := rect1.contains(&rect2); yes {
		t.Errorf("Expected %v doesn't contain %v", rect1, rect2)
	}
}

func TestDoesNotContainRectDisjoint(t *testing.T) {

	rect1, _ := NewRect(Point{3.7, -2.4}, Point{9.9, -1.3})
	rect2, _ := NewRect(Point{1.2, -19.6}, Point{3.4, -13.7})

	if yes := rect1.contains(&rect2); yes {
		t.Errorf("Expected %v doesn't contain %v", rect1, rect2)
	}
}

func TestNoIntersection(t *testing.T) {

	rect1, _ := NewRect(Point{1, 2}, Point{2, 3})
	rect2, _ := NewRect(Point{-1, -2}, Point{1.5, 1})

	// rect1 and rect2 fail to overlap in just one dimension (second)

	if intersect(&rect1, &rect2) {
		t.Errorf("Expected intersect(%v, %v) == false", rect1, rect2)
	}
}

// consider touching an intersection
// func TestNoIntersectionJustTouches(t *testing.T) {

// 	rect1, _ := NewRect(Point{1, 2}, Point{2, 3})
// 	rect2, _ := NewRect(Point{-1, -2}, Point{1.5, 2})

// 	// rect1 and rect2 fail to overlap in just one dimension (second)

// 	if intersect(&rect1, &rect2) {
// 		t.Errorf("Expected intersect(%v, %v) == nil", rect1, rect2)
// 	}
// }

func TestContainmentIntersection(t *testing.T) {

	rect1, _ := NewRect(Point{1, 2}, Point{2, 3})
	rect2, _ := NewRect(Point{1, 2.2}, Point{1.5, 2.7})

	r := Point{1, 2.2}
	s := Point{1.5, 2.7}

	if !intersect(&rect1, &rect2) {
		t.Errorf("intersect(%v, %v) != %v, %v", rect1, rect2, r, s)
	}
}

func TestOverlapIntersection(t *testing.T) {

	rect1, _ := NewRect(Point{1, 2}, Point{2, 4.5})
	rect2, _ := NewRect(Point{1, 4}, Point{4, 6})

	r := Point{1, 4}
	s := Point{2, 4.5}

	if !intersect(&rect1, &rect2) {
		t.Errorf("intersect(%v, %v) != %v, %v", rect1, rect2, r, s)
	}
}

func TestToRect(t *testing.T) {
	x := Point{3.7, -2.4}
	tol := 0.05
	rect := x.ToRect(tol)

	p := Point{3.65, -2.45}
	q := Point{3.75, -2.35}
	d1 := p.dist(rect.lower)
	d2 := q.dist(rect.upper)
	if d1 > EPS || d2 > EPS {
		t.Errorf("Expected %v.ToRect(%v) == %v, %v, got %v", x, tol, p, q, rect)
	}
}

func TestToCenter(t *testing.T) {

	rect, _ := NewRect(Point{2, 2}, Point{2, 4})
	center := rect.ToCenter()
	expectedx := (rect.lower[0] + rect.upper[0]) / 2
	expectedy := (rect.lower[1] + rect.upper[1]) / 2

	if center[0] != expectedx {
		t.Errorf("Expected %v.ToCenter() == %v, %v, got %v", rect, center, expectedx, center[0])
	}

	if center[1] != expectedy {
		t.Errorf("Expected %v.ToCenter() == %v, %v, got %v", rect, center, expectedy, center[1])
	}

}

func TestToCenterToRect(t *testing.T) {

	rect, _ := NewRect(Point{2, 2}, Point{4, 4})
	center := rect.ToCenter()
	rect2 := *center.ToRect(1)

	if rect != rect2 {
		t.Errorf("Expected %v, got %v", rect, rect2)
	}

}

func TestEqual(t *testing.T) {

	rect, _ := NewRect(Point{2, 2}, Point{4, 4})
	center := rect.ToCenter()
	rect2 := center.ToRect(1)

	if !rect.equal(rect2) {
		t.Errorf("Expected %v == %v", rect, rect2)
	}
}

func TestMinDistZero(t *testing.T) {
	p := Point{1, 2}
	r := p.ToRect(1)
	if d := p.minDist(r); d > EPS {
		t.Errorf("Expected %v.minDist(%v) == 0, got %v", p, r, d)
	}
}

func TestMinMaxdist(t *testing.T) {
	p := Point{-3, -2}
	r := Rect{Point{0, 0}, Point{1, 2}}

	q1 := Point{0, 2}
	q2 := Point{1, 0}
	q3 := Point{1, 2}

	d1 := p.dist(q1)
	d2 := p.dist(q2)
	d3 := p.dist(q3)
	expected := math.Min(d1*d1, math.Min(d2*d2, d3*d3))

	if d := p.minMaxDist(&r); math.Abs(d-expected) > EPS {
		t.Errorf("Expected %v.minMaxDist(%v) == %v, got %v", p, r, expected, d)
	}
}
