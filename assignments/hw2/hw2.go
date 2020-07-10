package hw2

import (
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

// DeltaStep ...
// Apply the delta-stepping algorihtm to Graph and return
// a shortest path tree.
//
// Note that this uses Shortest to make it easier for you,
// but you can use another struct if that makes more sense
// for the concurrency model you chose.
func DeltaStep(s graph.Node, g graph.Graph) Shortest {
	// Your code goes here.
	// return newShortestFrom(s, g.Nodes())
	return DijkstraFrom(s, g)
}
