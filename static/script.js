let boardEl=document.getElementById('board');
function initBoard(board, cols, rows, turn){
renderBoard(board, turn);
}
function renderBoard(board, turn){
boardEl.innerHTML='';
for(let r=0;r<board.length;r++){
for(let c=
