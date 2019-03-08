package hrtree

import (
	"errors"
	"fmt"
	h "github.com/jtejido/hilbert"
	"math"
	"math/big"
	"sort"
)

const (
	DefaultMaxNodeEntries = 1000
	DefaultMinNodeEntries = 20
	Dim                   = 2
	SiblingsNumber        = 2  // minimum number of cooperating siblings used for moving entries before split is considered
	DefaultResolution     = 32 // minimum resolution required for hilbert computation's resolution
)

var ErrMinGTMax = errors.New("Minimum number of nodes should be less than Maximum number of nodes and not vice versa.")

// HRtree represents a Hilbert R-tree, a balanced search tree for storing and querying
// spatial objects.  MinChildren/MaxChildren specify the minimum/maximum branching factors.
type HRtree struct {
	min, max, bits int
	root           *node
	hf             *h.Hilbert
	size           int
}

// NewTree creates a new HRtree instance.
func NewTree(min, max, bits int) (*HRtree, error) {
	hf, err := h.New(uint32(bits), 2)

	if err != nil {
		return nil, err
	}

	if min < 0 {
		min = DefaultMinNodeEntries
	}

	if max < 0 {
		max = DefaultMaxNodeEntries
	}

	if max < min {
		return nil, ErrMinGTMax
	}

	rt := HRtree{min: min, max: max, bits: bits, hf: hf}
	rt.root = newNode(min, max)
	rt.root.leaf = true
	return &rt, nil
}

// Size returns the number of objects currently stored in tree.
func (tree *HRtree) Size() int {
	return tree.size
}

func (tree *HRtree) String() string {
	return "(HRtree)"
}

// node represents a tree node of the tree.
type node struct {
	min, max    int
	parent      *node
	left, right *node
	leaf        bool
	entries     *entryList
	lhv         *big.Int
	bb          *rectangle // bounding-box of all children of this entry
}

func newNode(min, max int) *node {
	return &node{
		min:     min,
		max:     max,
		lhv: 	 big.NewInt(0),
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

		if n.lhv.Cmp(en.getLHV()) < 0 {
			n.lhv = en.getLHV()
		}
	}
}

// adjustMBR adjusts the bounding box of the node
func (n *node) adjustMBR() {
	var bb rectangle
	for i, e := range n.getEntries() {
		if i == 0 {
			bb = *e.getMBR()
		} else {
			bb.enlarge(e.getMBR())
		}
	}

	n.bb = &bb
}

func (n *node) isOverflowing() bool {
	return n.entries.len() == n.max
}

