package helper

// Truncate adds 3(.) ie: ... at the end of the string
func TruncateString(input string, maxLength int, truncate bool) string {
	inputRune := []rune(input)

	if maxLength < 0 {
		return input
	}

	if len(inputRune) < maxLength {
		return input
	}

	end := ""
	if truncate {
		maxLength = maxLength - 3
		end = "..."
	}

	return string(inputRune[:maxLength]) + end
}
