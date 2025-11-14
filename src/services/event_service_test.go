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

func TestEventService(t *testing.T) {
	// Skip test if ArangoDB is not available
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Create ArangoDB client
	client, err := clients.NewArangoDBClient()
	if err != nil {
		t.Skipf("Skipping test: failed to create ArangoDB client: %v", err)
	}

	// Create EventService
	service, err := NewEventService(client)
	if err != nil {
		t.Fatalf("Failed to create EventService: %v", err)
	}

	if service == nil {
		t.Error("Expected service to be created")
	}

	// Create PersonService
	personService, err := NewPersonService(client)
	if err != nil {
		t.Fatalf("Failed to create PersonService: %v", err)
	}

	// Create OrganizationService
	orgService, err := NewOrganizationService(client)
	if err != nil {
		t.Fatalf("Failed to create OrganizationService: %v", err)
	}

	// Create RelationshipService
	relationshipService, err := NewRelationshipService(client)
	if err != nil {
		t.Fatalf("Failed to create RelationshipService: %v", err)
	}

	// Test GetEvents validation
	t.Run("GetEvents Validation", func(t *testing.T) {
		// Test with missing start_time
		_, err := service.GetEvents(context.Background(), &geovision.GetEventsRequest{
			EndTime: 100,
		})
		if err == nil {
			t.Error("Expected error when start_time is missing")
		} else {
			if status.Code(err) != codes.InvalidArgument {
				t.Errorf("Expected InvalidArgument error, got %v", status.Code(err))
			}
		}

		// Test with missing end_time
		_, err = service.GetEvents(context.Background(), &geovision.GetEventsRequest{
			StartTime: 100,
		})
		if err == nil {
			t.Error("Expected error when end_time is missing")
		} else {
			if status.Code(err) != codes.InvalidArgument {
				t.Errorf("Expected InvalidArgument error, got %v", status.Code(err))
			}
		}

		// Test with start_time after end_time
		_, err = service.GetEvents(context.Background(), &geovision.GetEventsRequest{
			StartTime: 200,
			EndTime:   100,
		})
		if err == nil {
			t.Error("Expected error when start_time is after end_time")
		} else {
			if status.Code(err) != codes.InvalidArgument {
				t.Errorf("Expected InvalidArgument error, got %v", status.Code(err))
			}
		}
	})

	// Test GetEventRelatedEntities validation
	t.Run("GetEventRelatedEntities Validation", func(t *testing.T) {
		// Test with missing event key
		_, err := service.GetEventRelatedEntities(context.Background(), &geovision.GetEventRelatedEntitiesRequest{})
		if err == nil {
			t.Error("Expected error when event key is missing")
		} else {
			if status.Code(err) != codes.InvalidArgument {
				t.Errorf("Expected InvalidArgument error, got %v", status.Code(err))
			}
		}
	})

	// Test CRUD operations
	t.Run("CRUD Operations", func(t *testing.T) {
		// Create a person
		createPersonReq := &geovision.CreatePersonRequest{
			Person: &model.Person{
				Name: "Test Person",
			},
		}

		createPersonResp, err := personService.CreatePerson(context.Background(), createPersonReq)
		if err != nil {
			t.Fatalf("Failed to create person: %v", err)
		}

		if createPersonResp.Person == nil {
			t.Fatal("Expected person in create response")
		}

		if createPersonResp.Person.Key == "" {
			t.Error("Expected person to have a key")
		}

		// Create an organization
		createOrgReq := &geovision.CreateOrganizationRequest{
			Organization: &model.Organization{
				Name: "Test Organization",
			},
		}

		createOrgResp, err := orgService.CreateOrganization(context.Background(), createOrgReq)
		if err != nil {
			t.Fatalf("Failed to create organization: %v", err)
		}

		if createOrgResp.Organization == nil {
			t.Fatal("Expected organization in create response")
		}

		if createOrgResp.Organization.Key == "" {
			t.Error("Expected organization to have a key")
		}

		// Create multiple events
		createEvent1Req := &geovision.CreateEventRequest{
			Event: &model.Event{
				HappenedAt: 1000,
			},
		}

		createEvent1Resp, err := service.CreateEvent(context.Background(), createEvent1Req)
		if err != nil {
			t.Fatalf("Failed to create event 1: %v", err)
		}

		if createEvent1Resp.Event == nil {
			t.Fatal("Expected event 1 in create response")
		}

		if createEvent1Resp.Event.Key == "" {
			t.Error("Expected event 1 to have a key")
		}

		// Check the happened_at value of the created event
		if createEvent1Resp.Event.HappenedAt != 1000 {
			t.Errorf("Expected event 1 happened_at to be 1000, got %d", createEvent1Resp.Event.HappenedAt)
		}

		createEvent2Req := &geovision.CreateEventRequest{
			Event: &model.Event{
				HappenedAt: 2000,
			},
		}

		createEvent2Resp, err := service.CreateEvent(context.Background(), createEvent2Req)
		if err != nil {
			t.Fatalf("Failed to create event 2: %v", err)
		}

		if createEvent2Resp.Event == nil {
			t.Fatal("Expected event 2 in create response")
		}

		if createEvent2Resp.Event.Key == "" {
			t.Error("Expected event 2 to have a key")
		}

		// Check the happened_at value of the created event
		if createEvent2Resp.Event.HappenedAt != 2000 {
			t.Errorf("Expected event 2 happened_at to be 2000, got %d", createEvent2Resp.Event.HappenedAt)
		}

		// Create outbound relationships from events to organization
		// This is what GetEventRelatedEntities looks for
		createRel1Req := &geovision.CreateRelationshipRequest{
			Relationship: &model.Relation{
				From: "events/" + createEvent1Resp.Event.Key,
				To:   "organizations/" + createOrgResp.Organization.Key,
				Name: "hosted_by",
			},
		}

		createRel1Resp, err := relationshipService.CreateRelationship(context.Background(), createRel1Req)
		if err != nil {
			t.Fatalf("Failed to create relationship 1: %v", err)
		}

		if createRel1Resp.Relationship == nil {
			t.Fatal("Expected relationship 1 in create response")
		}

		createRel2Req := &geovision.CreateRelationshipRequest{
			Relationship: &model.Relation{
				From: "events/" + createEvent2Resp.Event.Key,
				To:   "organizations/" + createOrgResp.Organization.Key,
				Name: "hosted_by",
			},
		}

		createRel2Resp, err := relationshipService.CreateRelationship(context.Background(), createRel2Req)
		if err != nil {
			t.Fatalf("Failed to create relationship 2: %v", err)
		}

		if createRel2Resp.Relationship == nil {
			t.Fatal("Expected relationship 2 in create response")
		}

		// Create a relationship between the two events
		createEventRelReq := &geovision.CreateRelationshipRequest{
			Relationship: &model.Relation{
				From: "events/" + createEvent1Resp.Event.Key,
				To:   "events/" + createEvent2Resp.Event.Key,
				Name: "related_to",
			},
		}

		createEventRelResp, err := relationshipService.CreateRelationship(context.Background(), createEventRelReq)
		if err != nil {
			t.Fatalf("Failed to create event relationship: %v", err)
		}

		if createEventRelResp.Relationship == nil {
			t.Fatal("Expected event relationship in create response")
		}

		// Test GetEvents with valid time range
		getEventsResp, err := service.GetEvents(context.Background(), &geovision.GetEventsRequest{
			StartTime: 1,
			EndTime:   9999999999, // Far in the future
		})
		if err != nil {
			t.Fatalf("GetEvents failed: %v", err)
		}

		// Check that we got a response
		if getEventsResp == nil {
			t.Error("Expected getEventsResp to not be nil")
		}

		// Check that we have events
		if len(getEventsResp.Events) == 0 {
			t.Error("Expected at least one event")
		} else {
			t.Logf("Found %d events", len(getEventsResp.Events))
			for i, event := range getEventsResp.Events {
				t.Logf("Event %d: Key=%s, HappenedAt=%d", i, event.Key, event.HappenedAt)
			}
		}

		// Check that we have relations
		if len(getEventsResp.Relations) == 0 {
			t.Error("Expected at least one relation between events")
		} else {
			t.Logf("Found %d relations", len(getEventsResp.Relations))
		}

		// Test GetEventRelatedEntities with valid event key
		getRelatedResp, err := service.GetEventRelatedEntities(context.Background(), &geovision.GetEventRelatedEntitiesRequest{
			Key: createEvent1Resp.Event.Key,
		})
		if err != nil {
			t.Fatalf("Failed to get related entities: %v", err)
		}

		// Check that we got a response
		if getRelatedResp == nil {
			t.Error("Expected getRelatedResp to not be nil")
		}

		// Now check that we have at least one related entity (the organization we created)
		if len(getRelatedResp.Entities) == 0 {
			t.Error("Expected at least one related entity")
		} else {
			t.Logf("Found %d related entities", len(getRelatedResp.Entities))
		}

		// Store the keys for later use
		event1Key := createEvent1Resp.Event.Key
		event2Key := createEvent2Resp.Event.Key

		// Delete the relationships
		deleteRel1Req := &geovision.DeleteRelationshipRequest{
			Id: createRel1Resp.Relationship.Id,
		}

		_, err = relationshipService.DeleteRelationship(context.Background(), deleteRel1Req)
		if err != nil {
			t.Fatalf("Failed to delete relationship 1: %v", err)
		}

		deleteRel2Req := &geovision.DeleteRelationshipRequest{
			Id: createRel2Resp.Relationship.Id,
		}

		_, err = relationshipService.DeleteRelationship(context.Background(), deleteRel2Req)
		if err != nil {
			t.Fatalf("Failed to delete relationship 2: %v", err)
		}

		deleteEventRelReq := &geovision.DeleteRelationshipRequest{
			Id: createEventRelResp.Relationship.Id,
		}

		_, err = relationshipService.DeleteRelationship(context.Background(), deleteEventRelReq)
		if err != nil {
			t.Fatalf("Failed to delete event relationship: %v", err)
		}

		// Delete the person
		deletePersonReq := &geovision.DeletePersonRequest{
			Key: createPersonResp.Person.Key,
		}

		_, err = personService.DeletePerson(context.Background(), deletePersonReq)
		if err != nil {
			t.Fatalf("Failed to delete person: %v", err)
		}

		// Delete the organization
		deleteOrgReq := &geovision.DeleteOrganizationRequest{
			Key: createOrgResp.Organization.Key,
		}

		_, err = orgService.DeleteOrganization(context.Background(), deleteOrgReq)
		if err != nil {
			t.Fatalf("Failed to delete organization: %v", err)
		}

		// Delete the events
		deleteEvent1Req := &geovision.DeleteEventRequest{
			Key: event1Key,
		}

		_, err = service.DeleteEvent(context.Background(), deleteEvent1Req)
		if err != nil {
			t.Fatalf("Failed to delete event 1: %v", err)
		}

		deleteEvent2Req := &geovision.DeleteEventRequest{
			Key: event2Key,
		}

		_, err = service.DeleteEvent(context.Background(), deleteEvent2Req)
		if err != nil {
			t.Fatalf("Failed to delete event 2: %v", err)
		}
	})
}
