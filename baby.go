package main

import (
	"fmt"
	"strconv"
)

const (
	wordSize = 32
	words    = 32
)

// Instruction opcodes
const (
	JMP  = iota // Jump (000)
	SUB         // Subtract (001)
	LDN         // Load negative (010)
	CMP         // Compare (011)
	JRP         // Jump relative (100)
	SUB2        // Subtract (101)
	STO         // Store (110)
	STP         // Stop (111)
)

type instruction struct {
	op   int32
	data int32
}

func instFromWord(word int32) *instruction {
	i := &instruction{
		op:   (word & 0x0000F000) >> 12,
		data: word & 0x00000FFF,
	}

	return i
}

var names = []string{"JMP", "SUB", "LDN", "CMP", "JRP", "SUB", "STO", "STP"}

func (i *instruction) String() string {
	return names[i.op]
}

type register int32

type baby struct {
	memory      [words]int32
	ci, pi, acc register // registers (ci == pc -> program counter, pi == present instruction, acc == accumulator)
	running     bool
}

func NewBaby() *baby {
	return &baby{running: true}
}

func (b *baby) Display() {
	for row := 0; row < words; row++ {
		fmt.Printf("%032s\n", strconv.FormatInt(int64(b.memory[row]), 2))
	}
	fmt.Println()
}

func (b *baby) Step() {
	inst := instFromWord(b.memory[b.ci])
	fmt.Println(inst)
	b.running = false
}

func (b *baby) Run() {
	for {
		if !b.running {
			break
		}

		b.Step()
	}
}

func main() {
	b := NewBaby()
	b.Display()
	b.Run()
	b.Display()
}
