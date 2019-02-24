package hrtree

import (
	"fmt"
	h "github.com/jtejido/hilbert"
	"math"
	"sort"
)

const (
	DefaultMaxNodeEntries = 1000
	DefaultMinNodeEntries = 20
	Dim                   = 2  // default spatial dimension
	SiblingsNumber        = 2  // minimum number of cooperating siblings used for moving entries before split is considered
	DefaultBits           = 32 // minimum bits required for hilbert computation's resolution
)

// Rtree represents a Hilbert R-tree, a balanced search tree for storing and querying
// spatial objects.  MinChildren/MaxChildren specify the minimum/maximum branching factors.
type Rtree struct {
	MinChildren, MaxChildren, Bits int
	root                           *node
	HilbertFunc                    *h.Hilbert
	size                           int
}

// NewTree creates a new R-tree instance.
func NewTree(MinChildren, MaxChildren, Bits int) (*Rtree, error) {
	hf, err := h.New(uint32(Bits), 2)

	if err != nil {
		return nil, err
	}

	if MinChildren < 0 {
		MinChildren = DefaultMinNodeEntries
	}

	if MaxChildren < 0 {
		MaxChildren = DefaultMaxNodeEntries
	}

	if MaxChildren < MinChildren {
		MaxChildren++
	}

	rt := Rtree{MinChildren: MinChildren, MaxChildren: MaxChildren, Bits: Bits, HilbertFunc: hf}
	rt.root = newNode(MinChildren, MaxChildren)
	rt.root.leaf = true
	return &rt, nil
}

// Size returns the number of objects currently stored in tree.
func (tree *Rtree) Size() int {
	return tree.size
}

func (tree *Rtree) String() string {
	return "(H-Rtree)"
}

// node represents a tree node of the tree.
type node struct {
	min, max    int
	parent      *node
	left, right *node
	leaf        bool
	entries     *entryList
	lhv         uint64
	bb          *Rect // bounding-box of all children of this entry
}

func newNode(min, max int) *node {
	return &node{
		min:     min,
		max:     max,
		entries: newList(max),
	}
}

func (n *node) String() string {
	return fmt.Sprintf("node{leaf: %v, entries: %v, lhv : %v}", n.leaf, n.entries, n.lhv)
}

func (n *node) getEntries() []entry {
	return n.entries.getEntries()
}

// adjustLHV gets the largest Hilbert value among the node's entries
func (n *node) adjustLHV() {
	for _, en := range n.getEntries() {
		if en.h > n.lhv {
			n.lhv = en.h
		}
	}
}

// adjustMBR adjusts the bounding box of the node
func (n *node) adjustMBR() {
	var bb Rect
	for i, e := range n.getEntries() {
		if i == 0 {
			bb = *e.getMBR()
		} else {
			bb.enlarge(e.getMBR())
		}
	}

	n.bb = &bb
}

func (n *node) isOverflow() bool {
	return n.entries.len() == n.max
}

func (n *node) isUnderflow() bool {
	return n.entries.len() < n.max
}

func (n *node) getSiblings(siblingsNum int) []*node {
	nodes := make([]*node, 0)
	nodes = append(nodes, n)
	right := n.right
	for len(nodes) < siblingsNum && right != nil {
		nodes = append(nodes, right)
		right = right.right
	}

	return nodes
}

func (n *node) removeLeaf(obj Spatial) bool {
	if !n.leaf {
		panic("Cannot remove entry from nonleaf node.")
	}

	ind := -1
	for i, en := range n.entries.getEntries() {

		if en.obj.Bounds().equal(obj.Bounds()) {
			ind = i
		}
	}

	if ind < 0 {
		return false
	}

	n.entries.entries = append(n.entries.entries[:ind], n.entries.entries[ind+1:]...)

	return true
}

func (n *node) removeNonLeaf(node *node) bool {
	if n.leaf {
		panic("Cannot remove entry from leaf node.")
	}

	ind := -1
	for i, en := range n.entries.getEntries() {
		if node == en.node {
			ind = i
		}
	}

	if ind < 0 {
		return false
	}

	n.entries.entries = append(n.entries.entries[:ind], n.entries.entries[ind+1:]...)

	return true

}

