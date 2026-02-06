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
	mappaScope    map[int]int32
	scorePoints   [2]int32
	history       []*pb.Card
	tableTop      []*pb.Card
	victoryPoints int32
	hasStarted    bool
	isGameOver    bool
	ultimaPresa   pb.Actor
}

var gameManager = make(map[int]*GameSession)
var managerMU sync.Mutex

var valPrimiera = map[pb.Rank]int32{
	pb.Rank_ASSO:    16,
	pb.Rank_DUE:     12,
	pb.Rank_TRE:     13,
	pb.Rank_QUATTRO: 14,
	pb.Rank_CINQUE:  15,
	pb.Rank_SEI:     18,
	pb.Rank_SETTE:   21,
	pb.Rank_FANTE:   10,
	pb.Rank_CAVALLO: 10,
	pb.Rank_RE:      10,
}

func calcolaRisultati(game *GameSession) [2]int32 {

	punteggioFinale := [2]int32{game.scorePoints[0], game.scorePoints[1]}

	carteTotali := [2]int32{0, 0}
	denariTotali := [2]int32{0, 0}

	maxPrimiera := [2][4]int32{
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	for i := 0; i < 4; i++ {
		teamID := i % 2

		punteggioFinale[teamID] += game.mappaScope[i]

		for _, card := range game.scoreDeck[i] {
			carteTotali[teamID]++

			if card.Suit == pb.Suit_DENARI {
				denariTotali[teamID]++
				if card.Rank == pb.Rank_SETTE {
					punteggioFinale[teamID]++
				}
			}

			valore := valPrimiera[card.Rank]
			suitIdx := int(card.Suit)

			if valore > maxPrimiera[teamID][suitIdx] {
				maxPrimiera[teamID][suitIdx] = valore
			}
		}
	}

	if carteTotali[0] > carteTotali[1] {
		punteggioFinale[0]++
	} else if carteTotali[1] > carteTotali[0] {
		punteggioFinale[1]++
	}

	if denariTotali[0] > denariTotali[1] {
		punteggioFinale[0]++
	} else if denariTotali[1] > denariTotali[0] {
		punteggioFinale[1]++
	}

	totalePrimiera := [2]int32{0, 0}
	for i := 0; i < 2; i++ {
		for k := 0; k < 4; k++ {
			totalePrimiera[i] += maxPrimiera[i][k]
		}
	}

	if totalePrimiera[0] > totalePrimiera[1] {
		punteggioFinale[0]++
	} else if totalePrimiera[1] > totalePrimiera[0] {
		punteggioFinale[1]++
	}

	return punteggioFinale
}

func calcolaGiocata(hand []*pb.Card, tableTop []*pb.Card, history []*pb.Card) *pb.Card {

	mappaPunteggi := make(map[*pb.Card]int)
	for _, card := range hand {
		mappaPunteggi[card] = 0
	}

	// caso tavolo vuoto
	if len(tableTop) == 0 {
		conta := make(map[pb.Rank]int)
		for _, card := range hand {
			conta[card.Rank]++
		}
		for _, card := range history {
			conta[card.Rank]++
		}

		for rank, numero := range conta {
			switch numero {
			case 4:
				for _, card := range hand {
					if card.Rank == rank && card.Suit != pb.Suit_DENARI {
						return card
					}
				}

			case 3:
				for _, card := range hand {
					mappaPunteggi[card] += 100
					if card.Rank == pb.Rank_SETTE && card.Suit == pb.Suit_DENARI {
						mappaPunteggi[card] -= 1500
					}
					if card.Rank == pb.Rank_SETTE {
						mappaPunteggi[card] -= 40
					}
					if card.Suit == pb.Suit_DENARI {
						mappaPunteggi[card] -= 20
					}
				}

			case 2:
				for _, card := range hand {
					mappaPunteggi[card] += 59
					if card.Rank == pb.Rank_SETTE && card.Suit == pb.Suit_DENARI {
						mappaPunteggi[card] -= 1500
					}
					if card.Rank == pb.Rank_SETTE {
						mappaPunteggi[card] -= 40
					}
					if card.Suit == pb.Suit_DENARI {
						mappaPunteggi[card] -= 20
					}
				}

			default:
				for _, card := range hand {
					mappaPunteggi[card] += 18
					if card.Rank == pb.Rank_SETTE && card.Suit == pb.Suit_DENARI {
						mappaPunteggi[card] -= 1500
					}
					if card.Rank == pb.Rank_SETTE {
						mappaPunteggi[card] -= 40
					}
					if card.Suit == pb.Suit_DENARI {
						mappaPunteggi[card] -= 20
					}
				}
			}
		}
	} else { // caso tavolo non vuoto
		contaPunteggioTavolo := 0
		for _, card := range tableTop {
			contaPunteggioTavolo += int(card.Rank)
		}

		// se posso fare scopa
		if contaPunteggioTavolo <= 10 {
			possibileGiocata := make([]*pb.Card, 0)

			for _, card := range hand {
				if card.Rank == pb.Rank(contaPunteggioTavolo) {
					if card.Suit == pb.Suit_DENARI {
						return card
					} else {
						possibileGiocata = append(possibileGiocata, card)
					}
				}
			}
			return possibileGiocata[0]
		}

		// se non posso fare scopa (valuto le carte in mano)
		for _, card := range hand {
			combinazioni := trovaCombinazioni(tableTop, int32(card.Rank))
			if len(combinazioni) == 0 {
				if card.Suit == pb.Suit_DENARI {
					mappaPunteggi[card] += 100
				}
				if card.Rank == pb.Rank_SETTE {
					mappaPunteggi[card] += 500
				}

				punteggioMax := 0
				combinazioneScelta := make([]*pb.Card, 0)
				for _, combinazione := range combinazioni {
					punteggioAttuale := calcolaPunteggioCombinazione(combinazione)
					if punteggioAttuale > punteggioMax {
						punteggioMax = punteggioAttuale
						combinazioneScelta = combinazione
					}
				}

				coteggioGiocata := 0
				for _, c := range combinazioneScelta {
					coteggioGiocata += int(c.Rank)
				}

				if (contaPunteggioTavolo - coteggioGiocata) <= 10 {
					mappaPunteggi[card] -= 1500
				} else {
					mappaPunteggi[card] += 100
				}
			} else {
				if card.Suit == pb.Suit_DENARI {
					mappaPunteggi[card] -= 100
				}
				if card.Rank == pb.Rank_SETTE {
					mappaPunteggi[card] -= 500
				}
			}
		}
	}

	var cartaScelta *pb.Card
	var punteggioMax int = 0

	for card, punteggio := range mappaPunteggi {
		if punteggio > punteggioMax {
			punteggioMax = punteggio
			cartaScelta = card
		}
	}
	return cartaScelta
}

func calcolaPunteggioCombinazione(combinazione []*pb.Card) int {
	Punteggio := 0

	for _, card := range combinazione {
		Punteggio += 10

		if card.Suit == pb.Suit_DENARI {
			Punteggio += 50
		}

		switch card.Rank {
		case pb.Rank_SETTE:
			if card.Suit == pb.Suit_DENARI {
				Punteggio += 1500
			} else {
				Punteggio += 200
			}
		case pb.Rank_SEI:
			Punteggio += 60

		case pb.Rank_ASSO:
			Punteggio += 20
		default:
			Punteggio += 0
		}

	}
	return Punteggio
}

func (g *GameSession) subscribe() chan *pb.TurnUpdate {

	g.mu.Lock()
	defer g.mu.Unlock()
	ch := make(chan *pb.TurnUpdate, 100)
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
				g.mappaScope[int(player)]++
			}

			g.ultimaPresa = player
			return update
		}
	}

	combinazioniTotali := trovaCombinazioni(g.tableTop, int32(req.Rank))

	if len(combinazioniTotali) == 0 {
		g.tableTop = append(g.tableTop, req)
		update.CartePrese = nil
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
			g.mappaScope[int(player)]++
		}
		g.ultimaPresa = player
		return update
	} else if len(combinazioniTotali) > 1 {
		punteggiomax := 0
		for _, combinazione := range combinazioniTotali {
			punteggioAttuale := calcolaPunteggioCombinazione(combinazione)
			if punteggioAttuale > punteggiomax {
				punteggiomax = punteggioAttuale
				update.CartePrese = combinazione
			}
		}

		for _, card := range update.CartePrese {
			for i, tableCard := range g.tableTop {
				if tableCard.Game_ID == card.Game_ID && tableCard.Suit == card.Suit && tableCard.Rank == card.Rank {
					g.tableTop = append(g.tableTop[:i], g.tableTop[i+1:]...)
					g.history = append(g.history, tableCard)
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
			g.mappaScope[int(player)]++
		}
		g.ultimaPresa = player
		return update
	}

	return update
}

func (g *GameSession) runCPUPlayers(player pb.Actor, myChan chan *pb.TurnUpdate) {
	for update := range myChan {
		if update.NextPlayer_ID == player && !g.isGameOver && !update.IsMatchOver {
			g.mu.Lock()

			cpuHand := g.hands[player]

			if len(cpuHand) == 0 {
				updateFine := &pb.TurnUpdate{
					Actor:         g.ultimaPresa,
					NextPlayer_ID: -1,
					IsMatchOver:   true,
				}

				if len(g.tableTop) > 0 {
					for i, card := range g.tableTop {
						g.history = append(g.history, card)
						g.scoreDeck[int(g.ultimaPresa)] = append(g.scoreDeck[int(g.ultimaPresa)], card)
						updateFine.CartePrese = append(updateFine.CartePrese, card)
						g.tableTop = append(g.tableTop[:i], g.tableTop[i+1:]...)
					}
				} else if len(g.tableTop) == 0 {
					g.mappaScope[int(g.ultimaPresa)]--
				}

				g.broadcastUpdate(updateFine)
				g.mu.Unlock()
				continue
			}

			cartaScelta := calcolaGiocata(cpuHand, g.tableTop, g.history)

			upadateTurnoPlayer := g.tableManager(cartaScelta, player)

			go g.broadcastUpdate(upadateTurnoPlayer)

			g.mu.Unlock()
		}
	}
}
