package games

import "math/rand"

const (
	SLOTS_JACKPOT = 66666

	SLOTS_CHAR_BAD_1          = '🟪'
	SLOTS_CHAR_BAD_2          = '🟦'
	SLOTS_CHAR_BAD_3          = '🟫'
	SLOTS_CHAR_BAD_4          = '🟩'
	SLOTS_CHAR_BAD_5          = '🟥'
	SLOTS_CHAR_BAD_6          = '🟨'
	SLOTS_CHAR_BAD_7          = '🟧'
	SLOTS_CHAR_SPECIAL_MEDIUM = '😳'
	SLOTS_CHAR_SPECIAL_EPIC   = '❗'
)

var availableSlots = []rune{
	SLOTS_CHAR_BAD_1, SLOTS_CHAR_BAD_1, SLOTS_CHAR_BAD_1,
	SLOTS_CHAR_BAD_2, SLOTS_CHAR_BAD_2, SLOTS_CHAR_BAD_2,
	SLOTS_CHAR_BAD_3, SLOTS_CHAR_BAD_3, SLOTS_CHAR_BAD_3,
	SLOTS_CHAR_BAD_4, SLOTS_CHAR_BAD_4, SLOTS_CHAR_BAD_4,
	SLOTS_CHAR_BAD_5, SLOTS_CHAR_BAD_5, SLOTS_CHAR_BAD_5,
	SLOTS_CHAR_BAD_6, SLOTS_CHAR_BAD_6, SLOTS_CHAR_BAD_6,
	SLOTS_CHAR_BAD_7, SLOTS_CHAR_BAD_7, SLOTS_CHAR_BAD_7,
	SLOTS_CHAR_SPECIAL_MEDIUM, SLOTS_CHAR_SPECIAL_MEDIUM,
	SLOTS_CHAR_SPECIAL_EPIC,
}

func slotsRewards(inp [3]rune) int {
	if inp[0] == inp[1] && inp[1] == inp[2] {
		switch inp[0] {
		case SLOTS_CHAR_BAD_1,
			SLOTS_CHAR_BAD_2,
			SLOTS_CHAR_BAD_3,
			SLOTS_CHAR_BAD_4,
			SLOTS_CHAR_BAD_5,
			SLOTS_CHAR_BAD_6,
			SLOTS_CHAR_BAD_7:

			return 419
		case SLOTS_CHAR_SPECIAL_MEDIUM:
			return 11111
		case SLOTS_CHAR_SPECIAL_EPIC:
			return SLOTS_JACKPOT
		}
	}

	if inp[1] == inp[0] || inp[1] == inp[2] {
		switch inp[1] {
		case SLOTS_CHAR_BAD_1,
			SLOTS_CHAR_BAD_2,
			SLOTS_CHAR_BAD_3,
			SLOTS_CHAR_BAD_4,
			SLOTS_CHAR_BAD_5,
			SLOTS_CHAR_BAD_6,
			SLOTS_CHAR_BAD_7:

			return 34
		case SLOTS_CHAR_SPECIAL_MEDIUM:
			return 505
		case SLOTS_CHAR_SPECIAL_EPIC:
			return 1984
		}
	}

	otherChars := []rune{SLOTS_CHAR_SPECIAL_MEDIUM, SLOTS_CHAR_SPECIAL_EPIC}
	otherCharRewards := []int{69, 54}

	for i, r := range otherChars {
		for _, in := range inp {
			if r == in {
				return otherCharRewards[i]
			}
		}
	}

	return 0
}

func slots() []rune {
	slots := make([]rune, len(availableSlots))
	copy(slots, availableSlots)

	rand.Shuffle(len(slots), func(i, j int) {
		slots[i], slots[j] = slots[j], slots[i]
	})

	return slots
}
