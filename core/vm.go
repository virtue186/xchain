package core

import "fmt"

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a
	InstrAdd      Instruction = 0x0b
	InstrPushByte Instruction = 0x0c
	InstrPack     Instruction = 0x0d
	InstrSub      Instruction = 0x0e
)

type VM struct {
	data  []byte
	pc    int    // 指向下一条要执行的字节的位置
	stack *Stack // 栈
	sp    int    // 栈的指针
}

type Stack struct {
	data []any
	sp   int
}

func (s *Stack) Push(v any) {
	s.sp++
	s.data[s.sp] = v
}

func (s *Stack) Pop() any {
	if s.sp < 0 {
		panic("stack underflow")
	}
	value := s.data[s.sp]
	s.sp--
	return value
}

// Top 获取 Stack顶部元素
func (s *Stack) Top() (any, error) {
	if s.sp < 0 {
		return nil, fmt.Errorf("stack is empty")
	}
	return s.data[s.sp], nil
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sp:   -1,
	}
}

func NewVm(data []byte) *VM {
	return &VM{
		data:  data,
		stack: NewStack(1024),
		pc:    0,
	}
}

func (vm *VM) Run() error {
	for {
		instr := vm.data[vm.pc]
		err := vm.Exec(Instruction(instr))
		if err != nil {
			return err
		}
		if vm.pc > len(vm.data)-1 {
			break
		}
	}
	return nil

}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {
	case InstrPushInt:
		val := int(vm.data[vm.pc+1])
		vm.stack.Push(val)
		vm.pc += 2
	case InstrAdd:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		vm.stack.Push(a + b)
		vm.pc++
	case InstrPushByte:
		val := vm.data[vm.pc+1]
		vm.stack.Push(val)
		vm.pc += 2
	case InstrPack:
		n := vm.stack.Pop().(int)
		if n < 0 || n > vm.stack.sp+1 {
			return fmt.Errorf("invalid pack length: %d", n)
		}
		b := make([]byte, n)
		for i := n - 1; i >= 0; i-- {
			b[i] = vm.stack.Pop().(byte)
		}
		vm.stack.Push(b)
		vm.pc++
	case InstrSub:
		b := vm.stack.Pop().(int)
		a := vm.stack.Pop().(int)
		vm.stack.Push(a - b)
		vm.pc++
	}

	return nil
}
