package services

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/geovision/src/logging"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
	"github.com/bouncingmaxt/omniscent-library/gen/go/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SourceService struct {
	geovision.UnimplementedSourceServiceServer

	DBClient *clients.ArangoDBClient
}

func NewSourceService(client *clients.ArangoDBClient) (*SourceService, error) {
	service := &SourceService{
		DBClient: client,
	}

	// Create sources collection
	ctx := context.Background()
	exists, err := client.DB.CollectionExists(ctx, "sources")
	if err != nil {
		return nil, fmt.Errorf("failed to check sources collection existence: %v", err)
	}

	if !exists {
		_, err = client.DB.CreateCollection(ctx, "sources", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create sources collection: %v", err)
		}
		fmt.Printf("✅ Created sources collection\n")
	}

	return service, nil
}

func (s *SourceService) GetSource(ctx context.Context, req *geovision.GetSourceRequest) (*geovision.GetSourceResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Getting source with ID: %s", req.GetKey())

	// Get sources collection
	collection, err := s.DBClient.DB.Collection(ctx, "sources")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"id":    req.GetKey(),
		}).Error("failed to get sources collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Read document from collection
	var source model.Source
	meta, err := collection.ReadDocument(ctx, req.GetKey(), &source)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetKey(),
			}).Info("source not found")
			return nil, status.Errorf(codes.NotFound, "Source not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to read source document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	source.Id = meta.ID.String()
	source.Key = meta.Key
	source.Rev = meta.Rev
	return &geovision.GetSourceResponse{Source: &source}, nil
}

func (s *SourceService) CreateSource(ctx context.Context, req *geovision.CreateSourceRequest) (*geovision.CreateSourceResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Creating source with name: %s", req.GetSource().GetName())

	// Get sources collection
	collection, err := s.DBClient.DB.Collection(ctx, "sources")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
		}).Error("failed to get sources collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Create document in collection
	var source model.Source
	ctxWithReturnNew := driver.WithReturnNew(ctx, &source)
	meta, err := collection.CreateDocument(ctxWithReturnNew, req.GetSource())
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
		}).Error("failed to create source document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	source.Id = meta.ID.String()
	source.Key = meta.Key
	source.Rev = meta.Rev
	return &geovision.CreateSourceResponse{Source: &source}, nil
}

func (s *SourceService) UpdateSource(ctx context.Context, req *geovision.UpdateSourceRequest) (*geovision.UpdateSourceResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Updating source with Key: %s", req.GetSource().GetKey())

	// Get sources collection
	collection, err := s.DBClient.DB.Collection(ctx, "sources")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetSource().GetKey(),
		}).Error("failed to get sources collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Update document in collection
	var source model.Source
	ctxWithReturnNew := driver.WithReturnNew(ctx, &source)
	meta, err := collection.UpdateDocument(ctxWithReturnNew, req.GetSource().GetKey(), req.GetSource())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetSource().GetKey(),
			}).Info("source not found for update")
			return nil, status.Errorf(codes.NotFound, "Source not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetSource().GetKey(),
		}).Error("failed to update source document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	source.Id = meta.ID.String()
	source.Key = meta.Key
	source.Rev = meta.Rev
	return &geovision.UpdateSourceResponse{Source: &source}, nil
}

func (s *SourceService) DeleteSource(ctx context.Context, req *geovision.DeleteSourceRequest) (*geovision.DeleteSourceResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Deleting source with Key: %s", req.GetKey())

	// Get sources collection
	collection, err := s.DBClient.DB.Collection(ctx, "sources")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to get sources collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Remove document from collection
	_, err = collection.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetKey(),
			}).Info("source not found for deletion")
			return nil, status.Errorf(codes.NotFound, "Source not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to delete source document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &geovision.DeleteSourceResponse{}, nil
}
