// this is modified from this project:
// https://github.com/mattn/go-sixel
package sixel

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"ruppi/pkg/helper"
	"strings"

	"github.com/soniakeys/quant/median"
	xdraw "golang.org/x/image/draw" // For high-quality image resizing
)

// Encoder encode image to sixel format
type Encoder struct {
	w strings.Builder
	// Dither, if true, will dither the image when generating a paletted version
	// using the Floyd–Steinberg dithering algorithm.
	Dither bool

	// Width is the maximum width to draw to.
	Width int
	// Height is the maximum height to draw to.
	Height int

	// Colors sets the number of colors for the encoder to quantize if needed.
	// If the value is below 2 (e.g. the zero value), then 255 is used.
	// A color is always reserved for alpha, so 2 colors give you 1 color.
	Colors int
}

// EncodeFromURL fetches an image from URL and returns sixel-encoded string
// maxWidth and maxHeight control the maximum dimensions (use 0 for defaults)
func EncodeFromURL(url string, maxWidth, maxHeight int) (string, error) {
	if maxWidth <= 0 {
		maxWidth = 400
	}
	if maxHeight <= 0 {
		maxHeight = 300
	}

	e := &Encoder{
		w:      strings.Builder{},
		Width:  maxWidth,
		Height: maxHeight,
		Colors: 255,
	}

	if err := e.EncodeFromUrl(url); err != nil {
		return "", err
	}
	return e.w.String(), nil
}

const (
	specialChNr = byte(0x6d)
	specialChCr = byte(0x64)
)

func (e *Encoder) EncodeFromUrl(url string) error {
	img, err := helper.ImageFromURL(url)
	if err != nil {
		return err
	}

	return e.Encode(img)
}

// resizeImage scales an image to fit within maxWidth and maxHeight while preserving aspect ratio
func resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// If image is already smaller than max dimensions, return as-is
	if origWidth <= maxWidth && origHeight <= maxHeight {
		return img
	}

	// Calculate scaling factor to fit within max dimensions
	scaleW := float64(maxWidth) / float64(origWidth)
	scaleH := float64(maxHeight) / float64(origHeight)
	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}

	newWidth := int(float64(origWidth) * scale)
	newHeight := int(float64(origHeight) * scale)

	// Ensure minimum dimensions
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Use bilinear interpolation for good quality with reasonable speed
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, bounds, xdraw.Over, nil)

	return dst
}

