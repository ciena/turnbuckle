package graph

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

var (
	// ErrNoPathFound no path found in graph.
	ErrNoPathFound = errors.New("no path found")

	// ErrNotFound node not found in graph.
	ErrNotFound = errors.New("node not found")

	// ErrEdgeNotFound edge not found in graph.
	ErrEdgeNotFound = errors.New("edge not found")
)

const (
	// disable magic number lint error.
	two = 2
)

// Graph of node->neighbors with cost.
type Graph struct {
	nodes []string
	edges map[string]map[string]int64
	log   logr.Logger
	sync.Mutex
}

// NodeAndCost node and its cost.
type NodeAndCost struct {
	Node string
	Cost int64
}

// NodePeerCost node, cost along with its peer.
type NodePeerCost struct {
	NodeAndCost
	Peer string
}

// NewGraph instantiate a new graph for the nodes.
func NewGraph(nodes []string, log logr.Logger) *Graph {
	graph := &Graph{nodes: nodes, edges: make(map[string]map[string]int64), log: log}
	// add the nodes as edges with no neighbors
	for _, n := range nodes {
		graph.edges[n] = make(map[string]int64)
	}

	return graph
}

// GetNodesWithLock get nodes of the graph. called with lock held.
func (g *Graph) GetNodesWithLock() []string {
	nodeCopy := make([]string, len(g.nodes))
	copy(nodeCopy, g.nodes)

	return nodeCopy
}

// GetNodes get nodes of the graph.
func (g *Graph) GetNodes() []string {
	g.Lock()
	defer g.Unlock()

	return g.GetNodesWithLock()
}

// AddNode just add a node edge with no neighbor.
func (g *Graph) AddNode(node string) {
	g.Lock()
	defer g.Unlock()

	if _, ok := g.edges[node]; !ok {
		g.edges[node] = make(map[string]int64)
		g.nodes = append(g.nodes, node)
	}
}

// AddCostWithLock add a cost for neighbor with graph lock held.
func (g *Graph) AddCostWithLock(vertex string, neighbor string, cost int64) {
	if _, ok := g.edges[vertex]; !ok {
		g.edges[vertex] = make(map[string]int64)
		g.nodes = append(g.nodes, vertex)
	}

	g.edges[vertex][neighbor] = cost
}

// AddCost  add a neighbor to vertex with cost.
func (g *Graph) AddCost(vertex string, neighbor string, cost int64) {
	g.Lock()
	defer g.Unlock()
	g.AddCostWithLock(vertex, neighbor, cost)
}

// AddCostBoth add a cost from vertex to neighbor in both directions.
func (g *Graph) AddCostBoth(vertex string, neighbor string, cost int64) {
	g.Lock()
	defer g.Unlock()
	g.AddCostWithLock(vertex, neighbor, cost)
	g.AddCostWithLock(neighbor, vertex, cost)
}

// Remove remove a neighbor from the graph.
func (g *Graph) Remove(vertex string, neighbor string) {
	g.Lock()
	defer g.Unlock()

	if _, ok := g.edges[vertex]; ok {
		delete(g.edges[vertex], neighbor)
	}
}

// NoPathFoundError (from,to).
func NoPathFoundError(from, to string) error {
	return fmt.Errorf("NoPathFound %w: from vertex: %s to vertex: %s", ErrNoPathFound, from, to)
}

// ShortestPath find the shortest path from node to neighbor.
func (g *Graph) ShortestPath(node string, neighbor string, cost int64) (path []string, dist int64, err error) {
	path = []string{}

	g.Lock()
	defer g.Unlock()

	parents, distances := g.ShortestPathsWithLock(node, cost)

	dist, ok := distances[neighbor]
	if !ok {
		err = NoPathFoundError(node, neighbor)

		return
	}

	for v := neighbor; v != ""; v = parents[v] {
		path = append(path, v)
	}

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return
}

