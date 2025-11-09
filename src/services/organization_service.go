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

type OrganizationService struct {
	geovision.UnimplementedOrganizationServiceServer

	DBClient *clients.ArangoDBClient
}

func NewOrganizationService(client *clients.ArangoDBClient) (*OrganizationService, error) {
	service := &OrganizationService{
		DBClient: client,
	}

	// Create organizations collection
	ctx := context.Background()
	exists, err := client.DB.CollectionExists(ctx, "organizations")
	if err != nil {
		return nil, fmt.Errorf("failed to check organizations collection existence: %v", err)
	}

	if !exists {
		_, err = client.DB.CreateCollection(ctx, "organizations", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create organizations collection: %v", err)
		}
		fmt.Printf("✅ Created organizations collection\n")
	}

	return service, nil
}

func (s *OrganizationService) GetOrganization(ctx context.Context, req *geovision.GetOrganizationRequest) (*geovision.GetOrganizationResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Getting organization with ID: %s", req.GetKey())

	// Get organizations collection
	collection, err := s.DBClient.DB.Collection(ctx, "organizations")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"id":    req.GetKey(),
		}).Error("failed to get organizations collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Read document from collection
	var organization model.Organization
	meta, err := collection.ReadDocument(ctx, req.GetKey(), &organization)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetKey(),
			}).Info("organization not found")
			return nil, status.Errorf(codes.NotFound, "Organization not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to read organization document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	organization.Id = meta.ID.String()
	organization.Key = meta.Key
	organization.Rev = meta.Rev
	return &geovision.GetOrganizationResponse{Organization: &organization}, nil
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, req *geovision.CreateOrganizationRequest) (*geovision.CreateOrganizationResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Creating organization")

	// Get organizations collection
	collection, err := s.DBClient.DB.Collection(ctx, "organizations")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
		}).Error("failed to get organizations collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Create document in collection
	var organization model.Organization
	ctxWithReturnNew := driver.WithReturnNew(ctx, &organization)
	meta, err := collection.CreateDocument(ctxWithReturnNew, req.GetOrganization())
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
		}).Error("failed to create organization document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	organization.Id = meta.ID.String()
	organization.Key = meta.Key
	organization.Rev = meta.Rev
	return &geovision.CreateOrganizationResponse{Organization: &organization}, nil
}

func (s *OrganizationService) UpdateOrganization(ctx context.Context, req *geovision.UpdateOrganizationRequest) (*geovision.UpdateOrganizationResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Updating organization with Key: %s", req.GetOrganization().GetKey())

	// Get organizations collection
	collection, err := s.DBClient.DB.Collection(ctx, "organizations")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetOrganization().GetKey(),
		}).Error("failed to get organizations collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Update document in collection
	var organization model.Organization
	ctxWithReturnNew := driver.WithReturnNew(ctx, &organization)
	meta, err := collection.UpdateDocument(ctxWithReturnNew, req.GetOrganization().GetKey(), req.GetOrganization())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetOrganization().GetKey(),
			}).Info("organization not found for update")
			return nil, status.Errorf(codes.NotFound, "Organization not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetOrganization().GetKey(),
		}).Error("failed to update organization document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	organization.Id = meta.ID.String()
	organization.Key = meta.Key
	organization.Rev = meta.Rev
	return &geovision.UpdateOrganizationResponse{Organization: &organization}, nil
}

func (s *OrganizationService) DeleteOrganization(ctx context.Context, req *geovision.DeleteOrganizationRequest) (*geovision.DeleteOrganizationResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Deleting organization with Key: %s", req.GetKey())

	// Get organizations collection
	collection, err := s.DBClient.DB.Collection(ctx, "organizations")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to get organizations collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Remove document from collection
	_, err = collection.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetKey(),
			}).Info("organization not found for deletion")
			return nil, status.Errorf(codes.NotFound, "Organization not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to delete organization document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &geovision.DeleteOrganizationResponse{}, nil
}
