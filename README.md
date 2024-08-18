# War Card Game Simulation

This Go application simulates the card game War and provides statistics on game outcomes. It allows for customization of various parameters and can run multiple games in a single simulation.

Companion writeup on [michaelgruen.com/d/2024/what-am-i-in-for/](https://michaelgruen.com/d/2024/what-am-i-in-for/)

## Features

- Simulate multiple games of War
- Customizable game parameters
- Optional inclusion of jokers
- Detailed statistics on game outcomes
- CSV output of game results

## Prerequisites

- Go 1.15 or higher

## Installation

1. Clone this repository or download the source code.
2. Navigate to the project directory.

## Usage

Run the application using the `go run` command:

```
go run main.go [flags]
```

### Flags

- `-hand int`: Time to play a hand (in milliseconds, default 500  \[0.5 seconds\])
- `-shuffle int`: Time to shuffle (in milliseconds, default 15000 \[15 seconds\])
- `-jokers`: Include jokers in the deck (default false)
- `-seed int64`: Random seed (0 for current time, default 0)
- `-games int`: Number of games to play (default 100)
- `-maxtime int`: Maximum game time in milliseconds (default 3600000 \[1 hour == 60min * 60sec * 1000ms\])

### Example

To run 1000 games with jokers and a custom seed:

```
go run main.go -games 1000 -jokers -seed 12345
```

## Output

The application will print summary statistics to the console and generate a CSV file with detailed results for each game.

### Console Output

The console output includes:

- Total number of games played
- Statistics on tricks, wars, deep wars, shuffles, and game duration
- Percentage of finished games

### CSV Output

A CSV file named `war_results_[parameters].csv` will be generated in the same directory. It contains detailed results for each game, including:

- Game number
- Number of tricks
- Number of wars
- Number of deep wars
- Number of shuffles for each player
- Game duration
- Whether the game finished

## Understanding the Results

- **Tricks**: The number of rounds played in a game.
- **Wars**: Occurrences when both players play cards of the same rank.
- **Deep Wars**: Wars that result in another war.
- **Shuffles**: How many times each player had to shuffle their winnings pile.
- **Game Duration**: How long each game took (in simulated time).
- **Finished Games**: Games that didn't time out or exceed the maximum number of tricks.

## Customizing the Simulation

You can modify the source code to change core game mechanics or add new features. Key areas for customization include:

- `createDeck()`: Adjust the deck composition
- `handleWar()`: Modify war resolution mechanics
- `GameStats` struct: Add or remove tracked statistics

## Contributing

Contributions are welcome. Please feel free to submit a Pull Request with an accompanying explanation of changes/improvements.

### Limitations

- Double-counting drawTime during wars (in general, there's a 2x pause, so probably comes out in the wash.)
- This was coded with an LLM. I found one or two minor logical errors, but didn't effect game time too dramatically.

## License

This project is open source and available under the [MIT License](LICENSE).