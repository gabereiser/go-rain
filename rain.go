package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"golang.org/x/term"
)

var charset = "@#$%&?ABCDEF0123456789"

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
					r, b := 0, 0
					if g.v > 250 {
						r, b = g.v-50, g.v-50
					}
					fmt.Printf("%s", s("\x1b[38;2;%d;%d;%dm%s", r, g.v, b, g.r))
					time.Sleep(100 * time.Nanosecond) // give the tty a little time to catch up to us.
				}
				cursor_to_position(0, uint(y))
			}
			time.Sleep(16 * time.Millisecond) // 1000ms / 60fps = 16.666666
		}
	}()
	wg.Add(1)
	go func() {
		for {
			go func() {
				x := rand.Intn(w)
				s := rand.Intn(250) + 50
				for y := 0; y < h; y++ {
					r := rand.Intn(charset_len)
					lck.Lock()
					idx := x + (w * y)
					g := glyphs[idx]
					g.r = charset[r : r+1]
					g.v = 255
					glyphs[idx] = g
					lck.Unlock()
					time.Sleep(time.Duration(s) * time.Millisecond) // look like it's struggling...
				}
			}()
			time.Sleep(100 * time.Millisecond) // don't spawn too many too fast...
		}
	}()
	wg.Wait()
}
