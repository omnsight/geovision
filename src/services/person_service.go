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
	return nil, status.Errorf(codes.Unimplemented, "method GetPersons not implemented")
}

func (s *PersonService) GetPerson(ctx context.Context, req *geovision.GetPersonRequest) (*geovision.GetPersonResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Getting person with ID: %s", req.GetKey())

	// Get persons collection
	collection, err := s.DBClient.DB.Collection(ctx, "persons")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"id":    req.GetKey(),
		}).Error("failed to get persons collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Read document from collection
	var person model.Person
	meta, err := collection.ReadDocument(ctx, req.GetKey(), &person)
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetKey(),
			}).Info("person not found")
			return nil, status.Errorf(codes.NotFound, "Person not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to read person document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	person.Id = meta.ID.String()
	person.Key = meta.Key
	person.Rev = meta.Rev
	return &geovision.GetPersonResponse{Person: &person}, nil
}

func (s *PersonService) CreatePerson(ctx context.Context, req *geovision.CreatePersonRequest) (*geovision.CreatePersonResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Creating person")

	// Get persons collection
	collection, err := s.DBClient.DB.Collection(ctx, "persons")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
		}).Error("failed to get persons collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Create document in collection
	var person model.Person
	ctxWithReturnNew := driver.WithReturnNew(ctx, &person)
	meta, err := collection.CreateDocument(ctxWithReturnNew, req.GetPerson())
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
		}).Error("failed to create person document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	person.Id = meta.ID.String()
	person.Key = meta.Key
	person.Rev = meta.Rev
	return &geovision.CreatePersonResponse{Person: &person}, nil
}

func (s *PersonService) UpdatePerson(ctx context.Context, req *geovision.UpdatePersonRequest) (*geovision.UpdatePersonResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Updating person with Key: %s", req.GetPerson().GetKey())

	// Get persons collection
	collection, err := s.DBClient.DB.Collection(ctx, "persons")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetPerson().GetKey(),
		}).Error("failed to get persons collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Update document in collection
	var person model.Person
	ctxWithReturnNew := driver.WithReturnNew(ctx, &person)
	meta, err := collection.UpdateDocument(ctxWithReturnNew, req.GetPerson().GetKey(), req.GetPerson())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetPerson().GetKey(),
			}).Info("person not found for update")
			return nil, status.Errorf(codes.NotFound, "Person not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetPerson().GetKey(),
		}).Error("failed to update person document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	person.Id = meta.ID.String()
	person.Key = meta.Key
	person.Rev = meta.Rev
	return &geovision.UpdatePersonResponse{Person: &person}, nil
}

func (s *PersonService) DeletePerson(ctx context.Context, req *geovision.DeletePersonRequest) (*geovision.DeletePersonResponse, error) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Deleting person with Key: %s", req.GetKey())

	// Get persons collection
	collection, err := s.DBClient.DB.Collection(ctx, "persons")
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to get persons collection")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	// Remove document from collection
	_, err = collection.RemoveDocument(ctx, req.GetKey())
	if err != nil {
		if driver.IsNotFoundGeneral(err) {
			logger.WithFields(map[string]interface{}{
				"key": req.GetKey(),
			}).Info("person not found for deletion")
			return nil, status.Errorf(codes.NotFound, "Person not found")
		}

		logger.WithFields(map[string]interface{}{
			"error": err,
			"key":   req.GetKey(),
		}).Error("failed to delete person document")
		return nil, status.Errorf(codes.Internal, "Internal service error. Please try again later.")
	}

	return &geovision.DeletePersonResponse{}, nil
}