// BFSWithLock breadth-first order traversal calling do function for every unvisited vertex w.
func (g *Graph) BFSWithLock(vertex string, visit func(v, w string, c int64)) {
	visited := make(map[string]struct{})
	visited[vertex] = struct{}{}

	for queue := []string{vertex}; len(queue) > 0; {
		node := queue[0]
		queue = queue[1:]

		g.VisitWithLock(node, func(neighbor string, c int64) (skip bool) {
			if _, ok := visited[neighbor]; ok {
				return
			}
			visit(node, neighbor, c)
			visited[neighbor] = struct{}{}
			queue = append(queue, neighbor)

			return
		})
	}
}

// BFS breadth-first order traversal calling do function for every unvisited vertex w.
func (g *Graph) BFS(vertex string, do func(v, w string, c int64)) {
	g.Lock()
	defer g.Unlock()
	g.BFSWithLock(vertex, do)
}

func (g *Graph) dfsVisit(vertex string, visitedMap map[string]struct{}, visit func(v, w string, c int64)) {
	g.VisitWithLock(vertex, func(neighbor string, cost int64) (skip bool) {
		if _, ok := visitedMap[neighbor]; ok {
			return
		}
		visitedMap[neighbor] = struct{}{}
		g.dfsVisit(neighbor, visitedMap, visit)
		visit(vertex, neighbor, cost)

		return
	})
}

// DFSWithLock depth first traversal.
func (g *Graph) DFSWithLock(vertex string, do func(v, w string, c int64)) {
	visitedMap := make(map[string]struct{})
	visitedMap[vertex] = struct{}{}
	g.dfsVisit(vertex, visitedMap, do)
}

// DFS depth first traversal.
func (g *Graph) DFS(vertex string, do func(v, w string, c int64)) {
	g.Lock()
	defer g.Unlock()
	g.DFSWithLock(vertex, do)
}

func abs(val int64) int64 {
	if val < 0 {
		return -val
	}

	return val
}

func lessThan(cost map[string]int64, request int64) func(n1, n2 string) bool {
	// n1 and n2 map to heap values at indices i and j
	return func(n1, n2 string) bool {
		cost1, cost2 := cost[n1], cost[n2]

		gradedCost1 := abs(request - cost1)
		gradedCost2 := abs(request - cost2)

		return gradedCost1 < gradedCost2
	}
}

// ShortestPathsWithLock find all the vertices from node along with parents and cost. called with lock held.
func (g *Graph) ShortestPathsWithLock(node string, request int64) (map[string]string, map[string]int64) {
	parents := make(map[string]string)
	dist := make(map[string]int64)

	for _, n := range g.nodes {
		parents[n] = ""
	}

	less := lessThan(dist, request)
	queue := emptyPrioQueue(less)

	dist[node] = 0
	parents[node] = ""

	queue.push(node)

	for queue.len() > 0 {
		vertex := queue.pop()
		g.VisitWithLock(vertex, func(neighbor string, c int64) (skip bool) {
			neighborCost := dist[vertex] + c
			currentCost, present := dist[neighbor]
			if !present {
				parents[neighbor] = vertex
				dist[neighbor] = neighborCost
				queue.push(neighbor)

				return
			}

			// fix if new cost < current cost
			newGradedCost := abs(request - neighborCost)
			currentGradedCost := abs(request - currentCost)
			if newGradedCost < currentGradedCost {
				var p string

				// check for cycles
				for p = parents[vertex]; p != "" && p != neighbor; p = parents[p] {
				}

				if p != neighbor {
					parents[neighbor] = vertex
					dist[neighbor] = neighborCost
					queue.fix(neighbor)
				}
			}

			return
		})
	}

	parents[node] = ""

	return parents, dist
}

// ShortestPaths find all vertices from node along with their cost and parents.
func (g *Graph) ShortestPaths(node string, request int64) (map[string]string, map[string]int64) {
	g.Lock()
	defer g.Unlock()

	return g.ShortestPathsWithLock(node, request)
}

