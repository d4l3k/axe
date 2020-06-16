package axe

import (
	"log"
	"math"
	"math/rand"
	"sort"

	"github.com/pkg/errors"
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

type Partition struct {
	ExternalInputs []int
	Nodes          map[int]struct{}
	Fixed          bool
}

type Partitioning struct {
	Nodes      []Node
	EdgeCost   map[int]int
	Partitions []Partition
}

func (p Partitioning) Cost() (int, []int, error) {
	sources := map[int]int{}
	for group, partition := range p.Partitions {
		for nodeId := range partition.Nodes {
			for _, output := range p.Nodes[nodeId].Outputs {
				sources[output] = group
			}
		}
		for _, output := range partition.ExternalInputs {
			sources[output] = group
		}
	}

	costs := make([]int, len(p.Partitions))
	for group, partition := range p.Partitions {
		for _, output := range partition.ExternalInputs {
			costs[group] += p.EdgeCost[output]
		}
	}

	for nodeId, node := range p.Nodes {
		for group, partition := range p.Partitions {
			if _, ok := partition.Nodes[nodeId]; !ok {
				continue
			}
			costs[group] += node.Cost
			for _, input := range node.Inputs {
				source, ok := sources[input]
				if !ok {
					return 0, nil, errors.Errorf("failed to find source for input %d", input)
				}
				if source != group {
					costs[group] += p.EdgeCost[input]
					costs[source] += p.EdgeCost[input]
				}
			}
			for _, output := range node.Outputs {
				sources[output] = group
			}
		}
	}

	// Ignore costs from fixed nodes
	var filteredCosts []int
	for partitionId, cost := range costs {
		if !p.Partitions[partitionId].Fixed {
			filteredCosts = append(filteredCosts, cost)
		}
	}

	max := maxInts(filteredCosts)
	imbalance := 0
	for _, a := range filteredCosts {
		imbalance += max - a
	}

	totalCost := sumInts(filteredCosts)

	return totalCost + imbalance, costs, nil
}
func (p Partitioning) Move(id, from, to int) {
	delete(p.Partitions[from].Nodes, id)
	p.Partitions[to].Nodes[id] = struct{}{}
}

func (p Partitioning) PickOtherGroup(g int) int {
	var candidates []int
	for id, partition := range p.Partitions {
		if id == g || partition.Fixed {
			continue
		}
		candidates = append(candidates, id)
	}
	return candidates[rand.Intn(len(candidates))]
}

func (p Partition) MinID() int {
	if len(p.Nodes) == 0 {
		return 0
	}
	first := true
	min := 0
	for v := range p.Nodes {
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
		partition Partition
	}
	var sortable []entry
	for _, part := range p.Partitions {
		sortable = append(sortable, entry{
			min:       part.MinID(),
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

func MakePartitioning(nodes []Node, edgeCost map[int]int, n int) (Partitioning, error) {
	// Do an initial partition.
	p := Partitioning{
		Nodes:    nodes,
		EdgeCost: edgeCost,
	}
	partitionSize := len(nodes) / n
	for i := 0; i < n; i++ {
		p.Partitions = append(p.Partitions, Partition{Nodes: map[int]struct{}{}})
	}
	for i := 0; i < len(nodes); i++ {
		p.Partitions[i%partitionSize].Nodes[i] = struct{}{}
	}
	return p, nil
}

func (p Partitioning) Optimize(rounds int, initialTemperature float64) error {
	// Do N rounds and then continue until we can't improve any more.
	improved := false
	curCost, _, err := p.Cost()
	if err != nil {
		return err
	}
	for round := 0; (round < rounds) || improved; round++ {
		improved = false

		temperature := math.Max(initialTemperature*(1-float64(round)/float64(rounds)), 0)
		log.Printf("round %d: temp=%f cost=%d", round, temperature, curCost)

		for group, partition := range p.Partitions {
			if partition.Fixed {
				continue
			}

			for nodeId := range partition.Nodes {
				target := p.PickOtherGroup(group)
				if rand.Float64() < temperature {
					p.Move(nodeId, group, target)
					curCost, _, err = p.Cost()
					if err != nil {
						return err
					}
					continue
				}

				p.Move(nodeId, group, target)
				newCost, _, err := p.Cost()
				if err != nil {
					return err
				}
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
	return nil
}
