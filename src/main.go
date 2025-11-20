package main

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/omnsight/geovision/gen/geovision/v1"
	"github.com/omnsight/geovision/src/services"
	"github.com/omnsight/omniscent-library/src/clients"
	"github.com/omnsight/omniscent-library/src/constants"
	"github.com/omnsight/omniscent-library/src/logging"
	"github.com/omnsight/omniscent-library/src/middleware"
)

func main() {
	// ---- 1. Start the gRPC Server (your logic) ----
	// Get gRPC address from environment variable or use default
	grpcPort := os.Getenv(constants.GrpcPort)
	if grpcPort == "" {
		logrus.Fatalf("missing environment variable %s", constants.GrpcPort)
	}

	serverPort := os.Getenv(constants.ServerPort)
	if serverPort == "" {
		logrus.Fatalf("missing environment variable %s", constants.ServerPort)
	}

	clientId := os.Getenv(clients.KeycloakClientID)
	if clientId == "" {
		logrus.Fatalf("missing environment variable %s", clients.KeycloakClientID)
	}

	// Create a gRPC server
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(logging.LoggingInterceptor), grpc.UnaryInterceptor(middleware.GrpcGatewayIdentityInterceptor(clientId)),
	)

	// Create a new ArangoDB client
	client, err := clients.NewArangoDBClient()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to establish ArangoDB client")
	}

	// Register your business logic implementation with the gRPC server
	eventService, err := services.NewGeoService(client)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to create EventService")
	}
	geovision.RegisterGeoServiceServer(gRPCServer, eventService)

	// Enable reflection for debugging
	reflection.Register(gRPCServer)

	// Start the gRPC server in a separate goroutine
	go func() {
		lis, _ := net.Listen("tcp", ":"+grpcPort)
		gRPCServer.Serve(lis)
	}()

	// ---- 2. Start the gRPC-Gateway (the connection) ----
	ctx := context.Background()

	// Create a client connection to the gRPC server
	// The gateway acts as a client - using NewClient instead of deprecated DialContext
	conn, err := grpc.NewClient(
		grpcPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to create gRPC client")
	}
	defer conn.Close()

	// Create the gRPC-Gateway's multiplexer (router)
	// This mux knows how to translate HTTP routes (from proto definitions) to gRPC calls
	gwmux := gwRuntime.NewServeMux()

	// Register all service handlers with the gateway's router
	if err := geovision.RegisterGeoServiceHandler(ctx, gwmux, conn); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to register GeoService handler")
	}

	// ---- 3. Start the Gin Server (the HTTP entrypoint) ----
	// Create a Gin router
	r := gin.Default()

	// Tell Gin to proxy any requests on /v1/* to the gRPC-Gateway
	// THIS IS THE "CONNECTION"
	r.Any("/v1/*any", gin.WrapH(gwmux))

	// Add other Gin routes as needed
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Run the Gin server
	r.Run(":" + serverPort)
}
