package helper

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// Truncate adds 3(.) ie: ... at the end of the string
func TruncateString(input string, maxLength int, truncate bool) string {
	if maxLength < 0 {
		return input
	}

	inputRune := []rune(input)
	lengthOfInput := len(inputRune)

	if lengthOfInput == maxLength {
	}
	if lengthOfInput <= maxLength {
		return input + strings.Repeat(" ", maxLength-lengthOfInput)
	}

	end := ""
	if truncate {
		maxLength = maxLength - 3
		end = "..."
	}

	return string(inputRune[:maxLength]) + end
}

// Generates a random color and figures out to use black or white for foreground color
func ColorGenerator() (string, string) {
	color := "#"
	foregroundColor := "#ffffff"

	sensetivity := []float64{0.2126, 0.7152, 0.0722}
	var relativeLuminaceY float64

	for i := 0; i < 3; i++ {
		randColor := rand.Intn(255)
		color += fmt.Sprintf("%x", randColor)

		s := float64(randColor) / 255

		var lc float64
		if s <= 0.04045 {
			lc = s / 12.92
		} else {
			lc = math.Pow(((s + 0.055) / 1.055), 2.4)
		}

		relativeLuminaceY += sensetivity[i] * lc
	}

	if relativeLuminaceY > 0.179 {
		foregroundColor = "#000000"
	}

	return color, foregroundColor
}
