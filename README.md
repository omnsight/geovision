# GeoVision

GeoVision is a geospatial intelligence service that provides APIs for managing and analyzing geospatial data related to events, persons, organizations, locations, and various other entities. The service is built using gRPC with HTTP/JSON gateway support via Gin framework.

The service includes a complete backend implementation with placeholder methods for all defined APIs. It follows a microservices architecture pattern where each entity type has its own service implementation.

## Local Development

Tag is injested by a github action. Commit message including `#major`, `#minor`, `#patch`, or `#none` will bump the release and pre-release versions.

### Dependencies

To upgrade internal dependencies:

```bash
go clean -cache -modcache
GOPROXY=direct go get github.com/omnsight/omniscent-library@<branch>
GOPROXY=direct go get github.com/omnsight/omnibasement@<branch>
```

Buf build:

```bash
buf registry login buf.build

buf dep update

buf format -w
buf lint

buf generate
buf push

go mod tidy
```

### Testing

Run unit tests. You can view arangodb dashboard at http://localhost:8529.

```bash
docker-compose up -d arangodb
go test -v ./... -run <test name>
docker-compose down
```
