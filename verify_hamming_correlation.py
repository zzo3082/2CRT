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
        SgSet.append(indices)
        
    # Generate for g = p
    indices = []
    for j in range(w):
        val = (j * p) % pq
        indices.append(val)
    SgSet.append(indices)
    
    return SgSet

def calculate_hamming_cross_correlation_matrix(p, w):
    pq = p * p
    sequences = generate_prime_sequences(p, w)
    num_seq = len(sequences)
    
    # Initialize the cross-correlation matrix
    # H[i, j] will store the maximum cross-correlation between seq i and seq j over all shifts
    H_matrix = np.zeros((num_seq, num_seq), dtype=int)
    
    # For detailed output, we can also record the shift that achieves the maximum correlation
    for i in range(num_seq):
        set_i = set(sequences[i])
        for j in range(num_seq):
            max_corr = 0
            for shift in range(pq):
                # Shift sequence j by shift modulo pq
                shifted_j = { (x + shift) % pq for x in sequences[j] }
                corr = len(set_i.intersection(shifted_j))
                if corr > max_corr:
                    max_corr = corr
            H_matrix[i, j] = max_corr
            
    return H_matrix

def run_calculation():
    p = 5
    w = 5
    print("=" * 65)
    print(f"Calculating Hamming Cross-correlation for Case 1 Prime Sequences")
    print(f"Parameters: p = {p}, w = {w}, Period = {p*p} (2-CRT(5))")
    print("=" * 65)
    
    H_matrix = calculate_hamming_cross_correlation_matrix(p, w)
    
    # Display the sequences first
    sequences = generate_prime_sequences(p, w)
    print("\nGenerated Sequences (as characteristic sets):")
    for idx, seq in enumerate(sequences):
        print(f"  Sequence {idx} (g={idx if idx < p else 'p'}): {sorted(seq)}")
        
    print("\nHamming Cross-correlation Matrix H (Row i, Col j):")
    # Print headers
    print("      ", end="")
    for j in range(len(sequences)):
        print(f"Seq{j} ", end="")
    print()
    print("  " + "-" * 42)
    
    for i in range(len(sequences)):
        print(f"Seq{i} |", end="")
        for j in range(len(sequences)):
            if i == j:
                print("   - ", end="")
            else:
                print(f"{H_matrix[i, j]:4d} ", end="")
        print()
        
    print("\nSummary of Results:")
    print("1. For distinct g, h in [0, p-1] (Seq 0 to 4):")
    g_h_values = []
    for i in range(p):
        for j in range(i + 1, p):
            g_h_values.append(H_matrix[i, j])
    print(f"   Max cross-correlation values: {set(g_h_values)} (Expected: {{2}} according to Lemma 1)")
    
    print("\n2. For g in [0, p-1] and seq p (Seq 0-4 vs Seq 5):")
    g_p_values = [H_matrix[i, p] for i in range(p)]
    print(f"   Max cross-correlation values: {set(g_p_values)} (Expected: {{1}} according to Lemma 1)")

if __name__ == "__main__":
    run_calculation()
