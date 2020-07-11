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

// Apply the bellman-ford algorihtm to Graph and return
// a shortest path tree.
//
// Note that this uses Shortest to make it easier for you,
// but you can use another struct if that makes more sense
// for the concurrency model you chose.

// Distance ...
type distance struct {
	toIdx   int
	distNew float64
	fromIdx int
	changed bool
}

// UpdateDist ...
func UpdateDist(chnl chan distance, u graph.Node, v graph.Node, path Shortest, w float64) {
	k := path.indexOf[v.ID()]
	j := path.indexOf[u.ID()]
	var changed bool
	joint := path.dist[j] + w
	if joint < path.dist[k] {
		changed = true
		fmt.Println(joint)
	} else {
		changed = false
	}
	chnl <- distance{toIdx: k, distNew: joint, fromIdx: j, changed: changed}
}

// BellmanFord ...
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

	chnl := make(chan distance)

	for i := 1; i < len(nodes); i++ {
		changed := false
		for _, u := range nodes {
			for _, v := range g.From(u) {
				w, ok := weight(u, v)
				if !ok {
					panic("bellman-ford: unexpected invalid weight")
				}
				if w < 0 {
					panic("bellman-ford: negative weight")
				}
				go UpdateDist(chnl, u, v, path, w)
			}
			for range g.From(u) {
				dist, ok := <-chnl
				fmt.Println(dist)
				if !ok {
					panic("bellman-ford: bad channel read")
					// fmt.Println(v)
				}
				changed = dist.changed
				if changed {
					path.set(dist.toIdx, dist.distNew, dist.fromIdx)
				}
			}
		}
		if !changed {
			break
		}
	}

	//	for j, u := range nodes {
	//		for _, v := range g.From(u) {
	//			k := path.indexOf[v.ID()]
	//			w, ok := weight(u, v)
	//			if !ok {
	//				panic("bellman-ford: unexpected invalid weight")
	//			}
	///			if path.dist[j]+w < path.dist[k] {
	//			return path
	//		}
	//	}
	//	}
	close(chnl)
	// fmt.Println("BELLMAN FORD: ",path.dist)
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

	// relax the source node
	path := newShortestFrom(s, g.Nodes())
	relax(s, s, 0, path, B)

	allNodes := g.Nodes()

	// while there are any buckets, do
	for bucketIndex, bucket := range B {
		// // init structure S for remembering deleted nodes
		var S Bucket // kinda of a faux bucket
		S.nodes = nil

		requestedChannel := make(chan distance)

		// while Bucket i isn't empty:
		for len(bucket.nodes) != 0 {
			getReqLight(bucketIndex, bucket, path, g, requestedChannel) // find the light edges

			// add deleted nodes to S
			for _, bucketNode := range bucket.nodes {
				S.nodes = append(S.nodes, bucketNode)
			}

			// empty this bucket
			bucket.nodes = nil

			for range bucket.nodes {
				requested := <-requestedChannel
				if requested.changed {
					relax(allNodes[requested.fromIdx], allNodes[requested.toIdx], float64(requested.distNew), path, B)
					path.set(requested.toIdx, requested.distNew, requested.fromIdx)
				}
			}
		}

		getReqHeavy(bucketIndex, S, path, g, requestedChannel) // find the heavy edges
		for range bucket.nodes {
			requested := <-requestedChannel
			if requested.changed {
				relax(allNodes[requested.fromIdx], allNodes[requested.toIdx], float64(requested.distNew), path, B)
				path.set(requested.toIdx, requested.distNew, requested.fromIdx)
			}
		}
	}

	return path
}

func relax(u graph.Node, v graph.Node, c float64, path Shortest, B []Bucket) {
	if c < path.dist[path.indexOf[u.ID()]] {
		// what bucket should it be in?
		bucketIndex := int(c / DELTA)
		moveNodeToNewBucket(B, bucketIndex, v)
	}
}

func getReqLight(bucketIndex int, bucket Bucket, path Shortest, g graph.Graph, requested chan distance) {
	for _, from := range bucket.nodes {
		for _, to := range g.From(from) {
			evaluateLight(from, to, path, g, bucket.nodes, bucketIndex, requested)
		}
	}
}

func evaluateLight(from graph.Node, to graph.Node, path Shortest, g graph.Graph, bucketNodes []graph.Node, bucketIndex int, channel chan distance) {
	var weight Weighting
	if wg, ok := g.(graph.Weighter); ok {
		weight = wg.Weight
	} else {
		weight = UniformCost(g)
	}

	w, ok := weight(from, to)

	if ok {
		if w <= DELTA { // is it light?
			channel <- distance{toIdx: path.indexOf[to.ID()], distNew: w + path.dist[path.indexOf[to.ID()]], fromIdx: path.indexOf[from.ID()], changed: true}
		} else {
			channel <- distance{toIdx: path.indexOf[to.ID()], distNew: w + path.dist[path.indexOf[to.ID()]], fromIdx: path.indexOf[from.ID()], changed: false}
		}
	}
}

func getReqHeavy(bucketIndex int, s Bucket, path Shortest, g graph.Graph, requested chan distance) {
	for _, from := range s.nodes {
		for _, to := range g.From(from) {
			evaluateHeavy(from, to, path, g, s.nodes, bucketIndex, requested)
		}
	}
}

func evaluateHeavy(from graph.Node, to graph.Node, path Shortest, g graph.Graph, sNodes []graph.Node, bucketIndex int, channel chan distance) {
	var weight Weighting
	if wg, ok := g.(graph.Weighter); ok {
		weight = wg.Weight
	} else {
		weight = UniformCost(g)
	}

	w, ok := weight(from, to)

	if ok {
		if w > DELTA { // is it heavy?
			channel <- distance{toIdx: path.indexOf[to.ID()], distNew: w + path.dist[path.indexOf[to.ID()]], fromIdx: path.indexOf[from.ID()], changed: true}
		} else {
			channel <- distance{toIdx: path.indexOf[to.ID()], distNew: w + path.dist[path.indexOf[to.ID()]], fromIdx: path.indexOf[from.ID()], changed: false}
		}
	}
}

func removeNodeFromBuckets(buckets []Bucket, node graph.Node) {
	for _, bucket := range buckets {
		for i, searchNode := range bucket.nodes {
			if searchNode.ID() == node.ID() {
				// remove it!
				copy(bucket.nodes[i:], bucket.nodes[i+1:])
				bucket.nodes[len(bucket.nodes)-1] = nil
			}
		}
	}
}

func moveNodeToNewBucket(buckets []Bucket, bucketIndex int, node graph.Node) {
	// remove from its old bucket, if it's in one
	removeNodeFromBuckets(buckets, node)

	// add to its new bucket
	buckets[bucketIndex].nodes = append(buckets[bucketIndex].nodes, node)
}
