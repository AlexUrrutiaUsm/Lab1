package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	pb "caza/proto/grpc-server/proto"

	"google.golang.org/grpc"
)

var (
	reputacion        map[string]int
	balanceFinanciero int64 = 500000
	mu                sync.Mutex
)

func main() {
	log.Println("Esperando 5 segundos para iniciar el Cazarrecompensas...")
	time.Sleep(5 * time.Second)

	// Inicializar el mapa reputacion
	reputacion = make(map[string]int)

	connGobierno, err := grpc.Dial("gobierno:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error al conectar al servidor del Gobierno Mundial: %v", err)
	}
	defer connGobierno.Close()

	connMarina, err := grpc.Dial("marina:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error al conectar al servidor de la Marina: %v", err)
	}
	defer connMarina.Close()

	connSubmundo, err := grpc.Dial("submundo:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error al conectar al servidor del Submundo: %v", err)
	}
	defer connSubmundo.Close()

	gobiernoClient := pb.NewGobiernoServiceClient(connGobierno)
	marinaClient := pb.NewMarinaServiceClient(connMarina)
	submundoClient := pb.NewSubmundoServiceClient(connSubmundo)

	go simularCapturas(gobiernoClient, marinaClient, submundoClient)

	select {}
}

func consultarPiratasBuscados(client pb.GobiernoServiceClient) []*pb.Pirata {
	resp, err := client.PublicarListaPiratas(context.Background(), &pb.Empty{})
	if err != nil {
		log.Printf("Error al consultar la lista de piratas buscados: %v", err)
		return nil
	}

	log.Printf("Lista de piratas recibida: %d piratas encontrados", len(resp.Piratas))
	return resp.Piratas
}

func registrarCaptura(client pb.GobiernoServiceClient, idPirata string) {
	request := &pb.EntregaRequest{
		IdPirata: idPirata,
		Destino:  "Gobierno",
	}

	resp, err := client.RegistrarCaptura(context.Background(), request)
	if err != nil {
		log.Printf("Error al registrar la captura: %v", err)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if resp.Exito {
		balanceFinanciero += int64(resp.Recompensa)
		reputacion["Gobierno"]++
		fmt.Printf("Captura registrada con éxito: %s\n", resp.Mensaje)
	} else {
		fmt.Printf("Error al registrar la captura: %s\n", resp.Mensaje)
	}
}

func decidirEntrega(marinaClient pb.MarinaServiceClient, submundoClient pb.SubmundoServiceClient, idPirata string) {
	destino := "Marina"
	if rand.Intn(2) == 0 {
		destino = "Submundo"
	}

	if destino == "Marina" {
		request := &pb.EntregaRequest{
			IdPirata: idPirata,
			Destino:  destino,
		}

		resp, err := marinaClient.RecibirPirata(context.Background(), request)
		if err != nil {
			log.Printf("Error al entregar el pirata a la Marina: %v", err)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if resp.Exito {
			balanceFinanciero += int64(resp.Recompensa)
			reputacion["Marina"]++
			fmt.Printf("Pirata entregado con éxito a la Marina. Recompensa: %d Berries\n", resp.Recompensa)
		} else {
			fmt.Printf("Error al entregar el pirata a la Marina: %s\n", resp.Mensaje)
		}
	} else {
		request := &pb.EntregaRequest{
			IdPirata: idPirata,
			Destino:  destino,
		}

		resp, err := submundoClient.ComprarPirata(context.Background(), request)
		if err != nil {
			log.Printf("Error al vender el pirata al Submundo: %v", err)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if resp.Exito {
			balanceFinanciero += int64(resp.Recompensa)
			reputacion["Submundo"]++
			fmt.Printf("Pirata vendido con éxito al Submundo. Recompensa: %d Berries\n", resp.Recompensa)
		} else {
			fmt.Printf("Error al vender el pirata al Submundo: %s\n", resp.Mensaje)
		}
	}
}

func simularCapturas(gobiernoClient pb.GobiernoServiceClient, marinaClient pb.MarinaServiceClient, submundoClient pb.SubmundoServiceClient) {
	for {
		time.Sleep(time.Duration(rand.Intn(10)+5) * time.Second)

		piratas := consultarPiratasBuscados(gobiernoClient)
		if len(piratas) == 0 {
			continue
		}

		pirata := piratas[rand.Intn(len(piratas))]
		fmt.Printf("Intentando capturar al pirata %s\n", pirata.Nombre)
		registrarCaptura(gobiernoClient, pirata.Id)
		decidirEntrega(marinaClient, submundoClient, pirata.Id)
	}
}
