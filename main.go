package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	pb "github.com/iogurth/agua-server/server" // Importa el paquete generado por protoc
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMiServicioServer
	heartbeatStarted bool
	mu               sync.Mutex
}

func CoordenGen(n int) string {
	coordenadas := ""
	for i := 0; i < n; i++ {
		x := rand.Float64()
		y := rand.Float64()
		coordenadas += fmt.Sprintf("%f;%f\n", x, y)
	}
	return coordenadas
}

func (s *server) Inicializador(ctx context.Context, in *pb.InicializadorRequest) (*pb.InicializadorResponse, error) {
	coordenadas := CoordenGen(int(in.GetInicializador()))
	// Escribe las coordenadas en un archivo .txt
	f, err := os.Create("coordenadas.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(coordenadas)
	if err != nil {
		log.Fatal(err)
	}
	s.mu.Lock()
	s.heartbeatStarted = true
	s.mu.Unlock()
	return &pb.InicializadorResponse{Respuesta: "Inicializador recibido y archivo de coordenadas generado"}, nil
}

func (s *server) EnviarCoordenadas(ctx context.Context, in *pb.CoordenadasRequest) (*pb.CoordenadasResponse, error) {
	// Lee el archivo .txt y lo envía como una cadena de texto
	f, err := os.Open("coordenadas.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	return &pb.CoordenadasResponse{Respuesta: string(b)}, nil
}

func (s *server) Heartbeat(in *pb.HeartbeatRequest, stream pb.MiServicio_HeartbeatServer) error {
	for {
		s.mu.Lock()
		started := s.heartbeatStarted
		s.mu.Unlock()
		if started {
			if err := stream.Send(&pb.HeartbeatResponse{Estado: "Heartbeat"}); err != nil {
				return err
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	srv := &server{}
	pb.RegisterMiServicioServer(s, srv)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
