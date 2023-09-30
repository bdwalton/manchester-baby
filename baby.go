package main

// Manchester Baby
// Details of the machine gathered from several sources:
// * https://blog.mark-stevens.co.uk/2017/02/manchester-baby-ssem-emulator/
// * https://en.wikipedia.org/wiki/Manchester_Baby
// * http://curation.cs.manchester.ac.uk/computer50/www.computer50.org/mark1/prog98/prizewinners.html
// * http://curation.cs.manchester.ac.uk/computer50/www.computer50.org/mark1/new.baby.html
// * https://www.icsa.inf.ed.ac.uk/research/groups/hase/models/ssem/index.html

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
	// Decoding a memory word to an instruction, we use the specification from:
	// https://www.icsa.inf.ed.ac.uk/research/groups/hase/models/ssem/index.html
	// | Line No.	| Not Used | Func. No. | Not Used |
	// | 0 1 2 3 4	| 5 .. 12  | 13 14 15  | 16 .. 31 |

	return &instruction{
		op:   (word & 0x0000E000) >> 13,
		data: word & 0x0000001F,
	}
}

type register int32

type memory [words]int32

func (m *memory) RawWord(i int) uint32 {
	return bits.Reverse32(uint32(m[i]))
}

type baby struct {
	mem     memory
	ci, acc register // registers (ci == pc -> program counter, acc == accumulator)
	running bool
}

func NewBaby(mem memory) *baby {
	return &baby{running: true, mem: mem}
}

func (b *baby) Display() {
	fmt.Println("\033[H\033[2J")
	fmt.Printf("ci: %d, acc: %d, running: %t\n", b.ci, b.acc, b.running)
	for row := 0; row < words; row++ {
		rw := b.mem.RawWord(row)
		i := instFromWord(b.mem[row])
		ind := ""
		if row == int(b.ci) {
			ind = " <=="
		}
		fmt.Printf("%04d:%032s | %4s [%-8s ; %12d]\n", row, strconv.FormatInt(int64(rw), 2), ind, i, b.mem[row])
	}
	fmt.Println()
}

func (b *baby) Reboot(mem memory) {
	b.mem = mem
	b.Reset()
}

func (b *baby) Reset() {
	b.ci = 0
	b.acc = 0
	b.running = true
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
		b.ci = register(b.mem[inst.data])
	case SUB:
		b.acc = b.acc - register(b.mem[inst.data])
	case CMP:
		if b.acc < 0 {
			b.ci += 1
		}
	case LDN:
		b.acc = register(-b.mem[inst.data])
	case JRP:
		b.ci = b.ci + register(b.mem[inst.data])
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
		time.Sleep(time.Second / 700) // Baby ran at ~700 instructions per second
	}
}

var (
	missingOp      = errors.New("invalid code - missing operand")
	badEntry       = errors.New("invalid code - missing address, binary or code")
	extraOp        = errors.New("invalid code - unexpected argument")
	badAddress     = errors.New("invalid address - unusable address")
	badMemory      = errors.New("invalid binary code - couldn't convert to integer")
	badOperand     = errors.New("invalid code - invalid operand")
	badInstruction = errors.New("invalid code - unknown instruction")
)

func instructionFromCode(code string) (int32, *instruction, error) {
	parts := strings.SplitN(code, " ", 3)

	n, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil || n >= words || n < 0 {
		return 0, nil, badAddress
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
		return 0, 0, badEntry
	}

	n, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil || n >= words || n < 0 {
		return 0, 0, badAddress
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
	for {
		b.Display()
		fmt.Printf("(R)un, (S)tep, R(e)set, Re(b)oot, (Q)uit: ")
		var input rune

		_, err := fmt.Scanf("%c\n", &input)
		if err != nil {
			fmt.Println("Invalid input: ", err)
		}
		switch input {
		case 'R', 'r':
			b.Run()
		case 'S', 's':
			b.Step()
		case 'B', 'b':
			b.Reboot(mem)
		case 'E', 'e':
			b.Reset()
		case 'Q', 'q':
			os.Exit(0)
		}
	}
}
