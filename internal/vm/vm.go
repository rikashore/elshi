package vm

import (
	"bufio"
	"bytes"
	"elshi/internal/opcodes"
	"elshi/internal/register"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type VM struct {
	registers register.Registers
	memory    [^uint16(0)]uint16
}

func NewVM(b []byte) (*VM, error) {
	v := VM{
		registers: register.NewRegisters(),
		memory:    [^uint16(0)]uint16{},
	}

	reader := bytes.NewReader(b)

	var address uint16

	err := binary.Read(reader, binary.BigEndian, &address)
	if err != nil {
		return nil, err
	}

	v.registers.PC = address

	for {
		var instruction uint16

		err := binary.Read(reader, binary.BigEndian, &instruction)

		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		v.memory[address] = instruction
		address += 1
	}

	return &v, nil
}

func (v *VM) handleKeyboard() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	read := uint16(scanner.Text()[0])

	if read != 0 {
		v.WriteMemory(register.Kbsr, 1<<15)
		v.WriteMemory(register.Kbdr, read)
	} else {
		v.WriteMemory(register.Kbsr, 0)
	}
}

func (v *VM) ReadMemory(index uint16) uint16 {
	if index == register.Kbsr {
		v.handleKeyboard()
	}
	return v.memory[index]
}

func (v *VM) WriteMemory(address uint16, value uint16) {
	v.memory[address] = value
}

func signExtend(x uint16, bitCount uint8) uint16 {
	if (x>>(bitCount-1))&1 != 0 {
		x |= 0xFFFF << bitCount
	}
	return x
}

func br(instruction uint16, vm *VM) {
	pcOffset := signExtend(instruction&0x1ff, 9)
	conditionFlag := (instruction >> 9) & 0x7

	if conditionFlag&vm.registers.GetConditionRegister() != 0 {
		vm.registers.PC = pcOffset
	}
}

func add(instruction uint16, vm *VM) {
	dest := (instruction >> 9) & 0x7
	source1 := (instruction >> 6) & 0x7

	immediateFlag := (instruction >> 5) & 0x1
	var value uint16

	if immediateFlag == 1 {
		immediate := signExtend(instruction&0x1f, 5)
		value = vm.registers.Get(source1) + immediate
	} else {
		source2 := instruction & 0x7
		value = vm.registers.Get(source1) + vm.registers.Get(source2)
	}

	vm.registers.Update(dest, value)
	vm.registers.UpdateConditionRegister(dest)
}

func ld(instruction uint16, vm *VM) {
	dest := (instruction >> 9) & 0x7
	pcOffset := signExtend(instruction&0x1ff, 9)
	mem := pcOffset + vm.registers.PC
	value := vm.ReadMemory(mem)

	vm.registers.Update(dest, value)
	vm.registers.UpdateConditionRegister(dest)
}

func st(instruction uint16, vm *VM) {
	source := (instruction >> 9) & 0x7
	pcOffset := signExtend(instruction&0x1ff, 9)

	val := vm.registers.PC + pcOffset
	vm.WriteMemory(val, vm.registers.Get(source))
}

func jsr(instruction uint16, vm *VM) {
	baseRegister := (instruction >> 6) & 0x7
	longPcOffset := signExtend(instruction&0x7ff, 11)
	longFlag := (instruction >> 11) & 1

	vm.registers.Update(7, vm.registers.PC)

	if longFlag != 0 {
		value := vm.registers.PC + longPcOffset
		vm.registers.PC = value
	} else {
		vm.registers.PC = vm.registers.Get(baseRegister)
	}
}

func and(instruction uint16, vm *VM) {
	dest := (instruction >> 9) & 0x7
	source1 := (instruction >> 6) & 0x7

	immediateFlag := (instruction >> 5) & 0x1
	var value uint16

	if immediateFlag == 1 {
		immediate := signExtend(instruction&0x1f, 5)
		value = vm.registers.Get(source1) & immediate
	} else {
		source2 := instruction & 0x7
		value = vm.registers.Get(source1) & vm.registers.Get(source2)
	}

	vm.registers.Update(dest, value)
	vm.registers.UpdateConditionRegister(dest)
}

