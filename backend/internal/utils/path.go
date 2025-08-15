package utils

import (
	"container/heap"

	"github.com/google/uuid"
)

type Path struct {
	StartID uuid.UUID   // ID of the starting point
	EndID   uuid.UUID   // ID of the ending point
	Steps   []uuid.UUID // List of steps in the path
}

type AdjacencyMap[T comparable] map[T]map[T]float64

const maxDistance = 1000000.0 // Arbitrary large value for distance

// Min-heap for Dijkstra
type item[T comparable] struct {
	node     T
	priority float64
}
type minHeap[T comparable] []item[T]

func (h minHeap[T]) Len() int            { return len(h) }
func (h minHeap[T]) Less(i, j int) bool  { return h[i].priority < h[j].priority }
func (h minHeap[T]) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap[T]) Push(x interface{}) { *h = append(*h, x.(item[T])) }
func (h *minHeap[T]) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func FindPath[T comparable](adjacency AdjacencyMap[T], start, end T, heuristicFunc func(T, T) float64) []T {

	openSet := &minHeap[T]{{node: start, priority: 0}}
	heap.Init(openSet)

	previous := make(map[T]T)

	gScore := make(map[T]float64)
	for node := range adjacency {
		gScore[node] = maxDistance
	}
	gScore[start] = 0

	fScore := make(map[T]float64)
	for node := range adjacency {
		fScore[node] = maxDistance
	}
	fScore[start] = heuristicFunc(start, end)

	for openSet.Len() > 0 {
		curr := heap.Pop(openSet).(item[T])
		u := curr.node
		if u == end {
			return reconstructPath(previous, end)
		}

		for v, weight := range adjacency[u] {
			alt := gScore[u] + weight
			if alt < gScore[v] {
				previous[v] = u
				gScore[v] = alt
				fScore[v] = alt + heuristicFunc(v, end)
				if !contains(openSet, v) {
					heap.Push(openSet, item[T]{node: v, priority: fScore[v]})
				}
			}
		}

	}

	return nil
}

func reconstructPath[T comparable](previous map[T]T, end T) []T {
	path := []T{}
	at := end
	for {
		prevAt, exists := previous[at]
		if !exists {
			break
		}
		path = append([]T{at}, path...)
		at = prevAt
	}

	return path
}

func contains[T comparable](h *minHeap[T], node T) bool {
	for _, item := range *h {
		if item.node == node {
			return true
		}
	}
	return false
}

// GenerateHeuristicMap computes the shortest path distances between all pairs of nodes using Dijkstra.
// Returns a map: heuristic[a][b] = shortest distance from a to b.
func GenerateHeuristicMap[T comparable](adjacency AdjacencyMap[T]) map[T]map[T]float64 {
	heuristic := make(map[T]map[T]float64, len(adjacency))
	for a := range adjacency {
		// Dijkstra from a to all others
		dist := make(map[T]float64)
		for node := range adjacency {
			dist[node] = maxDistance
		}
		dist[a] = 0

		visited := make(map[T]bool)
		type item struct {
			node     T
			priority float64
		}
		h := &[]item{{node: a, priority: 0}}
		for len(*h) > 0 {
			// Find min priority
			minIdx := 0
			for i := 1; i < len(*h); i++ {
				if (*h)[i].priority < (*h)[minIdx].priority {
					minIdx = i
				}
			}
			curr := (*h)[minIdx]
			*h = append((*h)[:minIdx], (*h)[minIdx+1:]...)
			u := curr.node
			if visited[u] {
				continue
			}
			visited[u] = true
			for v, weight := range adjacency[u] {
				if visited[v] {
					continue
				}
				alt := dist[u] + weight
				if alt < dist[v] {
					dist[v] = alt
					*h = append(*h, item{node: v, priority: alt})
				}
			}
		}
		heuristic[a] = dist
	}
	return heuristic
}
