package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
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
	GameMode       string
}

var (
	game      *GameState
	gameMutex sync.Mutex
	templates *template.Template
)

func main() {
	rand.Seed(time.Now().UnixNano())

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
		"eq": func(a, b interface{}) bool {
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
	http.HandleFunc("/ai-move", aiMoveHandler)

	log.Println("Server starting on http://localhost:8080")
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
	gameMode := r.FormValue("gamemode")

	if player1 == "" {
		player1 = "Joueur 1"
	}
	if player2 == "" {
		if gameMode == "ai" {
			player2 = "IA"
		} else {
			player2 = "Joueur 2"
		}
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
	if gameMode == "" {
		gameMode = "pvp"
	}

	gameMutex.Lock()
	game = initGame(player1, player2, difficulty, player1Color, player2Color, gameMode)
	gameMutex.Unlock()

	templates.ExecuteTemplate(w, "game.html", game)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	gameMutex.Lock()
	defer gameMutex.Unlock()

	if game == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	colStr := r.FormValue("column")
	col, err := strconv.Atoi(colStr)
	if err != nil || col < 0 || col >= game.Cols {
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	if game.GameOver {
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	row := placePiece(col)
	if row == -1 {
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	game.TurnCount++

	if game.TurnCount%5 == 0 {
		game.GravityInverse = !game.GravityInverse
	}

	if checkWin(row, col) {
		game.Winner = game.CurrentPlayer
		game.GameOver = true
		game.FinishHimMode = true
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	if checkDraw() {
		game.Winner = 0
		game.GameOver = true
		game.FinishHimMode = false
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	if game.CurrentPlayer == 1 {
		game.CurrentPlayer = 2
	} else {
		game.CurrentPlayer = 1
	}

	templates.ExecuteTemplate(w, "game.html", game)
}

func aiMoveHandler(w http.ResponseWriter, r *http.Request) {
	gameMutex.Lock()
	defer gameMutex.Unlock()

	if game == nil || game.GameOver || game.GameMode != "ai" || game.CurrentPlayer != 2 {
		if game != nil {
			templates.ExecuteTemplate(w, "game.html", game)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		return
	}

	col := getAIMove()
	if col == -1 {
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	row := placePiece(col)
	if row == -1 {
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	game.TurnCount++

	if game.TurnCount%5 == 0 {
		game.GravityInverse = !game.GravityInverse
	}

	if checkWin(row, col) {
		game.Winner = game.CurrentPlayer
		game.GameOver = true
		game.FinishHimMode = true
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	if checkDraw() {
		game.Winner = 0
		game.GameOver = true
		game.FinishHimMode = false
		templates.ExecuteTemplate(w, "game.html", game)
		return
	}

	game.CurrentPlayer = 1

	templates.ExecuteTemplate(w, "game.html", game)
}

func rematchHandler(w http.ResponseWriter, r *http.Request) {
	gameMutex.Lock()
	if game != nil {
		game = initGame(game.Player1Name, game.Player2Name, game.Difficulty, game.Player1Color, game.Player2Color, game.GameMode)
	}
	gameMutex.Unlock()

	if game != nil {
		templates.ExecuteTemplate(w, "game.html", game)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func initGame(player1, player2, difficulty, player1Color, player2Color, gameMode string) *GameState {
	var rows, cols, blockedCount int

	switch difficulty {
	case "easy":
		rows, cols, blockedCount = 6, 7, 3
	case "hard":
		rows, cols, blockedCount = 7, 8, 7
	default:
		rows, cols, blockedCount = 6, 9, 5
	}

	board := make([][]int, rows)
	for i := range board {
		board[i] = make([]int, cols)
	}

	blockedCells := make(map[string]bool)
	count := 0
	for count < blockedCount {
		row := count % rows
		col := (count * 3) % cols
		key := strconv.Itoa(row) + "," + strconv.Itoa(col)
		if !blockedCells[key] {
			board[row][col] = -1
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
		GameMode:       gameMode,
	}
}

func placePiece(col int) int {
	if game.GravityInverse {
		for row := 0; row < game.Rows; row++ {
			if game.Board[row][col] == 0 {
				game.Board[row][col] = game.CurrentPlayer
				return row
			}
		}
	} else {
		for row := game.Rows - 1; row >= 0; row-- {
			if game.Board[row][col] == 0 {
				game.Board[row][col] = game.CurrentPlayer
				return row
			}
		}
	}
	return -1
}

func checkWin(row, col int) bool {
	player := game.Board[row][col]

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

func getAIMove() int {
	for col := 0; col < game.Cols; col++ {
		row := simulatePlacePiece(col, 2)
		if row == -1 {
			continue
		}
		game.Board[row][col] = 2
		if checkWinForPosition(row, col, 2) {
			game.Board[row][col] = 0
			return col
		}
		game.Board[row][col] = 0
	}

	for col := 0; col < game.Cols; col++ {
		row := simulatePlacePiece(col, 1)
		if row == -1 {
			continue
		}
		game.Board[row][col] = 1
		if checkWinForPosition(row, col, 1) {
			game.Board[row][col] = 0
			return col
		}
		game.Board[row][col] = 0
	}

	centerCols := []int{game.Cols / 2, game.Cols/2 - 1, game.Cols/2 + 1}
	for _, col := range centerCols {
		if col >= 0 && col < game.Cols {
			row := simulatePlacePiece(col, 2)
			if row != -1 {
				return col
			}
		}
	}

	validCols := []int{}
	for col := 0; col < game.Cols; col++ {
		if simulatePlacePiece(col, 2) != -1 {
			validCols = append(validCols, col)
		}
	}

	if len(validCols) > 0 {
		return validCols[rand.Intn(len(validCols))]
	}

	return -1
}
dddd