// Copyright (c) 2024 michaelgruen.com (with help from Claude 3.5 Sonnet)
// This software is licensed under the MIT License.

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type Card struct {
	Rank int
}

type Player struct {
	DrawPile     []Card
	WinningsPile []Card
}

type GameStats struct {
    GameNumber   int
    Tricks       int
    Wars         int
    DeepWars     int
    TotalWarDepth int
    ShufflesA    int
    ShufflesB    int
    GameDuration time.Duration
    Finished     bool
}

func main() {
	handTime, shuffleTime, includeJokers, seed, gamesToPlay, maxGameTime := parseArgs()

	if seed != 0 {
		rand.Seed(int64(seed))
	} else {
		rand.Seed(time.Now().UnixNano())
	}

	deck := createDeck(includeJokers)
	fmt.Printf("Deck size: %d\n", len(deck))

	fmt.Printf("Starting simulation of %d games...\n", gamesToPlay)
	startTime := time.Now()
	stats := runSimulations(gamesToPlay, handTime, shuffleTime, includeJokers, maxGameTime)
	fmt.Printf("Simulation completed in %v\n", time.Since(startTime))

	writeResultsToFile(stats, handTime, shuffleTime, includeJokers, seed, gamesToPlay, maxGameTime)
	printSummaryStatistics(stats)
}

func parseArgs() (int, int, bool, int64, int, int) {
	handTime := flag.Int("hand", 500, "Time to play a hand (in milliseconds)")
	shuffleTime := flag.Int("shuffle", 15000, "Time to shuffle (in milliseconds)")
	includeJokers := flag.Bool("jokers", false, "Include jokers in the deck")
	seed := flag.Int64("seed", 0, "Random seed (0 for current time)")
	gamesToPlay := flag.Int("games", 100, "Number of games to play")
	maxGameTime := flag.Int("maxtime", 3600000, "Maximum game time in milliseconds (default 1 hour)")

	flag.Parse()

	return *handTime, *shuffleTime, *includeJokers, *seed, *gamesToPlay, *maxGameTime
}

func runSimulations(gamesToPlay, handTime, shuffleTime int, includeJokers bool, maxGameTime int) []GameStats {
	stats := make([]GameStats, gamesToPlay)
	for i := 0; i < gamesToPlay; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic occurred in game %d: %v\n", i+1, r)
					stats[i] = GameStats{GameNumber: i + 1, Tricks: -1, Finished: false} // Use -1 to indicate an error
				}
			}()
			stats[i] = playGame(handTime, shuffleTime, includeJokers, maxGameTime)
			stats[i].GameNumber = i + 1
		}()
	}
	return stats
}

func playGame(handTime, shuffleTime int, includeJokers bool, maxGameTime int) GameStats {
    deck := createDeck(includeJokers)
    shuffleDeck(deck)

    playerA := Player{DrawPile: deck[:len(deck)/2]}
    playerB := Player{DrawPile: deck[len(deck)/2:]}

    stats := GameStats{}
    totalTime := 0 // in milliseconds
    maxTricks := 100000 // Safety mechanism to prevent infinite games

    for len(playerA.DrawPile) + len(playerA.WinningsPile) > 0 && 
        len(playerB.DrawPile) + len(playerB.WinningsPile) > 0 && 
        stats.Tricks < maxTricks && totalTime < maxGameTime {
        
        stats.Tricks++
        totalTime += handTime

        cardA, shuffledA := drawCard(&playerA)
        cardB, shuffledB := drawCard(&playerB)
        stats.ShufflesA += shuffledA
        stats.ShufflesB += shuffledB
        totalTime += (shuffledA + shuffledB) * shuffleTime

        if cardA.Rank == cardB.Rank {
            warPile := []Card{cardA, cardB}
            winner := handleWar(&playerA, &playerB, warPile, &stats, &totalTime, handTime, shuffleTime, maxGameTime, 1) // Added depth parameter (1)
            if winner == 1 {
                playerA.WinningsPile = append(playerA.WinningsPile, warPile...)
            } else if winner == 2 {
                playerB.WinningsPile = append(playerB.WinningsPile, warPile...)
            }
        } else if cardA.Rank > cardB.Rank {
            playerA.WinningsPile = append(playerA.WinningsPile, cardA, cardB)
        } else {
            playerB.WinningsPile = append(playerB.WinningsPile, cardA, cardB)
        }
    }

    stats.Finished = len(playerA.DrawPile) + len(playerA.WinningsPile) == 0 || 
        len(playerB.DrawPile) + len(playerB.WinningsPile) == 0
    if !stats.Finished {
        fmt.Printf("Game exceeded %d tricks or %d ms. Possible infinite game.\n", maxTricks, maxGameTime)
        fmt.Printf("Final state: PlayerA: %d cards, PlayerB: %d cards\n", 
            len(playerA.DrawPile) + len(playerA.WinningsPile), 
            len(playerB.DrawPile) + len(playerB.WinningsPile))
    }

    stats.GameDuration = time.Duration(totalTime) * time.Millisecond
    return stats
}

