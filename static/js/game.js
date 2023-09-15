// state: "waiting" | "started" | "exited"

function displayElementById(id, display) {
    if (display) {
        document.getElementById(id).classList.remove('d-none');
    } else {
        document.getElementById(id).classList.add('d-none');
    }
}
function setGameState(state) {
    switch (state) {
        case "exited":
            displayElementById("game", false);
            displayElementById("waiting", false);
            displayElementById("startButton", true);
            displayElementById("startStage", true);
            break;
        case "waiting":
            displayElementById("game", false);
            displayElementById("startStage", true);
            displayElementById("waiting", true);
            displayElementById("startButton", false);
            break;
        case "started":
            displayElementById("startStage", false);
            displayElementById("game", true);
            break;
    }
}
class GameClient {
    constructor(ws, playerName) {
        this.ws = ws;
        this.status = "";
        this.myTile = '';
        this.myTurn = false;
        this.playerName = playerName;
        this.opponentNode = document.getElementById("opponentName");
        this.playerNode = document.getElementById("playerName");
    }

    start() {
        this.playerNode.innerText = this.playerName;
        this.status = "waiting";
        setGameState("waiting");
        this.ws.send(JSON.stringify({
            "type": "start",
            "payload": {
                "name": this.playerName,
            }
        }));
    }

    play(m) {
        if (this.status !== "started") return;
        this.myTurn = !this.myTurn;
        this.updatePlayerBar()
        this.ws.send(JSON.stringify({
            "type": "play",
            "payload": {
                "move": m,
            }
        }));
    }

    getOppositeTile(t) {
        return (t === 'X') ? 'O' : 'X';
    }

    updatePlayerBar() {
        if (this.myTurn) {
            this.playerNode.innerText = this.playerName + " " + this.myTile + " (your turn)";
        } else {
            this.playerNode.innerText = this.playerName + " " + this.myTile;
        }
    }
    update(msg) {
        switch (msg.type) {
            case "started":
                this.myTile = msg.payload.tile;
                this.myTurn = (this.myTile == 'X') ? true : false;
                this.opponentNode.innerText = msg.payload.opponent + " " + this.getOppositeTile(this.myTile);
                this.updatePlayerBar()
                this.status = "started";
                setGameState(this.status);
                console.log("game started")
                break;
            case "played":
                this.myTurn = true;
                this.updatePlayerBar()
                document.getElementById(`tile-${msg.payload.move}`).innerText = this.getOppositeTile(this.myTile);
                break;
            case "ended":
                if (msg.payload.winner === this.myTile) {
                    document.getElementById('gameWinner').innerText = "You won!";
                } else {
                    document.getElementById('gameWinner').innerText = "You lost";
                }
                this.status = "finished";
                this.ws.send(JSON.stringify({
                    "type": "exit",
                    "payload": ""
                }));
                break;
            default:
                break;
        }

        document.getElementById("gameStatus").innerText = this.status;
    }
    exit() {
        console.log("exiting the game");
        this.status = "exited";
        setGameState(this.status)
        this.ws.send(JSON.stringify({
            "type": "exit",
            "payload": ""
        }));
    }
}

window.onload = () => {
    const ws = new WebSocket("ws://localhost:8080/ws");

    ws.addEventListener("message", async (event) => {
        console.log("websocket new message: ", event.data);
    });

    ws.addEventListener("open", async (event) => {
        console.log("websocket connected");
    });
    ws.addEventListener("close", async (event) => {
        console.log("websocket connection closed");
    });
    ws.addEventListener("error", (event) => {
        console.log("websocket error: ", event);
    });

    document.getElementById("startGame").onclick = async (event) => await startGame(ws);
}
function getCookie(cname) {
    let name = cname + "=";
    let decodedCookie = decodeURIComponent(document.cookie);
    let ca = decodedCookie.split(';');
    for (let i = 0; i < ca.length; i++) {
        let c = ca[i];
        while (c.charAt(0) == ' ') {
            c = c.substring(1);
        }
        if (c.indexOf(name) == 0) {
            return c.substring(name.length, c.length);
        }
    }
    return "";
}
function setCookie(cname, cvalue, exdays) {
    const d = new Date();
    d.setTime(d.getTime() + (exdays * 24 * 60 * 60 * 1000));
    let expires = "expires=" + d.toUTCString();
    document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/";
}
function deleteAllCookies() {
    const cookies = document.cookie.split(";");

    for (let i = 0; i < cookies.length; i++) {
        const cookie = cookies[i];
        const eqPos = cookie.indexOf("=");
        const name = eqPos > -1 ? cookie.substr(0, eqPos) : cookie;
        document.cookie = name + "=;expires=Thu, 01 Jan 1970 00:00:00 GMT";
    }
}
function startNewGameClicked(event) {
    deleteAllCookies();
    const playerName = getCookie("playerName");
    if (!playerName) {
        let myModal = new bootstrap.Modal(document.getElementById('playerNameModal'), {})
        myModal.toggle()
    }
    else {
        document.getElementById("startGame").click()
    }
}
async function startGame(ws) {
    let playerName = getCookie("playerName");
    if (playerName === "") {
        playerName = document.getElementById("playerNameInput").value;
        setCookie("playerName", playerName, 1);
    }

    console.log(playerName);
    if (playerName === "") {
        return
    }

    if (ws.readyState !== WebSocket.OPEN) {
        document.write("websocket is closed")
        throw "websocket is closed";
    }
    try {
        const res = await fetch('game.html');
        document.getElementById("game").innerHTML = await res.text();
        setGameState("waiting");
    } catch {
        setGameState("ended");
        throw "something went wrong";
    }

    const game = new GameClient(ws, playerName);
    ws.addEventListener("message", async (event) => {
        game.update(JSON.parse(event.data))
    });

    const board = document.getElementById("gameBoard");
    for (let i = 0; i < 3; i++) {
        const row = document.createElement("tr");
        if (i == 1) {
            row.style.borderTop = "3px solid gray";
            row.style.borderBottom = "3px solid gray";
        }
        for (let j = 0; j < 3; j++) {
            const squire = document.createElement("td");
            squire.id = `tile-${i * 3 + j}`;
            if (j == 1) {
                squire.style.borderLeft = "3px solid gray";
                squire.style.borderRight = "3px solid gray";
            }

            squire.onclick = (event) => {
                if (event.target.innerText === "" && game.myTurn) {
                    game.play(i * 3 + j);
                    event.target.innerText = game.myTile
                } else {
                    console.log("squire is occupied or it's not your turn")
                }
            };
            squire.innerText = " ";
            squire.classList.add("squire");
            row.appendChild(squire);
        }
        board.appendChild(row);
    }

    const exitButton = document.getElementById("exitBtn");
    exitButton.onclick = (event) => game.exit();

    game.start();
}
