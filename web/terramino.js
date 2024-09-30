var score = 0;
var highScore = 0; // Initialize high score
var newHighScore = false;

const SCORE_NEWTETROMINO = 10;
const SCORE_CLEARLINE = 100;

updateScore();
getHighScore();

function updateScore() {
  const scoreSpan = document.getElementById("score");
  if (scoreSpan) {
    scoreSpan.innerText = score.toString().padStart(8, "0"); // Pad with leading zeros
  } else {
    console.error("Score element not found!");
  }
}

function updateHighScoreUI() {
  const highScoreSpan = document.getElementById("highscore");
  if (highScoreSpan) {
    highScoreSpan.innerText = highScore.toString().padStart(8, "0"); // Pad with leading zeros
  } else {
    highScoreSpan.innerText = "ERROR";
    console.error("High score element not found!");
  }
}

// Get the high score
async function getHighScore() {
  const getScoreData = await fetch("/score").then((res) => res.text());
  highScore = getScoreData;
  updateHighScoreUI();
}

// Set the high score
async function setHighScore() {
  if (newHighScore) {
    const setScoreData = await fetch("/score", {
      method: "POST",
      body: highScore
    }).then((res) => res.text());
  }
}

// Set a new high score if the current score is higher
function checkHighScore() {
  if (score > highScore) {
    highScore = score;
    newHighScore = true;
    updateHighScoreUI();
  }
}

// Add points and update the score
function addScore(points) {
  score += points;
  updateScore();
  checkHighScore(); // Check if the high score needs to be updated
}
// get a random integer between the range of [min,max]
function getRandomInt(min, max) {
  min = Math.ceil(min);
  max = Math.floor(max);

  return Math.floor(Math.random() * (max - min + 1)) + min;
}

// generate a new tetromino sequence
function generateSequence() {
  const sequence = ["I", "J", "L", "O", "S", "T", "Z"];

  while (sequence.length) {
    const rand = getRandomInt(0, sequence.length - 1);
    const name = sequence.splice(rand, 1)[0];
    tetrominoSequence.push(name);
  }
}

const nextTetrominoQueue = []; // Queue for the next tetrominos

function getNextTetromino() {
  if (tetrominoSequence.length === 0) {
    generateSequence();
  }

  // Ensure the next queue has 3 blocks at all times
  while (nextTetrominoQueue.length < 3) {
    nextTetrominoQueue.push(tetrominoSequence.pop());
  }

  const nextTetromino = nextTetrominoQueue.shift();
  const matrix = tetrominos[nextTetromino];

  const col = playfield[0].length / 2 - Math.ceil(matrix[0].length / 2);
  const row = nextTetromino === "I" ? -1 : -2;

  return {
    name: nextTetromino,
    matrix: matrix,
    row: row,
    col: col,
  };
}

// Render next tetrominos in the side panel (2 blocks)
function drawNextTetrominos() {
  const nextCanvasIds = ["next-block-1", "next-block-2"];
  for (let i = 0; i < nextTetrominoQueue.length; i++) {
    const nextCanvas = document.getElementById(nextCanvasIds[i]);
    const nextContext = nextCanvas.getContext("2d");

    nextContext.clearRect(0, 0, nextCanvas.width, nextCanvas.height);
    const next = nextTetrominoQueue[i];
    const nextMatrix = tetrominos[next];

    const blockSize = 12; // Adjust scale to fit in the next block
    const matrixWidth = nextMatrix[0].length;
    const matrixHeight = nextMatrix.length;

    const offsetX = (nextCanvas.width - matrixWidth * blockSize) / 2;

    // Adjust offsetY differently based on the tetromino type
    let offsetY = (nextCanvas.height - matrixHeight * blockSize) / 2;

    // Handle specific cases for Tetrominos with varying heights
    switch (next) {
      case "I":
        // The "I" piece is tall, shift it down slightly
        offsetY += blockSize / 2;
        break;
      case "O":
        // "O" block is already centered, no adjustment needed
        offsetY = (nextCanvas.height - matrixHeight * blockSize) / 2;
        break;
      case "J":
      case "L":
      case "T":
        // These pieces are taller but not as tall as "I", slight downward adjustment
        offsetY += blockSize / 4;
        break;
      case "S":
      case "Z":
        // "S" and "Z" are two rows tall but can look slightly off-center, adjust down
        offsetY += blockSize / 4;
        break;
    }

    nextContext.fillStyle = colors[next];

    for (let row = 0; row < nextMatrix.length; row++) {
      for (let col = 0; col < nextMatrix[row].length; col++) {
        if (nextMatrix[row][col]) {
          nextContext.fillRect(
            offsetX + col * blockSize,
            offsetY + row * blockSize,
            blockSize - 1,
            blockSize - 1
          );
        }
      }
    }
  }
}

