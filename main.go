package main

import (
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/lczm/chippy8/chip8"
)

const (
	fps   = 60
	scale = 10
)

func GetHex(ui uint16) string {
	return fmt.Sprintf("%02x", ui)
}

func InitDisplay(scaledWidth, scaledHeight int) {
	// Set the init window
	rl.InitWindow(int32(scaledWidth),
		int32(scaledHeight), "chippy8")

	// Set the fps
	rl.SetTargetFPS(fps)
}

func CloseDisplay() {
	rl.CloseWindow()
}

func main() {
	args := os.Args

	if len(args) != 2 {
		fmt.Println("Please call ./chip8 {rom}")
		os.Exit(0)
	}

	rom := os.Args[1]
	if _, err := os.Open(rom); err != nil {
		fmt.Println("{rom} does not exist")
		os.Exit(0)
	}

	c8 := chip8.New(scale)
	c8.ReadRom(rom)

	InitDisplay(c8.ScaledWidth, c8.ScaledHeight)

	for !rl.WindowShouldClose() {
		c8.Cycle()

		if c8.DrawFlag {
			rl.BeginDrawing()

			// Clear the background to be white
			// so that a new screen can be drawn
			rl.ClearBackground(rl.RayWhite)

			c8.Draw()

			rl.EndDrawing()
		}

		// Update keys array
		for k, v := range chip8.KeyMap {
			if rl.IsKeyDown(k) {
				c8.Key(v, true)
			} else if rl.IsKeyUp(k) {
				c8.Key(v, false)
			}
		}
	}

	CloseDisplay()
}
