package games

import (
	"errors"
	"fmt"
)

type Tile int

var ErrGameEnd = errors.New("game ended")

const (
	TileEmpty Tile = 0
	TileO     Tile = 1
	TileX     Tile = 2
)

func (t Tile) Opposite() Tile {
	switch t {
	case TileO:
		return TileX
	case TileX:
		return TileO
	default:
		return TileEmpty
	}
}
func (t Tile) String() string {
	switch t {
	case TileEmpty:
		return " "
	case TileO:
		return "O"
	case TileX:
		return "X"
	default:
		return ""
	}
}

func (t Tile) Validate() error {
	if t > 3 || t < 1 {
		return fmt.Errorf("invalid tile")
	}
	return nil
}

type Move int

func (m Move) Validate() error {
	if m > 8 || m < 0 {
		return fmt.Errorf("invalid move")
	}
	return nil
}

type Game interface {
	GetWinner() (string, error)
	InGame(playerId string) bool
	GetPlayers() []string
	Play(playerId string, m Move) error
	GetPlayerTile(playerId string) Tile
	Exit(playerId string) (string, error)
}
