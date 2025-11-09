package services

import (
	"context"
	"testing"

	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
	"github.com/bouncingmaxt/omniscent-library/gen/go/model"
)

func TestPersonService(t *testing.T) {
	// Skip test if ArangoDB is not available
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Create ArangoDB client
	client, err := clients.NewArangoDBClient()
	if err != nil {
		t.Skipf("Skipping test: failed to create ArangoDB client: %v", err)
	}

	// Create PersonService
	service, err := NewPersonService(client)
	if err != nil {
		t.Fatalf("Failed to create PersonService: %v", err)
	}

	if service == nil {
		t.Error("Expected service to be created")
	}

	// Test CRUD operations
	t.Run("CRUD Operations", func(t *testing.T) {
		// Create a person
		createReq := &geovision.CreatePersonRequest{
			Person: &model.Person{
				Name: "John Doe",
			},
		}

		createResp, err := service.CreatePerson(context.Background(), createReq)
		if err != nil {
			t.Fatalf("Failed to create person: %v", err)
		}

		if createResp.Person == nil {
			t.Fatal("Expected person in create response")
		}

		if createResp.Person.Name != "John Doe" {
			t.Errorf("Expected name to be 'John Doe', got '%s'", createResp.Person.Name)
		}

		if createResp.Person.Key == "" {
			t.Error("Expected person to have a key")
		}

		// Store the key for later use
		personKey := createResp.Person.Key

		// Get the person
		getReq := &geovision.GetPersonRequest{
			Key: personKey,
		}

		getResp, err := service.GetPerson(context.Background(), getReq)
		if err != nil {
			t.Fatalf("Failed to get person: %v", err)
		}

		if getResp.Person == nil {
			t.Fatal("Expected person in get response")
		}

		if getResp.Person.Key != personKey {
			t.Errorf("Expected key to be '%s', got '%s'", personKey, getResp.Person.Key)
		}

		if getResp.Person.Name != "John Doe" {
			t.Errorf("Expected name to be 'John Doe', got '%s'", getResp.Person.Name)
		}

		// Update the person
		updateReq := &geovision.UpdatePersonRequest{
			Person: &model.Person{
				Key:  personKey,
				Name: "Jane Doe",
			},
		}

		updateResp, err := service.UpdatePerson(context.Background(), updateReq)
		if err != nil {
			t.Fatalf("Failed to update person: %v", err)
		}

		if updateResp.Person == nil {
			t.Fatal("Expected person in update response")
		}

		if updateResp.Person.Key != personKey {
			t.Errorf("Expected key to be '%s', got '%s'", personKey, updateResp.Person.Key)
		}

		if updateResp.Person.Name != "Jane Doe" {
			t.Errorf("Expected name to be 'Jane Doe', got '%s'", updateResp.Person.Name)
		}

		// Delete the person
		deleteReq := &geovision.DeletePersonRequest{
			Key: personKey,
		}

		_, err = service.DeletePerson(context.Background(), deleteReq)
		if err != nil {
			t.Fatalf("Failed to delete person: %v", err)
		}

		// Try to get the deleted person (should fail)
		_, err = service.GetPerson(context.Background(), getReq)
		if err == nil {
			t.Error("Expected error when getting deleted person")
		}
	})
}