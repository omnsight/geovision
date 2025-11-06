package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
)

type PersonService struct {
	geovision.UnimplementedPersonServiceServer

	DBClient *clients.ArangoDBClient
}

func NewPersonService(client *clients.ArangoDBClient) (*PersonService, error) {
	service := &PersonService{
		DBClient: client,
	}

	// Create persons collection
	ctx := context.Background()
	exists, err := client.DB.CollectionExists(ctx, "persons")
	if err != nil {
		return nil, fmt.Errorf("failed to check persons collection existence: %v", err)
	}

	if !exists {
		_, err = client.DB.CreateCollection(ctx, "persons", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create persons collection: %v", err)
		}
		fmt.Printf("✅ Created persons collection\n")
	}

	return service, nil
}

func (s *PersonService) GetPersons(ctx context.Context, req *geovision.GetPersonRequest) (*geovision.GetPersonResponse, error) {
	return nil, errors.New("method GetPersons not implemented")
}

func (s *PersonService) GetPerson(ctx context.Context, req *geovision.GetPersonRequest) (*geovision.GetPersonResponse, error) {
	return nil, errors.New("method GetPerson not implemented")
}

func (s *PersonService) CreatePerson(ctx context.Context, req *geovision.CreatePersonRequest) (*geovision.CreatePersonResponse, error) {
	return nil, errors.New("method CreatePerson not implemented")
}

func (s *PersonService) UpdatePerson(ctx context.Context, req *geovision.UpdatePersonRequest) (*geovision.UpdatePersonResponse, error) {
	return nil, errors.New("method UpdatePerson not implemented")
}

func (s *PersonService) DeletePerson(ctx context.Context, req *geovision.DeletePersonRequest) (*geovision.DeletePersonResponse, error) {
	return nil, errors.New("method DeletePerson not implemented")
}
