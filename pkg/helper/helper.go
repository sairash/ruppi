package helper

import "strings"

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
