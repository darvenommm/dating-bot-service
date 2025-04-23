package main

import (
	"context"
	"log"
	"net"

	"github.com/darvenommm/dating-bot-service/internal/action"
	"github.com/darvenommm/dating-bot-service/internal/broker"
	"github.com/darvenommm/dating-bot-service/internal/cache"
	"github.com/darvenommm/dating-bot-service/internal/filter"
	"github.com/darvenommm/dating-bot-service/internal/match"
	"github.com/darvenommm/dating-bot-service/internal/orm"
	"github.com/darvenommm/dating-bot-service/internal/profile"
	actionv1 "github.com/darvenommm/dating-bot-service/pkg/api/action/v1"
	filterv1 "github.com/darvenommm/dating-bot-service/pkg/api/filter/v1"
	matchv1 "github.com/darvenommm/dating-bot-service/pkg/api/match/v1"
	profilev1 "github.com/darvenommm/dating-bot-service/pkg/api/profile/v1"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	createdORM, err := orm.NewFromEnvironments()
	if err != nil {
		log.Fatalln(err)
	}

	err = createdORM.Migrate([]orm.Migrator{profile.Migrate, filter.Migrate, action.Migrate})
	if err != nil {
		log.Fatalln(err)
	}

	cache, err := cache.NewCacheFromEnvironment()
	if err != nil {
		log.Fatalln(err)
	}

	err = broker.CreateTopics(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	grpcServer := grpc.NewServer()

	profileServer := profile.NewServer(createdORM, cache)
	profilev1.RegisterProfileServiceServer(grpcServer, profileServer)

	filterServer := filter.NewServer(createdORM)
	filterv1.RegisterFilterServiceServer(grpcServer, filterServer)

	actionServer := action.NewServer(createdORM)
	actionv1.RegisterActionServiceServer(grpcServer, actionServer)

	matchServer := match.NewServer()
	matchv1.RegisterMatchServiceServer(grpcServer, matchServer)

	action.CheckMatchesCron(createdORM)
	action.StartListeningAction(createdORM)

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":10000")
	log.Println("start listening")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
