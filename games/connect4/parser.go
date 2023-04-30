package connect4

import "strings"

func parseGame(desc string) *Game {
	g := &Game{
		TurnIsPlayer2: false,
		CurrentCol:    0,
		Board:         [6][7]TileState{},
	}
	spl := strings.Split(desc, "\n")

	for pIndex, line := range spl[:2] {
		id := ""
		for _, b := range line[12:] {
			if b < '0' || b > '9' {
				continue
			}
			id += string(b)
		}

		if pIndex == 0 {
			g.Player1 = id
		} else {
			g.Player2 = id
		}
	}

	g.TurnIsPlayer2 = spl[0][len(spl[0])-1] == '>'

	for i, b := range spl[3] {
		if b == CURSOR_TOP {
			g.CurrentCol = i - 1
		}
	}

	for y, row := range spl[4 : len(spl)-1] {
		for x, col := range row[1 : len(row)-1] {
			g.Board[y][x] = TileState(col)
		}
	}

	return g
}
