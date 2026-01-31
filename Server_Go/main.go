package main

import (
	"log"
	"net"
	pb "scopone_server/Proto_Files"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Errore: impossibile ascoltare sulla porta 50051: %v", err)
	}

	grpc_Server := grpc.NewServer()
	pb.RegisterGoBackendServer(grpc_Server, &server{})

	log.Println("Server in ascolto sulla porta :50051")

	if err := grpc_Server.Serve(lis); err != nil {
		log.Fatalf("Errore: impossibile avviare il server gRPC: %v", err)
	}

}
