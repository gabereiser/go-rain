package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

var charset = "@#$%&ABCDEF0123456789"

func colorstr(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")

	}
	return
}
func s(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
func prep(c color.RGBA) string {
	return s("\x1b[38;2;%d;%d;%dm", c.R, c.G, c.B)
}

type Tty struct{}

type Glyph struct {
	r     string
	color color.RGBA
}

func min(m int, v int) uint8 {
	if v > m {
		return uint8(m)
	}
	if v > 255 {
		v = 255
	}
	return uint8(v)
}
func clear_screen() {
	fmt.Print("\x1b[2J")
}
func cursor_to_position(x uint, y uint) {
	fmt.Printf("\x1b[%d;%dH", y+1, x+1)
}
func get_screen() (int, int) {
	w, h, err := terminal.GetSize(0)
	if err != nil {
		return 0, 0
	}
	return w, h
}

func main() {
	gr, e := colorstr("#000000")
	if e != nil {
		panic(e)
	}
	clear_screen()
	w, h := get_screen()
	clear_screen()
	glyphs := make([]*Glyph, (w * h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			glyphs[x+(w*y)] = &Glyph{
				r:     ".",
				color: gr,
			}
			cursor_to_position(uint(x), uint(y))
			fmt.Printf("%s%s", prep(glyphs[x+(w*y)].color), glyphs[x+(w*y)].r)
		}
	}
	for {
		go func() {
			x := rand.Intn(w)
			for y := 0; y < 1024; y++ {
				cursor_to_position(uint(x), uint(y))
				idx := x + (y * w)
				if idx >= len(glyphs) {
					continue
				}
				g := glyphs[x+(y*w)]
				g.color.R = 0
				g.color.B = 0
				g.color.G = 255
				r := rand.Intn(len(charset))
				g.r = charset[r : r+1]
				fmt.Printf("\b%s%s", prep(g.color), g.r)
				go func(x, y, w int) {
					for {
						g := glyphs[x+(y*w)]
						cursor_to_position(uint(x), uint(y))
						g.color.G -= 1
						r := rand.Intn(len(charset))
						g.r = charset[r : r+1]
						glyphs[x+(y*w)] = g
						fmt.Printf("\b%s%s", prep(g.color), g.r)
						time.Sleep(66 * time.Millisecond)
						if g.color.G == 0 {
							break
						}
					}

				}(x, y, w)
				time.Sleep(16 * time.Millisecond)
			}
		}()

		time.Sleep(160 * time.Millisecond)
	}

}
