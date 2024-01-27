package opcodes

type OpCode uint16

const (
	BR  OpCode = iota
	ADD        // add
	LD         // load
	ST
	JSR  // jump register
	AND  // bitwise and
	LDR  // load register
	STR  // store register
	RTI  // unused
	NOT  // bitwise not
	LDI  // load indirect
	STI  // store indirect
	JMP  // jump
	RES  // reserved (unused)
	LEA  // load effective address
	TRAP // execute trap
	UNKNOWN
)

func GetOpCode(instruction uint16) OpCode {
	switch instruction >> 12 {
	case 0:
		return BR
	case 1:
		return ADD
	case 2:
		return LD
	case 3:
		return ST
	case 4:
		return JSR
	case 5:
		return AND
	case 6:
		return LDR
	case 7:
		return STR
	case 8:
		return RTI
	case 9:
		return NOT
	case 10:
		return LDI
	case 11:
		return STI
	case 12:
		return JMP
	case 13:
		return RES
	case 14:
		return LEA
	case 15:
		return TRAP
	default:
		return UNKNOWN
	}
}
