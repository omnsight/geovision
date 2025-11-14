package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/bouncingmaxt/omniscent-library/gen/go/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
)

type DbQueryResult struct {
	Type   string          `json:"type"`
	Entity json.RawMessage `json:"entity"` // Keeps the JSON bytes raw
	Edge   json.RawMessage `json:"edge"`
}

func (qr *DbQueryResult) MapToRelatedEntity(cursor driver.Cursor, ctx context.Context) (*model.RelatedEntity, error) {
	// 1. Read the document directly into the receiver (qr)
	_, err := cursor.ReadDocument(ctx, qr)
	if err != nil {
		// Return the error (including IsNoMoreDocuments) directly to the caller
		return nil, err
	}
	logrus.Debugf("query result: %+v", qr)

	// 2. Initialize the Protobuf response with the Relation (Edge)
	// Use DiscardUnknown to prevent errors if DB has fields not in Proto
	unmarshalOpts := protojson.UnmarshalOptions{DiscardUnknown: true}

	relation := &model.Relation{}
	if err := unmarshalOpts.Unmarshal(qr.Edge, relation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relation: %w", err)
	}

	// Create the result struct
	result := &model.RelatedEntity{
		Relation: relation,
	}

	// 3. Switch on the 'Type' field from ArangoDB to hydrate the specific OneOf entity
	switch qr.Type {
	case "persons":
		p := &model.Person{}
		if err := unmarshalOpts.Unmarshal(qr.Entity, p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal person: %w", err)
		}
		result.Entity = &model.RelatedEntity_Person{Person: p}

	case "organizations":
		o := &model.Organization{}
		if err := unmarshalOpts.Unmarshal(qr.Entity, o); err != nil {
			return nil, fmt.Errorf("failed to unmarshal organization: %w", err)
		}
		result.Entity = &model.RelatedEntity_Organization{Organization: o}

	case "sources":
		s := &model.Source{}
		if err := unmarshalOpts.Unmarshal(qr.Entity, s); err != nil {
			return nil, fmt.Errorf("failed to unmarshal source: %w", err)
		}
		result.Entity = &model.RelatedEntity_Source{Source: s}

	case "websites":
		w := &model.Website{}
		if err := unmarshalOpts.Unmarshal(qr.Entity, w); err != nil {
			return nil, fmt.Errorf("failed to unmarshal website: %w", err)
		}
		result.Entity = &model.RelatedEntity_Website{Website: w}

	default:
		// Log a warning or return an error if strict typing is required
		return nil, fmt.Errorf("encountered unknown entity collection type: %s", qr.Type)
	}

	return result, nil
}
