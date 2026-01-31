package main

import (
	pb "scopone_server/Proto_Files"
	"sync"
)

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

func (g *GameSession) subscribe() chan *pb.TurnUpdate {

	g.mu.Lock()
	defer g.mu.Unlock()
	ch := make(chan *pb.TurnUpdate, 10)
	g.listeners = append(g.listeners, ch)

	return ch
}

func (g *GameSession) broadcastUpdate(update *pb.TurnUpdate) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.state = update

	for _, ch := range g.listeners {
		select {
		case ch <- update:
		default:
		}
	}
}

func trovaCombinazioni(cards []*pb.Card, target int32) [][]*pb.Card {
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

func (g *GameSession) tableManager(req *pb.Card, player pb.Actor) *pb.TurnUpdate {
	g.mu.Lock()
	defer g.mu.Unlock()

	for i, card := range g.hands[player] {
		if card.Game_ID == req.Game_ID && card.Suit == req.Suit && card.Rank == req.Rank {
			g.hands[player] = append(g.hands[player][:i], g.hands[player][i+1:]...) // eliminazione carta dalla mano utente
			break
		}
	}

	update := &pb.TurnUpdate{
		Actor:         player,
		NextPlayer_ID: pb.Actor((int(player) + 1) % 4),
		PlayedCard:    req,
		IsMatchOver:   false,
	}

	for i, card := range g.tableTop {
		if card.Game_ID == req.Game_ID && card.Rank == req.Rank {
			g.tableTop = append(g.tableTop[:i], g.tableTop[i+1:]...)
			g.history = append(g.history, card)
			g.history = append(g.history, req)
			update.CartePrese = append(update.CartePrese, card)
			update.CartePrese = append(update.CartePrese, req)
			g.scoreDeck[int(player)] = append(g.scoreDeck[int(player)], card, req)
			if len(g.tableTop) == 0 {
				update.Scopa = true
			}

			return update
		}
	}

	combinazioniTotali := trovaCombinazioni(g.tableTop, int32(req.Rank))

	if len(combinazioniTotali) == 0 {
		g.tableTop = append(g.tableTop, req)
		return update
	} else if len(combinazioniTotali) == 1 {
		combinazione := combinazioniTotali[0]
		for _, card := range combinazione {
			for i, tableCard := range g.tableTop {
				if tableCard.Game_ID == card.Game_ID && tableCard.Suit == card.Suit && tableCard.Rank == card.Rank {
					g.tableTop = append(g.tableTop[:i], g.tableTop[i+1:]...)
					g.history = append(g.history, tableCard)
					update.CartePrese = append(update.CartePrese, tableCard)
					g.scoreDeck[int(player)] = append(g.scoreDeck[int(player)], tableCard)
					break
				}
			}
		}
		g.history = append(g.history, req)
		update.CartePrese = append(update.CartePrese, req)
		g.scoreDeck[int(player)] = append(g.scoreDeck[int(player)], req)

		if len(g.tableTop) == 0 {
			update.Scopa = true
		}
		return update
	} else if len(combinazioniTotali) > 1 {
		// da implementare ancora come scegliere la combinazione con il punteggio migliore

		return update
	}

	return update
}

func (g *GameSession) runCPUPlayers(player pb.Actor, myChan chan *pb.TurnUpdate) {
	// ancora da implementare
}