func ldr(instruction uint16, vm *VM) {
	dest := (instruction >> 9) & 0x7
	baseRegister := (instruction >> 6) & 0x7

	offset := signExtend(instruction&0x3f, 6)

	value := vm.registers.Get(baseRegister) + offset
	memValue := vm.ReadMemory(value)

	vm.registers.Update(dest, memValue)
	vm.registers.UpdateConditionRegister(dest)
}

func str(instruction uint16, vm *VM) {
	dest := (instruction >> 9) & 0x7
	baseRegister := (instruction >> 6) & 0x7

	offset := signExtend(instruction&0x3f, 6)

	value := vm.registers.Get(baseRegister) + offset
	vm.WriteMemory(value, vm.registers.Get(dest))
}

func rti(instruction uint16, vm *VM) {}

func not(instruction uint16, vm *VM) {
	dest := (instruction >> 9) & 0x7
	source := (instruction >> 6) & 0x7

	vm.registers.Update(dest, ^vm.registers.Get(source))
	vm.registers.UpdateConditionRegister(dest)
}

func ldi(instruction uint16, vm *VM) {
	dest := (instruction >> 9) & 0x7
	pcOffset := signExtend(instruction&0x1ff9, 9)

	first := vm.ReadMemory(pcOffset)
	resulting := vm.ReadMemory(first)

	vm.registers.Update(dest, resulting)
	vm.registers.UpdateConditionRegister(dest)
}

func sti(instruction uint16, vm *VM) {
	source := (instruction >> 9) & 0x7
	pcOffset := signExtend(instruction&0x1ff, 9)

	val := vm.registers.PC + pcOffset
	address := vm.ReadMemory(val)

	vm.WriteMemory(address, vm.registers.Get(source))
}

func jmp(instruction uint16, vm *VM) {
	reg := (instruction >> 6) & 0x7
	vm.registers.PC = vm.registers.Get(reg)
}

func lea(instruction uint16, vm *VM) {
	dr := (instruction >> 9) & 0x7
	pcOffset := signExtend(instruction&0x1ff, 9)
	value := vm.registers.PC + pcOffset

	vm.registers.Update(dr, value)
	vm.registers.UpdateConditionRegister(dr)
}

func trap(instruction uint16, vm *VM) {
	switch instruction & 0xFF {
	case 0x20:
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}

		vm.registers.Update(0, uint16(scanner.Text()[0]))
	case 0x21:
		c := vm.registers.Get(0)
		fmt.Printf("%c\n", c)
	case 0x22:
		idx := vm.registers.Get(0)
		c := vm.ReadMemory(idx)

		for c != 0x0000 {
			fmt.Printf("%c", c)
			idx += 1
			c = vm.ReadMemory(idx)
		}
	case 0x23:
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}

		c := scanner.Text()[0]
		vm.registers.Update(0, uint16(c))
		fmt.Printf("%c\n", c)
	case 0x24:
		idx := vm.registers.Get(0)
		c := vm.ReadMemory(idx)

		for c != 0x0000 {
			c1 := c & 0xFF
			fmt.Printf("%c", c1)
			c2 := c >> 8
			if c2 != 0x00 {
				fmt.Printf("%c", c2)
			}

			idx += 1
			c = vm.ReadMemory(idx)
		}
	case 0x25:
		fmt.Println("HALTing execution")
		os.Exit(0)
	default:
		os.Exit(1)
	}
}

func (v *VM) executeInstruction(instruction uint16) {
	opcode := opcodes.GetOpCode(instruction)

	switch opcode {
	case 0:
		br(instruction, v)
	case 1:
		add(instruction, v)
	case 2:
		ld(instruction, v)
	case 3:
		st(instruction, v)
	case 4:
		jsr(instruction, v)
	case 5:
		and(instruction, v)
	case 6:
		ldr(instruction, v)
	case 7:
		str(instruction, v)
	case 8:
		rti(instruction, v)
	case 9:
		not(instruction, v)
	case 10:
		ldi(instruction, v)
	case 11:
		sti(instruction, v)
	case 12:
		jmp(instruction, v)
	case 13:
		return
	case 14:
		lea(instruction, v)
	case 15:
		trap(instruction, v)
	default:
		return
	}
}

func (v *VM) Execute() {
	for v.registers.PC < ^uint16(0) {
		instruction := v.ReadMemory(v.registers.PC)

		v.registers.PC += 1

		v.executeInstruction(instruction)
	}
}
