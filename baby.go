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
	"math/bits"
	"os"
	"strconv"
	"strings"
	"time"
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

type memory [words]int32

func (m *memory) RawWord(i int) uint32 {
	return bits.Reverse32(uint32(m[i]))
}

type baby struct {
	mem         memory
	ci, pi, acc register // registers (ci == pc -> program counter, pi == present instruction, acc == accumulator)
	running     bool
}

func NewBaby(mem memory) *baby {
	return &baby{running: true, mem: mem}
}

func (b *baby) Display() {
	fmt.Println("\033[H\033[2J")
	fmt.Printf("ci: %d, acc: %d\n", b.ci, b.acc)
	for row := 0; row < words; row++ {
		rw := b.mem.RawWord(row)
		i := instFromWord(b.mem[row])
		fmt.Printf("%04d:%032s | [%s ; %d]\n", row, strconv.FormatInt(int64(rw), 2), i, b.mem[row])
	}
	fmt.Println()
}

func (b *baby) Step() {
	// The Baby increments the ci (current instruction) counter
	// prior to loading the instruction, not after executing from
	// the current value.
	b.ci += 1

	inst := instFromWord(b.mem[b.ci])
	fmt.Println(inst)

	switch inst.op {
	case JMP:
		// Because we increment ci before executing, the jump
		// must go to the instruction prior to the one we
		// expect to execute
		b.ci = register(inst.data) - 1
	case SUB:
		b.acc = b.acc - register(b.mem[inst.data])
	case CMP:
		if b.acc < 0 {
			b.ci += 1
		}
	case LDN:
		b.acc = register(-b.mem[inst.data])
	case JRP:
		// Because we increment ci before executing, the jump
		// must go to the instruction prior to the one we
		// expect to execute
		b.ci = b.ci + register(b.mem[inst.data]) - 1
	case STO:
		b.mem[inst.data] = int32(b.acc)
	case STP:
		b.running = false
	}
}

func (b *baby) Run() {
	for {
		b.Display()
		if !b.running {
			break
		}

		b.Step()
		time.Sleep(1 * time.Second)
	}
}

var (
	missingOp      = errors.New("invalid code - missing operand")
	extraOp        = errors.New("invalid code - unexpected argument")
	badMemory      = errors.New("invalid binary code - couldn't convert to integer")
	badOperand     = errors.New("invalid code - invalid operand")
	badInstruction = errors.New("invalid code - unknown instruction")
)

func instructionFromCode(code string) (int32, *instruction, error) {
	parts := strings.SplitN(code, " ", 3)

	n, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return 0, nil, badMemory
	}

	switch parts[1] {
	case "CMP", "STP":
		if len(parts) > 2 {
			return 0, nil, extraOp
		}
		return int32(n), &instruction{op: nameOps[parts[1]]}, nil
	default:
		if len(parts) < 3 {
			return 0, nil, missingOp
		}

		operand, err := strconv.Atoi(parts[2])
		if err != nil {
			return 0, nil, badOperand
		}

		// This is syntactic sugar for allowing the input of
		// numbers. Special case it.
		if parts[1] == "NUM" {
			return int32(n), &instruction{op: JMP, data: int32(operand)}, nil
		}

		op, ok := nameOps[parts[1]]
		if !ok {
			return 0, nil, badInstruction
		}

		return int32(n), &instruction{op: op, data: int32(operand)}, nil
	}
}

func memFromBin(code string) (int32, int32, error) {
	parts := strings.SplitN(code, ":", 2)
	if len(parts) < 2 {
		return 0, 0, missingOp
	}

	n, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return 0, 0, badMemory
	}
	i, err := strconv.ParseUint(parts[1], 2, 32)
	if err != nil {
		return 0, 0, badMemory
	}

	return int32(n), int32(bits.Reverse32(uint32(i))), nil
}

// Function loadProgram takes a file path and reads a baby program from it.
// Programs may be written in either assembly or binary.
// Assembly format:
// INST DATA - JRP 24
// Binary format:
// WORD#:32-bit Binary - 0000:00000110101001000100000100000100
func loadProgram(programfile string) (memory, error) {
	var mem memory

	data, err := os.ReadFile(programfile)
	if err != nil {
		return mem, fmt.Errorf("error reading programfile: %v", err)
	}

	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		if line != "" {
			if strings.Contains(line, ":") {
				n, m, err := memFromBin(line)
				if err != nil {
					return mem, fmt.Errorf("error on line %d: %v", i+1, err)
				}
				mem[n] = m
			} else {
				n, inst, err := instructionFromCode(line)
				if err != nil {
					return mem, fmt.Errorf("error on line %d: %v", i+1, err)
				}
				mem[n] = inst.toInt32()
			}
		}
	}

	return mem, nil
}

func main() {
	flag.Parse()

	mem, err := loadProgram(*programfile)
	if err != nil {
		log.Fatalf("Couldn't load program from %q: %v", *programfile, err)
	}
	b := NewBaby(mem)
	b.Run()
}
