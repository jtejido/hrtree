package main

import (
	"fmt"
	. "github.com/jtejido/hrtree"
)

type Thing struct {
	r *Rect
}

func mustRect(upper, lower Point) *Thing {

	r, err := NewRect(upper, lower)
	if err != nil {
		panic(err)
	}
	return &Thing{&r}
}

func indexOf(objs []Spatial, obj Spatial) int {
	ind := -1
	for i, r := range objs {
		if r == obj {
			ind = i
			break
		}
	}
	return ind
}

func (r *Thing) String() string {
	return fmt.Sprintf("%v", r.r)
}

func (r *Thing) Bounds() *Rect {
	return r.r
}

func (r *Thing) Center() Point {
	return r.r.ToCenter()
}

func main() {
	things := []*Thing{
		mustRect(Point{1, 1}, Point{2, 2}),
		mustRect(Point{6, 6}, Point{7, 7}),
		mustRect(Point{3, 3}, Point{4, 4}),
		mustRect(Point{4, 4}, Point{5, 5}),
		mustRect(Point{5, 5}, Point{6, 6}),
		mustRect(Point{7, 7}, Point{8, 8}),
		mustRect(Point{2, 2}, Point{3, 3}),
		mustRect(Point{8, 8}, Point{9, 9}),
	}

	rt, err := NewTree(2, 4, 32)

	if err != nil {
		fmt.Println(err)
	}

	for _, thing := range things {
		rt.Insert(thing)
	}

	bb := mustRect(Point{5, 5}, Point{6, 6})
	q := rt.SearchIntersect(bb.Bounds())
	fmt.Println(q)

	rt.Delete(bb)

	q2 := rt.SearchIntersect(bb.Bounds())
	fmt.Println(q2)

	rt.Insert(bb)

	q3 := rt.SearchIntersect(mustRect(Point{8, 8}, Point{9, 9}).Bounds())
	fmt.Println(q3)

}
