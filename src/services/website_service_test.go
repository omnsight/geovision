package services

import (
	"context"
	"os"
	"testing"

	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
	"github.com/bouncingmaxt/omniscent-library/gen/go/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Set environment variables for testing
	os.Setenv("ARANGO_URL", "http://localhost:8529")
	os.Setenv("ARANGO_DB", "geovision")
	os.Setenv("ARANGO_USERNAME", "root")
	os.Setenv("ARANGO_PASSWORD", "0123")

	// Run tests
	code := m.Run()

	// Exit with the same code as the tests
	os.Exit(code)
}

func TestWebsiteService(t *testing.T) {
	// Skip test if ArangoDB is not available
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Create ArangoDB client
	client, err := clients.NewArangoDBClient()
	if err != nil {
		t.Skipf("Skipping test: failed to create ArangoDB client: %v", err)
	}

	// Create WebsiteService
	service, err := NewWebsiteService(client)
	if err != nil {
		t.Fatalf("Failed to create WebsiteService: %v", err)
	}

	if service == nil {
		t.Error("Expected service to be created")
	}

	// Test CRUD operations
	t.Run("CRUD Operations", func(t *testing.T) {
		// Create a website
		createReq := &geovision.CreateWebsiteRequest{
			Website: &model.Website{
				Url: "https://example.com",
			},
		}

		createResp, err := service.CreateWebsite(context.Background(), createReq)
		if err != nil {
			t.Fatalf("Failed to create website: %v", err)
		}

		if createResp.Website == nil {
			t.Fatal("Expected website in create response")
		}

		if createResp.Website.Url != "https://example.com" {
			t.Errorf("Expected URL to be 'https://example.com', got '%s'", createResp.Website.Url)
		}

		if createResp.Website.Key == "" {
			t.Error("Expected website to have a key")
		}

		// Store the key for later use
		websiteKey := createResp.Website.Key

		// Get the website
		getReq := &geovision.GetWebsiteRequest{
			Key: websiteKey,
		}

		getResp, err := service.GetWebsite(context.Background(), getReq)
		if err != nil {
			t.Fatalf("Failed to get website: %v", err)
		}

		if getResp.Website == nil {
			t.Fatal("Expected website in get response")
		}

		if getResp.Website.Key != websiteKey {
			t.Errorf("Expected key to be '%s', got '%s'", websiteKey, getResp.Website.Key)
		}

		if getResp.Website.Url != "https://example.com" {
			t.Errorf("Expected URL to be 'https://example.com', got '%s'", getResp.Website.Url)
		}

		// Update the website
		updateReq := &geovision.UpdateWebsiteRequest{
			Website: &model.Website{
				Key: websiteKey,
				Url: "https://updated-example.com",
			},
		}

		updateResp, err := service.UpdateWebsite(context.Background(), updateReq)
		if err != nil {
			t.Fatalf("Failed to update website: %v", err)
		}

		if updateResp.Website == nil {
			t.Fatal("Expected website in update response")
		}

		if updateResp.Website.Key != websiteKey {
			t.Errorf("Expected key to be '%s', got '%s'", websiteKey, updateResp.Website.Key)
		}

		if updateResp.Website.Url != "https://updated-example.com" {
			t.Errorf("Expected URL to be 'https://updated-example.com', got '%s'", updateResp.Website.Url)
		}

		// Delete the website
		deleteReq := &geovision.DeleteWebsiteRequest{
			Key: websiteKey,
		}

		_, err = service.DeleteWebsite(context.Background(), deleteReq)
		if err != nil {
			t.Fatalf("Failed to delete website: %v", err)
		}

		// Try to get the deleted website (should fail)
		_, err = service.GetWebsite(context.Background(), getReq)
		if err == nil {
			t.Error("Expected error when getting deleted website")
		} else {
			// Check that we get the expected NotFound error
			if status.Code(err) != codes.NotFound {
				t.Errorf("Expected NotFound error, got: %v", err)
			}
		}
	})
}
