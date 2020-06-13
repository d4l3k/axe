package axe

import (
	"log"
	"math"
	"math/rand"
	"sort"
)

type Node struct {
	Inputs  []int
	Outputs []int

	Cost int
}

type Edge struct {
	ID           int
	BoundaryCost int
}

func maxInts(a []int) int {
	if len(a) == 0 {
		return 0
	}
	max := a[0]
	for _, b := range a {
		if b > max {
			max = b
		}
	}
	return max
}

func sumInts(a []int) int {
	sum := 0
	for _, b := range a {
		sum += b
	}
	return sum
}

type Partitioning struct {
	Nodes      []Node
	EdgeCost   map[int]int
	Partitions []map[int]struct{}
}

func (p Partitioning) Cost() int {
	sources := map[int]int{}
	for group, nodes := range p.Partitions {
		for nodeId := range nodes {
			for _, output := range p.Nodes[nodeId].Outputs {
				sources[output] = group
			}
		}
	}

	costs := make([]int, len(p.Partitions))
	for group, nodes := range p.Partitions {
		for nodeId := range nodes {
			node := p.Nodes[nodeId]
			costs[group] += node.Cost
			for _, input := range node.Inputs {
				source := sources[input]
				if source != group {
					costs[group] += p.EdgeCost[input]
					costs[source] += p.EdgeCost[input]
				}
			}
		}
	}

	max := maxInts(costs)

	imbalance := 0
	for _, a := range costs {
		imbalance += max - a
	}

	totalCost := sumInts(costs)

	return totalCost + imbalance
}
func (p Partitioning) Move(id, from, to int) {
	delete(p.Partitions[from], id)
	p.Partitions[to][id] = struct{}{}
}

func (p Partitioning) PickOtherGroup(g int) int {
	randGroup := rand.Intn(len(p.Partitions))
	if randGroup == g {
		randGroup = (randGroup + 1) % len(p.Partitions)
	}
	return randGroup
}

func partitionMin(p map[int]struct{}) int {
	if len(p) == 0 {
		return 0
	}
	first := true
	min := 0
	for v := range p {
		if first {
			min = v
			first = false
		} else {
			if v < min {
				min = v
			}
		}
	}
	return min
}

func (p Partitioning) Normalize() {
	type entry struct {
		min       int
		partition map[int]struct{}
	}
	var sortable []entry
	for _, part := range p.Partitions {
		sortable = append(sortable, entry{
			min:       partitionMin(part),
			partition: part,
		})
	}
	sort.Slice(sortable, func(i, j int) bool {
		return sortable[i].min < sortable[j].min
	})
	for i, e := range sortable {
		p.Partitions[i] = e.partition
	}
}

func Optimize(nodes []Node, edgeCost map[int]int, n int, rounds int, initialTemperature float64) Partitioning {
	// Do an initial partition.
	p := Partitioning{
		Nodes:    nodes,
		EdgeCost: edgeCost,
	}
	partitionSize := len(nodes) / n
	for i := 0; i < n; i++ {
		p.Partitions = append(p.Partitions, map[int]struct{}{})
	}
	for i := 0; i < len(nodes); i++ {
		p.Partitions[i%partitionSize][i] = struct{}{}
	}

	// Do N rounds and then continue until we can't improve any more.
	improved := false
	curCost := p.Cost()
	for round := 0; (round < rounds) || improved; round++ {
		improved = false

		temperature := math.Max(initialTemperature*(1-float64(round)/float64(rounds)), 0)
		log.Printf("round %d: temp=%f cost=%d", round, temperature, curCost)

		for group, nodes := range p.Partitions {
			for nodeId := range nodes {
				target := p.PickOtherGroup(group)
				if rand.Float64() < temperature {
					p.Move(nodeId, group, target)
					curCost = p.Cost()
					continue
				}

				p.Move(nodeId, group, target)
				newCost := p.Cost()
				if newCost < curCost {
					improved = true
					curCost = newCost
				} else {
					// if new cost is higher move it back
					p.Move(nodeId, target, group)
				}
			}
		}
	}
	log.Printf("final cost %d", curCost)
	p.Normalize()
	return p
}