// VisitWithLock visit all neighbors for given vertex calling do function for each.
func (g *Graph) VisitWithLock(vertex string, visit func(w string, c int64) bool) bool {
	if _, ok := g.edges[vertex]; !ok {
		return false
	}

	for neighbor, cost := range g.edges[vertex] {
		if visit(neighbor, cost) {
			return true
		}
	}

	return false
}

// Visit visit all neighbors for given vertex calling do function for each.
func (g *Graph) Visit(vertex string, do func(w string, c int64) bool) bool {
	g.Lock()
	defer g.Unlock()

	return g.VisitWithLock(vertex, do)
}

// Degree expected to be called with lock held.
func (g *Graph) Degree(vertex string) int {
	if _, ok := g.edges[vertex]; ok {
		return len(g.edges[vertex])
	}

	return 0
}

// NotFoundError  for node.
func NotFoundError(node string) error {
	return fmt.Errorf("NotFound %w: %s", ErrNotFound, node)
}

// EdgeNotFoundError for (v,w).
func EdgeNotFoundError(v, w string) error {
	return fmt.Errorf("EdgeNotFound %w: from: %s, to: %s", ErrEdgeNotFound, v, w)
}

// GetCostWithLock get the cost of neighbor from vertex.
func (g *Graph) GetCostWithLock(vertex, neighbor string) (int64, error) {
	neighborMap, present := g.edges[vertex]
	if !present {
		return -1, NotFoundError(vertex)
	}

	cost, present := neighborMap[neighbor]
	if !present {
		return -1, EdgeNotFoundError(vertex, neighbor)
	}

	return cost, nil
}

// GetCost get the cost of neighbor from vertex.
func (g *Graph) GetCost(vertex, neighbor string) (int64, error) {
	g.Lock()
	defer g.Unlock()

	return g.GetCostWithLock(vertex, neighbor)
}

// Order expected to be called with lock held.
func (g *Graph) Order() int {
	return len(g.nodes)
}

// if cost is <= request, add candidate map
// we add the nodes by the one that fits the following criteria
// request <= avail <= limit
// if the cost is < request, we take the one that is closest to request
// or min(request-avail)
// So we build the nodeandcost list and sort it with the graded cost in the map.
func (g *Graph) findNodeFromCandidatesFilter(candidates map[string]NodeAndCost,
	applyFilter func(nodeName string) bool) *NodePeerCost {
	if len(candidates) == 0 {
		return nil
	}

	nodes := make([]NodePeerCost, 0, len(candidates))

	for node, peerCost := range candidates {
		nodes = append(nodes, NodePeerCost{Peer: peerCost.Node, NodeAndCost: NodeAndCost{Node: node, Cost: peerCost.Cost}})
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Cost < nodes[j].Cost
	})

	for i := range nodes {
		if applyFilter != nil && !applyFilter(nodes[i].Node) {
			continue
		}

		return &nodes[i]
	}

	return nil
}

// ComputeNodeCost euclidean distance.
func ComputeNodeCost(durations ...time.Duration) time.Duration {
	var sigma int64

	durationPairs := len(durations) / two
	for i, pair := 0, 0; pair < durationPairs; i, pair = i+two, pair+1 {
		diff := int64(durations[i]) - int64(durations[i+1])
		diffSquare := diff * diff
		sigma += diffSquare
	}

	return time.Duration(sigma)
}

func (g *Graph) findNode(vertexList []string, nodeList []string) map[string]NodeAndCost {
	candidateMap := make(map[string]NodeAndCost)
	nodeMap := make(map[string]struct{})

	for _, n := range nodeList {
		nodeMap[n] = struct{}{}
	}

	g.Lock()
	defer g.Unlock()

	// vertexList has the destination endpoints to start from looking for the best candidate
	for _, vertex := range vertexList {
		g.VisitWithLock(vertex, func(neighbor string, cost int64) (skip bool) {
			// skip if it isn't part of the candidate nodeList
			if _, ok := nodeMap[neighbor]; !ok {
				return
			}
			// we take the least graded cost
			if currentNodeCost, ok := candidateMap[neighbor]; !ok {
				// add the peer info
				candidateMap[neighbor] = NodeAndCost{Node: vertex, Cost: cost}
			} else if cost < currentNodeCost.Cost {
				candidateMap[neighbor] = NodeAndCost{Node: vertex, Cost: cost}
			}

			return
		})
	}

	return candidateMap
}

