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
		for id := range p.Partitions[g] {
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
	p := Optimize(nodes, edgeCost, 3, 100, 0.1)
	partitionsEqual(t, p, [][]int{
		[]int{0},
		[]int{1},
		[]int{2},
	})
}
