package main

import (
	"encoding/csv"
	"fmt"
	"os"
	pb "scopone_server/Proto_Files"
	"strconv"
	"time"
)

func isSameCard(c1, c2 *pb.Card) bool {
	return c1.Suit == c2.Suit && c1.Rank == c2.Rank
}

func convertToProtoOption(combos [][]*pb.Card) []*pb.Cobination {
	var option []*pb.Cobination
	for _, c := range combos {
		option = append(option, &pb.Cobination{Cards: c})
	}

	return option
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

func salvaStatistichePartite(gameID int32, userScore int32, cpuScore int32, limitepunti int32) {
	filename := "/data/match_history.csv"

	if _, err := os.Stat("/data"); os.IsNotExist(err) {
		filename = "match_history.csv"
	}

	fileExists := false
	if _, err := os.Stat(filename); err == nil {
		fileExists = true
	}
	// 0644 sono le modalità di accesso 0 numero in formato ottalw
	// 6 indica che il propietario può leggere e scrivere
	// i due 4 indicano che il gruppo,gli altri  può solo leggere
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Errore nell'apertura del file CSV: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !fileExists {
		header := []string{"Data", "GameID", "LimitePunti", "PuntiUser", "PuntiCPU", "Vincitore"}
		writer.Write(header)
	}

	vincitore := "CPU"

	if userScore > cpuScore {
		vincitore = "USER"
	}

	record := []string{
		time.Now().Format("02-01-2006 15:04:05"),
		strconv.Itoa(int(gameID)),
		strconv.Itoa(int(limitepunti)),
		strconv.Itoa(int(userScore)),
		strconv.Itoa(int(cpuScore)),
		vincitore,
	}

	err = writer.Write(record)
	if err != nil {
		fmt.Printf("Errore scrittura CSV: %v\n", err)
	} else {
		fmt.Println("Statistiche partitr salvate con successo")
	}
}
