package main

import (
	"fmt"
	"math/bits"
	"os"
	"strings"
	"sync"
	"time"
)

// Helper to convert user bitmask to letter sequence
func getLetterSeqString(mask uint32, letter string) string {
	var sb strings.Builder
	sb.WriteString("[")
	for t := 0; t < 25; t++ {
		if t > 0 {
			sb.WriteString(",")
		}
		if (mask & (1 << t)) != 0 {
			sb.WriteString(letter)
		} else {
			sb.WriteString("0")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// Helper to format the superposition result (res) with MPR r=2 constraint
func formatSuperpositionResult(masks [6]uint32) string {
	letters := []string{"a", "b", "c", "d", "e", "f"}
	var sb strings.Builder
	sb.WriteString("[")
	
	for t := 0; t < 25; t++ {
		if t > 0 {
			sb.WriteString(",")
		}
		
		// Collect who transmitted in this slot
		var active []string
		for u := 0; u < 6; u++ {
			if (masks[u] & (1 << t)) != 0 {
				active = append(active, letters[u])
			}
		}
		
		// Apply MPR r=2 constraint
		if len(active) == 0 {
			sb.WriteString("0")
		} else if len(active) == 1 {
			sb.WriteString(active[0])
		} else if len(active) == 2 {
			sb.WriteString("{" + active[0] + "," + active[1] + "}")
		} else {
			// Collision > 2, ignored (set to 0)
			sb.WriteString("0")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

func main() {
	p := 5
	pq := p * p // Period is 25

	// 1. Define the 6 prime sequences (characteristic sets)
	rawSeqs := [][]int{
		{0, 1, 2, 3, 4},      // Seq 0 (g=0) -> a
		{0, 6, 12, 18, 24},   // Seq 1 (g=1) -> b
		{0, 8, 11, 19, 22},   // Seq 2 (g=2) -> c
		{0, 7, 14, 16, 23},   // Seq 3 (g=3) -> d
		{0, 9, 13, 17, 21},   // Seq 4 (g=4) -> e
		{0, 5, 10, 15, 20},   // Seq 5 (g=p) -> f
	}

	// 2. Precalculate shifts as bitmasks
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
	fmt.Println("Starting MPR (r=2) UI Verification for K=6 users...")
	fmt.Println("Traversing all independent 25^5 = 9,765,625 states...")
	fmt.Println("=================================================================")
	startTime := time.Now()

	// Parallel traversal
	var wg sync.WaitGroup
	wg.Add(25)

	var mu sync.Mutex
	var totalSuccessStates int64
	var totalFailureStates int64

	// Structure to hold failure records
	type FailureRecord struct {
		shifts      [6]int
		activeMasks [6]uint32
		missing     string
	}
	var failureRecords []FailureRecord
	const maxFailureOutput = 100 // Limit failure details in log to prevent huge file

	d0Val := 0
	m0 := masks[0][d0Val]

	for d1 := 0; d1 < 25; d1++ {
		go func(d1Val int) {
			defer wg.Done()

			var localSuccess int64
			var localFailure int64
			var localFailures []FailureRecord

			m1 := masks[1][d1Val]

			for d2 := 0; d2 < 25; d2++ {
				m2 := masks[2][d2]

				for d3 := 0; d3 < 25; d3++ {
					m3 := masks[3][d3]

					for d4 := 0; d4 < 25; d4++ {
						m4 := masks[4][d4]

						for d5 := 0; d5 < 25; d5++ {
							m5 := masks[5][d5]

							// For each slot t, compute state
							var unionMask uint8
							for t := 0; t < 25; t++ {
								var slotEmissions uint8
								if (m0 & (1 << t)) != 0 { slotEmissions |= 1 << 0 }
								if (m1 & (1 << t)) != 0 { slotEmissions |= 1 << 1 }
								if (m2 & (1 << t)) != 0 { slotEmissions |= 1 << 2 }
								if (m3 & (1 << t)) != 0 { slotEmissions |= 1 << 3 }
								if (m4 & (1 << t)) != 0 { slotEmissions |= 1 << 4 }
								if (m5 & (1 << t)) != 0 { slotEmissions |= 1 << 5 }

								// If number of active users in this slot <= 2 (MPR r=2)
								if bits.OnesCount8(slotEmissions) <= 2 {
									unionMask |= slotEmissions
								}
							}

							// Check if all 6 letters are present (unionMask == 63)
							if unionMask == 63 {
								localSuccess++
							} else {
								localFailure++
								
								// Record failure if we haven't reached the limit
								if len(localFailures) < maxFailureOutput {
									// Determine missing letters
									var missing []string
									letters := []string{"a", "b", "c", "d", "e", "f"}
									for u := 0; u < 6; u++ {
										if (unionMask & (1 << u)) == 0 {
											missing = append(missing, letters[u])
										}
									}
									
									localFailures = append(localFailures, FailureRecord{
										shifts:      [6]int{d0Val, d1Val, d2, d3, d4, d5},
										activeMasks: [6]uint32{m0, m1, m2, m3, m4, m5},
										missing:     strings.Join(missing, ","),
									})
								}
							}
						}
					}
				}
			}

			// Aggregate to global counters
			mu.Lock()
			totalSuccessStates += localSuccess
			totalFailureStates += localFailure
			if len(failureRecords) < maxFailureOutput {
				availableSlot := maxFailureOutput - len(failureRecords)
				toAdd := len(localFailures)
				if toAdd > availableSlot {
					toAdd = availableSlot
				}
				failureRecords = append(failureRecords, localFailures[:toAdd]...)
			}
			mu.Unlock()

		}(d1)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	// Print Summary
	totalStates := totalSuccessStates + totalFailureStates
	fmt.Printf("\nTraversal completed in %v\n", elapsed)
	fmt.Printf("Total States Analyzed: %d\n", totalStates)
	fmt.Printf("Success States (Meet UI): %d (%.4f%%)\n", totalSuccessStates, float64(totalSuccessStates)/float64(totalStates)*100)
	fmt.Printf("Failure States (Fail UI): %d (%.4f%%)\n", totalFailureStates, float64(totalFailureStates)/float64(totalStates)*100)

	// Write failures to file
	filename := "mpr_failures.txt"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	// Write header to file
	file.WriteString("=================================================================\n")
	file.WriteString(fmt.Sprintf("MPR (r=2) UI Verification Failure Log for K=6, p=5\n"))
	file.WriteString(fmt.Sprintf("Total Failure States: %d out of %d (%.4f%%)\n", totalFailureStates, totalStates, float64(totalFailureStates)/float64(totalStates)*100))
	file.WriteString("=================================================================\n\n")

	if totalFailureStates == 0 {
		file.WriteString("CONGRATULATIONS! No failure states found. Every letter appeared at least once in all states!\n")
		fmt.Println("\nResult: Every state meets the MPR UI property! No failures found.")
	} else {
		file.WriteString(fmt.Sprintf("Showing the first %d failure states:\n\n", len(failureRecords)))
		
		letters := []string{"a", "b", "c", "d", "e", "f"}
		for idx, rec := range failureRecords {
			file.WriteString(fmt.Sprintf("Failure #%d:\n", idx+1))
			file.WriteString(fmt.Sprintf("  Relative Shifts d = (s0:%d, s1:%d, s2:%d, s3:%d, s4:%d, s5:%d)\n", 
				rec.shifts[0], rec.shifts[1], rec.shifts[2], rec.shifts[3], rec.shifts[4], rec.shifts[5]))
			file.WriteString(fmt.Sprintf("  Missing Letters: %s\n", rec.missing))
			
			// Print individual user sequences
			for u := 0; u < 6; u++ {
				file.WriteString(fmt.Sprintf("  s%d:  %s\n", u, getLetterSeqString(rec.activeMasks[u], letters[u])))
			}
			// Print superposition result (res)
			file.WriteString(fmt.Sprintf("  res: %s\n\n", formatSuperpositionResult(rec.activeMasks)))
		}
		
		fmt.Printf("\nResult: Found %d failure states (%.4f%%). Details exported to %s\n", totalFailureStates, float64(totalFailureStates)/float64(totalStates)*100, filename)
	}
}
