![hrtree](http://bukhantsov.org/wp-content/uploads/2012/04/r-tree-result.png)

# hrtree

<a href="https://travis-ci.org/jtejido/hrtree"><img src="https://img.shields.io/travis/jtejido/hrtree.svg?style=flat-square" alt="Build Status"></a>

### Hilbert R-tree

The Hilbert R-Tree uses Hilbert Curves on top of an R-Tree structure so as to impose the linear ordering among the points on each nodes. 

It has similar functions in comparison with Guttman's traditional R-Tree and therefore, supports all the underlying operations of R-Tree (search, insertion and deletion). The only difference being that the Hilbert value of the MBR is used instead of considering the area or the distances of the Bounding boxes.

With the help of this linear ordering, we can achieve almost 100% space utilization by deferring the split from a regular S+1 to S-to-(S+1).

### Hilbert Space-Filling Curve

A Hilbert curve is a continuous fractal space-filling curve, first described by the German mathematician David Hilbert in 1891. A space-filling curve visits all points in the grid exactly once without crossings. There are several such curves,like Gray-code curve, Hilbert curve and etc.. Of particular importance to us among them is Hilbert Curve, as it achieves the best clustering of the points in the grid. The Hilbert curve algorithm used here is from a paper by John Skilling titled "Programming the Hilbert curve", published in American Institute of Physics.