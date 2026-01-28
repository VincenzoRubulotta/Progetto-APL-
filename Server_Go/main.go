package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	pb "scopone_server/Proto_Files"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedGoBackendServer
}

type GameSession struct {
	mu        sync.Mutex
	state     *pb.TurnUpdate
	dealer_ID int32
	broadcast chan *pb.TurnUpdate
	deck      []*pb.Card

	hands map[pb.Actor][]*pb.Card
}

var gameManager = make(map[int]*GameSession)
var managerMU sync.Mutex

func (g *GameSession) runCPUPlayers(player pb.Actor) {}

func (s *server) StartGame(ctx context.Context, req *pb.GameSettings) (*pb.InitialState, error) {

	newGame := &GameSession{
		broadcast: make(chan *pb.TurnUpdate, 100),
		hands:     make(map[pb.Actor][]*pb.Card),

		deck: []*pb.Card{},
	}

	managerMU.Lock()

	maxIndex := -1
	for i := range gameManager {
		if i > maxIndex {
			maxIndex = i
		}
	}

	gameID := maxIndex + 1
	gameManager[gameID] = newGame

	managerMU.Unlock()

	dealerID := rand.Int31n(4)

	for i := 1; i <= 4; i++ {
		for k := 1; k <= 10; k++ {
			card := &pb.Card{Suit: pb.Suit(i), Rank: pb.Rank(k)}
			newGame.deck = append(newGame.deck, card)
		}
	}

	rand.Shuffle(len(newGame.deck), func(i, j int) {
		newGame.deck[i], newGame.deck[j] = newGame.deck[j], newGame.deck[i]
	})

	tempDealer := dealerID
	n := 0
	for i := 1; i <= 4; i++ {
		tempDealer = (tempDealer + 1) % 4
		newGame.hands[pb.Actor(tempDealer)] = newGame.deck[n : n+10]
		n = n + 10
	}

	go newGame.runCPUPlayers(pb.Actor_CPU_RIGHT)
	go newGame.runCPUPlayers(pb.Actor_CPU_PARTNER)
	go newGame.runCPUPlayers(pb.Actor_CPU_LEFT)
	return &pb.InitialState{
		Game_ID:       int32(gameID),
		Dealer_ID:     pb.Actor(dealerID),
		CurrentPlayer: pb.Actor((dealerID + 1) % 4),
		UserHand:      newGame.hands[pb.Actor_USER],
	}, nil
}

func (s *server) PlayCard(req *pb.Card, stream pb.GoBackend_PlayCardServer) error {
	return nil
}

func (s *server) ObserveTurn(req *pb.ObserveRequest, stream pb.GoBackend_ObserveTurnServer) error {
	return nil
}

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
