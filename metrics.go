package main

import (
	"sort"
)

type Metrics struct {
	Throughput float64
	Delay      float64
	AoI        float64
}

// GetOneIndex efficiently finds the first successful transmission in each interval of size T.
// sIndex must be sorted.
func GetOneIndex(sIndex []int, T int) []int {
	if len(sIndex) == 0 {
		return nil
	}
	var oneIndex []int
	lastGroup := -1
	for _, val := range sIndex {
		group := val / T
		if group > lastGroup {
			oneIndex = append(oneIndex, val)
			lastGroup = group
		}
	}
	return oneIndex
}

// CalculateMetrics calculates Throughput, Delay, and AoI for a single user's success indices.
func CalculateMetrics(successIndices []int, T, pq int) Metrics {
	throughput := float64(len(successIndices))

	if len(successIndices) == 0 {
		// If no successful transmissions, return throughput = 0, delay = 0, AoI = 0
		// (Matching MATLAB logic where delayCount=0 skips delay addition)
		return Metrics{0, 0, 0}
	}

	// 1. Calculate Delay
	delaySum := 0.0
	delayCount := 0

	// We iterate through groups from g=1 to pq/T
	numGroups := pq / T
	lastGroup := -1

	// Since successIndices is sorted, we can find the first element of each group
	for _, val := range successIndices {
		group := val / T
		if group < numGroups && group > lastGroup {
			modValue := val % T
			if modValue == 0 {
				delaySum += 1.0
			} else {
				delaySum += float64(modValue + 1)
			}
			delayCount++
			lastGroup = group
		}
	}

	delayAvg := 0.0
	if delayCount > 0 {
		delayAvg = delaySum / float64(delayCount)
	}

	// 2. Calculate AoI
	oneIndex := GetOneIndex(successIndices, T)

	// Create extendedMap array for O(1) lookup
	// Using []bool is dramatically faster than map[int]bool for dense integer ranges
	extendedMap := make([]bool, 2*pq+1) // +1 to safely access indices up to 2*pq
	for _, val := range oneIndex {
		if val < len(extendedMap) {
			extendedMap[val] = true
		}
		if val+pq < len(extendedMap) {
			extendedMap[val+pq] = true // second_group
		}
	}

	extendedAoI := make([]float64, 2*pq)
	for d := 1; d <= 2*pq; d++ {
		// To match MATLAB exactly: it checks ismember(d, extended_Index) where d starts at 1
		if extendedMap[d] {
			extendedAoI[d-1] = float64((d % T) + 1)
		} else if d == 1 {
			extendedAoI[d-1] = 0
		} else {
			extendedAoI[d-1] = extendedAoI[d-2] + 1
		}
	}

	// Calculate sum of the second half
	aoiSum := 0.0
	for d := pq; d < 2*pq; d++ {
		aoiSum += extendedAoI[d]
	}
	aoiAvg := aoiSum / float64(pq)

	return Metrics{Throughput: throughput, Delay: delayAvg, AoI: aoiAvg}
}

// CalculateSystemMetrics processes all users in one run.
// It mutates successCounts to ensure it's sorted, then calculates metrics.
func CalculateSystemMetrics(successCounts [][]int, T, pq int) []Metrics {
	se := len(successCounts)
	res := make([]Metrics, se)
	for i := 0; i < se; i++ {
		// Must sort because ApplyRandomShifts disrupts the order, and MATLAB's find() is sorted.
		sort.Ints(successCounts[i])
		res[i] = CalculateMetrics(successCounts[i], T, pq)
	}
	return res
}
