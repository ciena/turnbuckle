package graph

type prioQueue struct {
	heap   []string
	index  map[string]int
	lessFn func(n1, n2 string) bool
}

func emptyPrioQueue(less func(n1, n2 string) bool) *prioQueue {
	index := make(map[string]int)

	return &prioQueue{index: index, lessFn: less}
}

func (p *prioQueue) len() int {
	return len(p.heap)
}

func (p *prioQueue) push(node string) {
	n := p.len()
	p.heap = append(p.heap, node)
	p.index[node] = n
	p.up(n)
}

func (p *prioQueue) pop() string {
	n := p.len() - 1
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

// fix as the cost of node has changed.
func (p *prioQueue) fix(node string) {
	if _, ok := p.index[node]; !ok {
		return
	}

	if i := p.index[node]; !p.down(i, p.len()) {
		p.up(i)
	}
}

func (p *prioQueue) up(child int) {
	for {
		parent := (child - 1) / two
		if parent == child || !p.less(child, parent) {
			break
		}
		// swap child with parent
		p.swap(parent, child)
		child = parent
	}
}

func (p *prioQueue) down(start, num int) bool {
	root := start

	for {
		child1 := two*root + 1
		if child1 >= num || child1 < 0 {
			break
		}

		child := child1
		if child2 := child1 + 1; child2 < num && p.less(child2, child1) {
			child = child2
		}

		if !p.less(child, root) {
			break
		}

		p.swap(root, child)
		root = child
	}

	return root > start
}
