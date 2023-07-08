package main

import (
	"reflect"
	"testing"
)

func TestInstructionFromCode(t *testing.T) {
	cases := []struct {
		input   string
		want    *instruction
		wantErr error
	}{
		// Good
		{"JMP 22", &instruction{op: JMP, data: 22}, nil},
		{"SUB 21", &instruction{op: SUB, data: 21}, nil},
		{"LDN 21", &instruction{op: LDN, data: 21}, nil},
		{"CMP", &instruction{op: CMP}, nil},
		{"JRP 10", &instruction{op: JRP, data: 10}, nil},
		{"STO 2", &instruction{op: STO, data: 2}, nil},
		{"STP", &instruction{op: STP}, nil},
		{"NUM 10", &instruction{op: JMP, data: 10}, nil},

		// Bad
		{"JMP", nil, missingOp},
		{"SUB", nil, missingOp},
		{"LDN", nil, missingOp},
		{"CMP 21", nil, extraOp},
		{"JRP", nil, missingOp},
		{"STO", nil, missingOp},
		{"STP 21", nil, extraOp},

		// Ugly
		{"BAD 21", nil, badInstruction},
		{"21 BAD", nil, badOperand},
		{"21 21", nil, badInstruction},
	}

	for i, tc := range cases {
		got, err := instructionFromCode(tc.input)
		if !reflect.DeepEqual(got, tc.want) || err != tc.wantErr {
			t.Errorf("case %d: got(%v) != want(%v) || err(%v) != wantErr(%v)", i, got, tc.want, err, tc.wantErr)
		}
	}
}
