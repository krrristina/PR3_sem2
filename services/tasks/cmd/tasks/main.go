package main

import (
	"log"
	"net/http"
	"os"

	pb "github.com/krrristina/PR2_sem2/proto"
	"github.com/krrristina/PR2_sem2/services/tasks/internal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	authAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authAddr == "" {
		authAddr = "localhost:50051"
	}

	conn, err := grpc.Dial(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[Tasks] could not connect to auth: %v", err)
	}
	defer conn.Close()

	// Создаём Handler и передаём ему gRPC-клиент
	h := &internal.Handler{
		AuthClient: pb.NewAuthServiceClient(conn),
	}

	// Регистрируем маршрут ← этой строки не было!
	http.HandleFunc("/tasks", h.GetTasks)

	tasksPort := os.Getenv("TASKS_PORT")
	if tasksPort == "" {
		tasksPort = "8082"
	}

	log.Printf("[Tasks] HTTP server listening on :%s", tasksPort)
	log.Fatal(http.ListenAndServe(":"+tasksPort, nil))
}
