package main

import (
	"encoding/json"
	"html/template"
	"math/rand"
	"net/http"
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
}

type IndexData struct {
	*Game
	BoardJSON template.JS
}

var (
	game      = &Game{}
	tmplStart = template.Must(template.ParseFiles("templates/start.html"))
	tmplLevel = template.Must(template.ParseFiles("templates/level.html"))
	tmplPlay  = template.Must(template.ParseFiles("templates/index.html"))
)

func main() {
	rand.Seed(time.Now().UnixNano())
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", startHandler)
	http.HandleFunc("/level", levelHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/move", moveHandler)
	http.HandleFunc("/rematch", rematchHandler)
	http.ListenAndServe(":8080", nil)
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		game.PlayerNames[0] = r.FormValue("player1")
		game.PlayerNames[1] = r.FormValue("player2")
		http.Redirect(w, r, "/level", http.StatusSeeOther)
		return
	}
	tmplStart.Execute(w, nil)
}

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

func initGame(rows, cols, blocks int) {
	game.Rows = rows
	game.Cols = cols
	game.PlayerTurn = 1
	game.TurnCount = 0
	game.GravityDown = true
	game.GameOver = false
	game.Winner = 0
	game.CurrentPlayerIndex = 0
	game.Board = make([][]int, rows)
	for i := range game.Board {
		game.Board[i] = make([]int, cols)
	}
	for b := 0; b < blocks; b++ {
		r := rand.Intn(rows)
		c := rand.Intn(cols)
		if game.Board[r][c] == 0 {
			game.Board[r][c] = rand.Intn(2) + 1
		}
	}
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	boardBytes, _ := json.Marshal(game.Board)
	data := IndexData{
		Game:      game,
		BoardJSON: template.JS(boardBytes),
	}
	tmplPlay.Execute(w, data)
}

func moveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || game.GameOver {
		return
	}
	var data struct {
		Col int `json:"col"`
	}
	json.NewDecoder(r.Body).Decode(&data)
	placeToken(data.Col)
	game.TurnCount++
	if game.TurnCount%5 == 0 {
		game.GravityDown = !game.GravityDown
	}
	for r := 0; r < game.Rows; r++ {
		for c := 0; c < game.Cols; c++ {
			if game.Board[r][c] != 0 && checkWin(r, c, game.Board[r][c]) {
				game.GameOver = true
				game.Winner = game.Board[r][c]
			}
		}
	}
	if game.GameOver || isDraw() {
		game.GameOver = true
	}
	game.CurrentPlayerIndex = game.PlayerTurn - 1
	json.NewEncoder(w).Encode(game)
}

func rematchHandler(w http.ResponseWriter, r *http.Request) {
	initGame(game.Rows, game.Cols, 0)
	http.Redirect(w, r, "/play", http.StatusSeeOther)
}

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
