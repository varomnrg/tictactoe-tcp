package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

const (
	X        = "X"
	O        = "O"
	GameName = "Tic Tac Toe"
)

type Game struct {
	gameId        string
	board         [3][3]string
	playerOneTurn bool
	conns         []net.Conn
	currentPlayer int
	Usernames     []string
	mutex         *sync.Mutex
}

type GameHistory struct {
	GameID   string       `json:"game_id"`
	Player1  string       `json:"player1"`
	Player2  string       `json:"player2"`
	Winner   string       `json:"winner"`
	Finished bool         `json:"finished"`
	Board    [3][3]string `json:"board"`
}

type Server struct {
	games   map[string]*Game
	history []GameHistory
	mutex   *sync.Mutex
}

func NewGame(conns []net.Conn, usernames []string, gameId string) *Game {
	return &Game{
		gameId:        gameId,
		board:         [3][3]string{},
		playerOneTurn: true,
		conns:         conns,
		currentPlayer: 0,
		Usernames:     usernames,
		mutex:         &sync.Mutex{},
	}
}

func (g *Game) PrintBoard() {
	var builder strings.Builder
	builder.WriteString("\n    1 2 3\n")
	builder.WriteString("  --------\n")
	for i, row := range g.board {
		builder.WriteString(fmt.Sprintf("%d | ", i+1))
		for _, cell := range row {
			if cell == "" {
				builder.WriteString("_ ")
			} else {
				builder.WriteString(cell + " ")
			}
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n")

	// Send the board to all players
	for _, conn := range g.conns {
		fmt.Fprint(conn, builder.String())
	}
}

func (g *Game) ValidateMove(row, col int) error {
	if row < 0 || row >= 3 || col < 0 || col >= 3 {
		return errors.New("row and column must be between 1 and 3")
	}
	if g.board[row][col] != "" {
		return errors.New("cell already occupied")
	}
	return nil
}

func (g *Game) SetMove(row, col int) {
	mark := X
	if !g.playerOneTurn {
		mark = O
	}
	g.board[row][col] = mark
}

func (g *Game) SwitchTurn() {
	g.currentPlayer = 1 - g.currentPlayer
	g.playerOneTurn = !g.playerOneTurn
}

func (g *Game) CheckWinner() (winner string, finished bool) {
	for i := 0; i < 3; i++ {
		if g.board[i][0] != "" && g.board[i][0] == g.board[i][1] && g.board[i][1] == g.board[i][2] {
			return g.board[i][0], true
		}

		if g.board[0][i] != "" && g.board[0][i] == g.board[1][i] && g.board[1][i] == g.board[2][i] {
			return g.board[0][i], true
		}
	}

	if g.board[0][0] != "" && g.board[0][0] == g.board[1][1] && g.board[1][1] == g.board[2][2] {
		return g.board[0][0], true
	}
	if g.board[0][2] != "" && g.board[0][2] == g.board[1][1] && g.board[1][1] == g.board[2][0] {
		return g.board[0][2], true
	}

	for _, row := range g.board {
		for _, cell := range row {
			if cell == "" {
				return "", false
			}
		}
	}

	return "Draw", true
}

func (g *Game) Start(server *Server) {
	for {
		g.PrintBoard()

		conn := g.conns[g.currentPlayer]
		fmt.Fprint(conn, "Your turn! Enter row and column (e.g., 2 3):\n")
		scanner := bufio.NewScanner(conn)
		if !scanner.Scan() {
			fmt.Println("Player disconnected!")
			for _, conn := range g.conns {
				fmt.Fprintln(conn, "Player disconnected! game over!")
			}
			return
		}

		move := scanner.Text()
		row, col, err := parseMove(move)
		if err != nil {
			fmt.Fprintf(conn, "Invalid move: %v\n", err)
			continue
		}

		if err := g.ValidateMove(row, col); err != nil {
			fmt.Fprintf(conn, "Invalid move: %v\n", err)
			continue
		}

		g.SetMove(row, col)

		winner, finished := g.CheckWinner()
		if finished {

			g.mutex.Lock()
			server.SetGameHistory(*g, winner, true)
			server.ExportHistory()
			g.mutex.Unlock()

			for _, conn := range g.conns {
				if winner == "Draw" {
					fmt.Fprintln(conn, "Game over! It's a draw!")
				} else {
					fmt.Fprintf(conn, "Game over! Player %s wins!\n", winner)
				}
			}
			break
		}

		g.SwitchTurn()
	}

	for _, conn := range g.conns {
		conn.Close()
	}

	fmt.Println("Game over! Game ID:", g.gameId)
}

func main() {
	fmt.Println("Starting Tic Tac Toe server...")

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Server started on port 8080.")

	server := &Server{
		games:   make(map[string]*Game),
		history: []GameHistory{},
		mutex:   &sync.Mutex{},
	}

	for {
		gameId := uuid.New().String()

		// Accept first player connection
		conn1, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection for Player 1:", err)
			continue
		}
		fmt.Fprintln(conn1, "Enter your username: ")

		scanner1 := bufio.NewScanner(conn1)
		if !scanner1.Scan() {
			fmt.Println("Player 1 disconnected before entering username.")
			conn1.Close()
			continue
		}
		username1 := scanner1.Text()

		// Accept second player connection
		conn2, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection for Player 2:", err)
			conn1.Close()
			continue
		}
		fmt.Fprintln(conn2, "Enter your username: ")

		scanner2 := bufio.NewScanner(conn2)
		if !scanner2.Scan() {
			fmt.Println("Player 2 disconnected before entering username.")
			conn1.Close()
			conn2.Close()
			continue
		}
		username2 := scanner2.Text()

		// Create and start the game
		game := NewGame([]net.Conn{conn1, conn2}, []string{username1, username2}, gameId)

		server.mutex.Lock()
		server.games[gameId] = game
		server.mutex.Unlock()

		go func() {
			fmt.Printf("Game with ID %s is running...\n", gameId)
			game.Start(server)
			fmt.Printf("Game with ID %s has ended.\n", gameId)
		}()
	}
}

func parseMove(input string) (int, int, error) {
	parts := strings.Split(input, " ")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected two numbers separated by space")
	}
	row, err1 := strconv.Atoi(parts[0])
	col, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("invalid numbers")
	}
	return row - 1, col - 1, nil
}

func (s *Server) SetGameHistory(game Game, winner string, finished bool) {
	s.history = append(s.history, GameHistory{
		GameID:   game.gameId,
		Player1:  game.Usernames[0],
		Player2:  game.Usernames[1],
		Winner:   winner,
		Finished: finished,
		Board:    game.board,
	})
}

func (s *Server) ExportHistory() {
	file, err := os.Create("history.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(s.history); err != nil {
		fmt.Println("Error encoding history:", err)
		return
	}
}
