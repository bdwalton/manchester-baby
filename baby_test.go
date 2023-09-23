package main

import (
	"reflect"
	"testing"
)

func TestInstructionFromCode(t *testing.T) {
	cases := []struct {
		input   string
		wantN   int32
		want    *instruction
		wantErr error
	}{
		// Good
		{"0010 JMP 22", 10, &instruction{op: JMP, data: 22}, nil},
		{"0011 SUB 21", 11, &instruction{op: SUB, data: 21}, nil},
		{"0000 LDN 21", 0, &instruction{op: LDN, data: 21}, nil},
		{"0003 CMP", 3, &instruction{op: CMP}, nil},
		{"0000 JRP 10", 0, &instruction{op: JRP, data: 10}, nil},
		{"0000 STO 2", 0, &instruction{op: STO, data: 2}, nil},
		{"0031 STP", 31, &instruction{op: STP}, nil},
		{"0023 NUM 10", 23, &instruction{op: JMP, data: 10}, nil},

		// Bad
		{"000A JMP", 0, nil, badAddress},
		{"000A JMP 22", 0, nil, badAddress},
		{"0000 JMP", 0, nil, missingOp},
		{"0000 SUB", 0, nil, missingOp},
		{"0000 LDN", 0, nil, missingOp},
		{"0000 CMP 21", 0, nil, extraOp},
		{"000X CMP", 0, nil, badAddress},
		{"0000 JRP", 0, nil, missingOp},
		{"0000 STO", 0, nil, missingOp},
		{"0000 STP 21", 0, nil, extraOp},

		// Ugly
		{"", 0, nil, badAddress},

		{"0000 BAD 21", 0, nil, badInstruction},
		{"0000 21 BAD", 0, nil, badOperand},
		{"0000 21 21", 0, nil, badInstruction},
	}

	for i, tc := range cases {
		n, got, err := instructionFromCode(tc.input)
		if tc.wantN != n || !reflect.DeepEqual(got, tc.want) || err != tc.wantErr {
			t.Errorf("case %d: n = %d, want(%d) || got(%v) != want(%v) || err(%v) != wantErr(%v)", i, n, tc.wantN, got, tc.want, err, tc.wantErr)
		}
	}
}
