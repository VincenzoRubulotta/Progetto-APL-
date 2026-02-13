
using ScoponeScientifico;

namespace gameNamespace
{
    public class GameState
    {
        private readonly object _stateLock = new object();

        public List<card> MyHand { get; private set; } = new List<card>();
        public List<card> Table { get; private set; } = new List<card>();

        public int MyTeamScore { get; private set; }
        public int OpponentTeamScore { get; private set; }

        public int SelectedIndex { get; private set; } = 0;
        public string StatusMessage { get; private set; } = "In attesa di connessione...";
        public bool MyTurn { get; private set; } = false;


        public void ResetNewRound(IEnumerable<card> newHand)
        {
            lock (_stateLock)
            {
                MyHand.Clear();

                if (newHand != null)
                {
                    MyHand.AddRange(newHand);
                }

                Table.Clear();

                SelectedIndex = 0;
                StatusMessage = $"Nuova smazzarta! Hai {MyHand.Count} carte.";
            }
        }

        public void SpostaCartaSulTavolo(card cartaGiocata, Actor chiHaGiocato)
        {
            lock (_stateLock)
            {
                if (chiHaGiocato == Actor.User)
                {
                    RimuoviDallaMano(cartaGiocata);
                }

                Table.Add(cartaGiocata);

                StatusMessage = $"{chiHaGiocato} gioca {DescriviCarta(cartaGiocata)}";
            }
        }


        public void FinalizzaPresa(card cartaGiocata, List<card> cartePrese, bool ToccaAMeDopo)
        {
            lock (_stateLock)
            {
                if (cartePrese != null && cartePrese.Count > 0)
                {
                    foreach (var carta in cartePrese)
                    {
                        RimuoviDalTavolo(carta);
                    }

                    RimuoviDalTavolo(cartaGiocata);

                    StatusMessage = $"Presa! ({cartePrese.Count} carte catturate)";
                }
                else
                {
                    StatusMessage = "Nessuna presa.";
                }

                if (ToccaAMeDopo)
                    StatusMessage += "Tocca a te!";

                MyTurn = ToccaAMeDopo;
            }
        }


        public void AggiornaPunteggi(int myScore, int oppScore)
        {
            lock (_stateLock)
            {
                MyTeamScore = myScore;
                OpponentTeamScore = oppScore;
            }
        }

        private void RimuoviDallaMano(card target)
        {
            var cardToRemove = MyHand.FirstOrDefault(c => c.Suit == target.Suit && c.Rank == target.Rank);

            if (cardToRemove != null)
            {
                MyHand.Remove(cardToRemove);

                if (SelectedIndex >= MyHand.Count && MyHand.Count > 0)
                {
                    SelectedIndex = MyHand.Count - 1;
                }
            }
        }

        private void RimuoviDalTavolo(card target)
        {
            var cardToRemove = Table.FirstOrDefault(c => c.Suit == target.Suit && c.Rank == target.Rank);
            if (cardToRemove != null)
            {
                Table.Remove(cardToRemove);
            }
        }

        private string DescriviCarta(card c)
        {
            if (c == null) return "???";
            return $"{c.Rank} di {c.Suit}";
        }

        public void MoveSelectionRight()
        {
            lock (_stateLock)
            {
                if (MyHand.Count > 0 && SelectedIndex < MyHand.Count - 1)
                {
                    SelectedIndex++;
                }
            }
        }

        public void MoveSelectionLeft()
        {
            lock (_stateLock)
            {
                if (SelectedIndex > 0)
                {
                    SelectedIndex--;
                }
            }
        }

        public card? GetSelectedCard()
        {
            lock (_stateLock)
            {
                if (MyHand.Count == 0 || SelectedIndex >= MyHand.Count)
                    return null;

                return MyHand[SelectedIndex];
            }
        }

        public (List<card> hand, List<card> table, int selIndex, string msg, bool isMyTurn, int SC1, int SC2) GetSnapshot()
        {
            lock (_stateLock)
            {
                return (
                    new List<card>(MyHand),
                    new List<card>(Table),
                    SelectedIndex,
                    StatusMessage,
                    MyTurn,
                    MyTeamScore,
                    OpponentTeamScore
                );
            }
        }

        public void setStatusMessage(string stringa)
        {
            StatusMessage += stringa;
        }
    }
}

