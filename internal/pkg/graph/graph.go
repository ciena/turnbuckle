package graph

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

type Graph struct {
	nodes []string
	edges map[string]map[string]int64
	log   logr.Logger
	sync.Mutex
}

type NodeAndCost struct {
	Node string
	Cost int64
}

type NodePeerCost struct {
	NodeAndCost
	Peer string
}

func NewGraph(nodes []string, log logr.Logger) *Graph {
	g := &Graph{nodes: nodes, edges: make(map[string]map[string]int64), log: log}
	// add the nodes as edges with no neighbors
	for _, n := range nodes {
		g.edges[n] = make(map[string]int64)
	}
	return g
}

func (g *Graph) GetNodesWithLock() []string {
	nodeCopy := make([]string, len(g.nodes))
	copy(nodeCopy, g.nodes)
	return nodeCopy
}

func (g *Graph) GetNodes() []string {
	g.Lock()
	defer g.Unlock()
	return g.GetNodesWithLock()
}

// just add a node edge with no neighbor.
func (g *Graph) AddNode(node string) {
	g.Lock()
	defer g.Unlock()
	if _, ok := g.edges[node]; !ok {
		g.edges[node] = make(map[string]int64)
		g.nodes = append(g.nodes, node)
	}
}

func (g *Graph) AddCostWithLock(vertex string, neighbor string, cost int64) {
	if _, ok := g.edges[vertex]; !ok {
		g.edges[vertex] = make(map[string]int64)
		g.nodes = append(g.nodes, vertex)
	}
	g.edges[vertex][neighbor] = cost
}

func (g *Graph) AddCost(vertex string, neighbor string, cost int64) {
	g.Lock()
	defer g.Unlock()
	g.AddCostWithLock(vertex, neighbor, cost)
}

func (g *Graph) AddCostBoth(vertex string, neighbor string, cost int64) {
	g.Lock()
	defer g.Unlock()
	g.AddCostWithLock(vertex, neighbor, cost)
	g.AddCostWithLock(neighbor, vertex, cost)
}

func (g *Graph) Remove(vertex string, neighbor string) {
	g.Lock()
	defer g.Unlock()
	if _, ok := g.edges[vertex]; ok {
		delete(g.edges[vertex], neighbor)
	}
}

