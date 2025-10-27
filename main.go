package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Game struct {
	Board      [][]int
	Rows       int
	Columns    int
	Current    int
	Winner     int
	GameOver   bool
	Message    string
	Players    [3]string
	Gravity    bool
	TurnCount  int
	Difficulty string
}

var game Game
var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/start", handleStart)
	http.HandleFunc("/play", handlePlay)
	http.HandleFunc("/reset", handleReset)

	log.Println("Serveur lancé sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func newGame(p1, p2, difficulty string) {
	rand.Seed(time.Now().UnixNano())
	game.Difficulty = difficulty
	game.Players[1] = p1
	game.Players[2] = p2
	game.Gravity = true
	game.TurnCount = 0
	game.Winner = 0
	game.GameOver = false
	game.Message = ""
	game.Current = 1

	switch difficulty {
	case "normal":
		game.Rows, game.Columns = 6, 9
	case "hard":
		game.Rows, game.Columns = 7, 8
	default:
		game.Rows, game.Columns = 6, 7
	}

	game.Board = make([][]int, game.Rows)
	for r := range game.Board {
		game.Board[r] = make([]int, game.Columns)
	}

	fill := map[string]int{"easy": 3, "normal": 5, "hard": 7}[difficulty]
	for i := 0; i < fill; i++ {
		r := rand.Intn(game.Rows)
		c := rand.Intn(game.Columns)
		game.Board[r][c] = rand.Intn(2) + 1
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, game); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	p1 := r.FormValue("player1")
	p2 := r.FormValue("player2")
	if p1 == "" {
		p1 = "Rouge"
	}
	if p2 == "" {
		p2 = "Bleu"
	}
	d := strings.ToLower(r.FormValue("difficulty"))
	if d == "" {
		d = "easy"
	}
	newGame(p1, p2, d)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handlePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || game.GameOver {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	col, _ := strconv.Atoi(r.FormValue("column"))
	if col < 0 || col >= game.Columns {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	placed := false
	if game.Gravity {
		for row := game.Rows - 1; row >= 0; row-- {
			if game.Board[row][col] == 0 {
				game.Board[row][col] = game.Current
				handleAfterMove(row, col)
				placed = true
				break
			}
		}
	} else {
		for row := 0; row < game.Rows; row++ {
			if game.Board[row][col] == 0 {
				game.Board[row][col] = game.Current
				handleAfterMove(row, col)
				placed = true
				break
			}
		}
	}

	if !placed {
		game.Message = "Colonne pleine !"
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleAfterMove(r, c int) {
	player := game.Current
	if checkWin(r, c, player) {
		game.Winner = player
		game.GameOver = true
		game.Message = fmt.Sprintf("%s a gagné !", game.Players[player])
		return
	}
	if checkDraw() {
		game.GameOver = true
		game.Message = "Match nul !"
		return
	}

	game.TurnCount++
	if game.TurnCount%5 == 0 {
		game.Gravity = !game.Gravity
		if game.Gravity {
			game.Message = "Gravité normale rétablie."
		} else {
			game.Message = "Gravité inversée !"
		}
	}

	if game.Current == 1 {
		game.Current = 2
	} else {
		game.Current = 1
	}
}

func checkDraw() bool {
	for _, row := range game.Board {
		for _, cell := range row {
			if cell == 0 {
				return false
			}
		}
	}
	return true
}

func checkWin(r, c, player int) bool {
	dirs := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}
	for _, d := range dirs {
		count := 1
		for i := 1; i < 4; i++ {
			nr, nc := r+d[0]*i, c+d[1]*i
			if nr < 0 || nr >= game.Rows || nc < 0 || nc >= game.Columns || game.Board[nr][nc] != player {
				break
			}
			count++
		}
		for i := 1; i < 4; i++ {
			nr, nc := r-d[0]*i, c-d[1]*i
			if nr < 0 || nr >= game.Rows || nc < 0 || nc >= game.Columns || game.Board[nr][nc] != player {
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

func handleReset(w http.ResponseWriter, r *http.Request) {
	game = Game{}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
