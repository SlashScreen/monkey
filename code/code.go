package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type (
	Opcode     byte
	Definition struct {
		Name          string
		OperandWidths []int
	}
)

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
		}

		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

func Lookup(op byte) (*Definition, error) {
	if def, ok := definitions[Opcode(op)]; ok {
		return def, nil
	} else {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
	}

	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

const (
	OpConstant Opcode = iota
	OpPop
	OpJumpNotTruthy
	OpJump
	OpCall
	OpReturn
	OpReturnValue

	OpTrue
	OpFalse
	OpNull
	OpArray
	OpHash

	OpEqual
	OpNotEqual
	OpGreaterThan

	OpMinus
	OpBang

	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod

	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal
	OpGetBuiltin
	OpIndex
)

var definitions = map[Opcode]*Definition{
	OpConstant:      {"OpConstant", []int{2}},
	OpPop:           {"OpPop", []int{}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpJump:          {"OpJump", []int{2}},
	OpCall:          {"OpCall", []int{1}},
	OpReturn:        {"OpReturn", []int{}},
	OpReturnValue:   {"OpReturnValue", []int{}},

	OpTrue:  {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},
	OpNull:  {"OpNull", []int{}},
	OpArray: {"OpArray", []int{2}},
	OpHash:  {"OpHash", []int{2}},

	OpEqual:       {"OpEq", []int{}},
	OpNotEqual:    {"OpNeq", []int{}},
	OpGreaterThan: {"OpGreaterThan", []int{}},

	OpMinus: {"OpMinus", []int{}},
	OpBang:  {"OpBang", []int{}},

	OpAdd: {"OpAdd", []int{}},
	OpSub: {"OpSub", []int{}},
	OpMul: {"OpMul", []int{}},
	OpDiv: {"OpDiv", []int{}},
	OpMod: {"OpMod", []int{}},

	OpGetGlobal:  {"OpGetGlobal", []int{2}},
	OpSetGlobal:  {"OpSetGlobal", []int{2}},
	OpGetLocal:   {"OpGetLocal", []int{1}},
	OpSetLocal:   {"OpSetLocal", []int{1}},
	OpGetBuiltin: {"OpGetBuiltin", []int{1}},
	OpIndex:      {"OpIndex", []int{}},
}
