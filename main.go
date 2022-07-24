package main

import (
	"fmt"

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
	c8 := chip8.New(scale)
	c8.ReadRom("./roms/c8_test.c8")

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
	}

	CloseDisplay()
}
