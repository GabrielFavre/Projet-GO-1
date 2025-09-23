package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

const (
	rows    = 6
	columns = 7
)

type Game struct {
	Board    [rows][columns]int
	Current  int
	Winner   int
	GameOver bool
	Message  string
	BgColor  string
}

var game Game
var tmpl = template.Must(template.ParseFiles("./templates/index.html"))

func main() {
	resetGame()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/play", handlePlay)
	http.HandleFunc("/reset", handleReset)

	log.Println("Serveur lancé sur http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func resetGame() {
	game = Game{Current: 1, BgColor: "#ff4d4d"} // commence par le rouge
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, game); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || game.GameOver {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	colStr := r.FormValue("column")
	col, err := strconv.Atoi(colStr)
	if err != nil || col < 0 || col >= columns {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	for row := rows - 1; row >= 0; row-- {
		if game.Board[row][col] == 0 {
			game.Board[row][col] = game.Current
			if checkWin(row, col, game.Current) {
				game.Winner = game.Current
				game.GameOver = true
				game.Message = fmt.Sprintf("Le joueur %s a gagné !", playerName(game.Current))
			} else if checkDraw() {
				game.GameOver = true
				game.Message = "Match nul !"
			} else {
				if game.Current == 1 {
					game.Current = 2
					game.BgColor = "#4da6ff" // fond bleu
				} else {
					game.Current = 1
					game.BgColor = "#ff4d4d" // fond rouge
				}
			}
			break
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		resetGame()
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func checkDraw() bool {
	for c := 0; c < columns; c++ {
		if game.Board[0][c] == 0 {
			return false
		}
	}
	return true
}

func checkWin(r, c, player int) bool {
	directions := [][2]int{
		{0, 1},
		{1, 0},
		{1, 1},
		{1, -1},
	}

	for _, d := range directions {
		count := 1
		for i := 1; i < 4; i++ {
			nr, nc := r+d[0]*i, c+d[1]*i
			if nr < 0 || nr >= rows || nc < 0 || nc >= columns || game.Board[nr][nc] != player {
				break
			}
			count++
		}
		for i := 1; i < 4; i++ {
			nr, nc := r-d[0]*i, c-d[1]*i
			if nr < 0 || nr >= rows || nc < 0 || nc >= columns || game.Board[nr][nc] != player {
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

func playerName(player int) string {
	if player == 1 {
		return "Rouge"
	}
	return "Bleu"
}