// Rotate matrix 90 degrees
function rotate(matrix) {
  const N = matrix.length - 1;
  return matrix.map((row, i) => row.map((val, j) => matrix[N - j][i]));
}

// Check if the new matrix/row/col is valid
function isValidMove(matrix, cellRow, cellCol) {
  for (let row = 0; row < matrix.length; row++) {
    for (let col = 0; col < matrix[row].length; col++) {
      if (
        matrix[row][col] &&
        (cellCol + col < 0 ||
          cellCol + col >= playfield[0].length ||
          cellRow + row >= playfield.length ||
          playfield[cellRow + row][cellCol + col])
      ) {
        return false;
      }
    }
  }
  return true;
}

// Place tetromino on playfield
function placeTetromino() {
  for (let row = 0; row < tetromino.matrix.length; row++) {
    for (let col = 0; col < tetromino.matrix[row].length; col++) {
      if (tetromino.matrix[row][col]) {
        if (tetromino.row + row < 0) {
          return showGameOver();
        }
        playfield[tetromino.row + row][tetromino.col + col] = tetromino.name;
      }
    }
  }

  // Clear lines
  for (let row = playfield.length - 1; row >= 0; ) {
    if (playfield[row].every((cell) => !!cell)) {
      for (let r = row; r >= 0; r--) {
        playfield[r] = playfield[r - 1];
      }
      addScore(SCORE_CLEARLINE);
    } else {
      row--;
    }
  }

  addScore(SCORE_NEWTETROMINO);
  tetromino = getNextTetromino();
}

// Show game over
function showGameOver() {
  cancelAnimationFrame(rAF);
  gameOver = true;

  context.fillStyle = "black";
  context.globalAlpha = 0.75;
  context.fillRect(0, canvas.height / 2 - 30, canvas.width, 60);

  context.globalAlpha = 1;
  context.fillStyle = "white";
  context.font = "36px monospace";
  context.textAlign = "center";
  context.textBaseline = "middle";
  context.fillText("GAME OVER!", canvas.width / 2, canvas.height / 2);

  setHighScore();
}

const canvas = document.getElementById("game");
const context = canvas.getContext("2d");
const grid = 32;
const tetrominoSequence = [];

// Create an empty playfield
const playfield = [];
for (let row = -2; row < 20; row++) {
  playfield[row] = [];
  for (let col = 0; col < 10; col++) {
    playfield[row][col] = 0;
  }
}

// Tetrominos
const tetrominos = {
  I: [
    [0, 0, 0, 0],
    [1, 1, 1, 1],
    [0, 0, 0, 0],
    [0, 0, 0, 0],
  ],
  J: [
    [1, 0, 0],
    [1, 1, 1],
    [0, 0, 0],
  ],
  L: [
    [0, 0, 1],
    [1, 1, 1],
    [0, 0, 0],
  ],
  O: [
    [1, 1],
    [1, 1],
  ],
  S: [
    [0, 1, 1],
    [1, 1, 0],
    [0, 0, 0],
  ],
  Z: [
    [1, 1, 0],
    [0, 1, 1],
    [0, 0, 0],
  ],
  T: [
    [0, 1, 0],
    [1, 1, 1],
    [0, 0, 0],
  ],
};

// Colors
const colors = {
  I: "#7B42BC", // Terraform
  O: "#FFFC25", // Vault
  T: "#EC585D", // Boundary
  S: "#14C6CB", // Waypoint
  Z: "#DC477D", // Consul
  J: "#2E71E5", // Vagrant
  L: "#02A8EF", // Packer
};

let count = 0;
let tetromino = getNextTetromino();
let rAF = null;
let gameOver = false;

// Function to lighten a hex color by a percentage
function lightenHexColor(hex, percent) {
  const num = parseInt(hex.slice(1), 16),
    amt = Math.round(2.55 * percent),
    R = (num >> 16) + amt,
    G = ((num >> 8) & 0x00ff) + amt,
    B = (num & 0x0000ff) + amt;

  return `#${(
    0x1000000 +
    (R < 255 ? R : 255) * 0x10000 +
    (G < 255 ? G : 255) * 0x100 +
    (B < 255 ? B : 255)
  )
    .toString(16)
    .slice(1)}`;
}

