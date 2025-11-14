package clients

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/sirupsen/logrus"
)

// ArangoDBClient represents a connection to ArangoDB
type ArangoDBClient struct {
	Client     driver.Client
	DB         driver.Database
	OsintGraph driver.Graph
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

	ctx := context.Background()
	osintGraph, err := CreateOrGetGraph(db, ctx, "osint", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get osint graph: %v", err)
	}

	arangoClient := ArangoDBClient{
		Client:     client,
		DB:         db,
		OsintGraph: osintGraph,
	}

	return &arangoClient, nil
}

func (adb *ArangoDBClient) ParseDocID(id string) (collectionName string, key string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid document ID format: %s", id)
	}
	return parts[0], parts[1], nil
}

// GetDatabase returns the database instance
func (c *ArangoDBClient) GetDatabase() driver.Database {
	return c.DB
}

// GetClient returns the client instance
func (c *ArangoDBClient) GetClient() driver.Client {
	return c.Client
}

func (c *ArangoDBClient) GetCreateCollection(ctx context.Context, name string, options driver.CreateVertexCollectionOptions) (driver.Collection, error) {
	collection, err := c.DB.Collection(ctx, name)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			collection, err = c.OsintGraph.CreateVertexCollectionWithOptions(ctx, name, options)
			if err != nil {
				return nil, err
			}
		}
		return collection, nil
	} else {
		return collection, nil
	}
}

func (c *ArangoDBClient) GetCreateEdgeCollection(ctx context.Context, name string, constraints driver.VertexConstraints, options driver.CreateEdgeCollectionOptions) (driver.Collection, error) {
	collection, err := c.DB.Collection(ctx, name)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			collection, err = c.OsintGraph.CreateEdgeCollectionWithOptions(ctx, name, constraints, options)
			if err != nil {
				return nil, err
			}
		}
		return collection, nil
	} else {
		return collection, nil
	}
}

func CreateOrGetGraph(db driver.Database, ctx context.Context, name string, options *driver.CreateGraphOptions) (driver.Graph, error) {
	graph, err := db.Graph(ctx, name)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			graph, err = db.CreateGraphV2(ctx, name, options)
			if err != nil {
				return nil, err
			}
			return graph, nil
		}
		return nil, err
	} else {
		return graph, nil
	}
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
