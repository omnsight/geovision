package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/arangodb/go-driver"
	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/geovision/src/logging"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
	"github.com/bouncingmaxt/omniscent-library/gen/go/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RelationshipService struct {
	geovision.UnimplementedRelationshipServiceServer

	DBClient *clients.ArangoDBClient
}

func NewRelationshipService(client *clients.ArangoDBClient) (*RelationshipService, error) {
	service := &RelationshipService{
		DBClient: client,
	}

	// Create relationships collection as an edge collection
	ctx := context.Background()
	exists, err := client.DB.CollectionExists(ctx, "relationships")
	if err != nil {
		return nil, fmt.Errorf("failed to check relationships collection existence: %v", err)
	}

	if !exists {
		_, err = client.DB.CreateCollection(ctx, "relationships", &driver.CreateCollectionOptions{
			Type: driver.CollectionTypeEdge,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create relationships collection: %v", err)
		}
		fmt.Printf("✅ Created relationships collection as edge collection\n")
	}

	return service, nil
}

func (s *RelationshipService) CreateRelationship(ctx context.Context, req *geovision.CreateRelationshipRequest) (*geovision.CreateRelationshipResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Creating relationship")

	relationship := req.GetRelationship()
	if relationship == nil {
		logger.Error("relationship is nil")
		return nil, status.Errorf(codes.InvalidArgument, "Bad parameter")
	}

	fromColl, _, err := s.DBClient.ParseDocID(relationship.From)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"id":    relationship.From,
		}).Error("failed to parse from entitity id")
		return nil, status.Errorf(codes.InvalidArgument, "Bad parameter")
	}

	toColl, _, err := s.DBClient.ParseDocID(relationship.To)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"id":    relationship.From,
		}).Error("failed to parse to entitity id")
		return nil, status.Errorf(codes.InvalidArgument, "Bad parameter")
	}

	// Process relation name
	relationName := strings.ToLower(strings.ReplaceAll(relationship.Name, " ", "_"))
	if len(relationName) == 0 {
		logger.Error("invalid relation name")
		return nil, status.Errorf(codes.InvalidArgument, "invalid relation name")
	}

	collectionName := fmt.Sprintf("%s_%s_%s", fromColl, relationName, toColl)

	// Create the edge collection if it doesn't exist
	exists, err := s.DBClient.DB.CollectionExists(ctx, collectionName)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"name":  collectionName,
		}).Error("failed to check collection existence")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	if !exists {
		_, err = s.DBClient.DB.CreateCollection(ctx, collectionName, &driver.CreateCollectionOptions{
			Type: driver.CollectionTypeEdge,
		})
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err,
				"name":  collectionName,
			}).Error("failed to create edge collection")
			return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
		}
		logger.Infof("✅ Created edge collection: %s", collectionName)
	} else {
		logger.Infof("Edge collection already exists: %s", collectionName)
	}

	// Get the specific edge collection
	collection, err := s.DBClient.DB.Collection(ctx, collectionName)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"name":  collectionName,
		}).Error("failed to get edge collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Create document in collection
	relationship.Id = ""
	relationship.Key = ""
	relationship.Rev = ""

	var createdRelationship model.Relation
	ctxWithReturnNew := driver.WithReturnNew(ctx, &createdRelationship)
	_, err = collection.CreateDocument(ctxWithReturnNew, relationship)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"data":  relationship,
		}).Error("failed to create relationship document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &geovision.CreateRelationshipResponse{Relationship: &createdRelationship}, nil
}

func (s *RelationshipService) UpdateRelationship(ctx context.Context, req *geovision.UpdateRelationshipRequest) (*geovision.UpdateRelationshipResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Updating relationship with ID: %s", req.GetId())

	coll, key, err := s.DBClient.ParseDocID(req.GetId())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"id":    req.GetId(),
		}).Error("failed to parse relation id")
		return nil, status.Errorf(codes.InvalidArgument, "Invalid parameter")
	}

	// Using AQL query to update the document with arangodb ID
	query := `
		LET cleanPatch = UNSET(@patch, "_id", "_key", "_rev")
		UPDATE @key WITH cleanPatch IN @@collection
		RETURN NEW
	`

	cursor, err := s.DBClient.DB.Query(ctx, query, map[string]interface{}{
		"key":         key,
		"patch":       req.GetRelationship(),
		"@collection": coll,
	})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"data":  req.GetRelationship(),
		}).Error("failed to execute AQL query for updating relationship")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}
	defer cursor.Close()

	var relationship model.Relation
	meta, err := cursor.ReadDocument(ctx, &relationship)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			logger.WithFields(logrus.Fields{
				"id": req.GetId(),
			}).Info("relationship not found for update")
			return nil, status.Errorf(codes.NotFound, "Relation not found")
		}

		logger.WithFields(logrus.Fields{
			"error": err,
			"id":    req.GetId(),
		}).Error("failed to read updated relationship document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	relationship.Id = meta.ID.String()
	relationship.Key = meta.Key
	relationship.Rev = meta.Rev
	return &geovision.UpdateRelationshipResponse{Relationship: &relationship}, nil
}

func (s *RelationshipService) DeleteRelationship(ctx context.Context, req *geovision.DeleteRelationshipRequest) (*geovision.DeleteRelationshipResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Deleting relationship with ID: %s", req.GetId())

	coll, key, err := s.DBClient.ParseDocID(req.GetId())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"id":    req.GetId(),
		}).Error("failed to parse relation id")
		return nil, status.Errorf(codes.InvalidArgument, "Invalid parameter")
	}

	// Using AQL query to delete the document with arangodb ID
	query := `
		FOR doc IN @@collection
			FILTER doc._key == @key
			REMOVE doc IN @@collection
			RETURN OLD
	`

	cursor, err := s.DBClient.DB.Query(ctx, query, map[string]interface{}{
		"key":         key,
		"@collection": coll,
	})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"id":    req.GetId(),
		}).Error("failed to execute AQL query for deleting relationship")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}
	defer cursor.Close()

	var relationship model.Relation
	_, err = cursor.ReadDocument(ctx, &relationship)
	if err != nil {
		if driver.IsNoMoreDocuments(err) {
			logger.WithFields(logrus.Fields{
				"id": req.GetId(),
			}).Info("relationship not found for deletion")
			return nil, status.Errorf(codes.NotFound, "Relation not found")
		}

		logger.WithFields(logrus.Fields{
			"error": err,
			"id":    req.GetId(),
		}).Error("failed to read deleted relationship document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &geovision.DeleteRelationshipResponse{}, nil
}
