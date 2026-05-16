package main

// SimulateRun performs one iteration of the Monte Carlo simulation.
// It calculates the collisions, applies the MPR (Multi-Packet Reception) rule,
// and returns the successful transmission indices for each sequence.
func SimulateRun(shiftedSgSet [][]int, r int, pq int) [][]int {
	se := len(shiftedSgSet)

	// 1. Count occurrences of each slot (equivalent to addSg in MATLAB)
	slotCounts := make([]int, pq)
	for _, seqIndices := range shiftedSgSet {
		for _, idx := range seqIndices {
			slotCounts[idx]++
		}
	}

	// 2. Identify valid slots where 0 < count <= r (equivalent to rboth in MATLAB)
	validSlots := make([]bool, pq)
	for idx, count := range slotCounts {
		if count > 0 && count <= r {
			validSlots[idx] = true
		}
	}

	// 3. Find successful transmissions for each user (equivalent to intersect in MATLAB)
	// Since we only need to check if a user's transmission slot is valid,
	// we just filter their seqIndices array.
	sgSuccessCounts := make([][]int, se)
	for i, seqIndices := range shiftedSgSet {
		// Pre-allocate to optimize memory (capacity = len(seqIndices), length = 0)
		success := make([]int, 0, len(seqIndices))
		for _, idx := range seqIndices {
			if validSlots[idx] {
				success = append(success, idx)
			}
		}
		sgSuccessCounts[i] = success
	}

	return sgSuccessCounts
}
