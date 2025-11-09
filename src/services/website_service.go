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

type WebsiteService struct {
	geovision.UnimplementedWebsiteServiceServer

	DBClient *clients.ArangoDBClient
}

func NewWebsiteService(client *clients.ArangoDBClient) (*WebsiteService, error) {
	service := &WebsiteService{
		DBClient: client,
	}

	// Create websites collection
	ctx := context.Background()
	exists, err := client.DB.CollectionExists(ctx, "websites")
	if err != nil {
		return nil, fmt.Errorf("failed to check websites collection existence: %v", err)
	}

	if !exists {
		_, err = client.DB.CreateCollection(ctx, "websites", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create websites collection: %v", err)
		}
		fmt.Printf("✅ Created websites collection\n")
	}

	return service, nil
}

func (s *WebsiteService) GetWebsite(ctx context.Context, req *geovision.GetWebsiteRequest) (*geovision.GetWebsiteResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Getting website with ID: %s", req.GetKey())

	// Get websites collection
	collection, err := s.DBClient.DB.Collection(ctx, "websites")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"id":    req.GetKey(),
		}).Error("failed to get websites collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Read document from collection
	var website model.Website
	meta, err := collection.ReadDocument(ctx, req.GetKey(), &website)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetKey(),
			}).Info("website not found")
			return nil, status.Errorf(codes.NotFound, "Website not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to read website document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	website.Id = meta.ID.String()
	website.Key = meta.Key
	website.Rev = meta.Rev
	return &geovision.GetWebsiteResponse{Website: &website}, nil
}

func (s *WebsiteService) CreateWebsite(ctx context.Context, req *geovision.CreateWebsiteRequest) (*geovision.CreateWebsiteResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Creating website with URL: %s", req.GetWebsite().GetUrl())

	// Get websites collection
	collection, err := s.DBClient.DB.Collection(ctx, "websites")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
		}).Error("failed to get websites collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Create document in collection
	var website model.Website
	ctxWithReturnNew := driver.WithReturnNew(ctx, &website)
	meta, err := collection.CreateDocument(ctxWithReturnNew, req.GetWebsite())
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
		}).Error("failed to create website document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	website.Id = meta.ID.String()
	website.Key = meta.Key
	website.Rev = meta.Rev
	return &geovision.CreateWebsiteResponse{Website: &website}, nil
}

func (s *WebsiteService) UpdateWebsite(ctx context.Context, req *geovision.UpdateWebsiteRequest) (*geovision.UpdateWebsiteResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Updating website with Key: %s", req.GetWebsite().GetKey())

	// Get websites collection
	collection, err := s.DBClient.DB.Collection(ctx, "websites")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetWebsite().GetKey(),
		}).Error("failed to get websites collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Update document in collection
	var website model.Website
	ctxWithReturnNew := driver.WithReturnNew(ctx, &website)
	meta, err := collection.UpdateDocument(ctxWithReturnNew, req.GetWebsite().GetKey(), req.GetWebsite())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetWebsite().GetKey(),
			}).Info("website not found for update")
			return nil, status.Errorf(codes.NotFound, "Website not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetWebsite().GetKey(),
		}).Error("failed to update website document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	website.Id = meta.ID.String()
	website.Key = meta.Key
	website.Rev = meta.Rev
	return &geovision.UpdateWebsiteResponse{Website: &website}, nil
}

func (s *WebsiteService) DeleteWebsite(ctx context.Context, req *geovision.DeleteWebsiteRequest) (*geovision.DeleteWebsiteResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Deleting website with Key: %s", req.GetKey())

	// Get websites collection
	collection, err := s.DBClient.DB.Collection(ctx, "websites")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to get websites collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Remove document from collection
	_, err = collection.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetKey(),
			}).Info("website not found for deletion")
			return nil, status.Errorf(codes.NotFound, "Website not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to delete website document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &geovision.DeleteWebsiteResponse{}, nil
}
