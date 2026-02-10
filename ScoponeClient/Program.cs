using System;
using System.Collections.Generic;
using System.Reflection.Metadata;
using System.Threading;
using System.Threading.Tasks;
using Grpc.Net.Client;
using ScoponeScientifico;

class Program
{
    static int _gameID;
    static List<card> _myHand = new List<card>();
    static go_backend.go_backendClient _client = null!;
    static bool _isMyTurn = false;
    static bool _isMatchOver = false;

    static readonly object _consoleLock = new object();
    static async Task Main(string[] args)
    {
       Log("--- SCOPONE SCIENTIFICO C# ---");

        var channel = GrpcChannel.ForAddress("http://localhost:50051");
        _client = new go_backend.go_backendClient(channel);

        try
        {
            Log("Avvio partita...");
            var startReq = new game_settings { UserName = "Hero", MaxPoints = 11 };
            var initData = await _client.start_gameAsync(startReq);

            _gameID = initData.GameID;
            _myHand.AddRange(initData.UserHand);

            Log($"Partita ID: {_gameID} iniziata.");
            Log($"Mazziere: {initData.DealerID}");

            _ = Task.Run(() => AscoltaServer());

            while (true)
            {
                if (_isMatchOver)
                {
                    bool continua = await GestisciFineSmazzata();
                    if (!continua) break;
                }
                else if (_isMyTurn && _myHand.Count > 0)
                {
                    await FaiLaTuaMossa();
                }

                await Task.Delay(100);
            }
        }
        catch (Exception ex)
        {
            Log($"ERRORE FATALE: {ex.Message}");
        }
    }


    static void Log(string message)
        {
            lock (_consoleLock)
            {
                Console.WriteLine(message);
            }
        }

        static void LogInline(string message)
        {
            lock (_consoleLock)
            {
                Console.Write(message);
            }
        }

    static async Task<bool> GestisciFineSmazzata()
    {
        try
        {
            var request = new observe_request { GameID = _gameID };
            var scoreUpdate = await _client.calcola_punteggioAsync(request);

            lock (_consoleLock)
            {
                Console.WriteLine("=================================");
                Console.WriteLine("       RISULTATI PARZIALI        ");
                Console.WriteLine($"NOI (User+Partner): {scoreUpdate.UserSqudScore}");
                Console.WriteLine($"LORO (CpuLeft+Right): {scoreUpdate.CpuSquadScore}");
                Console.WriteLine("=================================");
            }
            

            if (scoreUpdate.IsGameOver)
            {
               Log("\n*** PARTITA CONCLUSA ***");

                if (scoreUpdate.UserSqudScore > scoreUpdate.CpuSquadScore)
                   Log("HAI VINTO! COMPLIMENTI!");
                else
                    Log("HAI PERSO. RITENTA!");

                return false;
            }

            Log("\nLa partita continua! Distribuzione nuove carte...");

            _myHand.Clear();
            _myHand.AddRange(scoreUpdate.UserHand);

            Log($"Ho ricevuto {_myHand.Count} nuove carte.");

            _isMatchOver = false;

            if (scoreUpdate.NextPlayerID == Actor.User)
            {
                Log("!!! TOCCA A TE (Primo di mano) !!!");
            }
            else
            {
                Log($"Inizia il giocatore: {scoreUpdate.NextPlayerID}");
            }

            return true;
        }
        catch (Exception ex)
        {
            Log($"Errore nel calcolo punteggi: {ex.Message}");
            return false;
        }
    }
    static async Task AscoltaServer()
    {
        try
        {
            var request = new observe_request { GameID = _gameID };
            using var stream = _client.observe_turn(request);

            Log("In ascolto degli eventi di gioco...");

            while (await stream.ResponseStream.MoveNext(CancellationToken.None))
            {
                var update = stream.ResponseStream.Current;

                lock (_consoleLock)
                {
                    if (update.PlayedCard != null)
                    {
                        Console.WriteLine($"\n[EVENTO] Ha giocato: {update.Actor} -> {update.PlayedCard.Rank} di {update.PlayedCard.Suit}");
                    }
                    else
                    {
                        Console.Write(" (Nessuna carta giocata)");
                    }

                    if (update.CartePrese.Count > 0)
                    {
                        Console.WriteLine($"   >>> PRESA! {update.CartePrese.Count} carte catturate.");
                    }

                    if (update.Scopa)
                    {
                        Console.WriteLine("   *** SCOPA! ***");
                    }
                }

                if (update.IsMatchOver)
                    {
                        _isMatchOver = update.IsMatchOver;
                    }

                Log($"Prossimo giocatore: {update.NextPlayerID}");

                if (update.NextPlayerID == Actor.User)
                {
                    Log("!!! TOCCA A TE !!!");
                    _isMyTurn = true;
                }
                else
                {
                    _isMyTurn = false;
                }

            }
        }
        catch (Exception ex)
        {
            Log($"Errore Stream: {ex.Message}");
        }
    }

    static async Task FaiLaTuaMossa()
    {
        lock (_consoleLock)
        {
            Console.WriteLine("\n-------------------------");
            Console.WriteLine("       IL TUO TURNO      ");
            Console.WriteLine("-------------------------");

            for (int i = 0; i < _myHand.Count; i++)
            {
                Console.WriteLine($"[{i}] {_myHand[i].Rank} di {_myHand[i].Suit}");
            }
        }

        int index = -1; 
        
        while (Console.KeyAvailable) Console.ReadKey(true);

        while (true)
        {
            LogInline($"Scegli l'indice della carta da giocare (0-{_myHand.Count - 1}): ");

            string input = Console.ReadLine() ?? "";

            if (string.IsNullOrWhiteSpace(input))
            {
                continue;
            }

            if (int.TryParse(input, out int parsedIndex) && parsedIndex >= 0 && parsedIndex < _myHand.Count)
            {
                index = parsedIndex;
                break;
            }
            else
            {
                Log("Indice non valido. Riprova.");
            }
        }

        var cartaScelta = _myHand[index];

        Log($"Invio carta: {cartaScelta.Rank}...");
        
        _isMyTurn = false;

        try
        {
            var req = new PlayRequest
            {
                GameID = _gameID,
                PlayedCard = cartaScelta
            };

            var response = await _client.play_cardAsync(req);

            if (response.ConflictResolutionNeeded)
            {
                lock (_consoleLock)
                {
                    Console.WriteLine("\n!!! AMBIGUITÀ - SCEGLI PRESA !!!");
                    for (int i = 0; i < response.Option.Count; i++)
                    {
                        Console.Write($"[{i}]: ");
                        foreach (var c in response.Option[i].Cards) Console.Write($"[{c.Rank}{c.Suit}] ");
                        Console.WriteLine();
                    }
                }

                int opzIdx = -1;
                while (true)
                {
                    LogInline("Scegli Opzione: ");
                    string scelta = Console.ReadLine() ?? "";
                    if (int.TryParse(scelta, out int parsedOpz) && parsedOpz >= 0 && parsedOpz < response.Option.Count)
                    {
                        opzIdx = parsedOpz;
                        break;
                    }
                }

                var reqConScelta = new PlayRequest { GameID = _gameID, PlayedCard = cartaScelta };
                reqConScelta.TargetCard.AddRange(response.Option[opzIdx].Cards);

                Log("Invio scelta conflitto...");
                await _client.play_cardAsync(reqConScelta);
            }

            _myHand.RemoveAt(index);
            Log("Mossa completata. Aspetto gli altri...");
        }
        catch (Exception ex)
        {
            Log($"Errore Giocata: {ex.Message}");
        }
    }
}