function drawGhostPiece() {
  let ghostRow = tetromino.row;

  // Find where the piece would land
  while (isValidMove(tetromino.matrix, ghostRow + 1, tetromino.col)) {
    ghostRow++;
  }

  // Set a reduced opacity to blend the ghost piece with the background
  context.globalAlpha = 0.3; // Make the ghost piece 30% opaque

  for (let row = 0; row < tetromino.matrix.length; row++) {
    for (let col = 0; col < tetromino.matrix[row].length; col++) {
      if (tetromino.matrix[row][col]) {
        const originalColor = colors[tetromino.name];
        context.fillStyle = originalColor;
        context.fillRect(
          (tetromino.col + col) * grid,
          (ghostRow + row) * grid,
          grid - 1,
          grid - 1
        );
      }
    }
  }

  // Reset opacity back to default
  context.globalAlpha = 1.0;
}

// Game loop
function loop() {
  rAF = requestAnimationFrame(loop);
  context.clearRect(0, 0, canvas.width, canvas.height);

  // Draw the playfield
  for (let row = 0; row < 20; row++) {
    for (let col = 0; col < 10; col++) {
      if (playfield[row][col]) {
        const name = playfield[row][col];
        context.fillStyle = colors[name];
        context.fillRect(col * grid, row * grid, grid - 1, grid - 1);
      }
    }
  }

  // Draw the ghost piece
  drawGhostPiece();

  // Draw the active tetromino
  if (tetromino) {
    if (++count > 35) {
      tetromino.row++;
      count = 0;

      if (!isValidMove(tetromino.matrix, tetromino.row, tetromino.col)) {
        tetromino.row--;
        placeTetromino();
      }
    }

    context.fillStyle = colors[tetromino.name];
    for (let row = 0; row < tetromino.matrix.length; row++) {
      for (let col = 0; col < tetromino.matrix[row].length; col++) {
        if (tetromino.matrix[row][col]) {
          context.fillRect(
            (tetromino.col + col) * grid,
            (tetromino.row + row) * grid,
            grid - 1,
            grid - 1
          );
        }
      }
    }
  }

  // Draw the next tetrominos in the side panel
  drawNextTetrominos();
}

// Listen to keyboard events to move the active tetromino
// Listen to keyboard events to move the active tetromino
document.addEventListener("keydown", function (e) {
  if (gameOver) return;

  // Left arrow key (move left)
  if (e.which === 37) {
    const col = tetromino.col - 1;
    if (isValidMove(tetromino.matrix, tetromino.row, col)) {
      tetromino.col = col;
    }
  }

  // Right arrow key (move right)
  if (e.which === 39) {
    const col = tetromino.col + 1;
    if (isValidMove(tetromino.matrix, tetromino.row, col)) {
      tetromino.col = col;
    }
  }

  // Up arrow key (rotate)
  if (e.which === 38) {
    const matrix = rotate(tetromino.matrix);
    if (isValidMove(matrix, tetromino.row, tetromino.col)) {
      tetromino.matrix = matrix;
    }
  }

  // Down arrow key (soft drop)
  if (e.which === 40) {
    const row = tetromino.row + 1;
    if (!isValidMove(tetromino.matrix, row, tetromino.col)) {
      tetromino.row = row - 1;
      placeTetromino();
      return;
    }
    tetromino.row = row;
  }

  // Space bar key (hard drop)
  if (e.which === 32) {
    hardDrop();
  }
});

// Hard drop mechanism
function hardDrop() {
  while (isValidMove(tetromino.matrix, tetromino.row + 1, tetromino.col)) {
    tetromino.row++;
  }
  placeTetromino(); // Lock the piece in place
}

// Start game
rAF = requestAnimationFrame(loop);

// Ensure the modal is hidden by default
window.onload = function () {
  const debugModal = document.getElementById("debug-modal");
  debugModal.style.display = "none"; // Ensure it's hidden on page load

  const modeToggle = document.getElementById("mode-toggle");
  modeToggle.addEventListener("click", () => {
    document.body.classList.toggle("light-mode");
    updateColors();
  });
};

// Event listener for the Debug button
document.getElementById("debug-button").addEventListener("click", async () => {
  const debugModal = document.getElementById("debug-modal");
  const debugInfo = document.getElementById("debug-info");

  // Fetch the data from /env and /redis endpoints
  const envData = await fetch("/env").then((res) => res.text());
  const redisData = await fetch("/redis").then((res) => res.text());

  // Update the modal content with fetched data
  debugInfo.textContent = `Env Info:\n${envData}\nRedis Info:\n${redisData}`;

  // Show the modal only when the button is clicked
  debugModal.style.display = "block";
});

// Event listener for the Close button inside the modal
document.getElementById("close-debug").addEventListener("click", () => {
  const debugModal = document.getElementById("debug-modal");
  debugModal.style.display = "none"; // Hide the modal when Close is clicked
});
