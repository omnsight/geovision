package services

import (
	"context"
	"testing"

	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
	"github.com/bouncingmaxt/omniscent-library/gen/go/model"
)

func TestOrganizationService(t *testing.T) {
	// Skip test if ArangoDB is not available
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Create ArangoDB client
	client, err := clients.NewArangoDBClient()
	if err != nil {
		t.Skipf("Skipping test: failed to create ArangoDB client: %v", err)
	}

	// Create OrganizationService
	service, err := NewOrganizationService(client)
	if err != nil {
		t.Fatalf("Failed to create OrganizationService: %v", err)
	}

	if service == nil {
		t.Error("Expected service to be created")
	}

	// Test CRUD operations
	t.Run("CRUD Operations", func(t *testing.T) {
		// Create an organization
		createReq := &geovision.CreateOrganizationRequest{
			Organization: &model.Organization{
				Name: "Test Organization",
			},
		}

		createResp, err := service.CreateOrganization(context.Background(), createReq)
		if err != nil {
			t.Fatalf("Failed to create organization: %v", err)
		}

		if createResp.Organization == nil {
			t.Fatal("Expected organization in create response")
		}

		if createResp.Organization.Name != "Test Organization" {
			t.Errorf("Expected name to be 'Test Organization', got '%s'", createResp.Organization.Name)
		}

		if createResp.Organization.Key == "" {
			t.Error("Expected organization to have a key")
		}

		// Store the key for later use
		orgKey := createResp.Organization.Key

		// Get the organization
		getReq := &geovision.GetOrganizationRequest{
			Key: orgKey,
		}

		getResp, err := service.GetOrganization(context.Background(), getReq)
		if err != nil {
			t.Fatalf("Failed to get organization: %v", err)
		}

		if getResp.Organization == nil {
			t.Fatal("Expected organization in get response")
		}

		if getResp.Organization.Key != orgKey {
			t.Errorf("Expected key to be '%s', got '%s'", orgKey, getResp.Organization.Key)
		}

		if getResp.Organization.Name != "Test Organization" {
			t.Errorf("Expected name to be 'Test Organization', got '%s'", getResp.Organization.Name)
		}

		// Update the organization
		updateReq := &geovision.UpdateOrganizationRequest{
			Organization: &model.Organization{
				Key:  orgKey,
				Name: "Updated Test Organization",
			},
		}

		updateResp, err := service.UpdateOrganization(context.Background(), updateReq)
		if err != nil {
			t.Fatalf("Failed to update organization: %v", err)
		}

		if updateResp.Organization == nil {
			t.Fatal("Expected organization in update response")
		}

		if updateResp.Organization.Key != orgKey {
			t.Errorf("Expected key to be '%s', got '%s'", orgKey, updateResp.Organization.Key)
		}

		if updateResp.Organization.Name != "Updated Test Organization" {
			t.Errorf("Expected name to be 'Updated Test Organization', got '%s'", updateResp.Organization.Name)
		}

		// Delete the organization
		deleteReq := &geovision.DeleteOrganizationRequest{
			Key: orgKey,
		}

		_, err = service.DeleteOrganization(context.Background(), deleteReq)
		if err != nil {
			t.Fatalf("Failed to delete organization: %v", err)
		}

		// Try to get the deleted organization (should fail)
		_, err = service.GetOrganization(context.Background(), getReq)
		if err == nil {
			t.Error("Expected error when getting deleted organization")
		}
	})
}