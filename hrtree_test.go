package hrtree

import (
	"fmt"
	h "github.com/jtejido/hilbert"
	"testing"
)

var hf, _ = h.New(uint32(5), 2)

func (r *rectangle) LowerLeft() Point {
	return r.lowerLeft
}

func (r *rectangle) UpperRight() Point {
	return r.upperRight
}

func rect(lower, upper Point) *rectangle {
	r, err := newRect(lower, upper)

	if err != nil {
		fmt.Println(err)
	}
	return &r
}

func index(objs []Rectangle, obj Rectangle) int {
	ind := -1
	for i, r := range objs {
		if r == obj {
			ind = i
			break
		}
	}
	return ind
}

func TestChooseNode(t *testing.T) {
	rt, _ := NewTree(DefaultMinNodeEntries, DefaultMaxNodeEntries, 5)

	rect1 := rect(Point{2, 1}, Point{2, 1})
	h1 := hf.Encode(getCenter(rect1)...)

	rect2 := rect(Point{2, 2}, Point{2, 2})
	h2 := hf.Encode(getCenter(rect2)...)

	rect3 := rect(Point{2, 3}, Point{2, 3})
	h3 := hf.Encode(getCenter(rect3)...)

	rect4 := rect(Point{2, 4}, Point{2, 4})
	h4 := hf.Encode(getCenter(rect4)...)

	l1 := entry{bb: rect1, obj: rect1, h: h2, leaf: true}
	l2 := entry{bb: rect2, obj: rect2, h: h3, leaf: true}
	l3 := entry{bb: rect3, obj: rect3, h: h4, leaf: true}

	leaf := newNode(DefaultMinNodeEntries, DefaultMaxNodeEntries)
	leaf.leaf = true

	if leaf := rt.chooseNode(rt.root, h1); leaf != rt.root {
		t.Errorf("expected chooseNode of empty tree to return root")
	}

	childNode1 := newNode(DefaultMinNodeEntries, DefaultMaxNodeEntries)
	childNode1.leaf = true
	childNode1.insertLeaf(l1)
	childNode1.adjustLHV()
	childNode1.adjustMBR()
	entry1 := entry{node: childNode1}

	childNode2 := newNode(DefaultMinNodeEntries, DefaultMaxNodeEntries)
	childNode2.leaf = true
	childNode2.insertLeaf(l2)
	childNode2.adjustLHV()
	childNode2.adjustMBR()
	entry2 := entry{node: childNode2}

	childNode3 := newNode(DefaultMinNodeEntries, DefaultMaxNodeEntries)
	childNode3.leaf = true
	childNode3.insertLeaf(l3)
	childNode3.adjustLHV()
	childNode3.adjustMBR()
	entry3 := entry{node: childNode3}

	nonLeaf := newNode(DefaultMinNodeEntries, DefaultMaxNodeEntries)

	nonLeaf.insertNonLeaf(entry3)
	nonLeaf.insertNonLeaf(entry2)
	nonLeaf.insertNonLeaf(entry1)

	if childNode3 != rt.chooseNode(nonLeaf, h2) {
		t.Errorf("incorrect chooseNode")
	}

}

func TestInsertNonLeafEntry(t *testing.T) {

	n := newNode(2, 4)
	nonLeafNode := newNode(2, 4)

	nonLeafEntry := entry{node: n}

	nonLeafNode.insertNonLeaf(nonLeafEntry)

	if nonLeafNode.entries.len() != 1 {
		t.Errorf("no entry added.")
	}

	if nonLeafNode != nonLeafEntry.node.parent {
		t.Errorf("incorrect parent.")
	}

}

func TestInsertNonLeafEntrySiblings(t *testing.T) {
	childNode := newNode(2, 4)
	childNode.leaf = true
	rect := rect(Point{2, 2}, Point{2, 4})
	h := hf.Encode(getCenter(rect)...)
	leafEntry := entry{bb: rect, obj: rect, h: h, leaf: true}
	childNode.insertLeaf(leafEntry)

	nonLeafEntry := entry{node: childNode}
	parent := newNode(2, 4)
	parent.insertNonLeaf(nonLeafEntry)

	if parent != childNode.parent {
		t.Errorf("incorrect parent.")
	}

	if parent != nonLeafEntry.node.parent {
		t.Errorf("incorrect parent.")
	}

	if childNode.right != nil {
		t.Errorf("incorrect right sibling.")
	}

	if childNode.left != nil {
		t.Errorf("incorrect left sibling.")
	}
}

