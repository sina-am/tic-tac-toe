package games

import (
	"fmt"
)

type ticTacToe struct {
	turn    Tile
	board   [3][3]Tile
	winner  Tile
	players map[Tile]string
}

func NewTicTacToe(players []string) *ticTacToe {
	return &ticTacToe{
		turn:  TileX,
		board: [3][3]Tile{},
		players: map[Tile]string{
			TileX: players[0],
			TileO: players[1],
		},
	}
}
func (g *ticTacToe) GetPlayers() []string {
	return []string{g.players[TileX], g.players[TileO]}
}

func (g *ticTacToe) GetOpponent(playerId string) string {
	for _, id := range g.players {
		if id != playerId {
			return id
		}
	}
	return ""
}

func (g *ticTacToe) InGame(playerId string) bool {
	for _, id := range g.players {
		if id == playerId {
			return true
		}
	}
	return false
}

func (g *ticTacToe) GetPlayerTile(playerId string) Tile {
	for t, p := range g.players {
		if p == playerId {
			return t
		}
	}
	return TileEmpty
}

func (g *ticTacToe) Play(playerId string, move Move) error {
	if g.winner > 0 {
		return ErrGameEnd
	}

	tile := g.GetPlayerTile(playerId)

	if err := move.Validate(); err != nil {
		return err
	}
	if err := tile.Validate(); err != nil {
		return err
	}
	if tile != g.turn {
		return fmt.Errorf("this it not %s turn", tile.String())
	}

	if g.board[move/3][move%3] != TileEmpty {
		return fmt.Errorf("a tile already placed")
	}

	g.board[move/3][move%3] = tile
	if winner, err := g.CheckEndGame(); err == nil {
		g.winner = winner
		return ErrGameEnd
	}

	g.turn = tile.Opposite()
	return nil
}

func (g *ticTacToe) GetWinner() (string, error) {
	if g.winner > 0 {
		return g.winner.String(), nil
	}
	return TileEmpty.String(), fmt.Errorf("no winner yet")
}

func (g *ticTacToe) Exit(playerId string) (string, error) {
	t := g.GetPlayerTile(playerId)
	return t.Opposite().String(), nil
}

func (g *ticTacToe) CheckEndGame() (Tile, error) {
	for i := 0; i < 3; i++ {
		if rowEqual(g.board, i) && g.board[i][0] != TileEmpty {
			return g.board[i][0], nil
		}
		if colEqual(g.board, i) && g.board[0][i] != TileEmpty {
			return g.board[0][i], nil
		}
	}
	if axeXEqual(g.board) && g.board[0][0] != TileEmpty {
		return g.board[0][0], nil
	}
	if axeYEqual(g.board) && g.board[0][2] != TileEmpty {
		return g.board[0][2], nil
	}
	return TileEmpty, fmt.Errorf("no winner yet")
}

func axeYEqual(b [3][3]Tile) bool {
	equal := true
	for i := 1; i < len(b); i++ {
		if b[0][2] != b[i][2-i] {
			equal = false
			break
		}
	}
	return equal
}
func axeXEqual(b [3][3]Tile) bool {
	equal := true
	for i := 1; i < len(b); i++ {
		if b[0][0] != b[i][i] {
			equal = false
			break
		}
	}
	return equal
}
func rowEqual(b [3][3]Tile, r int) bool {
	equal := true
	for i := 1; i < len(b[r]); i++ {
		if b[r][0] != b[r][i] {
			equal = false
			break
		}
	}
	return equal
}

func colEqual(b [3][3]Tile, c int) bool {
	equal := true
	for i := 1; i < len(b); i++ {
		if b[0][c] != b[i][c] {
			equal = false
			break
		}
	}
	return equal
}
