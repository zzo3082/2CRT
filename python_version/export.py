import pandas as pd

def export_to_xlsx(filename, avg_throughput, avg_delay, avg_aoi):
    """
    Exports the aggregated metrics to an XLSX file using pandas.
    The arrays are of length se + 1 (the last element is the total average).
    """
    se = len(avg_throughput) - 1
    
    # Define columns to match MATLAB / Go formatting
    columns = ["Metric"] + [f"User_{i+1}" for i in range(se)] + ["Total_Average"]
    
    # Prepare data rows
    data = [
        ["Throughput"] + list(avg_throughput),
        ["Delay"] + list(avg_delay),
        ["AoI"] + list(avg_aoi)
    ]
    
    # Create DataFrame and save
    df = pd.DataFrame(data, columns=columns)
    df.to_excel(filename, index=False)
