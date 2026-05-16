package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

func main() {
	// 0. Setup Logging to both file and stdout
	logFile, err := os.OpenFile("simulation.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	// Use MultiWriter so we see logs in terminal and save to file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	// 1. System Parameters
	p := 127
	r := 5
	se := 120
	q := int(math.Ceil(float64(2*se) / float64(r))) // 48
	pq := p * q

	// Sequences chosen by user
	sequences := []int{
		3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
		23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
		41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58,
		59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76,
		77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94,
		95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109,
		110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122,
	}

	values := []int{120}
	T_Values := []int{1, 6096}
	ns := 100000 // Number of Monte Carlo simulations

	// Output directory
	outDir := fmt.Sprintf("Case2_newcombined_M%dR%d_GO", se, r)
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		panic(err)
	}

	log.Printf("Starting Simulation... ns=%d, CPUS=%d", ns, runtime.NumCPU())
	startTime := time.Now()

	for _, w := range values {
		log.Printf("--- Generating Sequence for w=%d ---", w)
		// 2. Generate CRT Sequences (only indices where value is 1)
		sgSet := GenerateSequences(p, q, w)

		for _, T := range T_Values {
			log.Printf("Simulating T=%d... ", T)
			tFolder := filepath.Join(outDir, fmt.Sprintf("T_%d", T))
			os.MkdirAll(tFolder, os.ModePerm)

			// 3. Parallel Execution setup
			numWorkers := runtime.NumCPU()
			runsPerWorker := ns / numWorkers
			remainder := ns % numWorkers

			var wg sync.WaitGroup

			// Mutex to safely aggregate results from all workers
			var mu sync.Mutex

			// Aggregate accumulators
			totalThroughput := make([]float64, se)
			totalDelay := make([]float64, se+1)
			totalAoI := make([]float64, se)

			for wIdx := 0; wIdx < numWorkers; wIdx++ {
				wg.Add(1)
				runs := runsPerWorker
				if wIdx < remainder {
					runs++
				}

				go func(runs int) {
					defer wg.Done()
					// Each worker has its own local random number generator
					rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(rand.Intn(100000))))

					localThroughput := make([]float64, se)
					localDelay := make([]float64, se+1)
					localAoI := make([]float64, se)

					for m := 0; m < runs; m++ {
						// a. Random Shift
						shiftedSgSet := ApplyRandomShifts(sgSet, sequences, pq, rng)

						// b. Collision Simulation (MPR)
						successCounts := SimulateRun(shiftedSgSet, r, pq)

						// c. Calculate Metrics
						metrics := CalculateSystemMetrics(successCounts, T, pq)

						// d. Aggregate Local
						currentDelaySum := 0.0
						for i := 0; i < se; i++ {
							localThroughput[i] += metrics[i].Throughput
							localDelay[i] += metrics[i].Delay
							localAoI[i] += metrics[i].AoI
							currentDelaySum += metrics[i].Delay
						}
						localDelay[se] += currentDelaySum / float64(se)
					}

					// Safely add to global totals
					mu.Lock()
					for i := 0; i < se; i++ {
						totalThroughput[i] += localThroughput[i]
						totalDelay[i] += localDelay[i]
						totalAoI[i] += localAoI[i]
					}
					totalDelay[se] += localDelay[se]
					mu.Unlock()
				}(runs)
			}

			// Wait for all workers to finish
			wg.Wait()

			// 4. Final Averaging (matching MATLAB logic)
			tmpm := float64(p * q * ns)
			nsFloat := float64(ns)

			avgThroughput := make([]float64, se+1)
			avgDelay := make([]float64, se+1)
			avgAoI := make([]float64, se+1)

			sumThroughput := 0.0
			sumAoI := 0.0

			for i := 0; i < se; i++ {
				avgThroughput[i] = totalThroughput[i] / tmpm
				sumThroughput += totalThroughput[i]

				avgDelay[i] = totalDelay[i] / nsFloat

				avgAoI[i] = totalAoI[i] / nsFloat
				sumAoI += totalAoI[i]
			}

			// se+1 element averages
			avgThroughput[se] = sumThroughput / (float64(se) * tmpm)
			avgDelay[se] = totalDelay[se] / nsFloat
			avgAoI[se] = sumAoI / (float64(se) * nsFloat)

			// 5. Output to CSV
			filename := filepath.Join(tFolder, fmt.Sprintf("%d結果%d.csv", T, w))
			if err := ExportToCSV(filename, avgThroughput, avgDelay, avgAoI); err != nil {
				log.Printf("Error writing CSV: %v", err)
			} else {
				log.Printf("Done.")
			}
		}
	}
	log.Printf("Total Time Elapsed: %v", time.Since(startTime))
}
