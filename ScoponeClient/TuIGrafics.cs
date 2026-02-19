using ScoponeScientifico;
using gameNamespace;
using System.Text.Json.Serialization.Metadata;
using System.Reflection.Metadata;

namespace TUINapespace
{
    public static class TUIRender
    {
        private const int TABLE_Y = 10;
        private const int HAND_Y = 20;
        private const int CARD_WIDTH = 9;
        private const int CARD_HEIGHT = 6;
        private const int CARD_SPACING = 2;

        public static void Draw(GameState state)
        {
            var snapshot = state.GetSnapshot();

            Console.BackgroundColor = ConsoleColor.DarkGreen;
            Console.Clear();

            DrawHeader(snapshot.SC1, snapshot.SC2);
            DrawTable(snapshot.table);
            DrawHand(snapshot.hand, snapshot.selIndex, snapshot.isMyTurn);
            DrawMessage(snapshot.msg);

            Console.ResetColor();
        }

        public static void DrawHeader(int myScore, int oppScore)
        {
            Console.ForegroundColor = ConsoleColor.White;
            Console.SetCursorPosition(2, 1);
            Console.Write($"NOI: {myScore}punti");

            Console.SetCursorPosition(Console.WindowWidth - 20, 1);
            Console.Write($"LORO: {oppScore} punti");

            Console.SetCursorPosition(Console.WindowWidth / 2 - 10, 2);
            Console.Write("Scopone Scientifico");

            Console.SetCursorPosition(0, 3);
            Console.Write(new string('=', Console.WindowWidth));
        }

        private static void DrawTable(List<card> cards)
        {
            Console.SetCursorPosition(2, TABLE_Y - 2);
            Console.ForegroundColor = ConsoleColor.Yellow;
            Console.Write("TAVOLO");

            if (cards.Count == 0)
            {
                Console.SetCursorPosition(10, TABLE_Y + 2);
                Console.ForegroundColor = ConsoleColor.Gray;
                Console.Write("(Tavolo Vuoto)");
                return;
            }

            int startX = (Console.WindowWidth - (cards.Count * (CARD_WIDTH + CARD_SPACING))) / 2;

            for (int i = 0; i < cards.Count; i++)
            {
                int x = startX + (i * (CARD_WIDTH + CARD_SPACING));
                DrawSingleCard(x, TABLE_Y, cards[i], false);
            }
        }

        private static void DrawHand(List<card> cards, int SelectedIndex, bool isMyTurn)
        {
            Console.SetCursorPosition(2, HAND_Y - 2);
            Console.ForegroundColor = ConsoleColor.Cyan;
            Console.Write("LE TUE CARTE:");

            if (cards.Count == 0) return;

            int startX = (Console.WindowWidth - (cards.Count * (CARD_WIDTH + CARD_SPACING))) / 2;

            for (int i = 0; i < cards.Count; i++)
            {
                int x = startX + (i * (CARD_WIDTH + CARD_SPACING));
                bool isSelected = (i == SelectedIndex);

                DrawSingleCard(x, HAND_Y, cards[i], isSelected);

                if (isSelected && isMyTurn)
                {
                    Console.SetCursorPosition(x + (CARD_WIDTH / 2), HAND_Y + CARD_HEIGHT);
                    Console.ForegroundColor = ConsoleColor.Yellow;
                    Console.Write("^");
                }
            }
        }

        private static void DrawMessage(string message)
        {
            int y = Console.WindowHeight - 2;
            Console.SetCursorPosition(2, y);
            Console.ForegroundColor = ConsoleColor.White;
            Console.BackgroundColor = ConsoleColor.Black;

            Console.Write(new string(' ', Console.WindowWidth - 4));
            Console.SetCursorPosition(2, y);
            Console.Write($">{message}");

            Console.BackgroundColor = ConsoleColor.DarkGreen;
        }

        private static void DrawSingleCard(int x, int y, card card, bool isSelected)
        {
            ConsoleColor cardColor = ConsoleColor.Black;
            switch (card.Suit)
            {
                case Suit.Bastoni:
                    break;
                case Suit.Coppe:
                    cardColor = ConsoleColor.Red;
                    break;
                case Suit.Denari:
                    cardColor = ConsoleColor.Yellow;
                    break;
                case Suit.Spade:
                    cardColor = ConsoleColor.Blue;
                    break;
                default:
                    break;
            }

            ConsoleColor bgColor = isSelected ? ConsoleColor.DarkYellow : ConsoleColor.White;

            Console.ForegroundColor = cardColor;
            Console.BackgroundColor = bgColor;

            string rank = GetRankSymbol(card.Rank);
            string suit = GetSuitSymbolcard(card.Suit);

            Console.SetCursorPosition(x, y);
            Console.Write("┌───────┐");

            Console.SetCursorPosition(x, y + 1);
            Console.Write($"│{rank,-2}     │");

            Console.SetCursorPosition(x, y + 2);
            Console.Write($"│   {suit}   │");

            Console.SetCursorPosition(x, y + 3);
            Console.Write("│       │");

            Console.SetCursorPosition(x, y + 4);
            Console.Write($"│     {rank,2}│");

            Console.SetCursorPosition(x, y + 5);
            Console.Write("└───────┘");

            Console.BackgroundColor = ConsoleColor.DarkGreen;
        }

        private static string GetSuitSymbolcard(Suit suit)
        {
            switch (suit)
            {
                case Suit.Denari: return "\U0001F7E1";
                case Suit.Bastoni: return "\U0001fab5";
                case Suit.Coppe: return "\U0001F3C6";
                case Suit.Spade: return "\u2694\ufe0f";
                default: return "?";
            }
        }
        

        private static string GetRankSymbol(Rank rank)
        {
            switch (rank)
            {
                case Rank.Asso:    return "A";
                case Rank.Due:     return "2";
                case Rank.Tre:     return "3";
                case Rank.Quattro: return "4";
                case Rank.Cinque:  return "5";
                case Rank.Sei:     return "6";
                case Rank.Sette:   return "7";
                
                case Rank.Fante:   return "8";  
                case Rank.Cavallo: return "9";
                case Rank.Re:      return "10";
                
                default:           return "?";
            }
        }
    }    
}