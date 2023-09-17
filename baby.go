package main

// Manchester Baby
// Details of the machine gathered from several sources:
// * https://blog.mark-stevens.co.uk/2017/02/manchester-baby-ssem-emulator/
// * https://en.wikipedia.org/wiki/Manchester_Baby
// * http://curation.cs.manchester.ac.uk/computer50/www.computer50.org/mark1/prog98/prizewinners.html
// * http://curation.cs.manchester.ac.uk/computer50/www.computer50.org/mark1/new.baby.html

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	programfile = flag.String("programfile", "", "path to program file")
)

const (
	wordSize = 32
	words    = 32
)

// Instruction opcodes
const (
	JMP  = iota // Jump (0; 000 in LSB first)
	JRP         // Jump relative (1; 100 in LSB first)
	LDN         // Load negative (2; 010 in LSB first)
	STO         // Store (3; 110 in LSB first)
	SUB         // Subtract (4; 001 in LSB first)
	SUB2        // Subtract (5; 101 in LSB first)
	CMP         // Compare (6; 011 in LSB first)
	STP         // Stop (7; 111 in LSB first)
)

var opNames = []string{"JMP", "JRP", "LDN", "STO", "SUB", "SUB", "CMP", "STP"}
var nameOps = map[string]int32{
	"JMP": JMP,
	"JRP": JRP,
	"LDN": LDN,
	"STO": STO,
	"SUB": SUB,
	// SUB2
	"CMP": CMP,
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
	return 0 | (i.op << 13) | i.data
}

func instFromWord(word int32) *instruction {
	i := &instruction{
		op:   (word & 0x0000F000) >> 13,
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

func NewBaby(prog []int32) *baby {
	b := &baby{running: true}
	for i, m := range prog {
		b.memory[i] = m
	}
	return b
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

		// This is syntactic sugar for allowing the input of
		// numbers. Special case it.
		if parts[0] == "NUM" {
			return &instruction{op: JMP, data: int32(operand)}, nil
		}

		op, ok := nameOps[parts[0]]
		if !ok {
			return nil, badInstruction
		}

		return &instruction{op: op, data: int32(operand)}, nil
	}
}

func loadProgram(programfile string) ([]int32, error) {
	data, err := os.ReadFile(programfile)
	if err != nil {
		return nil, fmt.Errorf("error reading programfile: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	prog := []int32{}
	for i, line := range lines {
		if line != "" {
			inst, err := instructionFromCode(line)
			if err != nil {
				return nil, fmt.Errorf("error on line %d: %v", i+1, err)
			}
			prog = append(prog, inst.toInt32())
		}
	}

	if len(prog) > words {
		return nil, fmt.Errorf("too many words (%d) for baby (max words: %d)", len(prog), words)
	}

	return prog, nil
}

func main() {
	flag.Parse()

	prog, err := loadProgram(*programfile)
	if err != nil {
		log.Fatalf("Couldn't load program from %q: %v", *programfile, err)
	}
	b := NewBaby(prog)
	b.Run()
}
