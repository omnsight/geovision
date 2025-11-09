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

	"github.com/bouncingmaxt/geovision/src/clients"
	"github.com/bouncingmaxt/geovision/src/logging"
	"github.com/bouncingmaxt/geovision/src/services"
	"github.com/bouncingmaxt/omniscent-library/gen/go/geovision"
	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func main() {
	// ---- 1. Start the gRPC Server (your logic) ----
	// Get gRPC address from environment variable or use default
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		logrus.Fatal("missing environment variable GRPC_PORT")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		logrus.Fatal("missing environment variable SERVER_PORT")
	}

	// Create a gRPC server
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(logging.LoggingInterceptor),
	)

	// Create a new ArangoDB client
	client, err := clients.NewArangoDBClient()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to establish ArangoDB client")
	}

	// Register your business logic implementation with the gRPC server
	eventService, err := services.NewEventService(client)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to create EventService")
	}
	geovision.RegisterEventServiceServer(gRPCServer, eventService)

	personService, err := services.NewPersonService(client)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to create PersonService")
	}
	geovision.RegisterPersonServiceServer(gRPCServer, personService)

	organizationService, err := services.NewOrganizationService(client)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to create OrganizationService")
	}
	geovision.RegisterOrganizationServiceServer(gRPCServer, organizationService)

	sourceService, err := services.NewSourceService(client)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to create SourceService")
	}
	geovision.RegisterSourceServiceServer(gRPCServer, sourceService)

	websiteService, err := services.NewWebsiteService(client)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to create WebsiteService")
	}
	geovision.RegisterWebsiteServiceServer(gRPCServer, websiteService)

	relationshipService, err := services.NewRelationshipService(client)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to create RelationshipService")
	}
	geovision.RegisterRelationshipServiceServer(gRPCServer, relationshipService)

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
	if err := geovision.RegisterEventServiceHandler(ctx, gwmux, conn); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to register EventService handler")
	}

	if err := geovision.RegisterPersonServiceHandler(ctx, gwmux, conn); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to register PersonService handler")
	}

	if err := geovision.RegisterOrganizationServiceHandler(ctx, gwmux, conn); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to register OrganizationService handler")
	}

	if err := geovision.RegisterSourceServiceHandler(ctx, gwmux, conn); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to register SourceService handler")
	}

	if err := geovision.RegisterWebsiteServiceHandler(ctx, gwmux, conn); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to register WebsiteService handler")
	}

	if err := geovision.RegisterRelationshipServiceHandler(ctx, gwmux, conn); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("failed to register RelationshipService handler")
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