func TestNodeOverflowing(t *testing.T) {
	rect := rect(Point{2, 2}, Point{2, 4})
	h := hf.Encode(getCenter(rect)...)

	leafEntry := entry{bb: rect, obj: rect, h: h, leaf: true}
	leafEntry2 := entry{bb: rect, obj: rect, h: h, leaf: true}

	n := newNode(1, 2)
	n.leaf = true

	if n.isOverflowing() {
		t.Errorf("should not be overflowing")
	}

	n.insertLeaf(leafEntry)
	n.insertLeaf(leafEntry2)

	if !n.isOverflowing() {
		t.Errorf("should be overflowing")
	}
}

func TestNodeUnderflowing(t *testing.T) {
	rect := rect(Point{2, 2}, Point{2, 4})
	h := hf.Encode(getCenter(rect)...)

	leafEntry := entry{bb: rect, obj: rect, h: h, leaf: true}
	leafEntry2 := entry{bb: rect, obj: rect, h: h, leaf: true}

	n := newNode(1, 2)
	n.leaf = true

	if !n.isUnderflowing() {
		t.Errorf("should not be overflowing")
	}

	n.insertLeaf(leafEntry)
	n.insertLeaf(leafEntry2)

	if n.isUnderflowing() {
		t.Errorf("should be overflowing")
	}

}

func TestAdjustMBR(t *testing.T) {
	rect1 := rect(Point{2, 0}, Point{2, 4})
	h1 := hf.Encode(getCenter(rect1)...)
	leafEntry1 := entry{bb: rect1, obj: rect1, h: h1, leaf: true}

	rect2 := rect(Point{2, 1}, Point{2, 5})
	h2 := hf.Encode(getCenter(rect2)...)
	leafEntry2 := entry{bb: rect2, obj: rect2, h: h2, leaf: true}

	rect3 := rect(Point{2, 5}, Point{2, 10})
	h3 := hf.Encode(getCenter(rect3)...)
	leafEntry3 := entry{bb: rect3, obj: rect3, h: h3, leaf: true}

	n := newNode(2, 4)
	n.leaf = true
	n.insertLeaf(leafEntry1)
	n.insertLeaf(leafEntry2)

	n.adjustMBR()
	r := n.getMBR()

	if 2 != r.lowerLeft[0] {
		t.Errorf("incorrect lower[x]")
	}

	if 0 != r.lowerLeft[1] {
		t.Errorf("incorrect lower[y]")
	}

	if 2 != r.upperRight[0] {
		t.Errorf("incorrect upper[x]")
	}

	if 5 != r.upperRight[1] {
		t.Errorf("incorrect upper[y]")
	}

	n1 := newNode(2, 4)
	n1.leaf = true
	n1.insertLeaf(leafEntry1)
	n1.insertLeaf(leafEntry3)
	n1.adjustMBR()

	r1 := n1.getMBR()

	if 2 != r1.lowerLeft[0] {
		t.Errorf("incorrect lower[x]")
	}

	if 0 != r1.lowerLeft[1] {
		t.Errorf("incorrect lower[y]")
	}

	if 2 != r1.upperRight[0] {
		t.Errorf("incorrect upper[x]")
	}

	if 10 != r1.upperRight[1] {
		t.Errorf("incorrect upper[y]")
	}

}

