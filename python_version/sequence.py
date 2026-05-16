import numpy as np

def generate_sequences(p, q, w):
    pq = p * q
    SgSet = []
    
    # Generate for g = 0 to p-1
    for g in range(p):
        indices = []
        gSet_p = [(g * j) % p for j in range(w)]
        gSet_q = [j % q for j in range(w)]
        
        g_set = set(zip(gSet_p, gSet_q))
        
        for t in range(pq):
            if (t % p, t % q) in g_set:
                indices.append(t)
                
        SgSet.append(np.array(indices, dtype=np.int32))
        
    # Generate for g = p
    indices = []
    gSet_p = [j % p for j in range(w)]
    gSet_q = [0 for _ in range(w)]
    g_set = set(zip(gSet_p, gSet_q))
    
    for t in range(pq):
        if (t % p, t % q) in g_set:
            indices.append(t)
            
    SgSet.append(np.array(indices, dtype=np.int32))
    
    return SgSet

def get_base_matrix(sg_set, sequences):
    """
    Creates a 2D numpy array of shape (se, w) containing the selected sequences.
    This allows us to vectorize the random shift process.
    """
    # sequences is 1-indexed (from MATLAB)
    return np.array([sg_set[g-1] for g in sequences], dtype=np.int32)

def apply_random_shifts_vectorized(base_matrix, pq, rng):
    """
    Vectorized version of random shift.
    base_matrix: (se, w) numpy array
    Returns: (se, w) numpy array shifted by a random integer for each row.
    """
    se = base_matrix.shape[0]
    # rng.integers is exclusive of the upper bound, so pq + 1 gives [0, pq]
    shifts = rng.integers(0, pq + 1, size=(se, 1), dtype=np.int32)
    shifted_matrix = (base_matrix + shifts) % pq
    return shifted_matrix
