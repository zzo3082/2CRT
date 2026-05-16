import numpy as np

def calculate_metrics(success_indices, T, pq):
    throughput = len(success_indices)
    
    if throughput == 0:
        return 0.0, 0.0, 0.0
        
    # 1. Delay Calculation
    # Find the first successful index in each interval of size T
    group_ids = success_indices // T
    _, first_idxs = np.unique(group_ids, return_index=True)
    first_vals = success_indices[first_idxs]
    
    mod_values = first_vals % T
    delays = np.where(mod_values == 0, 1, mod_values + 1)
    delay_avg = np.mean(delays) if len(delays) > 0 else 0.0
    
    # 2. AoI Calculation
    # We must match the MATLAB/Go logic exactly.
    # extended_indices (0-indexed)
    extended_indices = np.concatenate([first_vals, first_vals + pq])
    
    # Create a boolean map for O(1) vectorized lookups
    # Size 2*pq + 1 to safely access up to 2*pq
    extended_map = np.zeros(2 * pq + 1, dtype=bool)
    # In Python, indices out of bounds will throw, but extended_indices <= 2pq-1
    extended_map[extended_indices] = True
    
    # d goes from 1 to 2pq
    d = np.arange(1, 2 * pq + 1)
    
    # Check if 'd' is in extended_map
    valid_mask = extended_map[d]
    
    # We want to fill an array aoi of size 2*pq.
    base_values = np.zeros(2 * pq, dtype=np.float64)
    base_values[valid_mask] = (d[valid_mask] % T) + 1
    
    # To vectorize the loop: 
    # if valid_mask[i]: aoi[i] = base_values[i]
    # elif i == 0: aoi[i] = 0
    # else: aoi[i] = aoi[i-1] + 1
    
    # We can use the fill-forward technique with np.maximum.accumulate
    # We mark the reset points. i=0 is a reset point unconditionally (if not valid_mask).
    reset_points = valid_mask.copy()
    reset_points[0] = True
    
    # Indices of reset points
    reset_indices = np.where(reset_points)[0]
    
    # For each position, find the index of the last reset point
    idx = np.searchsorted(reset_indices, np.arange(2 * pq), side='right') - 1
    last_reset_pos = reset_indices[idx]
    
    # Calculate AoI = Base value at last reset + distance from last reset
    aoi = base_values[last_reset_pos] + (np.arange(2 * pq) - last_reset_pos)
    
    # Calculate average of the second half
    aoi_avg = np.mean(aoi[pq:])
    
    return float(throughput), float(delay_avg), float(aoi_avg)

def calculate_system_metrics(success_counts, T, pq):
    """
    Processes all users in one run.
    """
    se = len(success_counts)
    throughputs = np.zeros(se, dtype=np.float64)
    delays = np.zeros(se, dtype=np.float64)
    aois = np.zeros(se, dtype=np.float64)
    
    for i in range(se):
        th, de, aoi = calculate_metrics(success_counts[i], T, pq)
        throughputs[i] = th
        delays[i] = de
        aois[i] = aoi
        
    return throughputs, delays, aois
