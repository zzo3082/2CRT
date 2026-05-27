package main

import (
	"fmt"
	"math/bits"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

// Helper to convert user bitmask (uint64) to letter sequence of length 49
func getLetterSeqString(mask uint64, letter string) string {
	var sb strings.Builder
	sb.WriteString("[")
	for t := 0; t < 49; t++ {
		if t > 0 {
			sb.WriteString(",")
		}
		if (mask & (1 << uint(t))) != 0 {
			sb.WriteString(letter)
		} else {
			sb.WriteString("0")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// Helper to format the superposition result (res) with MPR r=2 constraint
func formatSuperpositionResult(masks [8]uint64) string {
	letters := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	var sb strings.Builder
	sb.WriteString("[")
	
	for t := 0; t < 49; t++ {
		if t > 0 {
			sb.WriteString(",")
		}
		
		var active []string
		for u := 0; u < 8; u++ {
			if (masks[u] & (1 << uint(t))) != 0 {
				active = append(active, letters[u])
			}
		}
		
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
	p := 7
	pq := p * p // Period is 49

	// 1. Define the 8 prime sequences for p=7
	rawSeqs := [][]int{
		{0, 1, 2, 3, 4, 5, 6},            // Seq 0 (g=0) -> a
		{0, 8, 16, 24, 32, 40, 48},       // Seq 1 (g=1) -> b
		{0, 15, 30, 45, 11, 26, 41},      // Seq 2 (g=2) -> c
		{0, 22, 44, 17, 39, 12, 34},      // Seq 3 (g=3) -> d
		{0, 29, 9, 38, 18, 47, 27},       // Seq 4 (g=4) -> e
		{0, 36, 23, 10, 46, 33, 20},      // Seq 5 (g=5) -> f
		{0, 43, 37, 31, 25, 19, 13},      // Seq 6 (g=6) -> g
		{0, 7, 14, 21, 28, 35, 42},       // Seq 7 (g=p) -> h
	}

	// 2. Precalculate shifts as bitmasks (uint64, since period is 49)
	var masks [8][49]uint64
	for i := 0; i < 8; i++ {
		for d := 0; d < 49; d++ {
			var mask uint64
			for _, x := range rawSeqs[i] {
				shifted := (x + d) % pq
				mask |= 1 << uint(shifted)
			}
			masks[i][d] = mask
		}
	}

	totalTrials := int64(10000000) // 10 million trials
	numCPU := 16                    // Run in parallel

	fmt.Println("=================================================================")
	fmt.Println("Starting MPR (r=2) Monte Carlo Verification for K=8 users...")
	fmt.Printf("Simulating %d random trials in parallel...\n", totalTrials)
	fmt.Println("=================================================================")
	startTime := time.Now()

	var wg sync.WaitGroup
	wg.Add(numCPU)

	var mu sync.Mutex
	var totalSuccessStates int64
	var totalFailureStates int64

	type FailureRecord struct {
		shifts      [8]int
		activeMasks [8]uint64
		missing     string
	}
	var failureRecords []FailureRecord
	const maxFailureOutput = 100

	trialsPerWorker := totalTrials / int64(numCPU)

	for cpu := 0; cpu < numCPU; cpu++ {
		go func(workerID int) {
			defer wg.Done()

			// Local thread-safe random generator
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			var localSuccess int64
			var localFailure int64
			var localFailures []FailureRecord

			for step := int64(0); step < trialsPerWorker; step++ {
				// Generate random shifts for Users 1 to 7 (User 0 is fixed at 0)
				var shifts [8]int
				shifts[0] = 0
				for u := 1; u < 8; u++ {
					shifts[u] = rng.Intn(49)
				}

				// Get the masks
				var activeMasks [8]uint64
				for u := 0; u < 8; u++ {
					activeMasks[u] = masks[u][shifts[u]]
				}

				// Check each slot t in 0..48
				var unionMask uint8
				for t := 0; t < 49; t++ {
					var slotEmissions uint8
					for u := 0; u < 8; u++ {
						if (activeMasks[u] & (1 << uint(t))) != 0 {
							slotEmissions |= 1 << uint(u)
						}
					}

					// If number of active users in this slot <= 2 (MPR r=2)
					if bits.OnesCount8(slotEmissions) <= 2 {
						unionMask |= slotEmissions
					}
				}

				// Check if all 8 letters are present (unionMask == 255)
				if unionMask == 255 {
					localSuccess++
				} else {
					localFailure++

					// Record details if we haven't reached output limit
					if len(localFailures) < maxFailureOutput {
						var missing []string
						letters := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
						for u := 0; u < 8; u++ {
							if (unionMask & (1 << uint(u))) == 0 {
								missing = append(missing, letters[u])
							}
						}
						localFailures = append(localFailures, FailureRecord{
							shifts:      shifts,
							activeMasks: activeMasks,
							missing:     strings.Join(missing, ","),
						})
					}
				}
			}

			// Aggregate to global
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

		}(cpu)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	// Summary output
	fmt.Printf("\nVerification completed in %v\n", elapsed)
	fmt.Printf("Total Trials Run: %d\n", totalSuccessStates+totalFailureStates)
	fmt.Printf("Success Trials (Meet UI): %d (%.4f%%)\n", totalSuccessStates, float64(totalSuccessStates)/float64(totalSuccessStates+totalFailureStates)*100.0)
	fmt.Printf("Failure Trials (Fail UI): %d (%.4f%%)\n", totalFailureStates, float64(totalFailureStates)/float64(totalSuccessStates+totalFailureStates)*100.0)

	// Export to file
	filename := "mpr_failures_k7.txt"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	file.WriteString("=================================================================\n")
	file.WriteString(fmt.Sprintf("MPR (r=2) UI Verification Failure Log for K=8, p=7 (10M Trials)\n"))
	file.WriteString(fmt.Sprintf("Total Failure Trials: %d out of %d (%.4f%%)\n", totalFailureStates, totalSuccessStates+totalFailureStates, float64(totalFailureStates)/float64(totalSuccessStates+totalFailureStates)*100.0))
	file.WriteString("=================================================================\n\n")

	if totalFailureStates == 0 {
		file.WriteString("CONGRATULATIONS! No failure states found in 10M random trials!\n")
		fmt.Println("\nResult: Every trial met the MPR UI property! No failures found.")
	} else {
		file.WriteString(fmt.Sprintf("Showing the first %d failure trials:\n\n", len(failureRecords)))
		
		letters := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
		for idx, rec := range failureRecords {
			file.WriteString(fmt.Sprintf("Failure #%d:\n", idx+1))
			file.WriteString(fmt.Sprintf("  Relative Shifts d = (s0:%d, s1:%d, s2:%d, s3:%d, s4:%d, s5:%d, s6:%d, s7:%d)\n", 
				rec.shifts[0], rec.shifts[1], rec.shifts[2], rec.shifts[3], rec.shifts[4], rec.shifts[5], rec.shifts[6], rec.shifts[7]))
			file.WriteString(fmt.Sprintf("  Missing Letters: %s\n", rec.missing))
			
			for u := 0; u < 8; u++ {
				file.WriteString(fmt.Sprintf("  s%d:  %s\n", u, getLetterSeqString(rec.activeMasks[u], letters[u])))
			}
			file.WriteString(fmt.Sprintf("  res: %s\n\n", formatSuperpositionResult(rec.activeMasks)))
		}
		
		fmt.Printf("\nResult: Found %d failure trials (%.4f%%). Details exported to %s\n", totalFailureStates, float64(totalFailureStates)/float64(totalSuccessStates+totalFailureStates)*100.0, filename)
	}
}
