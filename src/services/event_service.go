package services

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/geovision/src/helpers"
	"github.com/bouncingmaxt/geovision/src/logging"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
	"github.com/bouncingmaxt/omniscent-library/gen/go/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventService struct {
	geovision.UnimplementedEventServiceServer

	DBClient   *clients.ArangoDBClient
	Collection driver.Collection
}

func NewEventService(client *clients.ArangoDBClient) (*EventService, error) {
	// Create events collection
	ctx := context.Background()
	collection, err := client.GetCreateCollection(ctx, "events", driver.CreateVertexCollectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get or create events collection: %v", err)
	}
	logrus.Infof("✅ Initialized collection %s", collection.Name())

	collection.EnsurePersistentIndex(ctx, []string{"happened_at"}, &driver.EnsurePersistentIndexOptions{
		InBackground: true,
	})

	service := &EventService{
		DBClient:   client,
		Collection: collection,
	}
	return service, nil
}

func (s *EventService) GetEvent(ctx context.Context, req *geovision.GetEventRequest) (*geovision.GetEventResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Getting event with ID: %s", req.GetKey())

	// Read document from collection
	var event model.Event
	meta, err := s.Collection.ReadDocument(ctx, req.GetKey(), &event)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetKey(),
			}).Info("event not found")
			return nil, status.Errorf(codes.NotFound, "Event not found")
		}

		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to read event document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	event.Id = meta.ID.String()
	event.Key = meta.Key
	event.Rev = meta.Rev
	return &geovision.GetEventResponse{Event: &event}, nil
}

func (s *EventService) GetEvents(ctx context.Context, req *geovision.GetEventsRequest) (*geovision.GetEventsResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Getting events")

	// Validate that both start_time and end_time are provided
	if req.GetStartTime() == 0 || req.GetEndTime() == 0 {
		logger.Error("both start_time and end_time are required")
		return nil, status.Errorf(codes.InvalidArgument, "both start time and end time are required")
	}

	// Validate that start_time is before end_time
	if req.GetStartTime() > req.GetEndTime() {
		logger.Error("start_time must be before end_time")
		return nil, status.Errorf(codes.InvalidArgument, "start time must be before end time")
	}

	// Build AQL query to fetch events within the time range
	query := `
		LET docs = (
            FOR doc IN @@collection
                FILTER doc.happened_at >= @start_time && doc.happened_at <= @end_time
                RETURN doc
        )

        LET doc_map = ZIP(docs[*]._id, docs[*]._id)

        LET internal_edges = (
            FOR start_node IN docs
                FOR v, e IN 1..1 OUTBOUND start_node GRAPH @graph
                FILTER HAS(doc_map, v._id)
                RETURN e
        )

        RETURN { events: docs, relations: internal_edges }
	`

	// Execute query
	cursor, err := s.DBClient.DB.Query(ctx, query, map[string]interface{}{
		"start_time":  req.GetStartTime(),
		"end_time":    req.GetEndTime(),
		"@collection": s.Collection.Name(),
		"graph":       s.DBClient.OsintGraph.Name(),
	})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to execute AQL query for getting events")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}
	defer cursor.Close()

	// Read all events from cursor
	var resp geovision.GetEventsResponse
	_, err = cursor.ReadDocument(ctx, &resp)

	if driver.IsNoMoreDocuments(err) {
		return &geovision.GetEventsResponse{}, nil
	} else if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *EventService) GetEventRelatedEntities(ctx context.Context, req *geovision.GetEventRelatedEntitiesRequest) (*geovision.GetEventRelatedEntitiesResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Getting related entities for event with ID: %s", req.GetKey())

	// Validate the event key
	if req.GetKey() == "" {
		logger.Error("event key is required")
		return nil, status.Errorf(codes.InvalidArgument, "event key is required")
	}

	query := `
		FOR v, e IN 1..1 OUTBOUND CONCAT(@collection, "/",@key) GRAPH @graph
			FILTER NOT IS_SAME_COLLECTION(@collection, v)
			LET type = PARSE_IDENTIFIER(v._id).collection
			RETURN DISTINCT {
				type: type,
				entity: v,
				edge: e
			}
	`

	// Execute the query
	binds := map[string]interface{}{
		"key":        req.GetKey(),
		"collection": s.Collection.Name(),
		"graph":      s.DBClient.OsintGraph.Name(),
	}
	logger.Debugf("Running query: %s with binds: %v", query, binds)
	cursor, err := s.DBClient.DB.Query(ctx, query, binds)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to execute AQL graph traversal query")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}
	defer cursor.Close()

	// Read all related entities from cursor
	var entities []*model.RelatedEntity
	var rowReader helpers.DbQueryResult

	for {
		entity, err := rowReader.MapToRelatedEntity(cursor, ctx)

		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			logger.WithError(err).Warn("skipping malformed entity in stream")
			continue
		}

		entities = append(entities, entity)
	}

	return &geovision.GetEventRelatedEntitiesResponse{Entities: entities}, nil
}

func (s *EventService) CreateEvent(ctx context.Context, req *geovision.CreateEventRequest) (*geovision.CreateEventResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Creating event")

	// Create document in collection
	var event model.Event
	ctxWithReturnNew := driver.WithReturnNew(ctx, &event)
	meta, err := s.Collection.CreateDocument(ctxWithReturnNew, req.GetEvent())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to create event document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	event.Id = meta.ID.String()
	event.Key = meta.Key
	event.Rev = meta.Rev
	return &geovision.CreateEventResponse{Event: &event}, nil
}

func (s *EventService) UpdateEvent(ctx context.Context, req *geovision.UpdateEventRequest) (*geovision.UpdateEventResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Updating event with Key: %s", req.GetEvent().GetKey())

	// Update document in collection
	var event model.Event
	ctxWithReturnNew := driver.WithReturnNew(ctx, &event)
	meta, err := s.Collection.UpdateDocument(ctxWithReturnNew, req.GetEvent().GetKey(), req.GetEvent())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetEvent().GetKey(),
			}).Info("event not found for update")
			return nil, status.Errorf(codes.NotFound, "Event not found")
		}

		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetEvent().GetKey(),
		}).Error("failed to update event document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	event.Id = meta.ID.String()
	event.Key = meta.Key
	event.Rev = meta.Rev
	return &geovision.UpdateEventResponse{Event: &event}, nil
}

func (s *EventService) DeleteEvent(ctx context.Context, req *geovision.DeleteEventRequest) (*geovision.DeleteEventResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Deleting event with Key: %s", req.GetKey())

	// Remove document from collection
	_, err := s.Collection.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetKey(),
			}).Info("event not found for deletion")
			return nil, status.Errorf(codes.NotFound, "Event not found")
		}

		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to delete event document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &geovision.DeleteEventResponse{}, nil
}
