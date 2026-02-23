# Progetto: Scopone Scientifico - Architettura Multiplayer e Data Analytics

Il presente repository contiene un'implementazione distribuita del tradizionale gioco di carte "Scopone Scientifico". Il sistema è stato progettato adottando un'architettura a microservizi, la quale integra un motore di gioco concorrente, un client con interfaccia testuale (TUI) e un sottosistema per l'analisi statistica dei dati di gioco in tempo reale.

## Architettura del Sistema

Il progetto è strutturato in tre moduli principali interconnessi:

1. **Backend (Go e gRPC):** Costituisce il nucleo applicativo. Gestisce la logica di dominio, l'applicazione delle regole, la transizione dei turni e il calcolo dei punteggi. Sfrutta le funzionalità di concorrenza del linguaggio Go (Goroutine e Mutex) per gestire accessi multipli in sicurezza (thread-safety). Il servizio è containerizzato e orchestrato tramite Docker.
2. **Modulo di Data Analytics (Python, Flask, Pandas):** Un server web preposto alla lettura asincrona dei dati di sessione e alla generazione dinamica di grafici statistici. Permette l'analisi dello storico delle vittorie, l'andamento dei singoli round e l'acquisizione dei punti di mazzo. Anch'esso viene eseguito all'interno di un container Docker dedicato.
3. **Client (C# e .NET):** Rappresenta l'interfaccia utente testuale (TUI). Comunica bidirezionalmente con il server Go tramite protocollo gRPC ed elabora gli aggiornamenti di stato per il rendering a schermo. Viene eseguito nativamente sull'ambiente host dell'utente.

---

## Prerequisiti di Sistema

Per la corretta compilazione ed esecuzione dell'applicativo, è richiesta la presenza dei seguenti strumenti sull'ambiente host:

* **Docker e Docker Compose:** Necessari per l'orchestrazione dei servizi backend e analytics.
* **.NET SDK 9.0:** Il progetto in questione è configurato nativamente con .NET SDK 9.0. 
  * Qualora si disponga di una versione **precedente**, si consiglia caldamente l'aggiornamento.
  * Qualora si disponga di una versione **successiva** (es. .NET 10), è necessario modificare la versione di destinazione (`TargetFramework`) all'interno del file `ScoponeClient.csproj` situato nella directory `ScoponeClient`.

---

## Istruzioni per l'Avvio

La procedura di avvio del progetto è suddivisa in due fasi: l'inizializzazione dell'infrastruttura server e l'esecuzione del client.


Aprire un terminale posizionandosi nella directory principale (root) del progetto ed eseguire il seguente comando per avviare i container in background:  

```bash
docker-compose up --build -d
```


Avendo avviato il server in background (grazie al flag -d), è possibile rimanere nella sessione di terminale attuale (oppure aprirne una nuova). Posizionarsi all'interno della directory del client ed avviare l'eseguibile tramite i seguenti comandi:


```bash
cd ScoponeClient
dotnet run
```


## Nota Operativa: 
È possibile istanziare molteplici processi client eseguendo il comando dotnet run in finestre di terminale separate, al fine di simulare o gestire partite indipendenti.

## Funzionalità e Comandi di Gioco
Durante l'esecuzione, l'interazione dell'utente avviene tramite input da tastiera, seguendo le indicazioni fornite dall'interfaccia. Oltre ai comandi standard di gioco, sono state implementate due funzionalità di interrogazione statistica:

Tasto 'P' (Statistiche della Partita Corrente): Richiama l'apertura del browser di sistema predefinito per esporre una dashboard relativa alla sessione in corso. Include l'andamento dei round, il computo delle carte acquisite, le scope effettuate e l'elaborazione dei punti di mazzo. Affinché i dati siano disponibili, è necessario aver concluso almeno la prima smazzata.

Tasto 'S' (Statistiche Globali): Genera e visualizza nel browser lo storico aggregato di tutte le partite registrate dal server, fornendo metriche quali le percentuali di vittoria, le medie dei punteggi e i relativi scarti.

## Visualizzazione delle Statistiche di Fine Partita
Al termine di ogni match, il sistema espone automaticamente le statistiche finali relative alla sessione appena conclusa, rendendole accessibili tramite l'interfaccia web dedicata.

Si precisa che, a causa delle tempistiche di scrittura asincrona dei dati sul file system condiviso, i grafici potrebbero non essere immediatamente renderizzati all'apertura del browser. Qualora la pagina dovesse risultare incompleta o priva di dati, sarà sufficiente effettuare un aggiornamento della pagina web (refresh) per forzare il corretto caricamento delle informazioni aggiornate.

## Note Tecniche e Ambiente di Collaudo (TUI)
Essendo l'interfaccia utente basata su riga di comando, la corretta formattazione della griglia grafica (allineamento delle carte e rendering dei semi) è strettamente dipendente dalla configurazione del terminale host.

Il presente applicativo è stato ampiamente collaudato in ambiente macOS, utilizzando il terminale integrato di Visual Studio configurato con il font a spaziatura fissa Monaco.

Si segnala pertanto che:

Configurazione del Font: L'utilizzo di font differenti o a spaziatura variabile potrebbe generare anomalie nel rendering delle emoji e dei simboli speciali associati ai semi delle carte, compromettendo l'allineamento geometrico della TUI. Si raccomanda caldamente l'impiego del font "Monaco" o di alternative Monospace equivalenti (es. Consolas in ambiente Windows).

## Codifica dei Caratteri
È necessario assicurarsi che il terminale supporti nativamente la codifica UTF-8. In ambiente Windows, è fortemente consigliato l'utilizzo di Windows Terminal in sostituzione del prompt dei comandi legacy (cmd.exe).

## Documentazione delle API

Il sistema si basa su due tipologie di interfacce di comunicazione: chiamate gRPC per il core del gioco e rotte HTTP (REST-like) per l'analytics.

### 1. API gRPC (Backend Go)
Definite nel file Protobuf, gestiscono la comunicazione bidirezionale tra Client C# e Server Go:
* `StartGame (GameSettings) -> InitialState`: Inizializza una nuova sessione di gioco, mescola il mazzo e restituisce la mano iniziale e l'ID della partita.
* `PlayCard (PlayRequest) -> TurnUpdate`: Permette all'utente di giocare una carta. Gestisce la logica di presa e la risoluzione di eventuali conflitti di combinazione.
* `ObserveTurn (ObserveRequest) -> stream TurnUpdate`: Flusso dati in *Server-Side Streaming*. Il client rimane in ascolto passivo per ricevere le mosse effettuate dalla CPU e gli aggiornamenti del tavolo in tempo reale.
* `CalcolaPunteggio (ObserveRequest) -> ScoreUpdate`: Invocata al termine di una smazzata per calcolare e salvare i punti di mazzo (Primiera, Settebello, ecc.) e verificare le condizioni di vittoria.

### 2. API HTTP / Flask (Data Analytics)
Esposte dal container Python sulla porta 5000:
* `GET /stats/match/<game_id>`: Genera e restituisce un'immagine PNG contenente la dashboard statistica della singola partita (carte prese, scope, punti di mazzo) basandosi sul file `rounds_history.csv`.
* `GET /stats/global` *(o eventuale rotta per il tasto S)*: Genera la dashboard globale dello storico partite leggendo il file `match_history.csv`.