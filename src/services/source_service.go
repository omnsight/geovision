package services

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/geovision/src/logging"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
	"github.com/bouncingmaxt/omniscent-library/gen/go/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SourceService struct {
	geovision.UnimplementedSourceServiceServer

	DBClient   *clients.ArangoDBClient
	Collection driver.Collection
}

func NewSourceService(client *clients.ArangoDBClient) (*SourceService, error) {
	// Create sources collection
	ctx := context.Background()
	collection, err := client.GetCreateCollection(ctx, "sources", driver.CreateVertexCollectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get or create sources collection: %v", err)
	}
	logrus.Infof("✅ Initialized collection %s", collection.Name())

	service := &SourceService{
		DBClient:   client,
		Collection: collection,
	}
	return service, nil
}

func (s *SourceService) GetSource(ctx context.Context, req *geovision.GetSourceRequest) (*geovision.GetSourceResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Getting source with ID: %s", req.GetKey())

	// Read document from collection
	var source model.Source
	meta, err := s.Collection.ReadDocument(ctx, req.GetKey(), &source)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetKey(),
			}).Info("source not found")
			return nil, status.Errorf(codes.NotFound, "Source not found")
		}

		logger.WithFields(logrus.Fields{
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

	// Create document in collection
	var source model.Source
	ctxWithReturnNew := driver.WithReturnNew(ctx, &source)
	meta, err := s.Collection.CreateDocument(ctxWithReturnNew, req.GetSource())
	if err != nil {
		logger.WithFields(logrus.Fields{
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

	// Update document in collection
	var source model.Source
	ctxWithReturnNew := driver.WithReturnNew(ctx, &source)
	meta, err := s.Collection.UpdateDocument(ctxWithReturnNew, req.GetSource().GetKey(), req.GetSource())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetSource().GetKey(),
			}).Info("source not found for update")
			return nil, status.Errorf(codes.NotFound, "Source not found")
		}

		logger.WithFields(logrus.Fields{
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

	// Remove document from collection
	_, err := s.Collection.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetKey(),
			}).Info("source not found for deletion")
			return nil, status.Errorf(codes.NotFound, "Source not found")
		}

		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to delete source document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &geovision.DeleteSourceResponse{}, nil
}
