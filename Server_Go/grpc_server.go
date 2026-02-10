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
		mappaScope:    make(map[int]int32),
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

	for i := 0; i < 4; i++ {
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

	for i := 0; i < 2; i++ {
		newGame.scorePoints[i] = 0
	}

	newGame.state = &pb.TurnUpdate{
		Actor:         pb.Actor(dealerID),
		NextPlayer_ID: pb.Actor((dealerID + 1) % 4),
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

func (s *server) PlayCard(ctx context.Context, req *pb.PlayRequest) (*pb.TurnUpdate, error) {
	managerMU.Lock()
	game, exists := gameManager[int(req.Game_ID)]
	managerMU.Unlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "Sessione di gioco non trovata")
	}

	game.mu.Lock()

	if game.state.NextPlayer_ID != pb.Actor_USER {
		game.mu.Unlock()
		return nil, status.Error(codes.FailedPrecondition, "Non è il turno dell'utente")
	}

	update := game.tableManager(req, pb.Actor_USER)

	if update.ConflictResolutionNeeded {
		game.mu.Unlock()
		return update, nil
	}

	game.state = update

	game.mu.Unlock()

	go game.broadcastUpdate(update)

	return update, nil
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

func (s *server) CalcolaPunteggio(ctx context.Context, req *pb.ObserveRequest) (*pb.ScoreUpdate, error) {
	managerMU.Lock()
	game, exist := gameManager[int(req.Game_ID)]
	managerMU.Unlock()

	if !exist {
		return nil, status.Error(codes.NotFound, "Sessione di gioco non trovata")
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	scoreUpdate := &pb.ScoreUpdate{
		Game_ID: int32(req.Game_ID),
	}

	punteggio := calcolaRisultati(game)

	scoreUpdate.CpuSquadScore = punteggio[1]
	scoreUpdate.UserSqudScore = punteggio[0]
	for i := 0; i < 2; i++ {
		game.scorePoints[i] = punteggio[i]
	}

	if punteggio[0] > game.victoryPoints && punteggio[0] > punteggio[1] {
		scoreUpdate.IsGameOver = true
		game.isGameOver = true
		scoreUpdate.UserHand = nil
	} else if punteggio[1] > game.victoryPoints && punteggio[1] > punteggio[0] {
		scoreUpdate.IsGameOver = true
		game.isGameOver = true
		scoreUpdate.UserHand = nil
	} else {
		scoreUpdate.IsGameOver = false
		game.dealer_ID = (game.dealer_ID + 1) % 4
		scoreUpdate.NextPlayer_ID = pb.Actor((game.dealer_ID + 1) % 4)

		for i := range game.scorePoints {
			game.scoreDeck[i] = nil
			game.mappaScope[i] = 0
		}

		game.deck = game.history
		game.history = nil
		game.tableTop = nil

		rand.Shuffle(len(game.deck), func(i, j int) {
			game.deck[i], game.deck[j] = game.deck[j], game.deck[i]
		})

		tempDealer := game.dealer_ID
		n := 0
		for i := 1; i <= 4; i++ {
			tempDealer = (tempDealer + 1) % 4
			game.hands[pb.Actor(tempDealer)] = game.deck[n : n+10]
			n = n + 10
		}

		scoreUpdate.UserHand = game.hands[pb.Actor_USER]

		newState := &pb.TurnUpdate{
			Actor:         pb.Actor(game.dealer_ID),
			NextPlayer_ID: scoreUpdate.NextPlayer_ID,
			IsMatchOver:   false,
			PlayedCard:    nil,
			CartePrese:    nil,
			Scopa:         false,
		}

		game.state = newState

		go game.broadcastUpdate(newState)
	}

	return scoreUpdate, nil
}
