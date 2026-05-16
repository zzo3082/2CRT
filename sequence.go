package main

import (
	"math/rand"
)

type Pair struct {
	x, y int
}

// GenerateSequences creates the CRT sequence sets returning only the indices where the value is 1.
func GenerateSequences(p, q, w int) [][]int {
	SgSet := make([][]int, p+1)
	pq := p * q

	// Generate for g = 0 to p-1
	for g := 0; g < p; g++ {
		gSetMap := make(map[Pair]bool)
		for j := 0; j < w; j++ {
			gSetMap[Pair{(g * j) % p, j % q}] = true
		}

		var indices []int
		for t := 0; t < pq; t++ {
			st := Pair{t % p, t % q}
			if gSetMap[st] {
				indices = append(indices, t)
			}
		}
		SgSet[g] = indices
	}

	// Generate for g = p
	gSetMap := make(map[Pair]bool)
	for j := 0; j < w; j++ {
		gSetMap[Pair{j % p, 0 % q}] = true
	}

	var indices []int
	for t := 0; t < pq; t++ {
		st := Pair{t % p, t % q}
		if gSetMap[st] {
			indices = append(indices, t)
		}
	}
	SgSet[p] = indices

	return SgSet
}

// ApplyRandomShifts applies random circular shifts to the selected sequences.
func ApplyRandomShifts(sgSet [][]int, sequences []int, pq int, rng *rand.Rand) [][]int {
	se := len(sequences)
	randomSgSet := make([][]int, se)

	for i, g := range sequences {
		// g in MATLAB is 1-indexed, so we use g-1 for 0-indexed SgSet
		seqIndices := sgSet[g-1]

		shift := rng.Intn(pq + 1) // randi([0, p*q]) in MATLAB includes p*q

		shiftedIndices := make([]int, len(seqIndices))
		for j, idx := range seqIndices {
			// Circular shift right
			shiftedIndices[j] = (idx + shift) % pq
		}
		randomSgSet[i] = shiftedIndices
	}

	return randomSgSet
}