func (n *node) isUnderflowing() bool {
	return n.entries.len() <= n.min
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

func (n *node) removeLeaf(obj Rectangle) bool {
	if !n.leaf {
		panic("Cannot remove entry from nonleaf node.")
	}

	ind := -1
	for i, en := range n.entries.getEntries() {

		if equal(en.obj, obj) {
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

	if n.isOverflowing() {
		panic("The node is overflowing.")
	}

	n.entries.insert(e)
}

func (n *node) insertNonLeaf(e entry) {

	if n.leaf {
		panic("The current node is a leaf node.")
	}

	if n.isOverflowing() {
		panic("The node is overflowing.")
	}

	it := n.entries.insert(e)

	e.node.parent = n

	assert(n.right == nil || (n.right != nil && n.leaf == n.right.leaf))
	assert(n.left == nil || (n.left != nil && n.leaf == n.left.leaf))

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
		assert(e.node.leaf == prevSib.leaf)
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
		assert(e.node.leaf == nextSib.leaf)
	}
}

// reset entries, bounding-box and largest hilbert value.
func (n *node) reset() {
	n.entries = newList(n.max)
	n.bb = nil
	n.lhv = big.NewInt(0)
}

func (n *node) getMBR() *rectangle {
	return n.bb
}

// entry represents a spatial index record stored in a tree node.
// this is shared between non-leaf and leaf entries.
// non-leaf has node, leaf has obj
type entry struct {
	bb   *rectangle // bounding-box of of this entry
	node *node
	obj  Rectangle
	h    *big.Int // hilbert value
	leaf bool
}

func (e entry) String() string {
	if !e.leaf {
		return fmt.Sprintf("entry{bb: %v}", e.bb)
	}
	return fmt.Sprintf("entry{bb: %v, obj: %v, hilbert: %v}", e.bb, e.obj, e.h)
}

func (e entry) getMBR() *rectangle {
	if e.leaf {
		return e.bb
	} else {
		return e.node.bb
	}
}

func (e entry) getLHV() *big.Int {
	if e.leaf {
		return e.h
	} else {
		return big.NewInt(0)
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

	index := sort.Search(len(l.entries), func(i int) bool { return l.entries[i].getLHV().Cmp(el.getLHV()) ==1 })
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

// Insert inserts a spatial object into the tree. Through Center(), we compute the hilbert value
// from the uncollapsed n-dimensional coordinates.
func (tree *HRtree) Insert(obj Rectangle) {

	hv := tree.hf.Encode(getCenter(obj)...)
	e := entry{&rectangle{obj.LowerLeft(), obj.UpperRight()}, nil, obj, hv, true}
	tree.insert(e)
	tree.size++
}

// insert adds the specified entry to the tree at the specified level.
func (tree *HRtree) insert(e entry) {
	siblings := make([]*node, 0)
	leaf := tree.chooseNode(tree.root, e.h)
	var split *node

	if !leaf.isOverflowing() {
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
func (tree *HRtree) chooseNode(n *node, h *big.Int) *node {
	if n.leaf {
		return n
	}

	// choose the entry (R, ptr, LHV) with the minimum LHV value greater than h.
	var last entry
	for _, en := range n.getEntries() {
		assert(!en.leaf)
		if en.node.lhv.Cmp(h) >= 0 {
			return tree.chooseNode(en.node, h)
		}
		last = en
	}

	//if h is larger than all the LHV already in the node,
	//choose the last of the node entries
	return tree.chooseNode(last.node, h)
}

// TO-DO..unify with adjustTreeForRemove
func (tree *HRtree) adjustTreeForInsert(root, n, nn *node, siblings []*node) (newRoot *node) {
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
				newRoot = newNode(tree.min, tree.max)

				newRoot.insertNonLeaf(entry{node: n})
				newRoot.insertNonLeaf(entry{node: nn})
			}

			newRoot.adjustLHV()
			newRoot.adjustMBR()

		} else {
			if nn != nil {
				enn := entry{node: nn}
				if !np.isOverflowing() {
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
func (tree *HRtree) adjustTreeForRemove(n, nn *node, siblings []*node) {
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
				data := mainEntry.getEntries()
				n.reset()

				if mainEntry.leaf {
					n.leaf = true
					for _, en := range data {
						n.insertLeaf(en)
					}

				} else {
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

				if dnParent.isUnderflowing() {
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
		assert(node.leaf == e.leaf)
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
			assert(prevSib.leaf == nn.leaf)
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

func (tree *HRtree) handleUnderflow(target *node, nodes []*node) (*node, []*node) {

	var nn *node

	entries := newListUncapped()

	nodes = target.getSiblings(SiblingsNumber + 1)

	for _, node := range nodes {
		for _, e := range node.getEntries() {
			entries.insert(e)
		}

		node.reset()
	}

	if entries.len() < len(nodes)*tree.min && target.parent != nil {
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

func (tree *HRtree) Delete(obj Rectangle) (ok bool) {
	leaf := tree.findLeaf(tree.root, obj)
	if leaf == nil {
		return
	}

	var dl *node

	siblings := make([]*node, 0)

	if leaf.removeLeaf(obj) {

		tree.size--

		if leaf.isUnderflowing() {
			dl, siblings = tree.handleUnderflow(leaf, siblings)
		}

		tree.adjustTreeForRemove(leaf, dl, siblings)

		ok = true
	}

	return
}

// findLeaf finds the leaf node containing obj.
func (tree *HRtree) findLeaf(n *node, obj Rectangle) *node {
	if n.leaf {
		return n
	}
	// if not leaf, search all candidate subtrees
	for _, e := range n.getEntries() {

		if e.getMBR().contains(obj) {
			leaf := tree.findLeaf(e.node, obj)
			if leaf == nil {
				continue
			}
			// check if the leaf actually contains the object
			for _, leafEntry := range leaf.getEntries() {
				if equal(leafEntry.obj, obj) {
					return leaf
				}
			}
		}
	}

	return nil
}

// Searching
// SearchIntersect returns all objects that intersects the specified rectangle.
func (tree *HRtree) SearchIntersect(bb Rectangle) []Rectangle {
	results := []Rectangle{}
	return tree.searchIntersect(tree.root, bb, results)
}

func (tree *HRtree) searchIntersect(n *node, bb Rectangle, results []Rectangle) []Rectangle {

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
