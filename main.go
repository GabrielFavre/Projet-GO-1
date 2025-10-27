package main

import (
	"encoding/json"
	"html/template"
	"math/rand"
	"net/http"
	"time"
)

type Game struct {
	Rows       int
	Cols       int
	Board      [][]int
	PlayerTurn int
	GameOver   bool
	Winner     int
}

type TemplateData struct {
	Game      *Game
	BoardJSON template.JS
}

var (
	game      *Game
	tmplIndex = template.Must(template.ParseFiles("templates/index.html"))
	tmplStart = template.Must(template.ParseFiles("templates/start.html"))
	tmplLevel = template.Must(template.ParseFiles("templates/level.html"))
	tmplWin   = template.Must(template.ParseFiles("templates/win.html"))
	tmplDraw  = template.Must(template.ParseFiles("templates/draw.html"))
)

func main() {
	rand.Seed(time.Now().UnixNano())
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", startHandler)
	http.HandleFunc("/level", levelHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/move", moveHandler)
	http.HandleFunc("/rematch", rematchHandler)
	initGame(6, 7)
	http.ListenAndServe(":8080", nil)
}

func initGame(rows, cols int) {
	game = &Game{
		Rows:       rows,
		Cols:       cols,
		PlayerTurn: 1,
		Board:      make([][]int, rows),
	}
	for i := 0; i < rows; i++ {
		game.Board[i] = make([]int, cols)
	}
	game.GameOver = false
	game.Winner = 0
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	tmplStart.Execute(w, nil)
}

func levelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		level := r.FormValue("level")
		switch level {
		case "easy":
			initGame(6, 7)
		case "normal":
			initGame(6, 9)
		case "hard":
			initGame(7, 8)
		default:
			initGame(6, 7)
		}
		http.Redirect(w, r, "/play", http.StatusSeeOther)
		return
	}
	tmplLevel.Execute(w, nil)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	boardBytes, _ := json.Marshal(game.Board)
	data := TemplateData{
		Game:      game,
		BoardJSON: template.JS(boardBytes),
	}
	tmplIndex.Execute(w, data)
}

func moveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || game.GameOver {
		return
	}
	var req struct {
		Col int `json:"col"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	placeToken(req.Col)
	checkGameOver()
	json.NewEncoder(w).Encode(game)
}

func rematchHandler(w http.ResponseWriter, r *http.Request) {
	initGame(game.Rows, game.Cols)
	http.Redirect(w, r, "/play", http.StatusSeeOther)
}

func placeToken(col int) {
	for i := game.Rows - 1; i >= 0; i-- {
		if game.Board[i][col] == 0 {
			game.Board[i][col] = game.PlayerTurn
			game.PlayerTurn = 3 - game.PlayerTurn
			return
		}
	}
}

func checkGameOver() {
	for r := 0; r < game.Rows; r++ {
		for c := 0; c < game.Cols; c++ {
			player := game.Board[r][c]
			if player == 0 {
				continue
			}
			if checkDirection(r, c, 0, 1, player) || checkDirection(r, c, 1, 0, player) ||
				checkDirection(r, c, 1, 1, player) || checkDirection(r, c, 1, -1, player) {
				game.GameOver = true
				game.Winner = player
				return
			}
		}
	}
	full := true
	for r := 0; r < game.Rows; r++ {
		for c := 0; c < game.Cols; c++ {
			if game.Board[r][c] == 0 {
				full = false
				break
			}
		}
	}
	if full {
		game.GameOver = true
		game.Winner = 0
	}
}

func checkDirection(r, c, dr, dc, player int) bool {
	count := 0
	for i := 0; i < 4; i++ {
		nr, nc := r+i*dr, c+i*dc
		if nr < 0 || nr >= game.Rows || nc < 0 || nc >= game.Cols || game.Board[nr][nc] != player {
			return false
		}
		count++
	}
	return count == 4
}
