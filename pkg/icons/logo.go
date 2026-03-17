package icons

import "math/rand"

func RandIcon() string {
	n := rand.Intn(len(icons))
	return icons[n]
}

var icons = []string{
	"chess-bishop",
	"chess-king",
	"chess-knight",
	"chess-pawn",
	"chess-queen",
	"chess-rook",
}