func TestAdjustMBR2(t *testing.T) {
	rect1 := rect(Point{2, 2}, Point{2, 3})
	h1 := hf.Encode(getCenter(rect1)...)

	leafEntry1 := entry{bb: rect1, obj: rect1, h: h1, leaf: true}

	rect2 := rect(Point{2, 8}, Point{2, 8})
	h2 := hf.Encode(getCenter(rect2)...)

	leafEntry2 := entry{bb: rect2, obj: rect2, h: h2, leaf: true}

	n := newNode(2, 4)
	n.leaf = true
	n.insertLeaf(leafEntry1)
	n.insertLeaf(leafEntry2)
	n.adjustMBR()
	r := n.getMBR()

	if 2 != r.lowerLeft[0] {
		t.Errorf("incorrect upper[y]")
	}

	if 2 != r.lowerLeft[1] {
		t.Errorf("incorrect upper[y]")
	}

	if 2 != r.upperRight[0] {
		t.Errorf("incorrect upper[y]")
	}

	if 8 != r.upperRight[1] {
		t.Errorf("incorrect upper[y]")
	}

}

func TestAdjustLHV(t *testing.T) {
	rect1 := rect(Point{2, 0}, Point{2, 0})
	h1 := hf.Encode(getCenter(rect1)...)
	leafEntry1 := entry{bb: rect1, obj: rect1, h: h1, leaf: true}

	rect2 := rect(Point{2, 0}, Point{2, 2})
	h2 := hf.Encode(getCenter(rect2)...)
	leafEntry2 := entry{bb: rect2, obj: rect2, h: h2, leaf: true}

	n := newNode(2, 4)
	n.leaf = true
	n.insertLeaf(leafEntry1)
	n.insertLeaf(leafEntry2)

	n.adjustLHV()

	if h1.Cmp(h2) >= 0 {
		t.Errorf("incorrect hilbert value")
	}

	if h2.Cmp(n.lhv) != 0 {
		t.Errorf("incorrect hilbert value")
	}
}

func TestSiblings(t *testing.T) {

	right := newNode(2, 4)
	right.leaf = true

	left := newNode(2, 4)
	left.leaf = true

	main := newNode(2, 4)
	main.leaf = true

	if 1 != len(main.getSiblings(2)) {
		t.Errorf("incorrect number of siblings")
	}

	main.right = right
	right.left = main

	if 2 != len(main.getSiblings(2)) {
		t.Errorf("incorrect number of siblings")
	}

	main.left = left
	left.right = main

	if 2 != len(main.getSiblings(2)) {
		t.Errorf("incorrect number of siblings")
	}

	if 1 != len(main.getSiblings(1)) {
		t.Errorf("incorrect number of siblings")
	}

	siblings := main.getSiblings(3)

	if siblings[0] != main {
		t.Errorf("incorrect sibling")
	}

	if siblings[1] != right {
		t.Errorf("incorrect sibling")
	}

}

func TestHandleOverflow(t *testing.T) {
	node1 := newNode(DefaultMinNodeEntries, DefaultMaxNodeEntries)
	node1.leaf = true
	siblings := make([]*node, 0)
	hf2, _ := h.New(uint32(5), 32)

	for i := 0; i < DefaultMaxNodeEntries; i++ {
		rect := rect(Point{2, uint64(i)}, Point{2, uint64(i)})
		h := hf2.Encode(getCenter(rect)...)
		entry := entry{bb: rect, obj: rect, h: h, leaf: true}
		node1.insertLeaf(entry)
	}

	rect2 := rect(Point{2, 0}, Point{2, 0})
	h2 := hf2.Encode(getCenter(rect2)...)
	entry2 := entry{bb: rect2, obj: rect2, h: h2, leaf: true}

	node2, _ := handleOverflow(node1, entry2, siblings)

	if DefaultMaxNodeEntries/2 != node1.entries.len() {
		t.Errorf("incorrect number of entries at node1")
	}

	if DefaultMaxNodeEntries/2+1 != node2.entries.len() {
		t.Errorf("incorrect number of entries at node2")
	}

	if node1 != node2.right {
		t.Errorf("incorrect right sibling at node2")
	}

	if nil != node2.left {
		t.Errorf("incorrect left sibling at node2")
	}

	if nil != node1.right {
		t.Errorf("incorrect right sibling at node1")
	}

	if node2 != node1.left {
		t.Errorf("incorrect left sibling at node1")
	}

}

