package server

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sina-am/tic-tac-toe/games"
)

type PlayerStatus int

const (
	StatusWaiting   PlayerStatus = 0
	StatusPlaying   PlayerStatus = 1
	StatusConnected PlayerStatus = 2
)

type message struct {
	Type    string `json:"type"`
	Payload json.RawMessage
}

type joinMsg struct {
	GameId   string `json:"game_id"`
	Opponent string `json:"opponent"`
	Tile     string `json:"title"`
}

type endMsg struct {
	Reason string `json:"reason"`
	Winner string `json:"winner"`
	Score  int    `json:"score"`
}

type playerInfo struct {
	mu            sync.RWMutex
	name          string
	currentGameId string
	status        PlayerStatus
}

type player struct {
	conn *websocket.Conn
	id   string
	info playerInfo
	Send chan any
	Join chan joinMsg
	Err  chan error
	End  chan endMsg

	gameHandler *gameHandler
	msgHandler  map[string]func(message) error
}

func NewPlayer(conn *websocket.Conn, gamHandler *gameHandler) *player {
	p := &player{
		conn:        conn,
		gameHandler: gamHandler,
		Send:        make(chan any),
		Join:        make(chan joinMsg),
		Err:         make(chan error),
		End:         make(chan endMsg),
		id:          uuid.New().String(),
		info: playerInfo{
			mu:     sync.RWMutex{},
			status: StatusConnected,
		},
	}
	p.msgHandler = map[string]func(message) error{
		"start": p.handleStart,
		"play":  p.handlePlay,
		"exit":  p.handleExit,
	}
	return p
}
func (p *player) GetId() string {
	return p.id
}

func (p *player) GetName() string {
	p.info.mu.RLock()
	defer p.info.mu.RUnlock()

	return p.info.name
}

func (p *player) ReadConn() {
	defer func() {
		log.Printf("player %s disconnected", p.conn.RemoteAddr())
		p.gameHandler.unregister <- p
		p.conn.Close()
	}()

	for {
		msg := message{}
		if err := p.conn.ReadJSON(&msg); err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				break
			}
			log.Printf("websocket error: %v", err)
			p.Err <- err
			break
		}

		if err := p.handleMessage(msg); err != nil {
			p.Err <- err
			continue
		}
	}
}

func (p *player) WriteConn() {
	defer func() {
		p.conn.Close()
		p.gameHandler.unregister <- p
	}()

	for {
		select {
		case msg, ok := <-p.Send:
			if !ok {
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := p.conn.WriteJSON(msg); err != nil {
				return
			}

		case err, ok := <-p.Err:
			if !ok {
				break
			}
			p.conn.WriteJSON(map[string]string{"error": err.Error()})
		case msg, ok := <-p.Join:
			if !ok {
				break
			}
			p.info.mu.Lock()
			p.info.currentGameId = msg.GameId
			p.info.status = StatusPlaying
			p.info.mu.Unlock()

			p.conn.WriteJSON(map[string]any{
				"type": "started",
				"payload": map[string]string{
					"tile":     msg.Tile,
					"opponent": msg.Opponent,
				},
			})
		case msg, ok := <-p.End:
			if !ok {
				break
			}
			p.info.mu.Lock()
			p.info.currentGameId = ""
			p.info.status = StatusConnected
			p.info.mu.Unlock()

			p.conn.WriteJSON(map[string]any{
				"type":    "ended",
				"payload": msg,
			})
		}
	}
}

type StartGameMessage struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name"`
}

func (p *player) handleMessage(msg message) error {
	handler, found := p.msgHandler[msg.Type]
	if !found {
		return fmt.Errorf("invalid message type")
	}
	return handler(msg)
}

func (p *player) handleStart(msg message) error {
	p.info.mu.Lock()
	defer p.info.mu.Unlock()

	payload := StartGameMessage{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}

	if p.info.status == StatusWaiting {
		return fmt.Errorf("already in a waiting list")
	}
	if p.info.status == StatusPlaying {
		return fmt.Errorf("already in a game")
	}

	p.info.status = StatusWaiting
	p.info.name = payload.Name
	p.gameHandler.wait <- p

	p.Send <- map[string]string{
		"message": fmt.Sprintf("Id: %s", p.id),
	}
	return nil
}

type PlayGameMessage struct {
	Move games.Move `json:"move"`
}

func (p *player) handlePlay(msg message) error {
	p.info.mu.Lock()
	defer p.info.mu.Unlock()

	payload := PlayGameMessage{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}
	if p.info.status != StatusPlaying {
		return fmt.Errorf("you're not in any game")
	}

	p.gameHandler.playMove <- playMoveMsg{
		player: p,
		gameId: p.info.currentGameId,
		move:   payload.Move,
	}
	return nil
}

func (p *player) handleExit(msg message) error {
	p.info.mu.Lock()
	defer p.info.mu.Unlock()

	if p.info.status == StatusWaiting {
		p.gameHandler.exitQueue <- p
	} else if p.info.status == StatusPlaying {
		p.gameHandler.exitGame <- exitGameMsg{
			gameId: p.info.currentGameId,
			player: p,
		}
	} else {
		return fmt.Errorf("you don't have a game")
	}

	p.info.status = StatusConnected
	p.Send <- map[string]string{"message": "deleted"}
	return nil
}