func (n *node) insertLeaf(e entry) {

	if !n.leaf {
		panic("The current node is not a leaf node.")
	}

	if n.isOverflow() {
		panic("The node is overflowing.")
	}

	n.entries.insert(e)
}

func (n *node) insertNonLeaf(e entry) {

	if n.leaf {
		panic("The current node is a leaf node.")
	}

	if n.isOverflow() {
		panic("The node is overflowing.")
	}

	it := n.entries.insert(e)

	e.node.parent = n

	var nextSib, prevSib *node

	if n.entries.get(it) != n.entries.first() {
		prevIt := it
		prevIt--
		prev := n.entries.get(prevIt)
		prevSib = prev.node
	}

	e.node.left = prevSib

	if prevSib != nil {
		prevSib.right = e.node
	}

	aux := n.entries.len()
	aux--

	if it != aux {
		nextIt := it
		nextIt++
		next := n.entries.get(nextIt)
		nextSib = next.node
	}

	e.node.right = nextSib

	if nextSib != nil {
		nextSib.left = e.node
	}
}

// reset entries, bounding-box and largest hilbert value.
func (n *node) reset() {
	n.entries = newList(n.max)
	n.bb = nil
	n.lhv = 0
}

func (n *node) getMBR() *Rect {
	return n.bb
}

// entry represents a spatial index record stored in a tree node.
// this is shared between non-leaf and leaf entries.
// non-leaf has node, leaf has obj
type entry struct {
	bb   *Rect // bounding-box of of this entry
	node *node
	obj  Spatial
	h    uint64 // hilbert value
	leaf bool
}

func (e entry) String() string {
	if !e.leaf {
		return fmt.Sprintf("entry{bb: %v}", e.bb)
	}
	return fmt.Sprintf("entry{bb: %v, obj: %v, hilbert: %v}", e.bb, e.obj, e.h)
}

func (e entry) getMBR() *Rect {
	if e.leaf {
		return e.bb
	} else {
		return e.node.bb
	}
}

// wrapper struct for entries
// this is used for abstracting utilities
type entryList struct {
	entries []entry
}

func newList(max int) *entryList {
	return &entryList{
		entries: make([]entry, 0, max),
	}
}

func newListUncapped() *entryList {
	return &entryList{
		entries: make([]entry, 0),
	}
}

func (l *entryList) insert(el entry) int {

	index := sort.Search(len(l.entries), func(i int) bool { return l.entries[i].h > el.h })
	l.entries = append(l.entries, entry{})
	copy(l.entries[index+1:], l.entries[index:])
	l.entries[index] = el
	return index

}

func (l *entryList) first() entry {
	return l.entries[0]
}

func (l *entryList) last() entry {
	return l.entries[l.len()-1]
}

func (l *entryList) len() int {
	return len(l.entries)
}

func (l *entryList) get(i int) entry {
	return l.entries[i]
}

func (l entryList) getEntries() []entry {
	return l.entries
}

// Any type that implements Spatial can be stored in an Rtree and queried.
type Spatial interface {
	Bounds() *Rect
	Center() Point
}

// Insert inserts a spatial object into the tree. Through Center(), we compute the hilbert value
// from the uncollapsed 3-dimensional coordinates. Through Bounds(), we get to operate on the bounding box.
func (tree *Rtree) Insert(obj Spatial) {
	p := obj.Center()

	hv := tree.HilbertFunc.Encode(uint64(p[0]), uint64(p[1]))

	e := entry{obj.Bounds(), nil, obj, hv.Uint64(), true}
	tree.insert(e)
	tree.size++
}

