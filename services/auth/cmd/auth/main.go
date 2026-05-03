package main

import (
	"log"
	"net"
	"os"

	pb "github.com/krrristina/PR2_sem2/proto"
	grpcserver "github.com/krrristina/PR2_sem2/services/auth/internal/grpc"
	"google.golang.org/grpc"
)

func main() {
	port := os.Getenv("AUTH_GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &grpcserver.AuthGRPCServer{})

	log.Printf("gRPC Auth server listening on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
