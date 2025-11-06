package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
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
	return nil, errors.New("method GetSource not implemented")
}

func (s *SourceService) CreateSource(ctx context.Context, req *geovision.CreateSourceRequest) (*geovision.CreateSourceResponse, error) {
	return nil, errors.New("method CreateSource not implemented")
}

func (s *SourceService) UpdateSource(ctx context.Context, req *geovision.UpdateSourceRequest) (*geovision.UpdateSourceResponse, error) {
	return nil, errors.New("method UpdateSource not implemented")
}

func (s *SourceService) DeleteSource(ctx context.Context, req *geovision.DeleteSourceRequest) (*geovision.DeleteSourceResponse, error) {
	return nil, errors.New("method DeleteSource not implemented")
}
