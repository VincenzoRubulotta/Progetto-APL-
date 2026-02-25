
using Grpc.Net.Client;
using ScoponeScientifico;
using gameNamespace;
using TUINapespace;
using System.Text;
using System.Diagnostics;
using System.IO.Enumeration;

class Program
{
    static GameState _gameState = new GameState();
    static go_backend.go_backendClient _client = null!;
    static int _GameID;

    static string _myId = "";
    static bool _gameRunning = true;

    static async Task Main(string[] args)
    {
        Console.OutputEncoding = Encoding.UTF8;
        Console.Title = "Scopone Scientifico - Client TUI";

        var channel = GrpcChannel.ForAddress("http://localhost:50051");
        _client = new go_backend.go_backendClient(channel);

        Console.Clear();
        Console.WriteLine("--- BENVENUTO ALLO SCOPONE SCIENTIFICO ---");
        Console.WriteLine("--- Legenda Comandi ---");
        Console.WriteLine("--- Muoviti con le frecce direzionali per scorrere le carte in mano ---");
        Console.WriteLine("--- Premi INVIO per giocare la carta durante il tuo turno ---");
        Console.WriteLine("--- Premi S per guardare le statistiche ---");
        Console.WriteLine("--- Premi P per guardare le statistiche parziali della partita ---");
        Console.WriteLine("--- Premi ESC per uscire dal gioco ---");

        Console.Write("\n\nInserisci il tuo Nome Utente: ");
        string userName = Console.ReadLine() ?? "Hero";

        int maxPoints = 0;
        while (maxPoints != 11 && maxPoints != 21)
        {
            Console.Write("A quanto si gioca? (11 o 21): ");
            string input = Console.ReadLine() ?? "";
            int.TryParse(input, out maxPoints);
        }

        Console.WriteLine("Connessione in corso...");

        try
        {
            var startReq = new game_settings { UserName = userName, MaxPoints = maxPoints };
            var initData = await _client.start_gameAsync(startReq);

            _GameID = initData.GameID;
            _myId = userName;

            _gameState.ResetNewRound(initData.UserHand);

            Console.CursorVisible = false;
            TUIRender.Draw(_gameState);

            _ = Task.Run(() => AscoltaServer());

            await GestisciInputUtente();


        }
        catch (Exception ex)
        {
            Console.Clear();
            Console.WriteLine($"ERRORE FATALE DI CONNESSIONE: {ex.Message}");
            Console.WriteLine("Premi un tasto per uscire...");
            Console.ReadKey();
        }
    }

    static async Task GestisciInputUtente()
    {
        while (_gameRunning)
        {
            var keyInfo = Console.ReadKey(intercept: true);

            if (keyInfo.Key == ConsoleKey.RightArrow)
            {
                _gameState.MoveSelectionRight();
                TUIRender.Draw(_gameState);
            }
            else if (keyInfo.Key == ConsoleKey.LeftArrow)
            {
                _gameState.MoveSelectionLeft();
                TUIRender.Draw(_gameState);
            }
            else if (keyInfo.Key == ConsoleKey.Enter)
            {
                if (_gameState.MyTurn)
                {
                    var cartaSelezionata = _gameState.GetSelectedCard();
                    if (cartaSelezionata != null)
                    {
                        await EseguiGiocata(cartaSelezionata);
                    }
                }
            }
            else if (keyInfo.Key == ConsoleKey.Escape)
            {
                _gameRunning = false;
            }
            else if (keyInfo.Key == ConsoleKey.S)
            {
                MostraStatistiche();
            }
            else if (keyInfo.Key == ConsoleKey.P)
            {
                MostraStatistichePartitaCorrente();
            }
        }
    }