// Encode do encoding
func (e *Encoder) Encode(img image.Image) error {
	nc := e.Colors // (>= 2, 8bit, index 0 is reserved for transparent key color)
	if nc < 2 {
		nc = 255
	}

	// Resize image first if max dimensions are set (major performance optimization)
	if e.Width > 0 || e.Height > 0 {
		maxW := e.Width
		maxH := e.Height
		if maxW <= 0 {
			maxW = 800 // Default max width
		}
		if maxH <= 0 {
			maxH = 600 // Default max height
		}
		img = resizeImage(img, maxW, maxH)
	}

	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	if width == 0 || height == 0 {
		return nil
	}

	// Update encoder dimensions to actual resized dimensions
	e.Width = width
	e.Height = height

	var paletted *image.Paletted

	// fast path for paletted images
	if p, ok := img.(*image.Paletted); ok && len(p.Palette) < int(nc) {
		paletted = p
	} else {
		// make adaptive palette using median cut alogrithm
		q := median.Quantizer(nc - 1)
		paletted = q.Paletted(img)

		if e.Dither {
			// copy source image to new image with applying floyd-stenberg dithering
			draw.FloydSteinberg.Draw(paletted, img.Bounds(), img, image.Point{})
		} else {
			draw.Draw(paletted, img.Bounds(), img, image.Point{}, draw.Over)
		}
	}

	// DECSIXEL Introducer(\033P0;0;8q) + DECGRA ("1;1;W;H): Set Raster Attributes

	if _, err := e.w.WriteString(fmt.Sprintf("\033P0;0;8q\"1;1;%d;%d", width, height)); err != nil {
		return err
	}

	for n, v := range paletted.Palette {
		r, g, b, _ := v.RGBA()
		r = r * 100 / 0xFFFF
		g = g * 100 / 0xFFFF
		b = b * 100 / 0xFFFF

		if _, err := e.w.WriteString(fmt.Sprintf("#%d;2;%d;%d;%d", n+1, r, g, b)); err != nil {
			return err
		}

		// DECGCI (#): Graphics Color Introducer
	}

	buf := make([]byte, width*nc)
	cset := make([]bool, nc)
	ch0 := specialChNr
	for z := 0; z < (height+5)/6; z++ {
		// DECGNL (-): Graphics Next Line
		if z > 0 {
			if err := e.w.WriteByte(0x2d); err != nil {
				return err
			}
		}
		for p := range 6 {
			y := z*6 + p
			for x := 0; x < width; x++ {
				_, _, _, alpha := img.At(x, y).RGBA()
				if alpha != 0 {
					idx := paletted.ColorIndexAt(x, y) + 1
					cset[idx] = false // mark as used
					buf[width*int(idx)+x] |= 1 << uint(p)
				}
			}
		}
		for n := 1; n < nc; n++ {
			if cset[n] {
				continue
			}
			cset[n] = true
			// DECGCR ($): Graphics Carriage Return
			if ch0 == specialChCr {
				if err := e.w.WriteByte(0x24); err != nil {
					return err
				}
			}
			// select color (#%d)
			if n >= 100 {
				digit1 := n / 100
				digit2 := (n - digit1*100) / 10
				digit3 := n % 10
				c1 := byte(0x30 + digit1)
				c2 := byte(0x30 + digit2)
				c3 := byte(0x30 + digit3)
				if _, err := e.w.Write([]byte{0x23, c1, c2, c3}); err != nil {
					return nil
				}
			} else if n >= 10 {
				c1 := byte(0x30 + n/10)
				c2 := byte(0x30 + n%10)
				if _, err := e.w.Write([]byte{0x23, c1, c2}); err != nil {
					return err
				}
			} else {
				if _, err := e.w.Write([]byte{0x23, byte(0x30 + n)}); err != nil {
					return err
				}
			}
			cnt := 0
			for x := 0; x < width; x++ {
				// make sixel character from 6 pixels
				ch := buf[width*n+x]
				buf[width*n+x] = 0
				if ch0 < 0x40 && ch != ch0 {
					// output sixel character
					s := 63 + ch0
					for ; cnt > 255; cnt -= 255 {
						if _, err := e.w.Write([]byte{0x21, 0x32, 0x35, 0x35, s}); err != nil {
							return err
						}
					}
					if cnt == 1 {
						if err := e.w.WriteByte(s); err != nil {
							return err
						}
					} else if cnt == 2 {
						if _, err := e.w.Write([]byte{s, s}); err != nil {
							return err
						}
					} else if cnt == 3 {
						if _, err := e.w.Write([]byte{s, s, s}); err != nil {
							return err
						}
					} else if cnt >= 100 {
						digit1 := cnt / 100
						digit2 := (cnt - digit1*100) / 10
						digit3 := cnt % 10
						c1 := byte(0x30 + digit1)
						c2 := byte(0x30 + digit2)
						c3 := byte(0x30 + digit3)
						// DECGRI (!): - Graphics Repeat Introducer
						if _, err := e.w.Write([]byte{0x21, c1, c2, c3, s}); err != nil {
							return err
						}
					} else if cnt >= 10 {
						c1 := byte(0x30 + cnt/10)
						c2 := byte(0x30 + cnt%10)
						// DECGRI (!): - Graphics Repeat Introducer
						if _, err := e.w.Write([]byte{0x21, c1, c2, s}); err != nil {
							return err
						}
					} else if cnt > 0 {
						// DECGRI (!): - Graphics Repeat Introducer
						if _, err := e.w.Write([]byte{0x21, byte(0x30 + cnt), s}); err != nil {
							return err
						}
					}
					cnt = 0
				}
				ch0 = ch
				cnt++
			}
			if ch0 != 0 {
				// output sixel character
				s := 63 + ch0
				for ; cnt > 255; cnt -= 255 {
					if _, err := e.w.Write([]byte{0x21, 0x32, 0x35, 0x35, s}); err != nil {
						return err
					}
				}
				if cnt == 1 {
					if err := e.w.WriteByte(s); err != nil {
						return err
					}
				} else if cnt == 2 {
					if _, err := e.w.Write([]byte{s, s}); err != nil {
						return err
					}
				} else if cnt == 3 {
					if _, err := e.w.Write([]byte{s, s, s}); err != nil {
						return err
					}
				} else if cnt >= 100 {
					digit1 := cnt / 100
					digit2 := (cnt - digit1*100) / 10
					digit3 := cnt % 10
					c1 := byte(0x30 + digit1)
					c2 := byte(0x30 + digit2)
					c3 := byte(0x30 + digit3)
					// DECGRI (!): - Graphics Repeat Introducer
					if _, err := e.w.Write([]byte{0x21, c1, c2, c3, s}); err != nil {
						return err
					}
				} else if cnt >= 10 {
					c1 := byte(0x30 + cnt/10)
					c2 := byte(0x30 + cnt%10)
					// DECGRI (!): - Graphics Repeat Introducer
					if _, err := e.w.Write([]byte{0x21, c1, c2, s}); err != nil {
						return err
					}
				} else if cnt > 0 {
					// DECGRI (!): - Graphics Repeat Introducer
					if _, err := e.w.Write([]byte{0x21, byte(0x30 + cnt), s}); err != nil {
						return err
					}
				}
			}
			ch0 = specialChCr
		}
	}
	// string terminator(ST)
	if _, err := e.w.Write([]byte{0x1b, 0x5c}); err != nil {
		return err
	}

	return nil
}

