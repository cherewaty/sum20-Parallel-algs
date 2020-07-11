package hw2

import (
	"fmt"

	"github.com/gonum/graph"
)

// Dijkstra ...
// Runs dijkstra from gonum to make sure that the tests are correct.
func Dijkstra(s graph.Node, g graph.Graph) Shortest {
	return DijkstraFrom(s, g)
}

// BellmanFord ...
// Apply the bellman-ford algorihtm to Graph and return
// a shortest path tree.
//
// Note that this uses Shortest to make it easier for you,
// but you can use another struct if that makes more sense
// for the concurrency model you chose.
func BellmanFord(u graph.Node, g graph.Graph) (path Shortest) {
	// Your code goes here.
	// sequential version from https://github.com/gonum/graph/blob/master/path/bellman_ford_moore.go
	if !g.Has(u) {
		return Shortest{from: u}
	}
	var weight Weighting
	if wg, ok := g.(graph.Weighter); ok {
		weight = wg.Weight
	} else {
		weight = UniformCost(g)
	}

	nodes := g.Nodes()

	path = newShortestFrom(u, nodes)
	path.dist[path.indexOf[u.ID()]] = 0

	// make this parallel
	for i := 1; i < len(nodes); i++ {
		changed := false
		for j, u := range nodes {
			for _, v := range g.From(u) {
				k := path.indexOf[v.ID()]
				w, ok := weight(u, v)
				if !ok {
					panic("bellman-ford: unexpected invalid weight")
				}
				joint := path.dist[j] + w
				if joint < path.dist[k] {
					path.set(k, joint, j)
					changed = true
				}
			}
		}
		if !changed {
			break
		}
	}

	for j, u := range nodes {
		for _, v := range g.From(u) {
			k := path.indexOf[v.ID()]
			w, ok := weight(u, v)
			if !ok {
				panic("bellman-ford: unexpected invalid weight")
			}
			if path.dist[j]+w < path.dist[k] {
				return path
			}
		}
	}

	return path
}

// Apply the delta-stepping algorihtm to Graph and return
// a shortest path tree.
//
// Note that this uses Shortest to make it easier for you,
// but you can use another struct if that makes more sense
// for the concurrency model you chose.

// Bucket ...
type Bucket struct {
	nodes []graph.Node
}

/// distance ...
type distance struct {
	toIdx   int
	distNew float64
	fromIdx int
	changed bool
}

// DELTA hyperparameter
const DELTA float64 = 3

// DeltaStep ...
func DeltaStep(s graph.Node, g graph.Graph) Shortest {
	// Your code goes here.
	if !g.Has(s) {
		return Shortest{from: s}
	}

	// initialize bucket data structure
	var B []Bucket // sequence of buckets

	// set up the channel
	chnl := make(chan distance)

	// relax the source node
	path := newShortestFrom(s, g.Nodes())
	relax(s, s, 0, path, chnl, B)

	// while there are any buckets, do
	for bucketIndex, bucket := range B {
		// // init structure S for remembering deleted nodes
		var S []graph.Node
		S = nil

		fmt.Println(S)

		// while Bucket i isn't empty:
		for len(bucket.nodes) != 0 {
			getReqLight(bucketIndex, bucket, path, g) // find the light edges, store in req

			// 	S = append(S, B[i]) 	// add deleted nodes to S
			// 		// empty this bucket
			// 	for _, v in req { // relax all the edges in req (parallel)
			// 			// so relaxed
			// 	}
		}

		// // find the heavy edges
		// for _, v in req { // relax all the edges in req (parallel)
		// 	// so relaxed
		// }
	}

	return newShortestFrom(s, g.Nodes())
}

func relax(u graph.Node, v graph.Node, c float64, path Shortest, chnl chan distance, B []Bucket) {

	from := path.indexOf[u.ID()]
	to := path.indexOf[v.ID()]

	if c < path.dist[path.indexOf[u.ID()]] {
		chnl <- distance{toIdx: to, distNew: c, fromIdx: from, changed: true}
		// move to bucket
		bucketIndex := int(c / DELTA)
		B[bucketIndex].nodes = append(B[bucketIndex].nodes, v)

	} else {
		chnl <- distance{toIdx: to, distNew: c, fromIdx: from, changed: false}
	}
}

func getReqLight(bucketIndex int, bucket Bucket, path Shortest, g graph.Graph) {
	// make a channel
	lightChannel := make(chan graph.Node)

	for _, from := range bucket.nodes {
		for _, to := range g.From(from) {
			evaluateLight(from, to, path, g, bucket.nodes, bucketIndex, lightChannel)
		}
	}
}

func evaluateLight(from graph.Node, to graph.Node, path Shortest, g graph.Graph, bucketNodes []graph.Node, bucketIndex int, channel chan graph.Node) {
	var weight Weighting
	if wg, ok := g.(graph.Weighter); ok {
		weight = wg.Weight
	} else {
		weight = UniformCost(g)
	}

	w, ok := weight(from, to)

	if ok {
		if w <= DELTA { // is it light?
			channel <- to
		}
	}
}