// insert adds the specified entry to the tree at the specified level.
func (tree *Rtree) insert(e entry) {
	siblings := make([]*node, 0)
	leaf := tree.chooseNode(tree.root, e.h)
	var split *node

	if !leaf.isOverflow() {
		leaf.insertLeaf(e)
		leaf.adjustLHV()
		leaf.adjustMBR()
		siblings = append(siblings, leaf)

	} else {
		// split leaf if overflows
		split, siblings = handleOverflow(leaf, e, siblings)
	}

	// TO-DO.. make the caller handle root adjustments
	tree.root = tree.adjustTreeForInsert(tree.root, leaf, split, siblings)

}

// chooseNode finds the node to which e should be added.
func (tree *Rtree) chooseNode(n *node, h uint64) *node {
	if n.leaf {
		return n
	}

	// choose the entry (R, ptr, LHV) with the minimum LHV value greater than h.
	var last entry
	for _, en := range n.getEntries() {
		if !en.leaf {
			if en.node.lhv >= h {
				return tree.chooseNode(en.node, h)
			}
			last = en
		}
	}

	//if h is larger than all the LHV already in the node,
	//choose the last of the node entries
	return tree.chooseNode(last.node, h)
}

// TO-DO..unify with adjustTreeForRemove
func (tree *Rtree) adjustTreeForInsert(root, n, nn *node, siblings []*node) (newRoot *node) {
	var pp *node
	var ok bool = true

	newRoot = root
	newSiblings := make([]*node, 0)

	s := siblings

	for ok {
		np := n.parent
		if np == nil {
			ok = false
			if nn != nil {
				newRoot = newNode(tree.MinChildren, tree.MaxChildren)

				newRoot.insertNonLeaf(entry{node: n})
				newRoot.insertNonLeaf(entry{node: nn})
			}

			newRoot.adjustLHV()
			newRoot.adjustMBR()

		} else {
			if nn != nil {
				enn := entry{node: nn}
				if !np.isOverflow() {
					np.insertNonLeaf(enn)
					np.adjustLHV()
					np.adjustMBR()

					newSiblings = append(newSiblings, np)

				} else {
					pp, newSiblings = handleOverflow(np, enn, newSiblings)
				}
			} else {
				newSiblings = append(newSiblings, np)
			}

			for _, node := range s {
				node.parent.adjustLHV()
				node.parent.adjustMBR()
			}

			n = np
			nn = pp

			s = newSiblings

		}
	}

	return newRoot

}

// TO-DO..unify with adjustTreeForInsert
func (tree *Rtree) adjustTreeForRemove(n, nn *node, siblings []*node) {
	var keepRunning bool = true

	newSiblings := make([]*node, 0)

	s := siblings

	for keepRunning {
		np := n.parent
		var dpParent *node

		if np == nil {
			keepRunning = false
			if n.entries.len() == 1 && !n.leaf {
				mainEntry := n.entries.get(0).node

				if mainEntry.leaf {
					n.leaf = true
					data := mainEntry.getEntries()
					n.reset()
					for _, en := range data {
						n.insertLeaf(en)
					}

				} else {
					data := mainEntry.getEntries()
					n.reset()
					for _, en := range data {
						n.insertNonLeaf(en)
					}
				}

			}

			n.adjustLHV()
			n.adjustMBR()
		} else {

			if nn != nil {
				dnParent := nn.parent
				dnParent.removeNonLeaf(nn)

				if dnParent.entries.len() < tree.MinChildren {
					dpParent, newSiblings = tree.handleUnderflow(dnParent, newSiblings)
				} else {
					newSiblings = append(newSiblings, dnParent)
				}
			}

			newSiblings = append(newSiblings, np)

			for _, node := range s {
				node.parent.adjustLHV()
				node.parent.adjustMBR()
			}

			n = np
			nn = dpParent

			s = newSiblings

		}
	}
}

