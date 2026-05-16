package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

// ExportToCSV writes the aggregated average metrics to a CSV file.
func ExportToCSV(filename string, avgThroughput, avgDelay, avgAoI []float64) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	se := len(avgThroughput) - 1 // Last element is the total average

	// Helper function to write a row
	writeRow := func(name string, data []float64) error {
		row := make([]string, len(data)+1)
		row[0] = name
		for i, val := range data {
			// Formatting to 6 decimal places like MATLAB often does
			row[i+1] = fmt.Sprintf("%.6f", val)
		}
		return writer.Write(row)
	}

	// Write Headers (User 1 ... User SE, Average)
	headers := make([]string, se+2)
	headers[0] = "Metric"
	for i := 1; i <= se; i++ {
		headers[i] = fmt.Sprintf("User_%d", i)
	}
	headers[se+1] = "Total_Average"
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write Data Rows
	if err := writeRow("Throughput", avgThroughput); err != nil {
		return err
	}
	if err := writeRow("Delay", avgDelay); err != nil {
		return err
	}
	if err := writeRow("AoI", avgAoI); err != nil {
		return err
	}

	return nil
}
