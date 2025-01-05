# TicTacToe Multiplayer

A multiplayer Tic-Tac-Toe game implemented in Go using TCP sockets. Players can connect to the server, play against each other in real-time, and game history is automatically saved.

## Features

- Real-time multiplayer gameplay using TCP sockets
- Username system for player identification
- Interactive game board display
- Move validation and error handling
- Game history tracking with JSON export
- Concurrent game handling (multiple games can run simultaneously)
- Clean disconnection handling

## Prerequisites

- Go 1.19 or higher
- github.com/google/uuid package

## Installation

1. Clone the repository:
```bash
git clone https://github.com/varomnrg/tictactoe-tcp.git
cd tictactoe-tcp
```

## Usage

1. Start the server:
```bash
go run server/main.go
```

2. In separate terminals, start two clients:
```bash
go run client/main.go
```

3. Follow the prompts to play

## Author