// The overflow handling algorithm in the Hilbert R-tree treats the overflowing nodes
// either by moving some of the entries to one of the s - 1 cooperating siblings or by splitting
// s nodes into s+1 nodes (2-3 splitting).
func handleOverflow(n *node, e entry, nodes []*node) (*node, []*node) {

	min := n.min

	max := n.max

	var targetPos int

	var nn *node

	nodes = n.getSiblings(SiblingsNumber)

	entries := newListUncapped()

	entries.insert(e)

	for i, node := range nodes {
		for _, e := range node.getEntries() {
			entries.insert(e)
		}

		node.reset()
		if node == n {
			targetPos = i
		}
	}

	if entries.len() > len(nodes)*max {
		nn = newNode(min, max)
		nn.leaf = e.leaf

		prevSib := n.left
		nn.left = prevSib

		if prevSib != nil {
			prevSib.right = nn
		}

		nn.right = n
		n.left = nn

		nodes = append(nodes, nil)
		copy(nodes[targetPos+1:], nodes[targetPos:])
		nodes[targetPos] = nn

	}

	redistributeEntries(entries, nodes)

	return nn, nodes
}

func (tree *Rtree) handleUnderflow(target *node, nodes []*node) (*node, []*node) {

	var nn *node

	entries := newListUncapped()

	nodes = target.getSiblings(SiblingsNumber + 1)

	for _, node := range nodes {
		for _, e := range node.getEntries() {
			entries.insert(e)
		}

		node.reset()
	}

	if entries.len() < len(nodes)*tree.MinChildren && target.parent != nil {
		nn = nodes[0]
		prevSib := nn.left
		nextSib := nn.right

		if prevSib != nil {
			prevSib.right = nextSib
		}

		if nextSib != nil {
			nextSib.left = prevSib
		}

		nodes = append(nodes[:0], nodes[0+1:]...)

	}

	redistributeEntries(entries, nodes)

	return nn, nodes
}

func redistributeEntries(entries *entryList, siblings []*node) {
	batchSize := int(math.Ceil(float64(entries.len()) / float64(len(siblings))))

	var currentBatch int

	j := 0
	for _, sibling := range siblings {

		for i := j; i < entries.len(); i++ {
			ee := entries.get(i)

			if ee.leaf {
				sibling.insertLeaf(ee)
			} else {
				sibling.insertNonLeaf(ee)
			}

			currentBatch++

			if currentBatch == batchSize {
				currentBatch = 0
				j = i + 1
				break
			}
		}

		sibling.adjustLHV()
		sibling.adjustMBR()
	}
}

func (tree *Rtree) Delete(obj Spatial) (ok bool) {
	leaf := tree.findLeaf(tree.root, obj)
	if leaf == nil {
		return
	}

	var dl *node

	siblings := make([]*node, 0)

	if leaf.removeLeaf(obj) {

		tree.size--

		if leaf.entries.len() < tree.MinChildren {
			dl, siblings = tree.handleUnderflow(leaf, siblings)
		}

		tree.adjustTreeForRemove(leaf, dl, siblings)

		ok = true
	}

	return
}

// findLeaf finds the leaf node containing obj.
func (tree *Rtree) findLeaf(n *node, obj Spatial) *node {
	if n.leaf {
		return n
	}
	// if not leaf, search all candidate subtrees
	for _, e := range n.getEntries() {

		if e.getMBR().contains(obj.Bounds()) {
			leaf := tree.findLeaf(e.node, obj)
			if leaf == nil {
				continue
			}
			// check if the leaf actually contains the object
			for _, leafEntry := range leaf.getEntries() {
				if leafEntry.obj.Bounds().equal(obj.Bounds()) {
					return leaf
				}
			}
		}
	}

	return nil
}

// Searching
// SearchIntersect returns all objects that intersect the specified rectangle.
func (tree *Rtree) SearchIntersect(bb *Rect) []Spatial {
	results := []Spatial{}
	return tree.searchIntersect(tree.root, bb, results)
}

func (tree *Rtree) searchIntersect(n *node, bb *Rect, results []Spatial) []Spatial {

	for _, e := range n.getEntries() {

		if intersect(e.getMBR(), bb) {
			if n.leaf {
				results = append(results, e.obj)
			} else {
				results = tree.searchIntersect(e.node, bb, results)
			}
		}
	}
	return results
}

// NearestNeighbor returns the closest object to the specified point.
func (tree *Rtree) NearestNeighbor(p Point) Spatial {
	obj, _ := tree.nearestNeighbor(p, tree.root, math.MaxFloat64, nil)
	return obj
}

