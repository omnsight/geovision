package services

import (
	"context"
	"testing"

	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
	"github.com/bouncingmaxt/omniscent-library/gen/go/model"
)

func TestSourceService(t *testing.T) {
	// Skip test if ArangoDB is not available
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Create ArangoDB client
	client, err := clients.NewArangoDBClient()
	if err != nil {
		t.Skipf("Skipping test: failed to create ArangoDB client: %v", err)
	}

	// Create SourceService
	service, err := NewSourceService(client)
	if err != nil {
		t.Fatalf("Failed to create SourceService: %v", err)
	}

	if service == nil {
		t.Error("Expected service to be created")
	}

	// Test CRUD operations
	t.Run("CRUD Operations", func(t *testing.T) {
		// Create a source
		createReq := &geovision.CreateSourceRequest{
			Source: &model.Source{
				Name: "Test Source",
			},
		}

		createResp, err := service.CreateSource(context.Background(), createReq)
		if err != nil {
			t.Fatalf("Failed to create source: %v", err)
		}

		if createResp.Source == nil {
			t.Fatal("Expected source in create response")
		}

		if createResp.Source.Name != "Test Source" {
			t.Errorf("Expected name to be 'Test Source', got '%s'", createResp.Source.Name)
		}

		if createResp.Source.Key == "" {
			t.Error("Expected source to have a key")
		}

		// Store the key for later use
		sourceKey := createResp.Source.Key

		// Get the source
		getReq := &geovision.GetSourceRequest{
			Key: sourceKey,
		}

		getResp, err := service.GetSource(context.Background(), getReq)
		if err != nil {
			t.Fatalf("Failed to get source: %v", err)
		}

		if getResp.Source == nil {
			t.Fatal("Expected source in get response")
		}

		if getResp.Source.Key != sourceKey {
			t.Errorf("Expected key to be '%s', got '%s'", sourceKey, getResp.Source.Key)
		}

		if getResp.Source.Name != "Test Source" {
			t.Errorf("Expected name to be 'Test Source', got '%s'", getResp.Source.Name)
		}

		// Update the source
		updateReq := &geovision.UpdateSourceRequest{
			Source: &model.Source{
				Key:  sourceKey,
				Name: "Updated Test Source",
			},
		}

		updateResp, err := service.UpdateSource(context.Background(), updateReq)
		if err != nil {
			t.Fatalf("Failed to update source: %v", err)
		}

		if updateResp.Source == nil {
			t.Fatal("Expected source in update response")
		}

		if updateResp.Source.Key != sourceKey {
			t.Errorf("Expected key to be '%s', got '%s'", sourceKey, updateResp.Source.Key)
		}

		if updateResp.Source.Name != "Updated Test Source" {
			t.Errorf("Expected name to be 'Updated Test Source', got '%s'", updateResp.Source.Name)
		}

		// Delete the source
		deleteReq := &geovision.DeleteSourceRequest{
			Key: sourceKey,
		}

		_, err = service.DeleteSource(context.Background(), deleteReq)
		if err != nil {
			t.Fatalf("Failed to delete source: %v", err)
		}

		// Try to get the deleted source (should fail)
		_, err = service.GetSource(context.Background(), getReq)
		if err == nil {
			t.Error("Expected error when getting deleted source")
		}
	})
}
