package axe

import (
	"reflect"
	"sort"
	"testing"
)

func partitionsEqual(t *testing.T, p Partitioning, want [][]int) {
	t.Logf("partitions %+v", p.Partitions)
	if len(p.Partitions) != 3 {
		t.Fatalf("expected %d partitions; got %d", len(want), len(p.Partitions))
	}
	for g, ids := range want {
		var idList []int
		for id := range p.Partitions[g].Nodes {
			idList = append(idList, id)
		}

		sort.Ints(idList)
		sort.Ints(ids)
		if !reflect.DeepEqual(ids, idList) {
			t.Fatalf("group %d: wanted %+v; got %+v", g, ids, idList)
		}
	}
}

func TestSimple(t *testing.T) {
	nodes := []Node{
		{
			Inputs:  nil,
			Outputs: []int{1},
			Cost:    10,
		},
		{
			Inputs:  []int{1},
			Outputs: []int{2},
			Cost:    10,
		},
		{
			Inputs:  []int{2},
			Outputs: nil,
			Cost:    10,
		},
	}
	edgeCost := map[int]int{
		1: 1,
		2: 1,
	}
	p, err := MakePartitioning(nodes, edgeCost, 3)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Optimize(100, 0.1); err != nil {
		t.Fatal(err)
	}
	p.Normalize()
	partitionsEqual(t, p, [][]int{
		[]int{0},
		[]int{1},
		[]int{2},
	})
}

func TestExternalInputs(t *testing.T) {
	p := Partitioning{
		EdgeCost: map[int]int{
			1: 1,
			2: 2,
		},
		Partitions: []Partition{
			{
				ExternalInputs: []int{1, 2},
			},
			{
				Nodes: map[int]struct{}{0: struct{}{}},
			},
		},
		Nodes: []Node{
			{Inputs: []int{1}},
		},
	}
	cost, _, err := p.Cost()
	if err != nil {
		t.Fatal(err)
	}
	// Partition 0: 1 + 2 + 1 = 4
	// Partition 1: 1 = 1
	// Total cost = 5
	// Imbalance = 3
	// Final = Total Cost + Imbalance = 8
	want := 8
	if cost != want {
		t.Fatalf("expected cost = %d; got %d", want, cost)
	}
}

func TestMutations(t *testing.T) {
	p := Partitioning{
		EdgeCost: map[int]int{
			1: 1,
		},
		Partitions: []Partition{
			{
				Nodes: map[int]struct{}{0: struct{}{}},
			},
			{
				Nodes: map[int]struct{}{1: struct{}{}},
			},
			{
				Nodes: map[int]struct{}{2: struct{}{}},
			},
		},
		Nodes: []Node{
			{Outputs: []int{1}},
			{Inputs: []int{1}, Outputs: []int{1}},
			{Inputs: []int{1}},
		},
	}
	cost, _, err := p.Cost()
	if err != nil {
		t.Fatal(err)
	}
	// Partition 0: 1
	// Partition 1: 1 + 1
	// Partition 2: 1
	// Total cost = 4
	// Imbalance = 2
	// Final = Total Cost + Imbalance = 6
	want := 6
	if cost != want {
		t.Fatalf("expected cost = %d; got %d", want, cost)
	}
}