// GetNodeWithNeighborsAndCandidatesFilter get the node with the best cost given its neighbors and eligible nodes.
func (g *Graph) GetNodeWithNeighborsAndCandidatesFilter(neighbors []string, candidateNodes []string,
	applyFilter func(nodeName string) bool) (*NodePeerCost, error) {
	if len(neighbors) == 0 {
		neighbors = g.nodes
	}

	if len(candidateNodes) == 0 {
		candidateNodes = g.nodes
	}

	candidates := g.findNode(neighbors, candidateNodes)
	if len(candidates) > 0 {
		g.log.V(1).Info("pod-candidates", "candidates", candidates)

		node := g.findNodeFromCandidatesFilter(candidates, applyFilter)
		if node != nil {
			g.log.V(1).Info("candidate-assignment", "node", node.Node, "cost", node.Cost, "peer", node.Peer)

			return node, nil
		}
	}

	return nil, ErrNotFound
}

// GetNodeWithNeighborsAndCandidates get node with best cost using neighbor nodes as vertices.
func (g *Graph) GetNodeWithNeighborsAndCandidates(neighbors []string, candidateNodes []string) (*NodePeerCost, error) {
	return g.GetNodeWithNeighborsAndCandidatesFilter(neighbors, candidateNodes, nil)
}

// GetNodeWithNeighborsFilter get node with best cost using neighbor nodes as vertices.
func (g *Graph) GetNodeWithNeighborsFilter(neighbors []string,
	applyFilter func(nodeName string) bool) (*NodePeerCost, error) {
	return g.GetNodeWithNeighborsAndCandidatesFilter(neighbors, g.nodes, applyFilter)
}

// GetNodeWithNeighbors get node with best cost using neighbor nodes as vertices.
func (g *Graph) GetNodeWithNeighbors(neighbors []string) (*NodePeerCost, error) {
	return g.GetNodeWithNeighborsAndCandidatesFilter(neighbors, g.nodes, nil)
}

// GetNode find the best node.
func (g *Graph) GetNode() (*NodePeerCost, error) {
	return g.GetNodeWithNeighbors(g.nodes)
}

// IsFullyConnected check if the graph is fully connected.
func (g *Graph) IsFullyConnected(nodes []string) bool {
	g.Lock()
	defer g.Unlock()

	degreeMap := make(map[string]int)
	nodeMap := make(map[string]struct{})

	numNodes := len(nodes)
	if currentNodes := g.Order(); currentNodes < numNodes {
		return false
	}

	targetDegree := numNodes * (numNodes - 1)

	for _, n := range nodes {
		nodeMap[n] = struct{}{}
	}

	for _, v := range nodes {
		g.VisitWithLock(v, func(w string, cost int64) (skip bool) {
			// check if the neighbor is part of the node map
			if _, ok := nodeMap[w]; ok {
				degreeMap[v]++
			}

			return
		})
	}

	degrees := 0
	for v := range degreeMap {
		degrees += degreeMap[v]
	}

	g.log.V(1).Info("degrees", "connected", degrees, "target", targetDegree, "order", g.Order())

	return degrees >= targetDegree
}

// WaitForNodesToBeConnected wait for the nodes in the graph to be connected.
func (g *Graph) WaitForNodesToBeConnected(nodes []string, done <-chan struct{}) (bool, <-chan struct{}) {
	if g.IsFullyConnected(nodes) {
		return true, nil
	}

	waitch := make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Second)

		defer close(waitch)
		defer ticker.Stop()

		for {
			if g.IsFullyConnected(nodes) {
				g.log.V(1).Info("graph-fully-connected")

				return
			}

			select {
			case <-ticker.C:
			case <-done:
				return
			}
		}
	}()

	return false, waitch
}
