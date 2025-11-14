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

type OrganizationService struct {
	geovision.UnimplementedOrganizationServiceServer

	DBClient   *clients.ArangoDBClient
	Collection driver.Collection
}

func NewOrganizationService(client *clients.ArangoDBClient) (*OrganizationService, error) {
	// Create organizations collection
	ctx := context.Background()
	collection, err := client.GetCreateCollection(ctx, "organizations", driver.CreateVertexCollectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get or create organizations collection: %v", err)
	}
	logrus.Infof("✅ Initialized collection %s", collection.Name())

	service := &OrganizationService{
		DBClient:   client,
		Collection: collection,
	}
	return service, nil
}

func (s *OrganizationService) GetOrganization(ctx context.Context, req *geovision.GetOrganizationRequest) (*geovision.GetOrganizationResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Getting organization with ID: %s", req.GetKey())

	// Read document from collection
	var organization model.Organization
	meta, err := s.Collection.ReadDocument(ctx, req.GetKey(), &organization)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetKey(),
			}).Info("organization not found")
			return nil, status.Errorf(codes.NotFound, "Organization not found")
		}

		logger.WithFields(logrus.Fields{
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

	// Create document in collection
	var organization model.Organization
	ctxWithReturnNew := driver.WithReturnNew(ctx, &organization)
	meta, err := s.Collection.CreateDocument(ctxWithReturnNew, req.GetOrganization())
	if err != nil {
		logger.WithFields(logrus.Fields{
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

	// Update document in collection
	var organization model.Organization
	ctxWithReturnNew := driver.WithReturnNew(ctx, &organization)
	meta, err := s.Collection.UpdateDocument(ctxWithReturnNew, req.GetOrganization().GetKey(), req.GetOrganization())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetOrganization().GetKey(),
			}).Info("organization not found for update")
			return nil, status.Errorf(codes.NotFound, "Organization not found")
		}

		logger.WithFields(logrus.Fields{
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

	// Remove document from collection
	_, err := s.Collection.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(logrus.Fields{
				"key": req.GetKey(),
			}).Info("organization not found for deletion")
			return nil, status.Errorf(codes.NotFound, "Organization not found")
		}

		logger.WithFields(logrus.Fields{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to delete organization document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &geovision.DeleteOrganizationResponse{}, nil
}
