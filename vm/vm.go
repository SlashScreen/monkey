package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const STACKSIZE = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]object.Object, STACKSIZE),
		sp:    0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			if err := vm.push(vm.constants[constIndex]); err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv, code.OpMod:
			if err := vm.executeBinOp(op); err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		}
	}

	return nil
}

func (vm *VM) executeBinOp(op code.Opcode) error {
	r := vm.pop()
	l := vm.pop() // order matters

	leftType := l.Type()
	rightType := r.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOp(op, l, r)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s",
			leftType, rightType)
	}
}

func (vm *VM) executeBinaryIntegerOp(op code.Opcode, l, r object.Object) error {
	lv := l.(*object.Integer).Value
	rv := r.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = lv + rv
	case code.OpSub:
		result = lv - rv
	case code.OpMul:
		result = lv * rv
	case code.OpDiv:
		result = lv / rv
	case code.OpMod:
		result = lv % rv
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= STACKSIZE {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}