func (g *Graph) ShortestPath(node string, neighbor string, cost int64) (path []string, dist int64, err error) {
	path = []string{}
	g.Lock()
	defer g.Unlock()
	parents, distances := g.ShortestPathsWithLock(node, cost)
	dist, ok := distances[neighbor]
	if !ok {
		err = fmt.Errorf("No path found from vertex %s to %s", node, neighbor)
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

// breadth-first order traversal calling do function for every unvisited vertex w.
func (g *Graph) BFSWithLock(vertex string, do func(v, w string, c int64)) {
	visited := make(map[string]struct{})
	visited[vertex] = struct{}{}
	for queue := []string{vertex}; len(queue) > 0; {
		v := queue[0]
		queue = queue[1:]
		g.VisitWithLock(v, func(w string, c int64) (skip bool) {
			if _, ok := visited[w]; ok {
				return
			}
			do(v, w, c)
			visited[w] = struct{}{}
			queue = append(queue, w)
			return
		})
	}
}

func (g *Graph) BFS(vertex string, do func(v, w string, c int64)) {
	g.Lock()
	defer g.Unlock()
	g.BFSWithLock(vertex, do)
}

// depth first traversal.
func (g *Graph) dfsVisit(vertex string, visitedMap map[string]struct{}, do func(v, w string, c int64)) {
	g.VisitWithLock(vertex, func(w string, c int64) (skip bool) {
		if _, ok := visitedMap[w]; ok {
			return
		}
		visitedMap[w] = struct{}{}
		g.dfsVisit(w, visitedMap, do)
		do(vertex, w, c)
		return
	})
}

func (g *Graph) DFSWithLock(vertex string, do func(v, w string, c int64)) {
	visitedMap := make(map[string]struct{})
	visitedMap[vertex] = struct{}{}
	g.dfsVisit(vertex, visitedMap, do)
}

func (g *Graph) DFS(vertex string, do func(v, w string, c int64)) {
	g.Lock()
	defer g.Unlock()
	g.DFSWithLock(vertex, do)
}

func lessThan(cost map[string]int64, request int64) func(n1, n2 string) bool {
	// n1 and n2 map to heap values at indices i and j
	return func(n1, n2 string) bool {
		c1, c2 := cost[n1], cost[n2]
		graded_c1 := request - c1
		if graded_c1 < 0 {
			graded_c1 = -graded_c1
		}
		graded_c2 := request - c2
		if graded_c2 < 0 {
			graded_c2 = -graded_c2
		}
		return graded_c1 < graded_c2
	}
}

func (g *Graph) ShortestPathsWithLock(node string, request int64) (parents map[string]string, dist map[string]int64) {
	parents = make(map[string]string)
	dist = make(map[string]int64)
	for _, n := range g.nodes {
		parents[n] = ""
	}
	less := lessThan(dist, request)
	q := EmptyPrioQueue(less)
	dist[node] = 0
	parents[node] = ""
	q.Push(node)
	for q.Len() > 0 {
		v := q.Pop()
		g.VisitWithLock(v, func(w string, c int64) (skip bool) {
			w_cost := dist[v] + c
			if current_w_cost, ok := dist[w]; !ok {
				parents[w] = v
				dist[w] = w_cost
				q.Push(w)
			} else {
				// fix if new cost < current cost
				new_graded_cost := request - w_cost
				if new_graded_cost < 0 {
					new_graded_cost = -new_graded_cost
				}
				current_graded_cost := request - current_w_cost
				if current_graded_cost < 0 {
					current_graded_cost = -current_graded_cost
				}
				if new_graded_cost < current_graded_cost {
					p := ""
					// check for cycles
					for p = parents[v]; p != "" && p != w; p = parents[p] {
					}
					if p != w {
						parents[w] = v
						dist[w] = w_cost
						q.Fix(w)
					}
				}
			}
			return
		})
	}
	parents[node] = ""
	return
}

func (g *Graph) ShortestPaths(node string, request int64) (parents map[string]string, dist map[string]int64) {
	g.Lock()
	defer g.Unlock()
	return g.ShortestPathsWithLock(node, request)
}

func (g *Graph) VisitWithLock(vertex string, do func(w string, c int64) bool) bool {
	if _, ok := g.edges[vertex]; !ok {
		return false
	}
	for neighbor, cost := range g.edges[vertex] {
		if do(neighbor, cost) {
			return true
		}
	}
	return false
}

func (g *Graph) Visit(vertex string, do func(w string, c int64) bool) bool {
	g.Lock()
	defer g.Unlock()
	return g.VisitWithLock(vertex, do)
}

// expected to be called with lock held.
func (g *Graph) Degree(vertex string) int {
	if _, ok := g.edges[vertex]; ok {
		return len(g.edges[vertex])
	}
	return 0
}

func (g *Graph) GetCostWithLock(vertex, neighbor string) (int64, error) {
	if neighborMap, ok := g.edges[vertex]; !ok {
		return -1, fmt.Errorf("Vertex %s not found in map", vertex)
	} else {
		if cost, ok := neighborMap[neighbor]; !ok {
			return -1, fmt.Errorf("Neighbor %s not found for vertex %s", neighbor, vertex)
		} else {
			return cost, nil
		}
	}
}

func (g *Graph) GetCost(vertex, neighbor string) (int64, error) {
	g.Lock()
	defer g.Unlock()
	return g.GetCostWithLock(vertex, neighbor)
}

// expected to be called with lock held.
func (g *Graph) Order() int {
	return len(g.nodes)
}

// if cost is <= request, add candidate map
// we add the nodes by the one that fits the following criteria
// request <= avail <= limit
// if the cost is < request, we take the one that is closest to request
// or min(request-avail)
// So we build the nodeandcost list and sort it with the graded cost in the map.
func (g *Graph) findNodeFromCandidatesFilter(candidates map[string]NodeAndCost, applyFilter func(nodeName string) bool) *NodePeerCost {
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

// euclidean distance.
func ComputeNodeCost(durations ...time.Duration) time.Duration {
	durationPairs := len(durations) / 2
	var sigma time.Duration
	for i, pair := 0, 0; pair < durationPairs; i, pair = i+2, pair+1 {
		diff := durations[i] - durations[i+1]
		diff_square := diff * diff
		sigma += diff_square
	}
	return sigma
}

func (g *Graph) findNode(vertexList []string, nodeList []string) (candidateMap map[string]NodeAndCost) {
	candidateMap = make(map[string]NodeAndCost)
	nodeMap := make(map[string]struct{})
	for _, n := range nodeList {
		nodeMap[n] = struct{}{}
	}
	g.Lock()
	defer g.Unlock()
	// vertexList has the destination endpoints to start from looking for the best candidate
	for _, v := range vertexList {
		g.VisitWithLock(v, func(w string, c int64) (skip bool) {
			// skip if it isn't part of the candidate nodeList
			if _, ok := nodeMap[w]; !ok {
				return
			}
			// we take the least graded cost
			if currentNodeCost, ok := candidateMap[w]; !ok {
				// add the peer info
				candidateMap[w] = NodeAndCost{Node: v, Cost: c}
			} else if c < currentNodeCost.Cost {
				candidateMap[w] = NodeAndCost{Node: v, Cost: c}
			}
			return
		})
	}
	return
}

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
	return nil, fmt.Errorf("Did not find any node from nodes %s for pod", neighbors)
}

// start the search for node using neighbor nodes as vertices.
func (g *Graph) GetNodeWithNeighborsAndCandidates(neighbors []string, candidateNodes []string) (*NodePeerCost, error) {
	return g.GetNodeWithNeighborsAndCandidatesFilter(neighbors, candidateNodes, nil)
}

func (g *Graph) GetNodeWithNeighborsFilter(neighbors []string, applyFilter func(nodeName string) bool) (*NodePeerCost, error) {
	return g.GetNodeWithNeighborsAndCandidatesFilter(neighbors, g.nodes, applyFilter)
}

func (g *Graph) GetNodeWithNeighbors(neighbors []string) (*NodePeerCost, error) {
	return g.GetNodeWithNeighborsAndCandidatesFilter(neighbors, g.nodes, nil)
}

// find the best node.
func (g *Graph) GetNode() (*NodePeerCost, error) {
	return g.GetNodeWithNeighbors(g.nodes)
}

func (g *Graph) IsFullyConnected(nodes []string) bool {
	g.Lock()
	defer g.Unlock()
	degreeMap := make(map[string]int)
	nodeMap := make(map[string]struct{})
	currentNodes := g.Order()
	numNodes := len(nodes)
	if currentNodes < numNodes {
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

func (g *Graph) WaitForNodesToBeConnected(nodes []string, done <-chan struct{}) (bool, <-chan struct{}) {
	if g.IsFullyConnected(nodes) {
		return true, nil
	}
	ch := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)
		defer close(ch)
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
	return false, ch
}
