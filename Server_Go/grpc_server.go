package main

import (
	"context"

	"math/rand"

	pb "scopone_server/Proto_Files"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedGoBackendServer
}

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
		return status.Error(codes.NotFound, "Sessione di gioco non trovata")
	}

	canaleRicezione := game.subscribe()

	game.mu.Lock()
	if game.state.NextPlayer_ID != pb.Actor_USER {
		game.mu.Unlock()
		return status.Error(codes.FailedPrecondition, "Non è il turno dell'utente")
	}

	update := game.tableManager(req, pb.Actor_USER)

	go game.broadcastUpdate(update)

	game.mu.Unlock()

	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-canaleRicezione:
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

func (s *server) ObserveTurn(req *pb.ObserveRequest, stream pb.GoBackend_ObserveTurnServer) error {
	managerMU.Lock()
	game, exists := gameManager[int(req.Game_ID)]
	managerMU.Unlock()
	if !exists {
		return status.Error(codes.NotFound, "Sessione di gioco non trovata")
	}

	canaleRicezione := game.subscribe()

	game.mu.Lock()

	stream.Send(game.state)

	if !game.hasStarted {
		game.hasStarted = true
		if game.state.NextPlayer_ID != pb.Actor_USER {
			go game.broadcastUpdate(game.state)
		}
	}

	game.mu.Unlock()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case update := <-canaleRicezione:
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}