func TestSearchIntersect(t *testing.T) {
	rt, _ := NewTree(3, 3, 12)
	things := []Rectangle{
		rect(Point{0, 0}, Point{2, 1}),
		rect(Point{3, 1}, Point{4, 3}),
		rect(Point{1, 2}, Point{3, 4}),
		rect(Point{8, 6}, Point{9, 7}),
		rect(Point{10, 3}, Point{11, 5}),
		rect(Point{11, 7}, Point{12, 8}),
		rect(Point{2, 6}, Point{3, 8}),
		rect(Point{3, 6}, Point{4, 8}),
		rect(Point{2, 8}, Point{3, 10}),
		rect(Point{3, 8}, Point{4, 10}),
	}

	for _, thing := range things {
		rt.Insert(thing)
	}

	bb := rect(Point{2, 1}, Point{12, 7})
	q := rt.SearchIntersect(bb)

	expected := []int{0, 1, 2, 3, 4, 5, 6, 7}

	if len(q) != len(expected) {
		t.Errorf("SearchIntersect failed to find all objects")
	}
	for _, ind := range expected {
		if index(q, things[ind]) < 0 {
			t.Errorf("SearchIntersect failed to find things[%d]", ind)
		}
	}
}

func TestDelete(t *testing.T) {
	rt, _ := NewTree(DefaultMinNodeEntries, DefaultMaxNodeEntries, 5)
	rect0 := rect(Point{2, 4}, Point{2, 8})

	rt.Insert(rect0)

	if !rt.root.leaf {
		t.Errorf("Root should be leaf")
	}

	if 1 != rt.root.entries.len() {
		t.Errorf("Root should have 1 entry")
	}

	if 1 != len(rt.SearchIntersect(rect0)) {
		t.Errorf("tree should have 1 result")
	}

	rect1 := rect(Point{2, 5}, Point{2, 7})
	rt.Delete(rect1)

	if !rt.root.leaf {
		t.Errorf("Root should be leaf")
	}

	if 1 != rt.root.entries.len() {
		t.Errorf("Root should have 1 entry")
	}

	if 1 != len(rt.SearchIntersect(rect0)) {
		t.Errorf("tree should have 1 result")
	}

	rect2 := rect(Point{2, 2}, Point{2, 10})
	rt.Delete(rect2)

	if !rt.root.leaf {
		t.Errorf("Root should be leaf")
	}

	if 1 != rt.root.entries.len() {
		t.Errorf("Root should have 1 entry")
	}

	if 1 != len(rt.SearchIntersect(rect0)) {
		t.Errorf("tree should have 1 result")
	}

	rt.Delete(rect0)

	if !rt.root.leaf {
		t.Errorf("Root should be leaf")
	}

	if 0 != rt.root.entries.len() {
		t.Errorf("Root should have 1 entry")
	}

	if 0 != len(rt.SearchIntersect(rect0)) {
		t.Errorf("tree should have 1 result")
	}
}

func TestDeleteAtMax(t *testing.T) {
	rt, _ := NewTree(DefaultMinNodeEntries, DefaultMaxNodeEntries, 12)

	for i := 0; i < DefaultMaxNodeEntries; i++ {
		r := rect(Point{2, uint64(i)}, Point{2, uint64(i)})
		rt.Insert(r)
	}

	for i := 0; i < (DefaultMaxNodeEntries - DefaultMinNodeEntries); i++ {
		r2 := rect(Point{2, uint64(i)}, Point{2, uint64(i)})
		rt.Delete(r2)
	}

	if !rt.root.leaf {
		t.Errorf("root should be leaf")
	}

	if DefaultMinNodeEntries != rt.root.entries.len() {
		t.Errorf("incorrect number of entries left")
	}

}

func TestDeleteAtMax2(t *testing.T) {
	nodeNo := DefaultMaxNodeEntries * 4
	rt, _ := NewTree(DefaultMinNodeEntries, DefaultMaxNodeEntries, 12)

	for i := 0; i < nodeNo; i++ {
		r := rect(Point{2, uint64(i)}, Point{2, uint64(i)})
		rt.Insert(r)
	}

	for i := 0; i < nodeNo; i++ {
		r2 := rect(Point{2, uint64(i)}, Point{2, uint64(i)})
		rt.Delete(r2)
	}

	if !rt.root.leaf {
		t.Errorf("root should be leaf")
	}

	if 0 != rt.root.entries.len() {
		t.Errorf("incorrect number of entries left")
	}
}

