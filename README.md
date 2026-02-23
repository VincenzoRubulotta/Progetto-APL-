Prerequisiti di Sistema
Per la corretta compilazione ed esecuzione dell'applicativo, è richiesta la presenza dei seguenti strumenti sull'ambiente host:

Docker e Docker Compose (necessari per l'orchestrazione dei servizi backend e analytics).

Il progetto in questione è configurto con .NET SDK 9.0 nel caso si dispone
di una versione più vechia è consigliato aggiornarla con la 9.0, se invece
la versione a dispozione è superiore alla 9.0 si consiglia di cambiare la versione corrente all'interno del file "ScoponeClient.csproj" che si trova nella directory "ScoponeClient"

struzioni per l'Avvio
La procedura di avvio del progetto è suddivisa in due fasi: l'inizializzazione dell'infrastruttura server e l'esecuzione del client.

Fase 1: Avvio dell'Infrastruttura Server (Docker)

Aprire un terminale posizionandosi nella directory principale (root) del progetto ed eseguire il seguente comando per avviare i container in background:  

docker-compose up --build -d

Fase 2: Avvio del Client di Gioco (C#)

Aprire una nuova sessione del terminale(o rimanere nella sessione attuale se si è avviato il server in backgorund), posizionarsi all'interno della directory del client ed avviare l'eseguibile:

cd ScoponeClient
dotnet run

Nota: È possibile istanziare molteplici processi client eseguendo il comando dotnet run in finestre di terminale separate, al fine di simulare o gestire partite indipendenti.

Funzionalità e Comandi di Gioco
Durante l'esecuzione, l'interazione dell'utente avviene tramite input da tastiera, seguendo le indicazioni fornite dall'interfaccia. Oltre ai comandi standard di gioco, sono state implementate due funzionalità di interrogazione statistica:

Tasto 'P' (Statistiche della Partita Corrente): Richiama l'apertura del browser di sistema predefinito per esporre una dashboard relativa alla sessione in corso. Include l'andamento dei round, il computo delle carte acquisite, le scope effettuate e l'elaborazione dei punti di mazzo. Affinché i dati siano disponibili, è necessario aver concluso almeno la prima smazzata.

Tasto 'S' (Statistiche Globali): Genera e visualizza nel browser lo storico aggregato di tutte le partite registrate dal server, fornendo metriche quali le percentuali di vittoria, le medie dei punteggi e i relativi scarti.

Visualizzazione delle Statistiche di Fine Partita
Al termine di ogni match, il sistema espone automaticamente le statistiche finali relative alla sessione appena conclusa, rendendole accessibili tramite l'interfaccia web dedicata. Si precisa che, a causa delle tempistiche di scrittura asincrona dei dati sul file system condiviso, i grafici potrebbero non essere immediatamente renderizzati all'apertura del browser. Qualora la pagina dovesse risultare incompleta o priva di dati, sarà sufficiente effettuare un aggiornamento della pagina web (refresh) per forzare il corretto caricamento delle informazioni aggiornate.

Note Tecniche e Ambiente di Collaudo (TUI)
Essendo l'interfaccia utente basata su riga di comando, la corretta formattazione della griglia grafica (allineamento delle carte e rendering dei semi) è strettamente dipendente dalla configurazione del terminale host.

Il presente applicativo è stato ampiamente collaudato in ambiente macOS, utilizzando il terminale integrato di Visual Studio configurato con il font a spaziatura fissa Monaco.

Si segnala pertanto che:

Configurazione del Font: L'utilizzo di font differenti o a spaziatura variabile potrebbe generare anomalie nel rendering delle emoji e dei simboli speciali associati ai semi delle carte, compromettendo l'allineamento geometrico della TUI. Si raccomanda caldamente l'impiego del font "Monaco" o di alternative Monospace equivalenti (es. Consolas in ambiente Windows).

Codifica dei Caratteri: È necessario assicurarsi che il terminale supporti nativamente la codifica UTF-8. In ambiente Windows, è fortemente consigliato l'utilizzo di Windows Terminal in sostituzione del prompt dei comandi legacy (cmd.exe).