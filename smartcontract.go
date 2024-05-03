package main // Package main, Do not change this line.

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Product represents the structure for a product entity
type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// SupplyChainContract defines the smart contract structure
type SupplyChainContract struct {
	contractapi.Contract
}

func (s *SupplyChainContract) getTimestamp(ctx contractapi.TransactionContextInterface) (string, error) {
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	return time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339), nil
}

func (s *SupplyChainContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	curTime, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	assets := []Product{
		{ID: "p1", Name: "Laptop", Status: "Manufactured", Owner: "CompanyA", CreatedAt: curTime, UpdatedAt: curTime, Description: "High-end gaming laptop", Category: "Electronics"},
		{ID: "p2", Name: "Smartphone", Status: "Manufactured", Owner: "CompanyB", CreatedAt: curTime, UpdatedAt: curTime, Description: "Latest model smartphone", Category: "Electronics"},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateProduct creates a new product in the ledger
func (s *SupplyChainContract) CreateProduct(ctx contractapi.TransactionContextInterface, id, name, owner, description, category string) error {
	// Check if the product already exists
	curTime, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	exists, err := s.ProductExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("product with ID %s already exists", id)
	}

	// Create a new product
	product := Product{
		ID:          id,
		Name:        name,
		Status:      "Manufactured",
		Owner:       owner,
		CreatedAt:   curTime,
		UpdatedAt:   curTime,
		Description: description,
		Category:    category,
	}

	// Add the product to the ledger
	assetJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// UpdateProduct allows updating a product's status, owner, description, and category
func (s *SupplyChainContract) UpdateProduct(ctx contractapi.TransactionContextInterface, id string, newStatus string, newOwner string, newDescription string, newCategory string) error {
	// Retrieve the existing product from the ledger
	curTime, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	asset, err := s.QueryProduct(ctx, id)
	if err != nil {
		return err
	}

	// Check if new values are empty, if not, update the corresponding fields
	if newStatus != "" {
		asset.Status = newStatus
	}
	if newOwner != "" {
		asset.Owner = newOwner
	}
	if newDescription != "" {
		asset.Description = newDescription
	}
	if newCategory != "" {
		asset.Category = newCategory
	}

	// Update the UpdatedAt field
	asset.UpdatedAt = curTime

	// Add the updated product to the ledger
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// TransferOwnership changes the owner of a product
func (s *SupplyChainContract) TransferOwnership(ctx contractapi.TransactionContextInterface, id, newOwner string) error {
	// Retrieve the existing product from the ledger
	curTime, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	asset, err := s.QueryProduct(ctx, id)
	if err != nil {
		return err
	}

	asset.Owner = newOwner
	asset.UpdatedAt = curTime
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// QueryProduct retrieves a single product from the ledger by ID
func (s *SupplyChainContract) QueryProduct(ctx contractapi.TransactionContextInterface, id string) (*Product, error) {
	// Retrieve the product from the ledger
	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if productJSON == nil {
		return nil, fmt.Errorf("product with ID %s does not exist", id)
	}

	var product Product
	if err := json.Unmarshal(productJSON, &product); err != nil {
		return nil, err
	}

	return &product, nil
}

// putProduct is a helper method for inserting or updating a product in the ledger
func (s *SupplyChainContract) putProduct(ctx contractapi.TransactionContextInterface, product *Product) error {
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(product.ID, productJSON)
}

// ProductExists is a helper method to check if a product exists in the ledger
func (s *SupplyChainContract) ProductExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return productJSON != nil, nil
}

// GetAllProducts is a helper method to retrieve all products from the ledger
func (s *SupplyChainContract) GetAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var products []*Product
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var product Product
		if err := json.Unmarshal(queryResponse.Value, &product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SupplyChainContract{})
	if err != nil {
		fmt.Printf("Error creating supply chain chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting supply chain chaincode: %s", err.Error())
	}
}
