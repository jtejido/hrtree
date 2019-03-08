// Copyright 2012 Daniel Connelly.  All rights reserved.
package hrtree

import (
	"testing"
)

func TestContains(t *testing.T) {
	rect1 := rect(Point{2, 0}, Point{2, 2})
	rect2 := rect(Point{2, 3}, Point{2, 5})

	if rect1.contains(rect2) {
		t.Errorf("Expected not %v.contains(%v", rect1, rect2)
	}

	if intersect(rect1, rect2) {
		t.Errorf("Expected not intersect(%v, %v)", rect1, rect2)
	}

	rect3 := rect(Point{3, 0}, Point{5, 2})

	if rect1.contains(rect3) {
		t.Errorf("Expected not %v.contains(%v", rect1, rect3)
	}

	if intersect(rect1, rect3) {
		t.Errorf("Expected not intersect(%v, %v)", rect1, rect3)
	}

	if rect2.contains(rect3) {
		t.Errorf("Expected not %v.contains(%v", rect2, rect3)
	}

	if intersect(rect2, rect3) {
		t.Errorf("Expected not intersect(%v, %v)", rect2, rect3)
	}
}

func TestNoIntersection(t *testing.T) {

	rect1 := rect(Point{2, 2}, Point{2, 3})
	rect2 := rect(Point{0, 0}, Point{2, 1})

	if intersect(rect1, rect2) {
		t.Errorf("Expected intersect(%v, %v) == false", rect1, rect2)
	}
}

func TestNoIntersectionJustTouches(t *testing.T) {

	rect1 := rect(Point{1, 2}, Point{2, 3})
	rect2 := rect(Point{3, 5}, Point{6, 6})

	if intersect(rect1, rect2) {
		t.Errorf("Expected intersect(%v, %v) == nil", rect1, rect2)
	}
}

func TestContainmentIntersection(t *testing.T) {

	rect1 := rect(Point{1, 2}, Point{2, 3})
	rect2 := rect(Point{1, 3}, Point{2, 3})

	r := Point{1, 3}
	s := Point{2, 3}

	if !intersect(rect1, rect2) {
		t.Errorf("intersect(%v, %v) != %v, %v", rect1, rect2, r, s)
	}
}

func TestOverlapIntersection(t *testing.T) {

	rect1 := rect(Point{1, 2}, Point{2, 5})
	rect2 := rect(Point{1, 4}, Point{4, 6})

	r := Point{1, 4}
	s := Point{2, 5}

	if !intersect(rect1, rect2) {
		t.Errorf("intersect(%v, %v) != %v, %v", rect1, rect2, r, s)
	}
}

func TestToCenter(t *testing.T) {

	rect := rect(Point{2, 2}, Point{2, 4})
	center := getCenter(rect)
	expectedx := (rect.LowerLeft()[0] + rect.UpperRight()[0]) / 2
	expectedy := (rect.LowerLeft()[1] + rect.UpperRight()[1]) / 2

	if center[0] != expectedx {
		t.Errorf("Expected %v.ToCenter() == %v, %v, got %v", rect, center, expectedx, center[0])
	}

	if center[1] != expectedy {
		t.Errorf("Expected %v.ToCenter() == %v, %v, got %v", rect, center, expectedy, center[1])
	}

}
