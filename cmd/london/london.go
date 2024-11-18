package main

import (
	"github.com/rgeraldes24/goevmlab/ops"
	"github.com/rgeraldes24/goevmlab/program"
)

func EfByte() []byte {
	inner := []byte{
		0xEF,
	}

	initcode := program.NewProgram()
	initcode.Mstore(inner, 0)
	initcode.Return(0, uint32(len(inner)))

	program := program.NewProgram()
	Create(program, initcode.Bytecode(), false, false)
	program.Op(ops.POP)
	Create(program, initcode.Bytecode(), true, true)
	program.Op(ops.POP)
	return program.Bytecode()
}

func Create(p *program.Program, code []byte, inMemory bool, isCreate2 bool) {
	var (
		value    = 0
		offset   = 0
		size     = len(code)
		salt     = 0
		createOp = ops.CREATE
	)
	// Load the code into mem
	if !inMemory {
		p.Mstore(code, 0)
	}
	// Create it
	if isCreate2 {
		p.Push(salt)
		createOp = ops.CREATE2
	}
	p.Push(size).Push(offset).Push(value).Op(createOp)
}
