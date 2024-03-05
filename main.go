package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Define the struct to match the JSON structure
type Response struct {
	Resources []Resource `json:"resources"`
	Metadata  Metadata   `json:"metadata"`
}

type Resource struct {
	ID                   string  `json:"id"`
	Nonce                int     `json:"nonce"`
	TxnType              string  `json:"txnType"`
	BurnHash             string  `json:"burnHash"`
	MintHash             *string `json:"mintHash"` // Pointer to string to handle null values
	TransferHash         string  `json:"transferHash"`
	From                 string  `json:"from"`
	Destination          string  `json:"destination"`
	Minter               *string `json:"minter"` // Pointer to string to handle null values
	FromNetwork          string  `json:"fromNetwork"`
	DestinationNetwork   string  `json:"destinationNetwork"`
	Amount               string  `json:"amount"`
	Denom                string  `json:"denom"`
	Status               string  `json:"status"`
	Timestamp            string  `json:"timestamp"`
	CreatedAt            string  `json:"createdAt"`
	Details              *string `json:"details"` // Pointer to string to handle null values
	DestinationTimestamp string  `json:"destinationTimestamp"`
}

type Metadata struct {
	Count int `json:"count"`
}

func fetchTransactions(baseURL string, offset, limit int) ([]Resource, error) {
	// Construct the URL with the current offset and limit
	url := fmt.Sprintf("%s&offset=%d&limit=%d", baseURL, offset, limit)

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the body of the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into the Response struct
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.Resources, nil
}

func main() {
	fetch()
	read()
}

func read() {
	records, _ := readFromCSV("3-4-2024.csv")
	toTotal := 0
	fromTotal := 0
	for _, record := range records {
		if record.FromNetwork == "noble" {
			res, _ := strconv.Atoi(record.Amount)
			fromTotal += res
		} else if record.DestinationNetwork == "noble" {
			res, _ := strconv.Atoi(record.Amount)
			toTotal += res
		}
	}
	println("to Noble")
	println(toTotal / 1000000)
	println("from Noble")
	println(fromTotal / 1000000)
}

func fetch() {
	baseURL := "https://usdc.range.org/api/usdcTrail/transactions?txnType=MAINNET&showPending=false&txnHash="
	offset := 0
	limit := 1000
	var allResources []Resource

	for {
		resources, err := fetchTransactions(baseURL, offset, limit)
		if err != nil {
			fmt.Println("Error fetching transactions: ", err)
			break
		}

		// Append the fetched resources to the allResources slice
		allResources = append(allResources, resources...)

		// Check if the number of resources fetched is less than the limit, indicating the last page
		if len(resources) < limit {
			break
		}

		// Increase the offset for the next iteration
		offset += limit

		// Optional: Sleep between requests to avoid hitting rate limits
		time.Sleep(1 * time.Second)
	}

	// Process the allResources slice as needed
	fmt.Printf("Total transactions fetched: %d\n", len(allResources))
	writeToCSV(allResources, "3-4-2024.csv")
}

func writeToCSV(resources []Resource, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header
	if err := writer.Write([]string{"ID", "Nonce", "TxnType", "BurnHash", "MintHash", "TransferHash", "From", "Destination", "Minter", "FromNetwork", "DestinationNetwork", "Amount", "Denom", "Status", "Timestamp", "CreatedAt", "Details", "DestinationTimestamp"}); err != nil {
		return err
	}

	// Write the data
	for _, resource := range resources {
		record := []string{
			resource.ID,
			fmt.Sprint(resource.Nonce),
			resource.TxnType,
			resource.BurnHash,
			stringOrNil(resource.MintHash),
			resource.TransferHash,
			resource.From,
			resource.Destination,
			stringOrNil(resource.Minter),
			resource.FromNetwork,
			resource.DestinationNetwork,
			resource.Amount,
			resource.Denom,
			resource.Status,
			resource.Timestamp,
			resource.CreatedAt,
			stringOrNil(resource.Details),
			resource.DestinationTimestamp,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to handle nil pointer for string fields
func stringOrNil(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func readFromCSV(filename string) ([]Resource, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var resources []Resource
	for i, record := range records {
		// Skip the header
		if i == 0 {
			continue
		}
		nonce, _ := strconv.Atoi(record[1])
		resources = append(resources, Resource{
			ID:                   record[0],
			Nonce:                nonce,
			TxnType:              record[2],
			BurnHash:             record[3],
			MintHash:             optionalString(record[4]),
			TransferHash:         record[5],
			From:                 record[6],
			Destination:          record[7],
			Minter:               optionalString(record[8]),
			FromNetwork:          record[9],
			DestinationNetwork:   record[10],
			Amount:               record[11],
			Denom:                record[12],
			Status:               record[13],
			Timestamp:            record[14],
			CreatedAt:            record[15],
			Details:              optionalString(record[16]),
			DestinationTimestamp: record[17],
		})
	}

	return resources, nil
}

// Helper function to convert empty string back to nil pointer for string fields
func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
