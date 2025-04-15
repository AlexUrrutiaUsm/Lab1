package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "gobierno/proto/grpc-server/proto"

	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	piratasBuscados []pb.Pirata
	capturas        []string
	mu              sync.Mutex
)

type GobiernoServer struct {
	pb.UnimplementedGobiernoServiceServer
}

func main() {
	// Cargar piratas desde el archivo
	cargarPiratasDesdeArchivo("piratas.txt")

	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterGobiernoServiceServer(server, &GobiernoServer{})
	reflection.Register(server)

	go generarPiratasBuscados()

	log.Println("Servidor del Gobierno Mundial escuchando en el puerto 50052")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

func cargarPiratasDesdeArchivo(ruta string) {
	file, err := os.Open(ruta)
	if err != nil {
		log.Fatalf("Error al abrir el archivo: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		linea := scanner.Text()
		partes := strings.Split(linea, ",")
		if len(partes) == 5 {
			recompensa, _ := strconv.Atoi(strings.TrimSpace(partes[2]))
			pirata := pb.Pirata{
				Id:                strings.TrimSpace(partes[0]),
				Nombre:            strings.TrimSpace(partes[1]),
				Recompensa:        int32(recompensa),
				NivelPeligrosidad: strings.TrimSpace(partes[3]),
				Estado:            strings.TrimSpace(partes[4]),
			}
			piratasBuscados = append(piratasBuscados, pirata)
			log.Printf("Pirata cargado: ID=%s, Nombre=%s, Recompensa=%d, Nivel=%s, Estado=%s",
				pirata.Id, pirata.Nombre, pirata.Recompensa, pirata.NivelPeligrosidad, pirata.Estado)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error al leer el archivo: %v", err)
	}
}

func (s *GobiernoServer) PublicarListaPiratas(ctx context.Context, req *pb.Empty) (*pb.ListaPiratas, error) {
	mu.Lock()
	defer mu.Unlock()

	piratas := make([]*pb.Pirata, len(piratasBuscados))
	for i := range piratasBuscados {
		piratas[i] = &piratasBuscados[i]
	}

	log.Printf("Enviando lista de piratas: %d piratas encontrados", len(piratasBuscados))
	return &pb.ListaPiratas{Piratas: piratas}, nil
}

func (s *GobiernoServer) RegistrarCaptura(ctx context.Context, req *pb.EntregaRequest) (*pb.EntregaResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	for i := range piratasBuscados {
		pirata := &piratasBuscados[i]
		if pirata.Id == req.IdPirata && pirata.Estado == "Buscado" {
			pirata.Estado = "Capturado"
			capturas = append(capturas, fmt.Sprintf("Pirata %s capturado por %s", pirata.Nombre, req.CazarrecompensasId))
			recompensa := pirata.Recompensa
			return &pb.EntregaResponse{
				Exito:      true,
				Mensaje:    "Captura registrada con éxito",
				Recompensa: recompensa,
			}, nil
		}
	}

	return &pb.EntregaResponse{
		Exito:   false,
		Mensaje: "Pirata no encontrado o ya capturado",
	}, nil
}

func generarPiratasBuscados() {
	for {
		time.Sleep(30 * time.Second)
		mu.Lock()
		id := fmt.Sprintf("%03d", rand.Intn(1000))
		nombre := fmt.Sprintf("Pirata%d", rand.Intn(100))
		recompensa := rand.Intn(500000000) + 10000000
		nivel := []string{"Bajo", "Medio", "Alto"}[rand.Intn(3)]
		piratasBuscados = append(piratasBuscados, pb.Pirata{
			Id:                id,
			Nombre:            nombre,
			Recompensa:        int32(recompensa),
			NivelPeligrosidad: nivel,
			Estado:            "Buscado",
		})
		log.Printf("Pirata generado: ID=%s, Nombre=%s, Recompensa=%d, Nivel=%s, Estado=%s",
			id, nombre, recompensa, nivel, "Buscado")
		mu.Unlock()
	}
}
