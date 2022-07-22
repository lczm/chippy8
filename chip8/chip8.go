package chip8

import (
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	width  = 64
	height = 32
)

var C8FontSet = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, //0
	0x20, 0x60, 0x20, 0x20, 0x70, //1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, //2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, //3
	0x90, 0x90, 0xF0, 0x10, 0x10, //4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, //5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, //6
	0xF0, 0x10, 0x20, 0x40, 0x40, //7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, //8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, //9
	0xF0, 0x90, 0xF0, 0x90, 0x90, //A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, //B
	0xF0, 0x80, 0x80, 0x80, 0xF0, //C
	0xE0, 0x90, 0x90, 0x90, 0xE0, //D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, //E
	0xF0, 0x80, 0xF0, 0x80, 0x80, //F
}

type Chip8 struct {
	scale        int
	ScaledWidth  int
	ScaledHeight int

	DrawFlag bool

	display [width][height]bool // display array
	memory  [4096]uint8         // 4k memory emulated

	opcode         uint16
	registers      [16]uint8 // 15 8-bit registers, 16th is 'carry flag'
	indexRegister  uint16
	programCounter uint16

	// Timers, countdown to 0
	delayTimer uint8
	soundTimer uint8 // Sound the system buzzer when soundTimer reaches 0

	// The internal stack of the chip8 vm
	stack        [16]uint16
	stackPointer uint8

	// Theres only 16 keys in chip8, use this to
	// store the state of the key
	key [16]uint8
}

func New(scale int) *Chip8 {
	c8 := &Chip8{
		scale:          scale,
		ScaledWidth:    scale * width,
		ScaledHeight:   scale * height,
		DrawFlag:       true,
		programCounter: 0x200,
		indexRegister:  0,
		stackPointer:   0,
		delayTimer:     0,
		soundTimer:     0,
	}

	// Load fonts
	for i := 0; i < 80; i++ {
		c8.memory[i] = C8FontSet[i]
	}

	// Temporary test
	for i := 0; i < width; i++ {
		c8.display[i][0] = true
	}

	return c8
}

func (c8 *Chip8) ReadRom(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, fileStat.Size())
	if _, err := file.Read(buffer); err != nil {
		panic(err)
	}

	for i := 0; i < len(buffer); i++ {
		c8.memory[i+512] = buffer[i]
	}
}

func (c8 *Chip8) Draw() {
	for i := 0; i < width; i++ {
		for h := 0; h < height; h++ {
			// If the array is "active", draw a "pixel"
			// TODO : Draw a rectangle instead
			if c8.display[i][h] {
				// For every pixel, render its "scaled" version
				for ws := 0; ws < c8.scale; ws++ {
					for wh := 0; wh < c8.scale; wh++ {
						rl.DrawPixel(int32((c8.scale*i)+ws),
							int32((c8.scale*h)+wh), rl.Red)
					}
				}
			}
		}
	}
}

// Emulate one cycle
func (c8 *Chip8) Cycle() {
	// Get opcode
	c8.opcode = (uint16(c8.memory[c8.programCounter]) << 8) |
		uint16(c8.memory[c8.programCounter+1])

	// Decode opcode
	switch c8.opcode & 0xF000 {
	case 0xA000: // ANNN
		c8.indexRegister = c8.opcode & 0x0FFF
		c8.programCounter += 2
	default:
		fmt.Println("Cannot find opcode")
	}

	// Update delay timer
	if c8.delayTimer > 0 {
		c8.delayTimer--
	}

	// Update sound timer
	if c8.soundTimer > 0 {
		if c8.soundTimer == 1 {
			fmt.Println("Beep")
		}
		c8.soundTimer--
	}
}
