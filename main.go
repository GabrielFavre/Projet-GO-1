package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type GameState struct {
	Board          [][]int
	CurrentPlayer  int
	Player1Name    string
	Player2Name    string
	Player1Color   string
	Player2Color   string
	Winner         int
	GameOver       bool
	Rows           int
	Cols           int
	Difficulty     string
	TurnCount      int
	GravityInverse bool
	BlockedCells   map[string]bool
	FinishHimMode  bool
}

var (
	game      *GameState
	gameMutex sync.Mutex
	templates *template.Template
)

func main() {
	// Initialize templates with custom functions
	funcMap := template.FuncMap{
		"iterate": func(count int) []int {
			var items []int
			for i := 0; i < count; i++ {
				items = append(items, i)
			}
			return items
		},
		"add": func(a, b int) int {
			return a + b
		},
		"eq": func(a, b int) bool {
			return a == b
		},
	}

	var err error
	templates, err = template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Error loading templates:", err)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/start", startGameHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/rematch", rematchHandler)

	log.Println("🎮 Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "home.html", nil)
}

func startGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	player1 := r.FormValue("player1")
	player2 := r.FormValue("player2")
	difficulty := r.FormValue("difficulty")
	player1Color := r.FormValue("player1color")
	player2Color := r.FormValue("player2color")

	if player1 == "" {
		player1 = "Joueur 1"
	}
	if player2 == "" {
		player2 = "Joueur 2"
	}
	if difficulty == "" {
		difficulty = "normal"
	}
	if player1Color == "" {
		player1Color = "#ef4444"
	}
	if player2Color == "" {
		player2Color = "#fbbf24"
	}

	gameMutex.Lock()
	game = initGame(player1, player2, difficulty, player1Color, player2Color)
	gameMutex.Unlock()

	templates.ExecuteTemplate(w, "game.html", game)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	colStr := r.FormValue("column")
	col, err := strconv.Atoi(colStr)
	if err != nil || col < 0 || col >= game.Cols {
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	gameMutex.Lock()
	defer gameMutex.Unlock()

	if game.GameOver {
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	// Place the piece
	row := placePiece(col)
	if row == -1 {
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	// Increment turn count
	game.TurnCount++

	// Check for gravity inversion every 5 turns
	if game.TurnCount%5 == 0 {
		game.GravityInverse = !game.GravityInverse
	}

	// Check for win
	if checkWin(row, col) {
		game.Winner = game.CurrentPlayer
		game.GameOver = true
		game.FinishHimMode = false
		templates.ExecuteTemplate(w, "result.html", game)
		return
	}

	// Check for draw
	if checkDraw() {
		game.Winner = 0
		game.GameOver = true
		game.FinishHimMode = false
		templates.ExecuteTemplate(w, "result.html", game)
		return
	}

	game.FinishHimMode = checkFinishHim()

	// Switch player
	if game.CurrentPlayer == 1 {
		game.CurrentPlayer = 2
	} else {
		game.CurrentPlayer = 1
	}

	templates.ExecuteTemplate(w, "game.html", game)
}

func rematchHandler(w http.ResponseWriter, r *http.Request) {
	gameMutex.Lock()
	if game != nil {
		game = initGame(game.Player1Name, game.Player2Name, game.Difficulty, game.Player1Color, game.Player2Color)
	}
	gameMutex.Unlock()

	templates.ExecuteTemplate(w, "game.html", game)
}

func initGame(player1, player2, difficulty, player1Color, player2Color string) *GameState {
	var rows, cols, blockedCount int

	switch difficulty {
	case "easy":
		rows, cols, blockedCount = 6, 7, 3
	case "hard":
		rows, cols, blockedCount = 7, 8, 7
	default: // normal
		rows, cols, blockedCount = 6, 9, 5
	}

	board := make([][]int, rows)
	for i := range board {
		board[i] = make([]int, cols)
	}

	// Add blocked cells
	blockedCells := make(map[string]bool)
	count := 0
	for count < blockedCount {
		row := count % rows
		col := (count * 3) % cols
		key := strconv.Itoa(row) + "," + strconv.Itoa(col)
		if !blockedCells[key] {
			board[row][col] = -1 // -1 represents blocked
			blockedCells[key] = true
			count++
		}
	}

	return &GameState{
		Board:          board,
		CurrentPlayer:  1,
		Player1Name:    player1,
		Player2Name:    player2,
		Player1Color:   player1Color,
		Player2Color:   player2Color,
		Winner:         0,
		GameOver:       false,
		Rows:           rows,
		Cols:           cols,
		Difficulty:     difficulty,
		TurnCount:      0,
		GravityInverse: false,
		BlockedCells:   blockedCells,
		FinishHimMode:  false,
	}
}

func placePiece(col int) int {
	if game.GravityInverse {
		// Inverse gravity: pieces fall upward
		for row := 0; row < game.Rows; row++ {
			if game.Board[row][col] == 0 {
				game.Board[row][col] = game.CurrentPlayer
				return row
			}
		}
	} else {
		// Normal gravity: pieces fall downward
		for row := game.Rows - 1; row >= 0; row-- {
			if game.Board[row][col] == 0 {
				game.Board[row][col] = game.CurrentPlayer
				return row
			}
		}
	}
	return -1 // Column is full
}

func checkWin(row, col int) bool {
	player := game.Board[row][col]

	// Check horizontal
	count := 1
	// Check left
	for c := col - 1; c >= 0 && game.Board[row][c] == player; c-- {
		count++
	}
	// Check right
	for c := col + 1; c < game.Cols && game.Board[row][c] == player; c++ {
		count++
	}
	if count >= 4 {
		return true
	}

	// Check vertical
	count = 1
	// Check up
	for r := row - 1; r >= 0 && game.Board[r][col] == player; r-- {
		count++
	}
	// Check down
	for r := row + 1; r < game.Rows && game.Board[r][col] == player; r++ {
		count++
	}
	if count >= 4 {
		return true
	}

	// Check diagonal (top-left to bottom-right)
	count = 1
	// Check up-left
	for r, c := row-1, col-1; r >= 0 && c >= 0 && game.Board[r][c] == player; r, c = r-1, c-1 {
		count++
	}
	// Check down-right
	for r, c := row+1, col+1; r < game.Rows && c < game.Cols && game.Board[r][c] == player; r, c = r+1, c+1 {
		count++
	}
	if count >= 4 {
		return true
	}

	// Check diagonal (top-right to bottom-left)
	count = 1
	// Check up-right
	for r, c := row-1, col+1; r >= 0 && c < game.Cols && game.Board[r][c] == player; r, c = r-1, c+1 {
		count++
	}
	// Check down-left
	for r, c := row+1, col-1; r < game.Rows && c >= 0 && game.Board[r][c] == player; r, c = r+1, c-1 {
		count++
	}
	if count >= 4 {
		return true
	}

	return false
}

func checkDraw() bool {
	for row := 0; row < game.Rows; row++ {
		for col := 0; col < game.Cols; col++ {
			if game.Board[row][col] == 0 {
				return false
			}
		}
	}
	return true
}

func checkFinishHim() bool {
	opponent := 3 - game.CurrentPlayer // Switch to opponent

	// Try each column to see if opponent can win
	for col := 0; col < game.Cols; col++ {
		// Simulate placing opponent's piece
		row := simulatePlacePiece(col, opponent)
		if row == -1 {
			continue
		}

		// Check if this would result in a win
		game.Board[row][col] = opponent
		wouldWin := checkWinForPosition(row, col, opponent)
		game.Board[row][col] = 0 // Undo simulation

		if wouldWin {
			return true
		}
	}
	return false
}

func simulatePlacePiece(col int, player int) int {
	if game.GravityInverse {
		for row := 0; row < game.Rows; row++ {
			if game.Board[row][col] == 0 {
				return row
			}
		}
	} else {
		for row := game.Rows - 1; row >= 0; row-- {
			if game.Board[row][col] == 0 {
				return row
			}
		}
	}
	return -1
}

func checkWinForPosition(row, col, player int) bool {
	// Check horizontal
	count := 1
	for c := col - 1; c >= 0 && game.Board[row][c] == player; c-- {
		count++
	}
	for c := col + 1; c < game.Cols && game.Board[row][c] == player; c++ {
		count++
	}
	if count >= 4 {
		return true
	}

	// Check vertical
	count = 1
	for r := row - 1; r >= 0 && game.Board[r][col] == player; r-- {
		count++
	}
	for r := row + 1; r < game.Rows && game.Board[r][col] == player; r++ {
		count++
	}
	if count >= 4 {
		return true
	}

	// Check diagonal (top-left to bottom-right)
	count = 1
	for r, c := row-1, col-1; r >= 0 && c >= 0 && game.Board[r][c] == player; r, c = r-1, c-1 {
		count++
	}
	for r, c := row+1, col+1; r < game.Rows && c < game.Cols && game.Board[r][c] == player; r, c = r+1, c+1 {
		count++
	}
	if count >= 4 {
		return true
	}

	// Check diagonal (top-right to bottom-left)
	count = 1
	for r, c := row-1, col+1; r >= 0 && c < game.Cols && game.Board[r][c] == player; r, c = r-1, c+1 {
		count++
	}
	for r, c := row+1, col-1; r < game.Rows && c >= 0 && game.Board[r][c] == player; r, c = r+1, c-1 {
		count++
	}
	if count >= 4 {
		return true
	}

	return false
}
