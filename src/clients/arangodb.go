package clients

import (
	"context"
	"fmt"
	"os"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/sirupsen/logrus"
)

// ArangoDBClient represents a connection to ArangoDB
type ArangoDBClient struct {
	Client driver.Client
	DB     driver.Database
}

// NewArangoDBClient creates a new ArangoDB client and connects to the database
func NewArangoDBClient() (*ArangoDBClient, error) {
	// Get ArangoDB connection details from environment variables
	arangoDBURL := os.Getenv("ARANGO_URL")
	if arangoDBURL == "" {
		logrus.Fatal("missing environment variable ARANGO_URL")
	}

	databaseName := os.Getenv("ARANGO_DB")
	if databaseName == "" {
		logrus.Fatal("missing environment variable ARANGO_DB")
	}

	username := os.Getenv("ARANGO_USERNAME")
	if username == "" {
		logrus.Fatal("missing environment variable ARANGO_USERNAME")
	}

	password := os.Getenv("ARANGO_PASSWORD")
	if password == "" {
		logrus.Fatal("missing environment variable ARANGO_PASSWORD")
	}

	// Connect to ArangoDB
	logrus.Infof("Connecting to ArangoDB at %s using (%s, %s)", arangoDBURL, username, password)
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{arangoDBURL},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %v", err)
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(username, password),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	// Create or get database
	db, err := createOrGetDatabase(client, databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %v", err)
	}

	return &ArangoDBClient{
		Client: client,
		DB:     db,
	}, nil
}

// GetDatabase returns the database instance
func (c *ArangoDBClient) GetDatabase() driver.Database {
	return c.DB
}

// GetClient returns the client instance
func (c *ArangoDBClient) GetClient() driver.Client {
	return c.Client
}

func createOrGetDatabase(client driver.Client, dbName string) (driver.Database, error) {
	ctx := context.Background()

	// Check if database exists
	exists, err := client.DatabaseExists(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %v", err)
	}

	if exists {
		fmt.Printf("📁 Using existing database: %s\n", dbName)
		return client.Database(ctx, dbName)
	}

	// Create new database
	fmt.Printf("🆕 Creating new database: %s\n", dbName)
	db, err := client.CreateDatabase(ctx, dbName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %v", err)
	}

	return db, nil
}
