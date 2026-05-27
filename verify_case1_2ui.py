import numpy as np

def generate_prime_sequences(p, w):
    pq = p * p  # Period is p^2
    SgSet = []
    
    # Generate for g = 0 to p-1
    for g in range(p):
        indices = []
        for j in range(w):
            val = (j * (1 + g * p)) % pq
            indices.append(val)
        SgSet.append(set(indices))
        
    # Generate for g = p (the p+1-th sequence)
    indices = []
    for j in range(w):
        val = (j * p) % pq
        indices.append(val)
    SgSet.append(set(indices))
    
    return SgSet

def check_2ui_pair(seq_A, seq_B, pq):
    """
    Checks if a pair of sequences (seq_A, seq_B) is 2UI-free.
    For any relative shift d in [0, pq-1], A shifted by d must have at least one element not in B,
    and B must have at least one element not in A shifted by d.
    """
    for d in range(pq):
        # Shift A by d
        shifted_A = { (x + d) % pq for x in seq_A }
        
        # Check if A has at least one collision-free slot
        a_has_free = any(x not in seq_B for x in shifted_A)
        # Check if B has at least one collision-free slot
        b_has_free = any(x not in shifted_A for x in seq_B)
        
        if not a_has_free or not b_has_free:
            return False, d
            
    return True, None

def verify_2ui_for_set(seq_list, pq):
    """
    Verifies if all pairs in the given set of sequences are 2UI-free.
    Returns (is_2ui, failed_pair)
    """
    n = len(seq_list)
    for i in range(n):
        for j in range(i + 1, n):
            is_2ui, failed_shift = check_2ui_pair(seq_list[i], seq_list[j], pq)
            if not is_2ui:
                return False, (i, j, failed_shift)
    return True, None

def run_verification():
    p = 5
    pq = p * p
    
    print("=" * 60)
    print(f"Verifying 2UI Property for Case 1 Prime Sequences (p={p}, Period={pq})")
    print("=" * 60)
    
    # Test w from 1 to p
    for w in range(1, p + 1):
        print(f"\n--- Testing for w = {w} ---")
        sequences = generate_prime_sequences(p, w)
        
        # 1. Test all 6 sequences together
        all_6_2ui, failed_info = verify_2ui_for_set(sequences, pq)
        if all_6_2ui:
            print(f"Result for ALL 6 sequences: Meets 2UI property!")
        else:
            i, j, shift = failed_info
            print(f"Result for ALL 6 sequences: Fails 2UI property!")
            print(f"  -> Failed pair: Sequence {i} and Sequence {j} at relative shift {shift}")
            
        # 2. Test subsets of 5 sequences (6 combinations)
        print("Result for choosing 5 out of 6 sequences:")
        all_5_subsets_pass = True
        for skip_idx in range(6):
            subset = [sequences[k] for k in range(6) if k != skip_idx]
            subset_2ui, failed_info = verify_2ui_for_set(subset, pq)
            if subset_2ui:
                print(f"  - Subset excluding sequence {skip_idx}: Meets 2UI property")
            else:
                i_sub, j_sub, shift = failed_info
                original_indices = [k for k in range(6) if k != skip_idx]
                i_orig = original_indices[i_sub]
                j_orig = original_indices[j_sub]
                print(f"  - Subset excluding sequence {skip_idx}: Fails 2UI property (Failed pair: seq {i_orig} and seq {j_orig} at shift {shift})")
                all_5_subsets_pass = False

if __name__ == "__main__":
    run_verification()
