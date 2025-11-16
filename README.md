# GeoVision

## Project Overview

GeoVision is a geospatial intelligence service that provides APIs for managing and analyzing geospatial data related to events, persons, organizations, locations, and various other entities. The service is built using gRPC with HTTP/JSON gateway support via Gin framework.

The service includes a complete backend implementation with placeholder methods for all defined APIs. It follows a microservices architecture pattern where each entity type has its own service implementation.

## Folder Structure

- `services/` - Contains the implementation of all gRPC services
- `main.go` - Entry point of the application that initializes all services and starts the gRPC server with HTTP gateway

## Infrastructure Overview

The GeoVision service follows a three-layer architecture:

1. **gRPC Services Layer**: Implements the business logic for each entity type
2. **gRPC Gateway Layer**: Translates HTTP/JSON requests to gRPC calls
3. **Gin HTTP Layer**: Provides the HTTP server and routing

Services are designed to be independent and follow the gRPC service definitions in the `idl/geovision` directory. Each service implements the `Unimplemented<ServiceName>Server` interface to ensure forward compatibility.

## Run Locally

### Buf Gen

```bash
buf registry login buf.build

buf dep update

buf format -w
buf lint

buf generate
buf push

go mod tidy
```

### Run Unit Tests

```bash
docker-compose up -d arangodb
go test -v ./... -run <test name>
docker-compose down
```

You can view arangodb dashboard at http://localhost:8529

### Run the Service

To run the service with ArangoDB using Docker Compose:

```bash
docker-compose up -d

docker-compose logs arangodb
docker inspect
```
