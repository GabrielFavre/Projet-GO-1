package main

import (
	"html/template"
	"net/http"
	"strconv"
)

const (
	rows    = 6
	columns = 7
)

type Game struct {
	Board    [rows][columns]int // 0 = vide, 1 = joueur 1, 2 = joueur 2
	Current  int                // joueur courant
	Winner   int                // gagnant éventuel
	GameOver bool
	Message  string
}

var game Game
var tmpl = template.Must(template.ParseFiles("./templates/index.html"))

func main() {
	resetGame()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/play", handlePlay)
	http.HandleFunc("/reset", handleReset)

	println("Serveur lancé sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func resetGame() {
	game = Game{Current: 1}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, game)
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

	// placer le pion dans la colonne choisie
	for row := rows - 1; row >= 0; row-- {
		if game.Board[row][col] == 0 {
			game.Board[row][col] = game.Current
			if checkWin(row, col, game.Current) {
				game.Winner = game.Current
				game.GameOver = true
				game.Message = "Le joueur " + strconv.Itoa(game.Current) + " a gagné !"
			} else if checkDraw() {
				game.GameOver = true
				game.Message = "Match nul !"
			} else {
				if game.Current == 1 {
					game.Current = 2
				} else {
					game.Current = 1
				}
			}
			break
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	resetGame()
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
		{0, 1},  // horizontal
		{1, 0},  // vertical
		{1, 1},  // diagonale ↘
		{1, -1}, // diagonale ↙
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
