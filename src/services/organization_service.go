package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
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
	return nil, errors.New("method GetOrganization not implemented")
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, req *geovision.CreateOrganizationRequest) (*geovision.CreateOrganizationResponse, error) {
	return nil, errors.New("method CreateOrganization not implemented")
}

func (s *OrganizationService) UpdateOrganization(ctx context.Context, req *geovision.UpdateOrganizationRequest) (*geovision.UpdateOrganizationResponse, error) {
	return nil, errors.New("method UpdateOrganization not implemented")
}

func (s *OrganizationService) DeleteOrganization(ctx context.Context, req *geovision.DeleteOrganizationRequest) (*geovision.DeleteOrganizationResponse, error) {
	return nil, errors.New("method DeleteOrganization not implemented")
}
