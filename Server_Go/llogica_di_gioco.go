package main

import (
	"fmt"
	pb "scopone_server/Proto_Files"
	"sync"
)

type GameSession struct {
	gameID             int
	userName           string
	mu                 sync.Mutex
	state              *pb.TurnUpdate
	dealer_ID          int32
	listeners          []chan *pb.TurnUpdate
	deck               []*pb.Card
	hands              map[pb.Actor][]*pb.Card
	scoreDeck          map[int][]*pb.Card
	mappaScope         map[int]int32
	scorePoints        [2]int32
	history            []*pb.Card
	tableTop           []*pb.Card
	victoryPoints      int32
	hasStarted         bool
	isGameOver         bool
	ultimaPresa        pb.Actor
	roundNum           int
	risultatiCalcolati bool
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
	if game.risultatiCalcolati {
		return [2]int32{game.scorePoints[0], game.scorePoints[1]}

	}

	game.risultatiCalcolati = true

	game.roundNum++

	punteggioFinale := [2]int32{game.scorePoints[0], game.scorePoints[1]}

	carteTotali := [2]int32{0, 0}
	denariTotali := [2]int32{0, 0}

	maxPrimiera := [2][4]int32{
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	}

	settebbelloUser := false
	settebbelloCPU := false

	for i := 0; i < 4; i++ {
		teamID := i % 2

		punteggioFinale[teamID] += game.mappaScope[i]

		for _, card := range game.scoreDeck[i] {
			carteTotali[teamID]++

			if card.Suit == pb.Suit_DENARI {
				denariTotali[teamID]++
				if card.Rank == pb.Rank_SETTE {
					punteggioFinale[teamID]++
					if teamID == 0 {
						settebbelloUser = true
					} else {
						settebbelloCPU = true
					}
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

	denariUser := denariTotali[0] > denariTotali[1]
	denariCpu := denariTotali[1] > denariTotali[0]

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

	primieraUser := totalePrimiera[0] > totalePrimiera[1]
	primieraCpu := totalePrimiera[1] > totalePrimiera[0]

	if totalePrimiera[0] > totalePrimiera[1] {
		punteggioFinale[0]++
	} else if totalePrimiera[1] > totalePrimiera[0] {
		punteggioFinale[1]++
	}

	scopeUser := game.mappaScope[0] + game.mappaScope[2]
	scopeCpu := game.mappaScope[1] + game.mappaScope[3]

	SalvStatisticheRound(game.userName, game.gameID, game.roundNum, int(carteTotali[0]), int(carteTotali[1]), int(scopeUser), int(scopeCpu), primieraUser, primieraCpu, settebbelloCPU, settebbelloUser, denariUser, denariCpu)

	for i := 0; i < 4; i++ {
		game.scoreDeck[i] = nil
		game.mappaScope[i] = 0
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

		for _, card := range hand {
			numero := conta[card.Rank]

			if numero == 4 && card.Suit != pb.Suit_DENARI {
				return card
			}

			switch numero {
			case 3:
				mappaPunteggi[card] += 100

			case 2:
				mappaPunteggi[card] += 59
			default:
				mappaPunteggi[card] += 18
			}

			if card.Rank == pb.Rank_SETTE && card.Suit == pb.Suit_DENARI {
				mappaPunteggi[card] -= 1500
			} else if card.Rank == pb.Rank_SETTE {
				mappaPunteggi[card] -= 40
			} else if card.Suit == pb.Suit_DENARI {
				mappaPunteggi[card] -= 20
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

			if len(possibileGiocata) > 0 {
				return possibileGiocata[0]
			}
		}

		// se non posso fare scopa (valuto le carte in mano)
		for _, card := range hand {
			combinazioni := trovaCombinazioni(tableTop, int32(card.Rank))
			if len(combinazioni) > 0 {

				punteggioMax := -9999999
				combinazioneScelta := make([]*pb.Card, 0)

				for _, combinazione := range combinazioni {
					punteggioAttuale := calcolaPunteggioCombinazione(combinazione)
					if punteggioAttuale > punteggioMax {
						punteggioMax = punteggioAttuale
						combinazioneScelta = combinazione
					}
				}

				mappaPunteggi[card] += punteggioMax

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

				if contaPunteggioTavolo+int(card.Rank) <= 10 {
					mappaPunteggi[card] -= 1500
				}
			}
		}
	}

	var cartaScelta *pb.Card
	var punteggioMax int = -9999999

	for card, punteggio := range mappaPunteggi {
		if punteggio > punteggioMax {
			punteggioMax = punteggio
			cartaScelta = card
		}
	}
	return cartaScelta
}

func (g *GameSession) controlloFineRound(update *pb.TurnUpdate, player pb.Actor) {

	if len(g.history)+len(g.tableTop) == 40 {
		update.IsMatchOver = true
		update.NextPlayer_ID = -1

		if len(g.tableTop) > 0 {
			ultimoID := int(g.ultimaPresa)

			g.history = append(g.history, g.tableTop...)
			g.scoreDeck[ultimoID] = append(g.scoreDeck[ultimoID], g.tableTop...)
			update.CartePrese = append(update.CartePrese, g.tableTop...)

			g.tableTop = nil
		} else {
			if update.Scopa {
				update.Scopa = false
				g.mappaScope[int(player)]--
			}
		}
	}

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
		ch <- update
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

func (g *GameSession) tableManager(req *pb.PlayRequest, player pb.Actor) *pb.TurnUpdate {

	cartaGiocata := req.PlayedCard

	if cartaGiocata == nil {
		fmt.Println("ERRORE: PlayedCard è nil!")
		return &pb.TurnUpdate{ConflictResolutionNeeded: false}
	}

	for i, card := range g.hands[player] {
		if isSameCard(card, cartaGiocata) {
			g.hands[player] = append(g.hands[player][:i], g.hands[player][i+1:]...) // eliminazione carta dalla mano utente
			break
		}
	}

	update := &pb.TurnUpdate{
		Actor:         player,
		NextPlayer_ID: pb.Actor((int(player) + 1) % 4),
		PlayedCard:    cartaGiocata,
		IsMatchOver:   false,
	}

	//logica nel caso in cui la carta giocata corrisponde con una a terra
	for i, card := range g.tableTop {
		if card.Rank == cartaGiocata.Rank {
			g.tableTop = append(g.tableTop[:i], g.tableTop[i+1:]...)
			g.history = append(g.history, card)
			g.history = append(g.history, cartaGiocata)
			update.CartePrese = append(update.CartePrese, card)
			update.CartePrese = append(update.CartePrese, cartaGiocata)
			g.scoreDeck[int(player)] = append(g.scoreDeck[int(player)], card, cartaGiocata)

			if len(g.tableTop) == 0 {
				update.Scopa = true
				g.mappaScope[int(player)]++
			}

			g.ultimaPresa = player

			g.controlloFineRound(update, player)

			return update
		}
	}

	combinazioniTotali := trovaCombinazioni(g.tableTop, int32(cartaGiocata.Rank))

	if len(combinazioniTotali) == 0 {

		g.tableTop = append(g.tableTop, cartaGiocata)
		update.CartePrese = nil

		g.controlloFineRound(update, player)

		return update

	} else if len(combinazioniTotali) == 1 {
		combinazione := combinazioniTotali[0]
		for _, card := range combinazione {
			for i, tableCard := range g.tableTop {
				if tableCard.Suit == card.Suit && tableCard.Rank == card.Rank {
					g.tableTop = append(g.tableTop[:i], g.tableTop[i+1:]...)
					g.history = append(g.history, tableCard)
					update.CartePrese = append(update.CartePrese, tableCard)
					g.scoreDeck[int(player)] = append(g.scoreDeck[int(player)], tableCard)
					break
				}
			}
		}
		g.history = append(g.history, cartaGiocata)
		update.CartePrese = append(update.CartePrese, cartaGiocata)
		g.scoreDeck[int(player)] = append(g.scoreDeck[int(player)], cartaGiocata)

		if len(g.tableTop) == 0 {
			update.Scopa = true
			g.mappaScope[int(player)]++
		}
		g.ultimaPresa = player

		g.controlloFineRound(update, player)

		return update

	} else if len(combinazioniTotali) > 1 {
		if player == pb.Actor_USER {
			if len(req.TargetCard) == 0 {
				g.hands[player] = append(g.hands[player], cartaGiocata)
				return &pb.TurnUpdate{
					ConflictResolutionNeeded: true,
					Option:                   convertToProtoOption(combinazioniTotali),
				}
			}
			update.CartePrese = req.TargetCard
		} else {
			punteggiomax := -9999
			for _, combinazione := range combinazioniTotali {
				punteggioAttuale := calcolaPunteggioCombinazione(combinazione)
				if punteggioAttuale > punteggiomax {
					punteggiomax = punteggioAttuale
					update.CartePrese = combinazione
				}
			}
		}

		for _, card := range update.CartePrese {
			for i, tableCard := range g.tableTop {
				if tableCard.Suit == card.Suit && tableCard.Rank == card.Rank {
					g.tableTop = append(g.tableTop[:i], g.tableTop[i+1:]...)
					g.history = append(g.history, tableCard)
					g.scoreDeck[int(player)] = append(g.scoreDeck[int(player)], tableCard)
					break
				}
			}
		}

		g.history = append(g.history, cartaGiocata)
		update.CartePrese = append(update.CartePrese, cartaGiocata)
		g.scoreDeck[int(player)] = append(g.scoreDeck[int(player)], cartaGiocata)
		if len(g.tableTop) == 0 {
			update.Scopa = true
			g.mappaScope[int(player)]++
		}
		g.ultimaPresa = player

		g.controlloFineRound(update, player)

		return update
	}

	return update
}

func (g *GameSession) runCPUPlayers(player pb.Actor, myChan chan *pb.TurnUpdate) {
	for update := range myChan {
		if update.NextPlayer_ID == player && !g.isGameOver && !update.IsMatchOver {
			g.mu.Lock()

			if len(g.hands[player]) == 0 {
				g.mu.Unlock()
				continue
			}

			cartaScelta := calcolaGiocata(g.hands[player], g.tableTop, g.history)

			fakePlayrequest := &pb.PlayRequest{
				PlayedCard: cartaScelta,
			}

			upadateTurnoPlayer := g.tableManager(fakePlayrequest, player)

			g.state = upadateTurnoPlayer

			g.mu.Unlock()

			go g.broadcastUpdate(upadateTurnoPlayer)
		}
	}
}
