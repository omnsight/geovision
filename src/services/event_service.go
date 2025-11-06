package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
)

type EventService struct {
	geovision.UnimplementedEventServiceServer

	DBClient *clients.ArangoDBClient
}

func NewEventService(client *clients.ArangoDBClient) (*EventService, error) {
	service := &EventService{
		DBClient: client,
	}

	// Create events collection
	ctx := context.Background()
	exists, err := client.DB.CollectionExists(ctx, "events")
	if err != nil {
		return nil, fmt.Errorf("failed to check events collection existence: %v", err)
	}

	if !exists {
		_, err = client.DB.CreateCollection(ctx, "events", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create events collection: %v", err)
		}
		fmt.Printf("✅ Created events collection\n")
	}

	return service, nil
}

func (s *EventService) GetEvents(ctx context.Context, req *geovision.GetEventsRequest) (*geovision.GetEventsResponse, error) {
	return nil, errors.New("method GetEvents not implemented")
}

func (s *EventService) GetEvent(ctx context.Context, req *geovision.GetEventRequest) (*geovision.GetEventResponse, error) {
	return nil, errors.New("method GetEvent not implemented")
}

func (s *EventService) GetRelatedEvents(ctx context.Context, req *geovision.GetRelatedEventsRequest) (*geovision.GetRelatedEventsResponse, error) {
	return nil, errors.New("method GetRelatedEvents not implemented")
}

func (s *EventService) CreateEvent(ctx context.Context, req *geovision.CreateEventRequest) (*geovision.CreateEventResponse, error) {
	return nil, errors.New("method CreateEvent not implemented")
}

func (s *EventService) UpdateEvent(ctx context.Context, req *geovision.UpdateEventRequest) (*geovision.UpdateEventResponse, error) {
	return nil, errors.New("method UpdateEvent not implemented")
}

func (s *EventService) DeleteEvent(ctx context.Context, req *geovision.DeleteEventRequest) (*geovision.DeleteEventResponse, error) {
	return nil, errors.New("method DeleteEvent not implemented")
}
