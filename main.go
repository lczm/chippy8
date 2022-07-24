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

		// b1 := uint16(buffer[programCounter])
		// b2 := uint16(buffer[programCounter+1])

		// Shift the first op left by 8 bits,
		// which will add 8 bits to the end
		// do and OR operation, to add 8 more bits
		// to the 8 bits that just got added
		// op := uint16(b1<<8) | uint16(b2)
		// programCounter += 2
	}

	CloseDisplay()
}
