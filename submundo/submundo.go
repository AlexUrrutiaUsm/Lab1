package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	pb "submundo/proto/grpc-server/proto"

	"google.golang.org/grpc"
)

var (
	balanceFinanciero int64 = 1000000
	transacciones     []string
	mu                sync.Mutex
)

func main() {
	log.Println("Esperando 5 segundos para iniciar el Submundo...")
	time.Sleep(5 * time.Second)

	listener, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Error al iniciar el servidor del Submundo: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterSubmundoServiceServer(server, &SubmundoServer{})
	log.Println("Servidor del Submundo escuchando en el puerto 50053")

	go simularEventos()

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Error al iniciar el servidor del Submundo: %v", err)
	}
}

type SubmundoServer struct {
	pb.UnimplementedSubmundoServiceServer
}

func (s *SubmundoServer) ComprarPirata(ctx context.Context, req *pb.EntregaRequest) (*pb.EntregaResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	// Determinar si ocurre un fraude
	if rand.Intn(10) < 3 { // 30% de probabilidad de fraude
		transacciones = append(transacciones, fmt.Sprintf("Fraude detectado. Pirata %s retenido sin pago", req.IdPirata))
		log.Printf("¡Fraude! Pirata %s retenido sin pagar al cazarrecompensas", req.IdPirata)
		return &pb.EntregaResponse{
			Exito:   false,
			Mensaje: "Fraude detectado. El Submundo se quedó con el pirata sin pagar.",
		}, nil
	}

	// Compra exitosa
	recompensa := rand.Intn(500000) + 100000
	if balanceFinanciero < int64(recompensa) {
		return &pb.EntregaResponse{
			Exito:   false,
			Mensaje: "Fondos insuficientes para comprar al pirata",
		}, nil
	}

	balanceFinanciero -= int64(recompensa)
	transacciones = append(transacciones, fmt.Sprintf("Pirata %s comprado por %d Berries", req.IdPirata, recompensa))
	log.Printf("Pirata comprado: ID=%s, Recompensa=%d", req.IdPirata, recompensa)

	return &pb.EntregaResponse{
		Exito:      true,
		Mensaje:    "Pirata comprado con éxito",
		Recompensa: int32(recompensa),
	}, nil
}

func simularEventos() {
	for {
		time.Sleep(time.Duration(rand.Intn(15)+10) * time.Second)
		enviarMercenarios()
	}
}

func enviarMercenarios() {
	mu.Lock()
	defer mu.Unlock()

	// Simular el envío de mercenarios para robar piratas
	roboExitoso := rand.Intn(10) < 5 // 50% de probabilidad de éxito
	if roboExitoso {
		log.Println("¡Mercenarios enviados! Robo de piratas exitoso.")
		transacciones = append(transacciones, "Mercenarios enviados. Robo de piratas exitoso.")
	} else {
		log.Println("¡Mercenarios enviados! Robo de piratas fallido.")
		transacciones = append(transacciones, "Mercenarios enviados. Robo de piratas fallido.")
	}
}
