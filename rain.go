package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"golang.org/x/term"
)

var charset = "@#$%&?ABCDEF0123456789"

//lint:ignore U1000 awesome way to pa
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

type Tty struct{}

type Glyph struct {
	r string
	x uint
	y uint
	v int
}

func clear_screen() {
	fmt.Print("\x1b[2J")
}
func cursor_to_position(x uint, y uint) {
	fmt.Printf("\x1b[%d;%dH", y+1, x+1)
}
func get_screen() (int, int) {
	w, h, err := term.GetSize(0)
	if err != nil {
		return 0, 0
	}
	return w, h
}

func main() {
	runtime.LockOSThread()
	fmt.Printf("\x1b[?25l")
	defer func() {
		fmt.Printf("\x1b[?25h")
	}()
	w, h := get_screen()
	clear_screen()
	glyphs := make([]*Glyph, (w * h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			glyphs[x+(w*y)] = &Glyph{
				r: " ",
				x: uint(x),
				y: uint(y),
				v: 0,
			}
			cursor_to_position(uint(x), uint(y))
			fmt.Printf("%s", s("\x1b[38;2;0;0;0m%s", glyphs[x+(w*y)].r))
		}
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	lck := sync.Mutex{}
	charset_len := len(charset)
	go func() {
		for {
			cursor_to_position(0, 0)
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					lck.Lock()
					idx := x + (w * y)
					g := glyphs[idx]
					g.v -= 1
					if g.v <= 0 {
						g.v = 0
					}
					if rand.Intn(10) == 9 {
						r := rand.Intn(charset_len)
						g.r = charset[r : r+1]
					}
					glyphs[idx] = g
					lck.Unlock()
					cursor_to_position(uint(x), uint(y))
					fmt.Printf("%s", s("\x1b[38;2;%d;%d;%dm%s", 0, g.v, 0, g.r))
					time.Sleep(1 * time.Microsecond)
				}
				cursor_to_position(0, uint(y))
			}
			time.Sleep(16 * time.Millisecond)
		}
	}()
	wg.Add(1)
	go func() {
		for {
			go func() {
				x := rand.Intn(w)
				for y := 0; y < h; y++ {
					r := rand.Intn(charset_len)
					lck.Lock()
					idx := x + (w * y)
					g := glyphs[idx]
					g.r = charset[r : r+1]
					g.v = 255
					glyphs[idx] = g
					lck.Unlock()
					time.Sleep(35 * time.Millisecond)
				}
			}()
			time.Sleep(105 * time.Millisecond)
		}
	}()
	wg.Wait()
	/* for {
		go func() {
			x := rand.Intn(w)
			speed := rand.Intn(50) + 50
			for y := 0; y < h; y++ {
				idx := x + (y * w)
				if idx >= len(glyphs) {
					continue
				}
				g := glyphs[x+(y*w)]
				g.color.R = 255
				g.color.B = 255
				g.color.G = 255
				r := rand.Intn(len(charset))
				g.r = charset[r : r+1]
				cursor_to_position(uint(x), uint(y))
				fmt.Printf("\b%s%s", prep(g.color), g.r)
				go func(g Glyph, x, y, w int) {
					g.color.R = 0
					g.color.B = 0
					for {
						g.color.G = g.color.G - 1
						if rand.Intn(40) == 40 {
							r := rand.Intn(len(charset))
							g.r = charset[r : r+1]
						}
						cursor_to_position(uint(x), uint(y))
						fmt.Printf("\b%s%s", prep(g.color), g.r)
						time.Sleep(100 * time.Millisecond)
						if g.color.G == 0 {
							break
						}
					}
					g.color.R = 0
					g.color.G = 0
					g.color.B = 0
					//glyphs[x+(y*w)] = &g

				}(*g, x, y, w)
				time.Sleep(time.Duration(speed) * time.Millisecond)
			}
		}()

		time.Sleep(360 * time.Millisecond)
	} */

}