    static async Task EseguiGiocata(card CartaScelta)
    {
        try
        {
            _gameState.setStatusMessage($"Invio {CartaScelta.Rank} {CartaScelta.Suit}...");
            TUIRender.Draw(_gameState);

            var req = new PlayRequest
            {
                GameID = _GameID,
                PlayedCard = CartaScelta
            };

            var response = await _client.play_cardAsync(req);

            if (response.ConflictResolutionNeeded)
            {
                StringBuilder sb = new StringBuilder("AMBIGUITÀ! Scegli numero:");
                for (int i = 0; i < response.Option.Count; i++)
                {
                    sb.Append($"[{i}]");
                    foreach (var c in response.Option[i].Cards) sb.Append($"{c.Rank.ToString().Substring(0, 1)}{c.Suit.ToString().Substring(0, 1)}");
                    sb.Append("|");
                }

                _gameState.setStatusMessage(sb.ToString());
                TUIRender.Draw(_gameState);


                int opzIdx = -1;
                while (true)
                {
                    var k = Console.ReadKey(intercept: true);

                    if (char.IsDigit(k.KeyChar))
                    {
                        int val = int.Parse(k.KeyChar.ToString());
                        if (val >= 0 && val < response.Option.Count)
                        {
                            opzIdx = val;
                            break;
                        }
                    }
                }

                var reqConScelta = new PlayRequest
                {
                    GameID = _GameID,
                    PlayedCard = CartaScelta,

                };

                reqConScelta.TargetCard.AddRange(response.Option[opzIdx].Cards);
                _gameState.setStatusMessage($"Hai scelto opzione {opzIdx}. Invio...");
                TUIRender.Draw(_gameState);

                await _client.play_cardAsync(reqConScelta);
            }
        }
        catch (Exception ex)
        {
            _gameState.setStatusMessage($"ERRORE GIOCATA: {ex.Message}");
            TUIRender.Draw(_gameState);
            await Task.Delay(2000);
        }

    }
    static async Task<bool> GestisciFineSmazzata()
    {
        try
        {
            var request = new observe_request { GameID = _GameID };
            var scoreUpdate = await _client.calcola_punteggioAsync(request);


            _gameState.AggiornaPunteggi((int)scoreUpdate.UserSqudScore, (int)scoreUpdate.CpuSquadScore);


            if (scoreUpdate.IsGameOver)
            {

                System.Threading.Thread.Sleep(1500);
                MostraStatistiche();
                System.Threading.Thread.Sleep(1500);
                MostraStatistichePartitaCorrente();

                Console.Clear();
                Console.WriteLine("\n\n\n");
                Console.WriteLine("=================================");
                Console.WriteLine("      PARTITA TERMINATA          ");
                Console.WriteLine("=================================");
                Console.WriteLine($"NOI:  {scoreUpdate.UserSqudScore}");
                Console.WriteLine($"LORO: {scoreUpdate.CpuSquadScore}");

                if (scoreUpdate.UserSqudScore > scoreUpdate.CpuSquadScore)
                    Console.WriteLine("\n     HAI VINTO!      ");
                else
                    Console.WriteLine("\n     HAI PERSO...      ");

                Console.WriteLine("\nPremi ESC per uscire.");
                return false;
            }

            _gameState.ResetNewRound(scoreUpdate.UserHand);

            TUIRender.Draw(_gameState);

            return true;
        }
        catch (Exception ex)
        {
            _gameState.setStatusMessage($"Errore Punteggi: {ex.Message}");
            TUIRender.Draw(_gameState);
            return false;
        }
    }
    static async Task AscoltaServer()
    {
        try
        {
            var request = new observe_request { GameID = _GameID };
            using var stream = _client.observe_turn(request);

            while (await stream.ResponseStream.MoveNext(CancellationToken.None))
            {
                var update = stream.ResponseStream.Current;



                if (update.IsMatchOver)
                {
                    _gameState.setStatusMessage("Fine smazzata! Calcolo Punteggi...");
                    TUIRender.Draw(_gameState);
                    bool continua = await GestisciFineSmazzata();

                    if (!continua)
                    {
                        _gameRunning = false;
                        break;
                    }
                    continue;
                }

                bool toccaAMeDopo = (update.NextPlayerID == Actor.User);

                if (update.PlayedCard != null)
                {
                    _gameState.SpostaCartaSulTavolo(update.PlayedCard, update.Actor);
                    TUIRender.Draw(_gameState);

                    await Task.Delay(2500);

                    var listaPrese = update.CartePrese.ToList();

                    _gameState.FinalizzaPresa(update.PlayedCard, listaPrese, toccaAMeDopo);

                    if (update.Scopa)
                    {
                        _gameState.setStatusMessage("!!! SCOPA !!! " + _gameState.StatusMessage);
                    }


                }
                else
                {
                    _gameState.AggiornaTurno(toccaAMeDopo);
                }

                TUIRender.Draw(_gameState);
            }
        }
        catch (Exception ex)
        {
            _gameState.setStatusMessage($"ERRORE STREAM: {ex.Message}");
            TUIRender.Draw(_gameState);
        }
    }


    static void MostraStatistiche()
    {
        string url = "http://localhost:5001/stats";

        _gameState.setStatusMessage("Apertura statistiche nel browser");
        TUIRender.Draw(_gameState);

        try
        {
            Process.Start(new ProcessStartInfo
            {
                FileName = url,
                UseShellExecute = true
            });
        }
        catch (Exception)
        {
            _gameState.setStatusMessage($"Impossibile aprire in automatico. Vai su {url}");
            TUIRender.Draw(_gameState);
        }
    }

    static void MostraStatistichePartitaCorrente()
    {
        string url = $"http://localhost:5001/stats/match/{_GameID}";

        _gameState.setStatusMessage($"Apertura statistiche del match {_GameID}...");
        TUIRender.Draw(_gameState);

        try 
        {
            Process.Start(new ProcessStartInfo { FileName = url, UseShellExecute = true });
        }
        catch (Exception)
        {
            _gameState.setStatusMessage($"Impossibile aprire in automatico. Vai su {url}");
            TUIRender.Draw(_gameState);
        }
    }
    
}