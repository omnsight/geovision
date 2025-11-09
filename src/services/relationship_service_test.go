package services

import (
	"context"
	"testing"

	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
	"github.com/bouncingmaxt/omniscent-library/gen/go/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRelationshipService(t *testing.T) {
	// Skip test if ArangoDB is not available
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Create ArangoDB client
	client, err := clients.NewArangoDBClient()
	if err != nil {
		t.Skipf("Skipping test: failed to create ArangoDB client: %v", err)
	}

	// Create required services
	personService, err := NewPersonService(client)
	if err != nil {
		t.Fatalf("Failed to create PersonService: %v", err)
	}

	orgService, err := NewOrganizationService(client)
	if err != nil {
		t.Fatalf("Failed to create OrganizationService: %v", err)
	}

	// Create RelationshipService
	service, err := NewRelationshipService(client)
	if err != nil {
		t.Fatalf("Failed to create RelationshipService: %v", err)
	}

	if service == nil {
		t.Error("Expected service to be created")
	}

	// Test validation
	t.Run("Validation", func(t *testing.T) {
		// Test with nil relationship
		_, err := service.CreateRelationship(context.Background(), &geovision.CreateRelationshipRequest{
			Relationship: nil,
		})
		if err == nil {
			t.Error("Expected error when relationship is nil")
		} else {
			if status.Code(err) != codes.InvalidArgument {
				t.Errorf("Expected InvalidArgument error, got %v", status.Code(err))
			}
		}

		// Test with missing from field
		_, err = service.CreateRelationship(context.Background(), &geovision.CreateRelationshipRequest{
			Relationship: &model.Relation{
				Name: "employment",
				To:   "organizations/acme_corp",
			},
		})
		if err == nil {
			t.Error("Expected error when from field is empty")
		} else {
			if status.Code(err) != codes.InvalidArgument {
				t.Errorf("Expected InvalidArgument error, got %v", status.Code(err))
			}
		}

		// Test with missing to field
		_, err = service.CreateRelationship(context.Background(), &geovision.CreateRelationshipRequest{
			Relationship: &model.Relation{
				Name: "employment",
				From: "persons/john_doe",
			},
		})
		if err == nil {
			t.Error("Expected error when to field is empty")
		} else {
			if status.Code(err) != codes.InvalidArgument {
				t.Errorf("Expected InvalidArgument error, got %v", status.Code(err))
			}
		}

		// Test with missing name field
		_, err = service.CreateRelationship(context.Background(), &geovision.CreateRelationshipRequest{
			Relationship: &model.Relation{
				From: "persons/john_doe",
				To:   "organizations/acme_corp",
			},
		})
		if err == nil {
			t.Error("Expected error when name field is empty")
		} else {
			if status.Code(err) != codes.InvalidArgument {
				t.Errorf("Expected InvalidArgument error, got %v", status.Code(err))
			}
		}
	})

	// Test CRUD operations
	t.Run("CRUD Operations", func(t *testing.T) {
		// First create a person
		personCreateReq := &geovision.CreatePersonRequest{
			Person: &model.Person{
				Name: "John Doe",
			},
		}

		personCreateResp, err := personService.CreatePerson(context.Background(), personCreateReq)
		if err != nil {
			t.Fatalf("Failed to create person: %v", err)
		}

		// Create an organization
		orgCreateReq := &geovision.CreateOrganizationRequest{
			Organization: &model.Organization{
				Name: "Acme Corp",
			},
		}

		orgCreateResp, err := orgService.CreateOrganization(context.Background(), orgCreateReq)
		if err != nil {
			t.Fatalf("Failed to create organization: %v", err)
		}

		// Create a relationship using the created entities
		createReq := &geovision.CreateRelationshipRequest{
			Relationship: &model.Relation{
				Name: "employment",
				From: "persons/" + personCreateResp.Person.Key,
				To:   "organizations/" + orgCreateResp.Organization.Key,
			},
		}

		createResp, err := service.CreateRelationship(context.Background(), createReq)
		if err != nil {
			t.Fatalf("Failed to create relationship: %v", err)
		}

		if createResp.Relationship == nil {
			t.Fatal("Expected relationship in create response")
		}

		if createResp.Relationship.Name != "employment" {
			t.Errorf("Expected name to be 'employment', got '%s'", createResp.Relationship.Name)
		}

		if createResp.Relationship.Id == "" {
			t.Error("Expected relationship to have an id")
		}

		// Store the id for later use
		relationshipId := createResp.Relationship.Id

		// Update the relationship
		updateReq := &geovision.UpdateRelationshipRequest{
			Id: relationshipId,
			Relationship: &model.Relation{
				Name: "contractor",
			},
		}

		updateResp, err := service.UpdateRelationship(context.Background(), updateReq)
		if err != nil {
			t.Fatalf("Failed to update relationship: %v", err)
		}

		if updateResp.Relationship == nil {
			t.Fatal("Expected relationship in update response")
		}

		if updateResp.Relationship.Id != relationshipId {
			t.Errorf("Expected id to be '%s', got '%s'", relationshipId, updateResp.Relationship.Id)
		}

		if updateResp.Relationship.Name != "contractor" {
			t.Errorf("Expected name to be 'contractor', got '%s'", updateResp.Relationship.Name)
		}

		// Delete the relationship
		deleteReq := &geovision.DeleteRelationshipRequest{
			Id: relationshipId,
		}

		_, err = service.DeleteRelationship(context.Background(), deleteReq)
		if err != nil {
			t.Fatalf("Failed to delete relationship: %v", err)
		}

		// Clean up - delete person and organization
		_, err = personService.DeletePerson(context.Background(), &geovision.DeletePersonRequest{
			Key: personCreateResp.Person.Key,
		})
		if err != nil {
			t.Fatalf("Failed to delete person: %v", err)
		}

		_, err = orgService.DeleteOrganization(context.Background(), &geovision.DeleteOrganizationRequest{
			Key: orgCreateResp.Organization.Key,
		})
		if err != nil {
			t.Fatalf("Failed to delete organization: %v", err)
		}
	})
}