func drawCard(player *Player) (Card, int) {
	if len(player.DrawPile) == 0 {
		if len(player.WinningsPile) == 0 {
			return Card{}, 0
		}
		player.DrawPile = player.WinningsPile
		player.WinningsPile = []Card{}
		shuffleDeck(player.DrawPile)
		return player.DrawPile[0], 1
	}
	card := player.DrawPile[0]
	player.DrawPile = player.DrawPile[1:]
	return card, 0
}

func handleWar(playerA, playerB *Player, warPile []Card, stats *GameStats, totalTime *int, handTime, shuffleTime, maxGameTime, depth int) int {
    stats.Wars++
    stats.TotalWarDepth += depth
    *totalTime += handTime

    if *totalTime >= maxGameTime {
        return 0 // No winner due to time out
    }

    // Function to draw cards and update stats
    drawCardsForWar := func(player *Player, shuffles *int, maxCards int) []Card {
        cards := make([]Card, 0, maxCards)
        for i := 0; i < maxCards; i++ {
            card, shuffled := drawCard(player)
            if shuffled > 0 {
                *shuffles++
                *totalTime += shuffleTime
            }
            if (card == Card{}) {
                break // No more cards available
            }
            cards = append(cards, card)
        }
        return cards
    }

    // Draw cards for both players (up to 4 cards each)
    cardsA := drawCardsForWar(playerA, &stats.ShufflesA, 4)
    cardsB := drawCardsForWar(playerB, &stats.ShufflesB, 4)

    // Check if either player has no cards left
    if len(cardsA) == 0 {
        return 2 // Player B wins
    }
    if len(cardsB) == 0 {
        return 1 // Player A wins
    }

    // Add face-down cards to the war pile
    warPile = append(warPile, cardsA[:len(cardsA)-1]...)
    warPile = append(warPile, cardsB[:len(cardsB)-1]...)

    // Compare the last card from each player
    cardA, cardB := cardsA[len(cardsA)-1], cardsB[len(cardsB)-1]
    warPile = append(warPile, cardA, cardB)

    if cardA.Rank == cardB.Rank {
        stats.DeepWars++
        return handleWar(playerA, playerB, warPile, stats, totalTime, handTime, shuffleTime, maxGameTime, depth+1)
    } else if cardA.Rank > cardB.Rank {
        return 1 // Player A wins
    } else {
        return 2 // Player B wins
    }
}


func createDeck(includeJokers bool) []Card {
	deck := make([]Card, 0, 52)
	for rank := 2; rank <= 14; rank++ { // 11=Jack, 12=Queen, 13=King, 14=Ace
		for suit := 0; suit < 4; suit++ {
			deck = append(deck, Card{Rank: rank})
		}
	}
	if includeJokers {
		deck = append(deck, Card{Rank: 15}, Card{Rank: 15}) // Two jokers
	}
	return deck
}

