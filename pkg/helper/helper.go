package helper

import (
	"fmt"
	"image"
	// Import image format decoders
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	_ "golang.org/x/image/webp" // WebP support
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

// Enhanced color palettes for better visual consistency
var colorPalettes = [][]string{
	// Modern pastels
	{"#ff9a9e", "#fecfef", "#fecfef"},
	{"#a8edea", "#fed6e3", "#d299c2"},
	{"#ffecd2", "#fcb69f", "#ff8a80"},
	{"#ff9a9e", "#fad0c4", "#ffd1ff"},

	// Ocean blues
	{"#667eea", "#764ba2", "#f093fb"},
	{"#4facfe", "#00f2fe", "#43e97b"},
	{"#30cfd0", "#91a7ff", "#a8e6cf"},

	// Warm sunset
	{"#fa709a", "#fee140", "#fa8072"},
	{"#ffeaa7", "#fab1a0", "#fd79a8"},
	{"#fdcb6e", "#fd79a8", "#6c5ce7"},

	// Cool mint
	{"#a8e6cf", "#dcedc1", "#ffd3a5"},
	{"#84fab0", "#8fd3f4", "#d4fc79"},

	// Purple gradients
	{"#c471ed", "#f64f59", "#12c2e9"},
	{"#667eea", "#764ba2", "#f093fb"},
}

func ColorGenerator() (string, string) {
	palette := colorPalettes[rand.Intn(len(colorPalettes))]
	color := palette[rand.Intn(len(palette))]

	foregroundColor := "#ffffff"

	hexColor := color[1:]

	r, _ := fmt.Sscanf(hexColor[0:2], "%x", new(int))
	g, _ := fmt.Sscanf(hexColor[2:4], "%x", new(int))
	b, _ := fmt.Sscanf(hexColor[4:6], "%x", new(int))

	sensitivity := []float64{0.2126, 0.7152, 0.0722}
	values := []int{r, g, b}

	var relativeLuminance float64
	for i, val := range values {
		s := float64(val) / 255.0
		var linearized float64
		if s <= 0.04045 {
			linearized = s / 12.92
		} else {
			linearized = math.Pow((s+0.055)/1.055, 2.4)
		}
		relativeLuminance += sensitivity[i] * linearized
	}

	if relativeLuminance > 0.5 {
		foregroundColor = "#000000"
	}

	return color, foregroundColor
}

func ResolveURL(baseStr, input string) (string, error) {
	base, err := url.Parse(baseStr)
	if err != nil {
		return "", err
	}

	ref, err := url.Parse(input)
	if err != nil {
		return "", err
	}

	// ResolveReference handles ALL cases correctly
	resolved := base.ResolveReference(ref)
	return resolved.String(), nil
}

// imageHTTPClient is a client with reasonable timeouts for image fetching
var imageHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

func ImageFromURL(url string) (image.Image, error) {
	resp, err := imageHTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: HTTP %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}
