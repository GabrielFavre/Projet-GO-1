package main

import (
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Game struct {
	Rows               int
	Cols               int
	Board              [][]int
	PlayerTurn         int
	PlayerNames        [2]string
	TurnCount          int
	GravityDown        bool
	GameOver           bool
	Winner             int
	CurrentPlayerIndex int
	CurrentMaxCol      int
}

var (
	tmplIndex = template.Must(template.ParseFiles("templates/index.html"))
	tmplStart = template.Must(template.ParseFiles("templates/start.html"))
	tmplLevel = template.Must(template.ParseFiles("templates/level.html"))
	tmplWin   = template.Must(template.ParseFiles("templates/win.html"))
	tmplDraw  = template.Must(template.ParseFiles("templates/draw.html"))
	game      = &Game{}
)

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", startHandler)
	http.HandleFunc("/level", levelHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/rematch", rematchHandler)
	http.ListenAndServe(":8080", nil)
}

// Page de démarrage
func startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		game.PlayerNames[0] = r.FormValue("player1")
		game.PlayerNames[1] = r.FormValue("player2")
		http.Redirect(w, r, "/level", http.StatusSeeOther)
		return
	}
	tmplStart.Execute(w, nil)
}

// Choix du niveau
func levelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		level := r.FormValue("level")
		switch level {
		case "easy":
			initGame(6, 7, 3)
		case "normal":
			initGame(6, 9, 5)
		case "hard":
			initGame(7, 8, 7)
		default:
			initGame(6, 7, 3)
		}
		http.Redirect(w, r, "/play", http.StatusSeeOther)
		return
	}
	tmplLevel.Execute(w, nil)
}

// Initialisation du jeu
func initGame(rows, cols, blocks int) {
	game.Rows = rows
	game.Cols = cols
	game.PlayerTurn = 1
	game.TurnCount = 0
	game.GravityDown = true
	game.GameOver = false
	game.Winner = 0
	game.CurrentPlayerIndex = 0
	game.CurrentMaxCol = cols - 1

	game.Board = make([][]int, rows)
	for i := range game.Board {
		game.Board[i] = make([]int, cols)
	}

	for b := 0; b < blocks; b++ {
		randRow := rand.Intn(rows)
		randCol := rand.Intn(cols)
		if game.Board[randRow][randCol] == 0 {
			game.Board[randRow][randCol] = rand.Intn(2) + 1
		}
	}
}

// Page du jeu
func playHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		colStr := r.FormValue("column")
		col, err := strconv.Atoi(colStr)
		if err == nil && col >= 0 && col < game.Cols {
			placeToken(col)
			game.TurnCount++
			if game.TurnCount%5 == 0 {
				game.GravityDown = !game.GravityDown
			}
		}
	}

	// Vérifier victoire
	for r := 0; r < game.Rows; r++ {
		for c := 0; c < game.Cols; c++ {
			if game.Board[r][c] != 0 && checkWin(r, c, game.Board[r][c]) {
				game.GameOver = true
				game.Winner = game.Board[r][c]
			}
		}
	}
	// Vérifier match nul
	game.GameOver = game.GameOver || isDraw()

	game.CurrentPlayerIndex = game.PlayerTurn - 1

	if game.GameOver {
		if game.Winner != 0 {
			data := struct {
				Game
				WinnerIndex int
			}{
				Game:        *game,
				WinnerIndex: game.Winner - 1,
			}
			tmplWin.Execute(w, data)
			return
		} else {
			tmplDraw.Execute(w, game)
			return
		}
	}

	tmplIndex.Execute(w, game)
}

// Rematch
func rematchHandler(w http.ResponseWriter, r *http.Request) {
	initGame(game.Rows, game.Cols, 0)
	http.Redirect(w, r, "/play", http.StatusSeeOther)
}

// Placer jeton selon gravité
func placeToken(col int) {
	if game.GravityDown {
		for i := game.Rows - 1; i >= 0; i-- {
			if game.Board[i][col] == 0 {
				game.Board[i][col] = game.PlayerTurn
				game.PlayerTurn = 3 - game.PlayerTurn
				return
			}
		}
	} else {
		for i := 0; i < game.Rows; i++ {
			if game.Board[i][col] == 0 {
				game.Board[i][col] = game.PlayerTurn
				game.PlayerTurn = 3 - game.PlayerTurn
				return
			}
		}
	}
}

// Vérifier victoire
func checkWin(row, col, player int) bool {
	dirs := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}
	for _, d := range dirs {
		count := 1
		for i := 1; i < 4; i++ {
			r, c := row+d[0]*i, col+d[1]*i
			if r < 0 || r >= game.Rows || c < 0 || c >= game.Cols || game.Board[r][c] != player {
				break
			}
			count++
		}
		for i := 1; i < 4; i++ {
			r, c := row-d[0]*i, col-d[1]*i
			if r < 0 || r >= game.Rows || c < 0 || c >= game.Cols || game.Board[r][c] != player {
				break
			}
			count++
		}
		if count >= 4 {
			return true
		}
	}
	return false
}

// Vérifier match nul
func isDraw() bool {
	for r := 0; r < game.Rows; r++ {
		for c := 0; c < game.Cols; c++ {
			if game.Board[r][c] == 0 {
				return false
			}
		}
	}
	return true
}
