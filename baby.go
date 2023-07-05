package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

var opNames = []string{"JMP", "SUB", "LDN", "CMP", "JRP", "SUB", "STO", "STP"}
var nameOps = map[string]int32{
	"JMP": JMP,
	"SUB": SUB,
	"LDN": LDN,
	"CMP": CMP,
	"JRP": JRP,
	// SUB2
	"STO": STO,
	"STP": STP,
}

type instruction struct {
	op   int32
	data int32
}

func (i *instruction) String() string {
	var sb strings.Builder

	sb.WriteString(opNames[i.op])

	switch i.op {
	case CMP, STP:
	default:
		sb.WriteString(fmt.Sprintf(" %d", i.data))
	}

	return sb.String()
}

func (i *instruction) toInt32() int32 {
	return 0 | (i.op << 12) | i.data
}

func instFromWord(word int32) *instruction {
	i := &instruction{
		op:   (word & 0x0000F000) >> 12,
		data: word & 0x00000FFF,
	}

	return i
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

var (
	missingOp      = errors.New("invalid code - missing operand")
	extraOp        = errors.New("invalid code - unexpected argument")
	badOperand     = errors.New("invalid code - invalid operand")
	badInstruction = errors.New("invalid code - unknown instruction")
)

func instructionFromCode(code string) (*instruction, error) {
	parts := strings.Split(code, " ")

	switch parts[0] {
	case "CMP", "STP":
		if len(parts) > 1 {
			return nil, extraOp
		}
		return &instruction{op: nameOps[parts[0]]}, nil
	default:
		if len(parts) < 2 {
			return nil, missingOp
		}

		operand, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, badOperand
		}

		op, ok := nameOps[parts[0]]
		if !ok {
			return nil, badInstruction
		}

		return &instruction{op: op, data: int32(operand)}, nil
	}
}

func main() {
	b := NewBaby()
	b.Display()
	b.Run()
	b.Display()
}
