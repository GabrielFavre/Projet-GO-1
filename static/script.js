const boardEl = document.getElementById('board');
const statusEl = document.getElementById('status');
let playerTurn;

function initBoard(board, turn) {
  playerTurn = turn;
  renderBoard(board);
}

function renderBoard(board) {
  boardEl.innerHTML = '';
  for (let r = 0; r < board.length; r++) {
    for (let c = 0; c < board[0].length; c++) {
      const cell = document.createElement('div');
      cell.className = 'cell';
      cell.dataset.col = c;
      boardEl.appendChild(cell);
      if (board[r][c] > 0) {
        const token = document.createElement('div');
        token.className = 'token ' + (board[r][c] === 1 ? 'player1' : 'player2');
        token.style.top = '-60px';
        setTimeout(() => token.style.top = '5px', 10);
        cell.appendChild(token);
      }
    }
  }
  statusEl.textContent = "Player " + playerTurn + "'s turn";
  boardEl.querySelectorAll('.cell').forEach(cell => {
    cell.onclick = () => {
      makeMove(parseInt(cell.dataset.col));
    };
  });
}

function makeMove(col) {
  fetch('/move', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({col})
  })
    .then(res => res.json())
    .then(data => {
      renderBoard(data.Board);
      playerTurn = data.PlayerTurn; // ⚡ correction : mettre à jour le joueur
      if (data.GameOver) {
        if (data.Winner === 0) alert('Draw!');
        else alert('Player ' + data.Winner + ' wins!');
      }
    });
}
