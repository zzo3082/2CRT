package main

import (
	"fmt"
	"math/bits"
	"sync"
	"time"
)

func main() {
	p := 5
	pq := p * p // Period is 25

	// 1. Define the 6 prime sequences (characteristic sets)
	rawSeqs := [][]int{
		{0, 1, 2, 3, 4},    // Seq 0 (g=0)
		{0, 6, 12, 18, 24}, // Seq 1 (g=1)
		{0, 8, 11, 19, 22}, // Seq 2 (g=2)
		{0, 7, 14, 16, 23}, // Seq 3 (g=3)
		{0, 9, 13, 17, 21}, // Seq 4 (g=4)
		{0, 5, 10, 15, 20}, // Seq 5 (g=p)
	}

	// 2. Precalculate shifts as bitmasks (uint32)
	var masks [6][25]uint32
	for i := 0; i < 6; i++ {
		for d := 0; d < 25; d++ {
			var mask uint32
			for _, x := range rawSeqs[i] {
				shifted := (x + d) % pq
				mask |= 1 << shifted
			}
			masks[i][d] = mask
		}
	}

	fmt.Println("=================================================================")
	fmt.Println("Starting traversal of all independent 25^5 = 9,765,625 states...")
	fmt.Println("  (Fixing User 0's shift d0 = 0 as the reference to avoid self-redundancy)")
	fmt.Println("=================================================================")
	startTime := time.Now()

	// 3. Parallel traversal using Goroutines over d1
	var wg sync.WaitGroup
	wg.Add(25)

	var mu sync.Mutex
	globalMaxCollisionDistribution := make([]int64, 6)
	globalUserCollisionDistribution := make([][6]int64, 6)

	// User 0's shift is fixed at 0 to avoid relative redundance (no self-reference shifts)
	d0Val := 0
	m0 := masks[0][d0Val]

	for d1 := 0; d1 < 25; d1++ {
		go func(d1Val int) {
			defer wg.Done()

			localMaxColDist := make([]int64, 6)
			localUserColDist := make([][6]int64, 6)

			m1 := masks[1][d1Val]

			// Loops for d2, d3, d4, d5
			for d2 := 0; d2 < 25; d2++ {
				m2 := masks[2][d2]

				for d3 := 0; d3 < 25; d3++ {
					m3 := masks[3][d3]

					for d4 := 0; d4 < 25; d4++ {
						m4 := masks[4][d4]

						for d5 := 0; d5 < 25; d5++ {
							m5 := masks[5][d5]

							// Compute collision count for each user (self is excluded from interference)
							c0 := bits.OnesCount32(m0 & (m1 | m2 | m3 | m4 | m5))
							c1 := bits.OnesCount32(m1 & (m0 | m2 | m3 | m4 | m5))
							c2 := bits.OnesCount32(m2 & (m0 | m1 | m3 | m4 | m5))
							c3 := bits.OnesCount32(m3 & (m0 | m1 | m2 | m4 | m5))
							c4 := bits.OnesCount32(m4 & (m0 | m1 | m2 | m3 | m5))
							c5 := bits.OnesCount32(m5 & (m0 | m1 | m2 | m3 | m4))

							localUserColDist[0][c0]++
							localUserColDist[1][c1]++
							localUserColDist[2][c2]++
							localUserColDist[3][c3]++
							localUserColDist[4][c4]++
							localUserColDist[5][c5]++

							maxCol := c0
							if c1 > maxCol {
								maxCol = c1
							}
							if c2 > maxCol {
								maxCol = c2
							}
							if c3 > maxCol {
								maxCol = c3
							}
							if c4 > maxCol {
								maxCol = c4
							}
							if c5 > maxCol {
								maxCol = c5
							}

							localMaxColDist[maxCol]++
						}
					}
				}
			}

			mu.Lock()
			for c := 0; c < 6; c++ {
				globalMaxCollisionDistribution[c] += localMaxColDist[c]
				for u := 0; u < 6; u++ {
					globalUserCollisionDistribution[u][c] += localUserColDist[u][c]
				}
			}
			mu.Unlock()

		}(d1)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	// 4. Print Results
	totalStates := int64(9765625)
	fmt.Printf("\nTraversal completed in %v\n", elapsed)
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println("DISTRIBUTION OF MAXIMUM COLLISION PER STATE:")
	fmt.Println("  (c = maximum number of slots collided for ANY user in a state)")
	fmt.Println("-----------------------------------------------------------------")

	var maxColEncountered int = 0
	for c := 0; c < 6; c++ {
		count := globalMaxCollisionDistribution[c]
		percentage := float64(count) / float64(totalStates) * 100.0
		fmt.Printf("  c = %d : %12d states (%8.4f%%)\n", c, count, percentage)
		if count > 0 && c > maxColEncountered {
			maxColEncountered = c
		}
	}

	fmt.Println("\n-----------------------------------------------------------------")
	fmt.Println("INDIVIDUAL USER COLLISION DISTRIBUTIONS:")
	fmt.Println("  (How many slots each user gets collided across all states)")
	fmt.Println("-----------------------------------------------------------------")
	for u := 0; u < 6; u++ {
		fmt.Printf("  Seq %d | ", u)
		for c := 0; c < 6; c++ {
			count := globalUserCollisionDistribution[u][c]
			percentage := float64(count) / float64(totalStates) * 100.0
			fmt.Printf("c=%d: %5.2f%% | ", c, percentage)
		}
		fmt.Println()
	}

	fmt.Println("\n================================================================-")
	fmt.Printf("CONCLUSION: The maximum Hamming collision count in all states is: %d\n", maxColEncountered)
	if maxColEncountered == 5 {
		fmt.Println("RESULT: Fails 2UI-free under 6 concurrent users!")
		fmt.Println("        There exist worst-case shift states where at least one user is")
		fmt.Println("        completely repressed (all 5 transmission slots collide).")
	} else {
		fmt.Printf("RESULT: Meets 2UI-free under 6 concurrent users!\n")
		fmt.Printf("        In ALL possible states, no user ever collides in all 5 slots.\n")
		fmt.Printf("        Every user is guaranteed at least %d collision-free slot(s).\n", 5-maxColEncountered)
	}
	fmt.Println("================================================================-")
}