// Decoder decode sixel format into image
type Decoder struct {
	r io.Reader
}

// NewDecoder return new instance of Decoder
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

// Decode do decoding from image
func (e *Decoder) Decode(img *image.Image) error {
	buf := bufio.NewReader(e.r)
	_, err := buf.ReadBytes('\x1B')
	if err != nil {
		if err == io.EOF {
			err = nil
		}
		return err
	}
	c, err := buf.ReadByte()
	if err != nil {
		return err
	}
	switch c {
	case 'P':
		_, err := buf.ReadString('q')
		if err != nil {
			return err
		}
	default:
		return errors.New("Invalid format: illegal header")
	}
	colors := map[uint]color.Color{
		// 16 predefined color registers of VT340
		0:  sixelRGB(0, 0, 0),
		1:  sixelRGB(20, 20, 80),
		2:  sixelRGB(80, 13, 13),
		3:  sixelRGB(20, 80, 20),
		4:  sixelRGB(80, 20, 80),
		5:  sixelRGB(20, 80, 80),
		6:  sixelRGB(80, 80, 20),
		7:  sixelRGB(53, 53, 53),
		8:  sixelRGB(26, 26, 26),
		9:  sixelRGB(33, 33, 60),
		10: sixelRGB(60, 26, 26),
		11: sixelRGB(33, 60, 33),
		12: sixelRGB(60, 33, 60),
		13: sixelRGB(33, 60, 60),
		14: sixelRGB(60, 60, 33),
		15: sixelRGB(80, 80, 80),
	}
	dx, dy := 0, 0
	dw, dh, w, h := 0, 0, 200, 200
	pimg := image.NewNRGBA(image.Rect(0, 0, w, h))
	var cn uint
data:
	for {
		c, err = buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		if c == '\r' || c == '\n' || c == '\b' {
			continue
		}
		switch c {
		case '\x1b':
			c, err = buf.ReadByte()
			if err != nil {
				return err
			}
			if c == '\\' {
				break data
			}
		case '"':
			params := []int{}
			for {
				var i int
				n, err := fmt.Fscanf(buf, "%d", &i)
				if err == io.EOF {
					return err
				}
				if n == 0 {
					i = 0
				}
				params = append(params, i)
				c, err = buf.ReadByte()
				if err != nil {
					return err
				}
				if c != ';' {
					break
				}
			}
			if len(params) >= 4 {
				if w < params[2] {
					w = params[2]
				}
				if h < params[3]+6 {
					h = params[3] + 6
				}
				pimg = expandImage(pimg, w, h)
			}
			err = buf.UnreadByte()
			if err != nil {
				return err
			}
		case '$':
			dx = 0
		case '!':
			err = buf.UnreadByte()
			if err != nil {
				return err
			}
			var nc uint
			var c byte
			n, err := fmt.Fscanf(buf, "!%d%c", &nc, &c)
			if err != nil {
				return err
			}
			if n != 2 || c < '?' || c > '~' {
				return fmt.Errorf("invalid format: illegal repeating data tokens '!%d%c'", nc, c)
			}
			if w <= dx+int(nc)-1 {
				w *= 2
				pimg = expandImage(pimg, w, h)
			}
			m := byte(1)
			c -= '?'
			for p := range 6 {
				if c&m != 0 {
					for q := 0; q < int(nc); q++ {
						pimg.Set(dx+q, dy+p, colors[cn])
					}
					if dh < dy+p+1 {
						dh = dy + p + 1
					}
				}
				m <<= 1
			}
			dx += int(nc)
			if dw < dx {
				dw = dx
			}
		case '-':
			dx = 0
			dy += 6
			if h <= dy+6 {
				h *= 2
				pimg = expandImage(pimg, w, h)
			}
		case '#':
			err = buf.UnreadByte()
			if err != nil {
				return err
			}
			var nc, csys uint
			var r, g, b uint
			var c byte
			n, err := fmt.Fscanf(buf, "#%d%c", &nc, &c)
			if err != nil {
				return err
			}
			if n != 2 {
				return fmt.Errorf("invalid format: illegal color specifier '#%d%c'", nc, c)
			}
			if c == ';' {
				n, err := fmt.Fscanf(buf, "%d;%d;%d;%d", &csys, &r, &g, &b)
				if err != nil {
					return err
				}
				if n != 4 {
					return fmt.Errorf("invalid format: illegal color specifier '#%d;%d;%d;%d;%d'", nc, csys, r, g, b)
				}
				if csys == 1 {
					colors[nc] = sixelHLS(r, g, b)
				} else {
					colors[nc] = sixelRGB(r, g, b)
				}
			} else {
				err = buf.UnreadByte()
				if err != nil {
					return err
				}
			}
			cn = nc
			if _, ok := colors[cn]; !ok {
				return fmt.Errorf("invalid format: undefined color number %d", cn)
			}
		default:
			if c >= '?' && c <= '~' {
				if w <= dx {
					w *= 2
					pimg = expandImage(pimg, w, h)
				}
				m := byte(1)
				c -= '?'
				for p := range 6 {
					if c&m != 0 {
						pimg.Set(dx, dy+p, colors[cn])
						if dh < dy+p+1 {
							dh = dy + p + 1
						}
					}
					m <<= 1
				}
				dx++
				if dw < dx {
					dw = dx
				}
				break
			}
			return errors.New("invalid format: illegal data tokens")
		}
	}
	rect := image.Rect(0, 0, dw, dh)
	tmp := image.NewNRGBA(rect)
	draw.Draw(tmp, rect, pimg, image.Point{0, 0}, draw.Src)
	*img = tmp
	return nil
}

