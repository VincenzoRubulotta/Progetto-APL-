package main

import (
	"context"
	std "fmt"
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
	mu            sync.Mutex
	state         *pb.TurnUpdate
	dealer_ID     int32
	listeners     []chan *pb.TurnUpdate
	deck          []*pb.Card
	hands         map[pb.Actor][]*pb.Card
	scoreDeck     map[int][]*pb.Card
	history       []*pb.Card
	tableTop      []*pb.Card
	victoryPoints int32
	hasStarted    bool
	isGameOver    bool
}

var gameManager = make(map[int]*GameSession)
var managerMU sync.Mutex

func TrovaCombinazioni(cards []*pb.Card, target int32) [][]*pb.Card {
	var risultati [][]*pb.Card
	var backtrack func(start int, currentSum int32, path []*pb.Card)

	backtrack = func(start int, currentSum int32, path []*pb.Card) {
		if currentSum == target {
			combo := make([]*pb.Card, len(path))
			copy(combo, path)
			risultati = append(risultati, combo)
			return
		}
		if currentSum > target {
			return
		}
		for i := start; i < len(cards); i++ {
			backtrack(i+1, currentSum+int32(cards[i].Rank), append(path, cards[i]))
		}
	}
	backtrack(0, 0, []*pb.Card{})

	return risultati
}
func (g *GameSession) tableManager(req *pb.Card) *pb.TurnUpdate {
	g.mu.Lock()
	defer g.mu.Unlock()

	for i, card := range g.hands[g.state.NextPlayer_ID] {
		if card.Game_ID == req.Game_ID && card.Suit == req.Suit && card.Rank == req.Rank {
			g.hands[g.state.NextPlayer_ID] = append(g.hands[g.state.NextPlayer_ID][:i], g.hands[g.state.NextPlayer_ID][i+1:]...) // eliminazione carta dalla mano utente
			break
		}
	}

	update := &pb.TurnUpdate{
		Actor:         g.state.NextPlayer_ID,
		NextPlayer_ID: pb.Actor((int(g.state.NextPlayer_ID) + 1) % 4),
		PlayedCard:    req,
		IsMatchOver:   false,
	}

	for i, card := range g.tableTop {
		if card.Game_ID == req.Game_ID && card.Rank == req.Rank {
			g.tableTop = append(g.tableTop[:i], g.tableTop[i+1:]...)
			g.history = append(g.history, card)
			g.history = append(g.history, req)
			update.CartePrese = append(update.CartePrese, card)
			if len(g.tableTop) == 0 {
				update.Scopa = true
			}

			return update
		}
	}

	combinazioniTotali := TrovaCombinazioni(g.tableTop, int32(req.Rank))

	if len(combinazioniTotali) == 0 {
		g.tableTop = append(g.tableTop, req)
		return update
	}
}
func (g *GameSession) subscribe() chan *pb.TurnUpdate {

	g.mu.Lock()
	defer g.mu.Unlock()
	ch := make(chan *pb.TurnUpdate, 10)
	g.listeners = append(g.listeners, ch)

	return ch
}

func (g *GameSession) runCPUPlayers(player pb.Actor, myChan chan *pb.TurnUpdate) {}

func (s *server) StartGame(ctx context.Context, req *pb.GameSettings) (*pb.InitialState, error) {

	newGame := &GameSession{

		deck:          make([]*pb.Card, 0),
		hands:         make(map[pb.Actor][]*pb.Card),
		scoreDeck:     make(map[int][]*pb.Card),
		victoryPoints: req.MaxPoints,
		hasStarted:    false,
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
			card := &pb.Card{Game_ID: int32(gameID), Suit: pb.Suit(i), Rank: pb.Rank(k)}
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

	newGame.state = &pb.TurnUpdate{
		Actor:         pb.Actor(dealerID),
		NextPlayer_ID: pb.Actor(dealerID + 1),
		IsMatchOver:   false,
	}

	go newGame.runCPUPlayers(pb.Actor_CPU_RIGHT, newGame.subscribe())
	go newGame.runCPUPlayers(pb.Actor_CPU_PARTNER, newGame.subscribe())
	go newGame.runCPUPlayers(pb.Actor_CPU_LEFT, newGame.subscribe())

	return &pb.InitialState{
		Game_ID:       int32(gameID),
		Dealer_ID:     pb.Actor(dealerID),
		CurrentPlayer: pb.Actor((dealerID + 1) % 4),
		UserHand:      newGame.hands[pb.Actor_USER],
	}, nil
}

func (s *server) PlayCard(req *pb.Card, stream pb.GoBackend_PlayCardServer) error {
	managerMU.Lock()
	game, exists := gameManager[int(req.Game_ID)]
	managerMU.Unlock()

	if !exists {
		std.Println("Partita non trovata")
		return nil
	}

	update := game.tableManager(req)
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
