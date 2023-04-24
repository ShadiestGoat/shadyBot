package connect4

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/shadyBot/characters"
)

type TileState rune

type Game struct {
	// ID of the players. Player 1 is the one who started it.
	Player1 string
	Player2 string

	TurnIsPlayer2 bool

	// 0-6
	CurrentCol int

	// Array of rows.
	// First row is the top row.
	Board [HEIGHT][WIDTH]TileState
}

const HEIGHT = 6
const WIDTH = 7

const (
	CURSOR_TOP = '🔽'
	CURSOR_BOT = '🔼'

	P0CHAR TileState = '🔲'
	P1CHAR TileState = '🟪'
	P2CHAR TileState = '🟧'
)

func NewGame(player1, player2 string) *Game {
	g := &Game{
		Player1:       player1,
		Player2:       player2,
		TurnIsPlayer2: false,
		CurrentCol:    3,
		Board:         [6][7]TileState{},
	}

	for i := range g.Board {
		for j := range g.Board[i] {
			g.Board[i][j] = P0CHAR
		}
	}

	return g
}

func playerString(char TileState, id string, isTurn bool) string {
	str := "Player " + string(char) + ": <@" + id + ">"
	if isTurn {
		str += " " + string(characters.HAND_LEFT)
	}
	return str
}

func (g Game) String() string {
	str := playerString(P1CHAR, g.Player1, !g.TurnIsPlayer2) + "\n" +
		playerString(P2CHAR, g.Player2, g.TurnIsPlayer2) + "\n\n"

	for i := 0; i < 7; i++ {
		if i == g.CurrentCol {
			str += string(characters.GAME_EMPTY_CHAR)
		} else {
			str += string(CURSOR_TOP)
		}
	}

	str += "\n"

	for _, rowRaw := range g.Board {
		str += string(characters.GAME_EMPTY_CHAR)

		for _, s := range rowRaw {
			str += string(s)
		}

		str += string(characters.GAME_EMPTY_CHAR) + "\n"
	}

	for i := 0; i < 7; i++ {
		if i == g.CurrentCol {
			str += string(characters.GAME_EMPTY_CHAR)
		} else {
			str += string(CURSOR_BOT)
		}
	}

	return str
}

func (g Game) buttons() []discordgo.MessageComponent {
	prefix := "c4_game_" + g.Player2

	return []discordgo.MessageComponent{
		discordgo.Button{
			Disabled: g.CurrentCol <= 1,
			Style:    discordgo.PrimaryButton,
			Emoji: discordgo.ComponentEmoji{
				Name: characters.SUPER_LEFT,
			},
			CustomID: prefix + "_superprev",
		},
		discordgo.Button{
			Disabled: g.CurrentCol == 0,
			Style:    discordgo.PrimaryButton,
			Emoji: discordgo.ComponentEmoji{
				Name: characters.ARROW_LL,
			},
			CustomID: prefix + "_prev",
		},
		discordgo.Button{
			Disabled: g.Board[0][g.CurrentCol] != P0CHAR,
			Style:    discordgo.SuccessButton,
			Emoji: discordgo.ComponentEmoji{
				Name: characters.ARROW_DD,
			},
			CustomID: prefix + "_place",
		},
		discordgo.Button{
			Disabled: g.CurrentCol == 6,
			Style:    discordgo.PrimaryButton,
			Emoji: discordgo.ComponentEmoji{
				Name: characters.ARROW_RR,
			},
			CustomID: prefix + "_next",
		},
		discordgo.Button{
			Disabled: g.CurrentCol >= 5,
			Style:    discordgo.PrimaryButton,
			Emoji: discordgo.ComponentEmoji{
				Name: characters.SUPER_RIGHT,
			},
			CustomID: prefix + "_supernext",
		},
	}
}

func isWon(inp []TileState) bool {
	return inp[0] == inp[1] && inp[1] == inp[2] && inp[2] == inp[3] && inp[0] != P0CHAR
}

type coords struct {
	x, y int
}

// Returns if the current player has won the round or not.
func (g *Game) Action(action string) bool {
	switch action {
	case "place":
		playerChar := P1CHAR
		if g.TurnIsPlayer2 {
			playerChar = P2CHAR
		}

		// [x, y] location
		loc := coords{
			x: g.CurrentCol,
		}

		for i := range g.Board {
			if g.Board[i][g.CurrentCol] != P0CHAR {
				g.Board[i-1][g.CurrentCol] = playerChar
				loc.y = i - 1

				break
			}
		}

		// horizontal check
		for i := 0; i <= WIDTH-4; i++ {
			if isWon(g.Board[loc.y][i : i+4]) {
				return true
			}
		}

		// vertical check
		for i := 0; i <= HEIGHT-4; i++ {
			if isWon([]TileState{
				g.Board[i+0][loc.x],
				g.Board[i+1][loc.x],
				g.Board[i+2][loc.x],
				g.Board[i+3][loc.x],
			}) {
				return true
			}
		}

		// plan for diagonal: find the 0th point of a diagonal and then check along the diagonal.
		// Before checking, make sure the diagonal length (measured in units in it) is *at least* 4
		// To check, make sure both the x and the y support it

		baseBT := coords{
			x: 0,
			y: loc.y - loc.x,
		}

		if baseBT.y >= HEIGHT {
			baseBT = coords{
				x: HEIGHT - 1 - loc.y + loc.x,
				y: HEIGHT - 1,
			}
		}

		if baseBT.y > 2 && baseBT.x <= WIDTH-4 {
			i := 0

			for {
				x := baseBT.x + i
				y := baseBT.y - i

				if y < 0 || x == WIDTH {
					break
				}

				if isWon([]TileState{
					g.Board[y-0][x+0],
					g.Board[y-1][x+1],
					g.Board[y-2][x+2],
					g.Board[y-3][x+3],
				}) {
					return true
				}
			}
		}

		baseTB := coords{
			x: 0,
			y: loc.y + loc.x,
		}

		if baseTB.y < 0 {
			baseBT = coords{
				x: loc.y + loc.x,
				y: 0,
			}
		}

		if baseTB.x <= WIDTH-4 && baseTB.y <= HEIGHT-4 {
			i := 0

			for {
				x := baseTB.x + i
				y := baseTB.y + i

				if y == HEIGHT || x == WIDTH {
					break
				}

				if isWon([]TileState{
					g.Board[y+0][x+0],
					g.Board[y+1][x+1],
					g.Board[y+2][x+2],
					g.Board[y+3][x+3],
				}) {
					return true
				}
			}
		}
	case "superprev":
		g.CurrentCol -= 2
	case "prev":
		g.CurrentCol -= 1
	case "next":
		g.CurrentCol += 1
	case "supernext":
		g.CurrentCol += 2
	}

	g.TurnIsPlayer2 = !g.TurnIsPlayer2

	return false
}