func sixelRGB(r, g, b uint) color.Color {
	return color.NRGBA{uint8(r * 0xFF / 100), uint8(g * 0xFF / 100), uint8(b * 0xFF / 100), 0xFF}
}

func sixelHLS(h, l, s uint) color.Color {
	var r, g, b, max, min float64

	/* https://wikimedia.org/api/rest_v1/media/math/render/svg/17e876f7e3260ea7fed73f69e19c71eb715dd09d */
	/* https://wikimedia.org/api/rest_v1/media/math/render/svg/f6721b57985ad83db3d5b800dc38c9980eedde1d */
	if l > 50 {
		max = float64(l) + float64(s)*(1.0-float64(l)/100.0)
		min = float64(l) - float64(s)*(1.0-float64(l)/100.0)
	} else {
		max = float64(l) + float64(s*l)/100.0
		min = float64(l) - float64(s*l)/100.0
	}

	/* sixel hue color ring is roteted -120 degree from nowdays general one. */
	h = (h + 240) % 360

	/* https://wikimedia.org/api/rest_v1/media/math/render/svg/937e8abdab308a22ff99de24d645ec9e70f1e384 */
	switch h / 60 {
	case 0: /* 0 <= hue < 60 */
		r = max
		g = min + (max-min)*(float64(h)/60.0)
		b = min
	case 1: /* 60 <= hue < 120 */
		r = min + (max-min)*(float64(120-h)/60.0)
		g = max
		b = min
	case 2: /* 120 <= hue < 180 */
		r = min
		g = max
		b = min + (max-min)*(float64(h-120)/60.0)
	case 3: /* 180 <= hue < 240 */
		r = min
		g = min + (max-min)*(float64(240-h)/60.0)
		b = max
	case 4: /* 240 <= hue < 300 */
		r = min + (max-min)*(float64(h-240)/60.0)
		g = min
		b = max
	case 5: /* 300 <= hue < 360 */
		r = max
		g = min
		b = min + (max-min)*(float64(360-h)/60.0)
	default:
	}
	return sixelRGB(uint(r), uint(g), uint(b))
}

func expandImage(pimg *image.NRGBA, w, h int) *image.NRGBA {
	b := pimg.Bounds()
	if w < b.Max.X {
		w = b.Max.X
	}
	if h < b.Max.Y {
		h = b.Max.Y
	}
	tmp := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(tmp, b, pimg, image.Point{0, 0}, draw.Src)
	return tmp
}
