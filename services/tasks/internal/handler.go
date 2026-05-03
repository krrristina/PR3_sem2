package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	pb "github.com/krrristina/PR2_sem2/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler содержит gRPC-клиент и HTTP-обработчики
type Handler struct {
	AuthClient pb.AuthServiceClient
}

// extractToken достаёт токен из заголовка Authorization: Bearer <token>
func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// verifyToken вызывает Auth через gRPC с дедлайном 2 секунды
func (h *Handler) verifyToken(token string) (string, error) {
	// Устанавливаем дедлайн — если Auth не ответит за 2 сек, получим ошибку
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	log.Println("[Tasks] calling grpc verify")

	resp, err := h.AuthClient.Verify(ctx, &pb.VerifyRequest{Token: token})
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.Unauthenticated:
			// Токен невалидный → 401
			log.Printf("[Tasks] grpc verify: unauthenticated: %v", st.Message())
			return "", fmt.Errorf("unauthorized")
		case codes.DeadlineExceeded:
			// Auth не ответил вовремя → 503
			log.Println("[Tasks] grpc verify: deadline exceeded → auth unavailable")
			return "", fmt.Errorf("auth unavailable")
		default:
			// Любая другая ошибка (недоступен, Internal и т.д.) → 503
			log.Printf("[Tasks] grpc verify: error %v → auth unavailable", st.Code())
			return "", fmt.Errorf("auth unavailable")
		}
	}

	if !resp.Valid {
		return "", fmt.Errorf("unauthorized")
	}

	return resp.Subject, nil
}

// GetTasks — HTTP-обработчик GET /tasks
func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized) // 401
		return
	}

	subject, err := h.verifyToken(token)
	if err != nil {
		if err.Error() == "unauthorized" {
			http.Error(w, "unauthorized", http.StatusUnauthorized) // 401
		} else {
			http.Error(w, "auth service unavailable", http.StatusServiceUnavailable) // 503
		}
		return
	}

	log.Printf("[Tasks] request authorized for subject: %s", subject)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"tasks": ["task1", "task2"], "user": "%s"}`, subject)
}
