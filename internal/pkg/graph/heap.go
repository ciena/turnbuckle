package graph

type prioQueue struct {
	heap   []string
	index  map[string]int
	lessFn func(n1, n2 string) bool
}

// create a priority queue with a set of nodes
func NewPrioQueue(nodes []string, less func(n1, n2 string) bool) *prioQueue {
	heap := make([]string, len(nodes))
	index := make(map[string]int)
	for i := range nodes {
		heap[i] = nodes[i]
		index[heap[i]] = i
	}
	return &prioQueue{heap: heap, index: index, lessFn: less}
}

// create an empty priority queue
func EmptyPrioQueue(less func(n1, n2 string) bool) *prioQueue {
	index := make(map[string]int)
	return &prioQueue{index: index, lessFn: less}
}

// get the length of the priority queue
func (p *prioQueue) Len() int {
	return len(p.heap)
}

// push a node into the priority queue
func (p *prioQueue) Push(node string) {
	n := p.Len()
	p.heap = append(p.heap, node)
	p.index[node] = n
	p.up(n)
}

// pop a node from the priority queue
func (p *prioQueue) Pop() string {
	n := p.Len() - 1
	p.swap(0, n)
	p.down(0, n)
	v := p.heap[n]
	delete(p.index, v)
	p.heap = p.heap[:n]
	return v
}

func (p *prioQueue) less(i, j int) bool {
	return p.lessFn(p.heap[i], p.heap[j])
}

func (p *prioQueue) swap(i, j int) {
	p.heap[i], p.heap[j] = p.heap[j], p.heap[i]
	p.index[p.heap[i]] = i
	p.index[p.heap[j]] = j
}

// check if node is part of the priority queue
func (p *prioQueue) Contains(node string) bool {
	if _, ok := p.index[node]; !ok {
		return false
	}
	return true
}

// fix as the cost of node has changed.
func (p *prioQueue) Fix(node string) {
	if _, ok := p.index[node]; !ok {
		return
	}
	if i := p.index[node]; !p.down(i, p.Len()) {
		p.up(i)
	}
}

func (p *prioQueue) up(j int) {
	for {
		i := (j - 1) / 2
		if i == j || !p.less(j, i) {
			break
		}
		// swap j with parent
		p.swap(i, j)
		j = i
	}
}

func (p *prioQueue) down(i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 {
			break
		}
		j := j1
		if j2 := j1 + 1; j2 < n && p.less(j2, j1) {
			j = j2
		}
		if !p.less(j, i) {
			break
		}
		p.swap(i, j)
		i = j
	}
	return i > i0
}