func shuffleDeck(deck []Card) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

func writeResultsToFile(stats []GameStats, handTime, shuffleTime int, includeJokers bool, seed int64, gamesToPlay, maxGameTime int) {
	filename := fmt.Sprintf("war_results_hand%d_shuffle%d_jokers%v_seed%d_games%d_maxtime%d.csv", handTime, shuffleTime, includeJokers, seed, gamesToPlay, maxGameTime)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Game Number", "Tricks", "Wars", "Deep Wars", "Shuffles A", "Shuffles B", "Game Duration (ms)", "Finished"}
	writer.Write(headers)

	for _, game := range stats {
		row := []string{
			strconv.Itoa(game.GameNumber),
			strconv.Itoa(game.Tricks),
			strconv.Itoa(game.Wars),
			strconv.Itoa(game.DeepWars),
			strconv.Itoa(game.ShufflesA),
			strconv.Itoa(game.ShufflesB),
			strconv.FormatInt(game.GameDuration.Milliseconds(), 10),
			strconv.FormatBool(game.Finished),
		}
		writer.Write(row)
	}
}

func printSummaryStatistics(stats []GameStats) {
    fmt.Printf("Total number of games played: %d\n", len(stats))

    tricks := make([]float64, len(stats))
    wars := make([]float64, len(stats))
    deepWars := make([]float64, len(stats))
    avgWarDepths := make([]float64, len(stats))
    shufflesA := make([]float64, len(stats))
    shufflesB := make([]float64, len(stats))
    gameTimes := make([]float64, len(stats))
    finishedGames := 0

    for i, game := range stats {
        tricks[i] = float64(game.Tricks)
        wars[i] = float64(game.Wars)
        deepWars[i] = float64(game.DeepWars)
        if game.Wars > 0 {
            avgWarDepths[i] = float64(game.TotalWarDepth) / float64(game.Wars)
        }
        shufflesA[i] = float64(game.ShufflesA)
        shufflesB[i] = float64(game.ShufflesB)
        gameTimes[i] = float64(game.GameDuration.Minutes())
        if game.Finished {
            finishedGames++
        }
    }

    printStatistic("Tricks", tricks)
    printStatistic("Wars", wars)
    printStatistic("Deep Wars", deepWars)
    printStatistic("Average War Depth", avgWarDepths)
    printStatistic("Shuffles A", shufflesA)
    printStatistic("Shuffles B", shufflesB)
    
    avgGameTime := average(gameTimes)
    minGameTime, maxGameTime := minMax(gameTimes)
    stdDevGameTime := standardDeviation(gameTimes, avgGameTime)
    
    fmt.Printf("Game Time (minutes): Avg %.2f (Min: %.2f, Max: %.2f, StdDev: %.2f)\n", 
               avgGameTime, minGameTime, maxGameTime, stdDevGameTime)
    fmt.Printf("Finished games: %d (%.2f%%)\n", finishedGames, float64(finishedGames)/float64(len(stats))*100)
}

func printStatistic(name string, data []float64) {
    avg := average(data)
    min, max := minMax(data)
    stdDev := standardDeviation(data, avg)

    fmt.Printf("%s: Avg %.2f (Min: %.0f, Max: %.0f, StdDev: %.2f)\n", name, avg, min, max, stdDev)
}

func average(data []float64) float64 {
    sum := 0.0
    for _, v := range data {
        sum += v
    }
    return sum / float64(len(data))
}

func minMax(data []float64) (float64, float64) {
    min, max := data[0], data[0]
    for _, v := range data[1:] {
        if v < min {
            min = v
        }
        if v > max {
            max = v
        }
    }
    return min, max
}

func standardDeviation(data []float64, mean float64) float64 {
    sum := 0.0
    for _, v := range data {
        sum += math.Pow(v-mean, 2)
    }
    variance := sum / float64(len(data))
    return math.Sqrt(variance)
}