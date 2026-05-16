import numpy as np

def simulate_run(shifted_matrix, r, pq):
    """
    Simulates collisions and applies MPR rule.
    Returns a list of 1D numpy arrays containing sorted successful indices for each user.
    """
    # 1. Count occurrences of each slot
    counts = np.bincount(shifted_matrix.ravel(), minlength=pq)
    
    # 2. Identify valid slots (where count > 0 and <= r)
    valid_slots = (counts > 0) & (counts <= r)
    
    # 3. Find successful transmissions for each user
    # valid_slots[shifted_matrix] gives a boolean mask of shape (se, w)
    mask = valid_slots[shifted_matrix]
    
    se = shifted_matrix.shape[0]
    success_counts = []
    
    # Extract the successful indices for each user and sort them
    for i in range(se):
        success_indices = shifted_matrix[i][mask[i]]
        success_indices.sort() # In-place sort, necessary for Delay/AoI calculation
        success_counts.append(success_indices)
        
    return success_counts
