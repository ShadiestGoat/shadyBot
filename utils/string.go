package utils

import (
	"fmt"
	"math"
	"strings"
)

var shortSuffixes = []string{"", "k", "m", "b", "t"}

func FormatShortInt(v int) string {
	if v < 1000 {
		return fmt.Sprint(v)
	}
	exp := math.Floor(math.Log10(float64(v)) / 3)
	full := float64(v) / math.Pow(1000, exp)
	top := fmt.Sprint(math.Floor(full))
	t := "%." + fmt.Sprint(3-len(top)) + "f"
	return fmt.Sprintf(t+"%v", full, shortSuffixes[int(exp)])
}

func FormatMinutes(minutes int) string {
	h := int(math.Floor(float64(minutes) / 60))
	d := int(math.Floor(float64(h) / 24))
	h %= 24

	m := minutes - h*60

	vcStr := []string{}

	if d != 0 {
		suf := ""
		if d != 1 {
			suf = "s"
		}

		vcStr = append(vcStr, fmt.Sprintf("%d day", d)+suf)
	}
	if h != 0 {
		vcStr = append(vcStr, fmt.Sprintf("%dh", h))
	}
	if m != 0 && d == 0 {
		vcStr = append(vcStr, fmt.Sprintf("%dm", m))
	}

	if len(vcStr) == 0 {
		return "No time"
	}

	return strings.Join(vcStr, " & ")
}
