import os
import time
import math
import multiprocessing
import numpy as np

from sequence import generate_sequences, get_base_matrix, apply_random_shifts_vectorized
from simulation import simulate_run
from metrics import calculate_system_metrics
from export import export_to_xlsx

# Worker function for multiprocessing
def worker_simulations(args):
    runs, base_matrix, pq, r, T, se = args
    
    # Initialize local random generator for thread-safety and unique randomness
    rng = np.random.default_rng()
    
    local_throughput = np.zeros(se, dtype=np.float64)
    local_delay = np.zeros(se + 1, dtype=np.float64)
    local_aoi = np.zeros(se, dtype=np.float64)
    
    for _ in range(runs):
        shifted_matrix = apply_random_shifts_vectorized(base_matrix, pq, rng)
        success_counts = simulate_run(shifted_matrix, r, pq)
        throughputs, delays, aois = calculate_system_metrics(success_counts, T, pq)
        
        local_throughput += throughputs
        local_delay[:se] += delays
        local_aoi += aois
        local_delay[se] += np.mean(delays)
        
    return local_throughput, local_delay, local_aoi

def main():
    p = 127
    r = 5
    se = 120
    q = math.ceil((2 * se) / r) # 48
    pq = p * q
    
    sequences = [
        3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
        23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
        41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58,
        59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76,
        77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94,
        95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109,
        110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122,
    ]
    
    values = [120]
    T_Values = [1, 6096]
    ns = 100000
    
    out_dir = f"Case2_newcombined_M{se}R{r}_PY"
    os.makedirs(out_dir, exist_ok=True)
    
    num_workers = multiprocessing.cpu_count()
    print(f"Starting Simulation... ns={ns}, CPUS={num_workers}")
    start_time = time.time()
    
    for w in values:
        print(f"\n--- Generating Sequence for w={w} ---")
        sg_set = generate_sequences(p, q, w)
        base_matrix = get_base_matrix(sg_set, sequences)
        
        for T in T_Values:
            print(f"Simulating T={T}...")
            t_folder = os.path.join(out_dir, f"T_{T}")
            os.makedirs(t_folder, exist_ok=True)
            
            # Prepare arguments for workers
            runs_per_worker = ns // num_workers
            remainder = ns % num_workers
            
            worker_args = []
            for i in range(num_workers):
                runs = runs_per_worker + (1 if i < remainder else 0)
                worker_args.append((runs, base_matrix, pq, r, T, se))
                
            # Execute in parallel
            with multiprocessing.Pool(processes=num_workers) as pool:
                results = pool.map(worker_simulations, worker_args)
                
            # Aggregate results
            total_throughput = np.zeros(se, dtype=np.float64)
            total_delay = np.zeros(se + 1, dtype=np.float64)
            total_aoi = np.zeros(se, dtype=np.float64)
            
            for res_th, res_de, res_aoi in results:
                total_throughput += res_th
                total_delay += res_de
                total_aoi += res_aoi
                
            # Final Averaging
            tmpm = p * q * ns
            
            avg_throughput = np.zeros(se + 1, dtype=np.float64)
            avg_delay = np.zeros(se + 1, dtype=np.float64)
            avg_aoi = np.zeros(se + 1, dtype=np.float64)
            
            avg_throughput[:se] = total_throughput / tmpm
            avg_delay[:se] = total_delay[:se] / ns
            avg_aoi[:se] = total_aoi / ns
            
            avg_throughput[se] = np.sum(total_throughput) / (se * tmpm)
            avg_delay[se] = total_delay[se] / ns
            avg_aoi[se] = np.sum(total_aoi) / (se * ns)
            
            # Output
            filename = os.path.join(t_folder, f"{T}結果{w}.xlsx")
            export_to_xlsx(filename, avg_throughput, avg_delay, avg_aoi)
            print(f"Done saving to {filename}.")
            
    print(f"\nTotal Time Elapsed: {time.time() - start_time:.2f} seconds")

if __name__ == '__main__':
    main()
