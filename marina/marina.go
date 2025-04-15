package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	pb "marina/proto/grpc-server/proto"

	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	balanceFinanciero int64 = 2000000
	transacciones     []string
	mu                sync.Mutex
)

type MarinaServer struct {
	pb.UnimplementedMarinaServiceServer
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterMarinaServiceServer(server, &MarinaServer{})
	reflection.Register(server)

	go simularRedadas()

	log.Println("Servidor de la Marina escuchando en el puerto 50051")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

func (s *MarinaServer) RecibirPirata(ctx context.Context, req *pb.EntregaRequest) (*pb.EntregaResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	recompensa := rand.Intn(500000) + 100000
	balanceFinanciero -= int64(recompensa)
	transacciones = append(transacciones, fmt.Sprintf("Pirata %s recibido. Recompensa pagada: %d Berries", req.IdPirata, recompensa))

	return &pb.EntregaResponse{
		Exito:      true,
		Mensaje:    "Pirata recibido con éxito",
		Recompensa: int32(recompensa),
	}, nil
}

func (s *MarinaServer) RealizarRedada(ctx context.Context, req *pb.RedadaRequest) (*pb.RedadaResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	exito := rand.Intn(2) == 0
	if exito {
		transacciones = append(transacciones, fmt.Sprintf("Redada exitosa en la región %s", req.Region))
		return &pb.RedadaResponse{
			Exito:   true,
			Mensaje: "Redada realizada con éxito",
		}, nil
	}

	transacciones = append(transacciones, fmt.Sprintf("Redada fallida en la región %s", req.Region))
	return &pb.RedadaResponse{
		Exito:   false,
		Mensaje: "Redada fallida",
	}, nil
}

func simularRedadas() {
	for {
		time.Sleep(time.Duration(rand.Intn(20)+10) * time.Second)
		region := fmt.Sprintf("Región-%d", rand.Intn(10)+1)
		mu.Lock()
		transacciones = append(transacciones, fmt.Sprintf("Simulación de redada en %s", region))
		mu.Unlock()
		fmt.Printf("Simulación de redada en %s\n", region)
	}
}
