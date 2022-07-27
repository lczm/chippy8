package chip8

import (
	"fmt"
	"math/rand"
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

var KeyMap = map[int32]byte{
	rl.KeyOne: 0x01, rl.KeyTwo: 0x02, rl.KeyThree: 0x03, rl.KeyFour: 0x0C,
	rl.KeyQ: 0x04, rl.KeyW: 0x05, rl.KeyE: 0x06, rl.KeyR: 0x0D,
	rl.KeyA: 0x07, rl.KeyS: 0x08, rl.KeyD: 0x09, rl.KeyF: 0x0E,
	rl.KeyZ: 0x0A, rl.KeyX: 0x00, rl.KeyC: 0x0B, rl.KeyV: 0x0F,
}

type Chip8 struct {
	scale        int
	ScaledWidth  int
	ScaledHeight int

	DrawFlag bool

	display [width][height]uint8 // display array
	memory  [4096]uint8          // 4k memory emulated

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
	key [16]bool
}

func New(scale int) *Chip8 {
	c8 := &Chip8{
		scale:          scale,
		ScaledWidth:    scale * width,
		ScaledHeight:   scale * height,
		DrawFlag:       false,
		programCounter: 0x200,
		indexRegister:  0,
		stackPointer:   0,
		delayTimer:     0,
		soundTimer:     0,
	}

	// Load fonts into memory
	for i := 0; i < 80; i++ {
		c8.memory[i] = C8FontSet[i]
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
			if c8.display[i][h] == 1 {
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

func (c8 *Chip8) Key(index byte, value bool) {
	c8.key[index] = value
}

// Emulate one cycle
func (c8 *Chip8) Cycle() {
	// Get opcode
	c8.opcode = (uint16(c8.memory[c8.programCounter]) << 8) |
		uint16(c8.memory[c8.programCounter+1])

	// Decode opcode
	// doing & 0xF000 makes it so that the switch case can go
	// on 0x10, 0x20 instead
	switch c8.opcode & 0xF000 {
	case 0x0000:
		switch c8.opcode {
		case 0x00E0: // 00E0 - Display - Clear the screen
			fmt.Println("0x00E0")
			for i := 0; i < width; i++ {
				for j := 0; j < height; j++ {
					c8.display[i][j] = 0
				}
			}
			c8.programCounter += 2
			c8.DrawFlag = true
		case 0x00EE: // 00EE - Flow - Return from subroutine
			fmt.Println("0x00EE")
			c8.programCounter = c8.stack[c8.stackPointer]
			c8.stackPointer--
			c8.programCounter += 2
		default:
			fmt.Printf("Cannot find opcode: %x\n", c8.opcode)
		}
	case 0x1000: // 1NNN - Flow - Jump to location NNN
		fmt.Println("0x1NNN")
		// Masked out by & 0xF000 at the start,
		// to mask it back in, use & 0x0FFF
		c8.programCounter = c8.opcode & 0x0FFF
	case 0x2000: // 2NNN - Flow - Call subroutine at NNN
		fmt.Println("0x2NNN")
		c8.stackPointer++
		c8.stack[c8.stackPointer] = c8.programCounter
		c8.programCounter = c8.opcode & 0x0FFF
	case 0x3000: // 3XKK - Cond - Skip next instruction if VX == NN
		fmt.Println("0x3XKK")
		if c8.registers[(c8.opcode&0x0F00)>>8] == uint8(c8.opcode)&0x00FF {
			c8.programCounter += 4
		} else {
			c8.programCounter += 2
		}
	case 0x4000: // 4XKK - Cond - Skip next instruction if VX != NN
		fmt.Println("0x4XKK")
		if c8.registers[(c8.opcode&0x0F00)>>8] != uint8(c8.opcode)&0x00FF {
			c8.programCounter += 4
		} else {
			c8.programCounter += 2
		}
	case 0x5000: // 5XY0 - Cond - Skip next instruction if VX == VY
		fmt.Println("0x5XY0")
		if c8.registers[(c8.opcode&0x0F00)>>8] == c8.registers[(c8.opcode&0x00F0)>>4] {
			c8.programCounter += 4
		} else {
			c8.programCounter += 2
		}
	case 0x6000: // 6XNN - Const - Sets VX to NN
		fmt.Println("0x6XNN")
		c8.registers[(c8.opcode&0x0F00)>>8] = uint8(c8.opcode) & 0x00FF
		c8.programCounter += 2
	case 0x7000: // 7XNN - 	Const - Adds NN to VX
		fmt.Println("0x7XNN")
		c8.registers[(c8.opcode&0x0F00)>>8] += uint8(c8.opcode) & 0x00FF
		c8.programCounter += 2
	case 0x8000:
		switch c8.opcode & 0x000F {
		case 0x0000: // 8XY0 - Assignment -  Set VX to VY
			fmt.Println("0x8XY0")
			c8.registers[(c8.opcode&0x0F00)>>8] = c8.registers[(c8.opcode&0x00F0)>>4]
			c8.programCounter += 2
		case 0x0001: // 8XY1 - BitOp - Sets VX to (VX OR VY), bitwise
			fmt.Println("0x8XY1")
			c8.registers[(c8.opcode&0x0F00)>>8] = c8.registers[(c8.opcode&0x0F00)>>8] | c8.registers[(c8.opcode&0x00F0)>>4]
			c8.programCounter += 2
		case 0x0002: // 8XY2 - BitOp - Sets VX to (VX AND VY), bitwise
			fmt.Println("0x8XY2")
			c8.registers[(c8.opcode&0x0F00)>>8] = c8.registers[(c8.opcode&0x0F00)>>8] & c8.registers[(c8.opcode&0x00F0)>>4]
			c8.programCounter += 2
		case 0x0003: // 8XY3 - BitOp - Sets VX to (VX XOR VY), bitwise
			fmt.Println("0x8XY3")
			c8.registers[(c8.opcode&0x0F00)>>8] = c8.registers[(c8.opcode&0x0F00)>>8] ^ c8.registers[(c8.opcode&0x00F0)>>4]
			c8.programCounter += 2
		case 0x0004: // 8XY4 - Math - Adds VY TO VX, VF set to 1 if there's a carry, to 0 if theres not
			fmt.Println("0x8XY4")
			if c8.registers[(c8.opcode&0x00F0)>>4] > 0xFF-c8.registers[(c8.opcode&0x0F00)>>8] {
				c8.registers[15] = 1
			} else {
				c8.registers[15] = 0
			}
			c8.registers[(c8.opcode&0x0F00)>>8] += c8.registers[(c8.opcode&0x00F0)>>4]
			c8.programCounter += 2
		case 0x0005: // 8XY5 - Math - Subtract VY FROM VX, VF set to 0 if there's a borrow, 1 when there is not
			fmt.Println("0x8XY5")
			if c8.registers[(c8.opcode&0x00F0)>>4] > 0xFF-c8.registers[(c8.opcode&0x0F00)>>8] {
				c8.registers[15] = 0
			} else {
				c8.registers[15] = 1
			}
			c8.registers[(c8.opcode&0x0F00)>>8] -= c8.registers[(c8.opcode&0x00F0)>>4]
			c8.programCounter += 2
		case 0x0006: // 8XY6 - BitOp - Store the least significant bit of VX in VF, then shift VX to the right by 1
			fmt.Println("0x8XY6")
			// Store least significant bit of VX in VF (1 is the least significant bit)
			c8.registers[15] = c8.registers[(c8.opcode&0x0F00)>>8] & 0x1
			// Shift VX to the right by 1
			c8.registers[(c8.opcode&0x0F00)>>8] = c8.registers[(c8.opcode&0x0F00)>>8] >> 1
			c8.programCounter += 2
		case 0x0007: // 8XY7 - Math - Sets VX to VF - VX, VF is set to 0 if there's a borrow, 1 when there is not
			fmt.Println("0x8XY7")
			if c8.registers[(c8.opcode&0x00F0)>>4] > 0xFF-c8.registers[(c8.opcode&0x0F00)>>8] {
				c8.registers[15] = 0
			} else {
				c8.registers[15] = 1
			}
			c8.registers[(c8.opcode&0x0F00)>>8] = c8.registers[(c8.opcode&0x00F0)>>4] - c8.registers[(c8.opcode&0x0F00)>>8]
			c8.programCounter += 2
		case 0x000E: // 8XYE - BitOp - Store the most significant bit of VX in VF, then shift VX to the left by 1
			fmt.Println("0x8XYE")
			// Store the most significant bit of VX in VF (7 is the most significant bit)
			c8.registers[15] = c8.registers[(c8.opcode&0x0F00)>>8] & 0x7
			c8.registers[(c8.opcode&0x0F00)>>8] = c8.registers[(c8.opcode&0x0F00)>>8] << 1
			c8.programCounter += 2
		}
	case 0x9000: // 9XY0 - Cond - Skip next instruction if VX != VY
		fmt.Println("0x9XY0")
		if c8.registers[(c8.opcode&0x0F00)>>8] != c8.registers[(c8.opcode&0x00F0)>>4] {
			c8.programCounter += 4
		} else {
			c8.programCounter += 2
		}
	case 0xA000: // ANNN - Mem - Sets address I = address NNN
		fmt.Println("0xA000")
		c8.indexRegister = c8.opcode & 0x0FFF
		c8.programCounter += 2
	case 0xB000: // BNNN - Flow - Jumps to address NNN + V0
		fmt.Println("0xB000")
		c8.programCounter = c8.opcode&0x0FFF + uint16(c8.registers[0])
	case 0xC000: // CXNN - Rand - Set VX = Rand(0-255) & NN
		fmt.Println("0xCXNN")
		c8.registers[(c8.opcode&0x0F00)>>8] = uint8(rand.Intn(256)) & uint8(c8.opcode&0x00FF)
		c8.programCounter += 2
	case 0xD000: // DXYN - Display - Draw sprite at coordinate VX, VY
		fmt.Println("0xDXYN")
		x := c8.registers[(c8.opcode&0x0F00)>>8]
		y := c8.registers[(c8.opcode&0x00F0)>>4]
		h := c8.opcode & 0x000F
		c8.registers[15] = 0

		var i uint16 = 0
		var j uint16 = 0
		for j = 0; j < h; j++ {
			pixel := c8.memory[c8.indexRegister+j]
			for i = 0; i < 8; i++ {
				if (pixel & (0x80 >> i)) != 0 {
					// if c8.display[(y + uint8(j))][x+uint8(i)] == 1 {
					if c8.display[(x + uint8(i))][y+uint8(j)] == 1 {
						c8.registers[15] = 1
					}
					c8.display[(x + uint8(i))][y+uint8(j)] ^= 1
				}
			}
		}

		c8.DrawFlag = true
		c8.programCounter += 2
	case 0xE000: // EX9E | EXA1
		switch c8.opcode & 0xFFF0 {
		case 0x000E: // EX9E - KeyOp - Skip next instruction if key stored in VX is pressed
			fmt.Println("0xEX9E")
			if c8.key[c8.registers[(c8.opcode&0x0F00)>>8]] {
				c8.programCounter += 4
			} else {
				c8.programCounter += 2
			}
		case 0x0001: // EXA1 - KeyOp - Skip next instruction if key stored in VX is NOT pressed
			fmt.Println("0xEXA1")
			if !c8.key[c8.registers[(c8.opcode&0x0F00)>>8]] {
				c8.programCounter += 4
			} else {
				c8.programCounter += 2
			}
		}
	case 0xF000: // FX07 | FX0A | FX15 | FX18 | FX1E | FX29 | FX33 | FX55 | FX65
		switch c8.opcode & 0x00FF {
		case 0x0007: // FX07 - Timer - Set VX to value of delay timer
			fmt.Println("0xFX07")
			c8.registers[(c8.opcode&0x0F00)>>8] = c8.delayTimer
			c8.programCounter += 2
		case 0x000A: // (TODO) FX0A - KeyOp - Wait for keypress, store in VX
			fmt.Println("0xFX0A")
			c8.programCounter += 2
		case 0x0015: // FX15 - Timer - Set delay timer to VX
			fmt.Println("0xFX15")
			c8.delayTimer = c8.registers[(c8.opcode&0x0F00)>>8]
			c8.programCounter += 2
		case 0x0018: // FX18 - Sound - Set sound timer to VX
			fmt.Println("0xFX18")
			c8.soundTimer = c8.registers[(c8.opcode&0x0F00)>>8]
			c8.programCounter += 2
		case 0x001E: // FX1E - MEM - Add VX to I
			fmt.Println("0xFX1E")
			c8.indexRegister += uint16(c8.registers[(c8.opcode&0x0F00)>>8])
			c8.programCounter += 2
		case 0x0029: // (TODO) FX29 - MEM - Set I to location of sprite for character in VX
			fmt.Println("0xFX29")
			c8.programCounter += 2
		case 0x0033: // (TODO) FX33 - BCD - Stores the binary-coded decimal representation of VX, with the most significant of three digits at the address in I, the middle digit at I plus 1, and the least significant digit at I plus 2
			fmt.Println("0xFX33")
			c8.programCounter += 2
		case 0x0055: // (TODO) FX55 - MEM - Stores from V0 to VX (including VX) in memory, starting at address I. The offset from I is increased by 1 for each value written, but I itself is left unmodified
			fmt.Println("0xFX55")
			c8.programCounter += 2
		case 0x0065: // (TODO) FX65 - MEM - Fills from V0 to VX (including VX) with values from memory, starting at address I. The offset from I is increased by 1 for each value read, but I itself is left unmodified
			fmt.Println("0xFX65")
			c8.programCounter += 2
		default:
			fmt.Printf("Cannot find opcode: %x\n", c8.opcode)
		}
	default:
		fmt.Printf("Cannot find opcode: %x\n", c8.opcode)
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
