package grpc

import (
	"context"
	"log"
	"strings"

	pb "github.com/krrristina/PR2_sem2/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRPCServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *AuthGRPCServer) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	log.Printf("[Auth gRPC] Verify called, token: %q", req.Token)

	// Токен пустой → Unauthenticated
	if req.Token == "" {
		log.Println("[Auth gRPC] empty token → Unauthenticated")
		return nil, status.Error(codes.Unauthenticated, "invalid token: token is empty")
	}

	// Токен начинается на "invalid" → Unauthenticated
	if strings.HasPrefix(req.Token, "invalid") {
		log.Println("[Auth gRPC] bad token → Unauthenticated")
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Всё ок
	log.Println("[Auth gRPC] token is valid")
	return &pb.VerifyResponse{
		Valid:   true,
		Subject: "user@example.com",
	}, nil
}
