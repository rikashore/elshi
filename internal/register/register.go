package register

const (
	Pos uint16 = 1 << iota
	Zro        = 1 << iota
	Neg        = 1 << iota
)

const (
	// Kbsr Keyboard status: The KBSR indicates whether a key has been pressed
	Kbsr = 0xFE00

	// Kbdr Keyboard data: The KBDR identifies which key was pressed
	Kbdr = 0xFE02
)

type Registers struct {
	regs [8]uint16
	PC   uint16
	cond uint16
}

func NewRegisters() Registers {
	return Registers{
		regs: [8]uint16{},
		PC:   0x3000,
		cond: Zro,
	}
}

func (r *Registers) Get(index uint16) uint16 {
	return r.regs[index]
}

func (r *Registers) GetConditionRegister() uint16 {
	return r.cond
}

func (r *Registers) Update(index uint16, value uint16) {
	r.regs[index] = value
}

func (r *Registers) UpdateConditionRegister(reg uint16) {
	if r.Get(reg) == 0 {
		r.cond = Zro
	} else if (r.Get(reg) >> 15) != 0 {
		r.cond = Neg
	} else {
		r.cond = Pos
	}
}