// utilities for sorting slices of entries
type entrySlice struct {
	entries []entry
	dists   []float64
	pt      Point
}

func (s entrySlice) Len() int { return len(s.entries) }

func (s entrySlice) Swap(i, j int) {
	s.entries[i], s.entries[j] = s.entries[j], s.entries[i]
	s.dists[i], s.dists[j] = s.dists[j], s.dists[i]
}

func (s entrySlice) Less(i, j int) bool {
	return s.dists[i] < s.dists[j]
}

func sortEntries(p Point, entries []entry) ([]entry, []float64) {
	sorted := make([]entry, len(entries))
	dists := make([]float64, len(entries))
	for i := 0; i < len(entries); i++ {
		sorted[i] = entries[i]
		dists[i] = p.minDist(entries[i].bb)
	}
	sort.Sort(entrySlice{sorted, dists, p})
	return sorted, dists
}

func pruneEntries(p Point, entries []entry, minDists []float64) []entry {
	minMinMaxDist := math.MaxFloat64
	for i := range entries {
		minMaxDist := p.minMaxDist(entries[i].bb)
		if minMaxDist < minMinMaxDist {
			minMinMaxDist = minMaxDist
		}
	}
	// remove all entries with minDist > minMinMaxDist
	pruned := []entry{}
	for i := range entries {
		if minDists[i] <= minMinMaxDist {
			pruned = append(pruned, entries[i])
		}
	}
	return pruned
}

func (tree *Rtree) nearestNeighbor(p Point, n *node, d float64, nearest Spatial) (Spatial, float64) {
	if n.leaf {
		for _, e := range n.getEntries() {
			dist := math.Sqrt(p.minDist(e.bb))
			if dist < d {
				d = dist
				nearest = e.obj
			}
		}
	} else {
		branches, dists := sortEntries(p, n.getEntries())
		branches = pruneEntries(p, branches, dists)
		for _, e := range branches {
			subNearest, dist := tree.nearestNeighbor(p, e.node, d, nearest)
			if dist < d {
				d = dist
				nearest = subNearest
			}
		}
	}

	return nearest, d
}

func (tree *Rtree) NearestNeighbors(k int, p Point) []Spatial {
	dists := make([]float64, k)
	objs := make([]Spatial, k)
	for i := 0; i < k; i++ {
		dists[i] = math.MaxFloat64
		objs[i] = nil
	}
	objs, _ = tree.nearestNeighbors(k, p, tree.root, dists, objs)
	return objs
}

// insert obj into nearest and return the first k elements in increasing order.
func insertNearest(k int, dists []float64, nearest []Spatial, dist float64, obj Spatial) ([]float64, []Spatial) {
	i := 0
	for i < k && dist >= dists[i] {
		i++
	}
	if i >= k {
		return dists, nearest
	}

	left, right := dists[:i], dists[i:k-1]
	updatedDists := make([]float64, k)
	copy(updatedDists, left)
	updatedDists[i] = dist
	copy(updatedDists[i+1:], right)

	leftObjs, rightObjs := nearest[:i], nearest[i:k-1]
	updatedNearest := make([]Spatial, k)
	copy(updatedNearest, leftObjs)
	updatedNearest[i] = obj
	copy(updatedNearest[i+1:], rightObjs)

	return updatedDists, updatedNearest
}

func (tree *Rtree) nearestNeighbors(k int, p Point, n *node, dists []float64, nearest []Spatial) ([]Spatial, []float64) {
	if n.leaf {
		for _, e := range n.getEntries() {
			dist := math.Sqrt(p.minDist(e.bb))
			dists, nearest = insertNearest(k, dists, nearest, dist, e.obj)
		}
	} else {
		branches, branchDists := sortEntries(p, n.getEntries())
		branches = pruneEntries(p, branches, branchDists)
		for _, e := range branches {
			nearest, dists = tree.nearestNeighbors(k, p, e.node, dists, nearest)
		}
	}
	return nearest, dists
}