func TestRedistributeEntries(t *testing.T) {
	entries := newListUncapped()
	nodes := make([]*node, 0)

	node1 := newNode(DefaultMinNodeEntries, DefaultMaxNodeEntries)
	node1.leaf = true
	nodes = append(nodes, node1)

	node2 := newNode(DefaultMinNodeEntries, DefaultMaxNodeEntries)
	node2.leaf = true
	nodes = append(nodes, node2)

	for i := 0; i < DefaultMaxNodeEntries*2-1; i++ {
		rect := rect(Point{2, 1}, Point{2, 1})

		h := hf.Encode(getCenter(rect)...)
		leafEntry := entry{bb: rect, obj: rect, h: h, leaf: true}
		entries.insert(leafEntry)
	}

	redistributeEntries(entries, nodes)

	if DefaultMaxNodeEntries != node1.entries.len() {
		t.Errorf("incorrect number of entries")
	}

	if DefaultMaxNodeEntries-1 != node2.entries.len() {
		t.Errorf("incorrect number of entries")
	}
}

func TestSearchIntersectNoResult(t *testing.T) {
	rt, _ := NewTree(3, 3, 12)
	things := []Rectangle{
		rect(Point{0, 0}, Point{2, 1}),
		rect(Point{3, 1}, Point{4, 3}),
		rect(Point{1, 2}, Point{3, 4}),
		rect(Point{8, 6}, Point{9, 7}),
		rect(Point{10, 3}, Point{11, 5}),
		rect(Point{11, 7}, Point{12, 8}),
		rect(Point{2, 6}, Point{3, 8}),
		rect(Point{3, 6}, Point{4, 8}),
		rect(Point{2, 8}, Point{3, 10}),
		rect(Point{3, 8}, Point{4, 10}),
	}

	for _, thing := range things {
		rt.Insert(thing)
	}

	bb := rect(Point{99, 99}, Point{109, 94})
	q := rt.SearchIntersect(bb)
	if len(q) != 0 {
		t.Errorf("SearchIntersect failed to return nil slice on failing query")
	}
}

func BenchmarkGetIntersect(b *testing.B) {
	b.StopTimer()
	rt, _ := NewTree(3, 3, 12)
	things := []Rectangle{
		rect(Point{0, 0}, Point{2, 1}),
		rect(Point{3, 1}, Point{4, 3}),
		rect(Point{1, 2}, Point{3, 4}),
		rect(Point{8, 6}, Point{9, 7}),
		rect(Point{10, 3}, Point{11, 5}),
		rect(Point{11, 7}, Point{12, 8}),
		rect(Point{2, 6}, Point{3, 8}),
		rect(Point{3, 6}, Point{4, 8}),
		rect(Point{2, 8}, Point{3, 10}),
		rect(Point{3, 8}, Point{4, 10}),
	}
	for _, thing := range things {
		rt.Insert(thing)
	}

	bb := rect(Point{2, 1}, Point{12, 7})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rt.SearchIntersect(bb)
	}
}

func BenchmarkInsert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rt, _ := NewTree(3, DefaultMaxNodeEntries, 5)
		things := []Rectangle{
			rect(Point{0, 0}, Point{2, 1}),
			rect(Point{3, 1}, Point{4, 3}),
			rect(Point{1, 2}, Point{3, 4}),
			rect(Point{8, 6}, Point{9, 7}),
			rect(Point{10, 3}, Point{11, 5}),
			rect(Point{11, 7}, Point{12, 8}),
			rect(Point{2, 6}, Point{3, 8}),
			rect(Point{3, 6}, Point{4, 8}),
			rect(Point{2, 8}, Point{3, 10}),
			rect(Point{3, 8}, Point{4, 10}),
		}
		for _, thing := range things {
			rt.Insert(thing)
		}
	}
}
