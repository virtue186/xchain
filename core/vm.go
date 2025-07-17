package core

import (
	"encoding/binary"
	"fmt"
)

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a
	InstrAdd      Instruction = 0x0b
	InstrPushByte Instruction = 0x0c
	InstrPack     Instruction = 0x0d
	InstrSub      Instruction = 0x0e
	InstrStore    Instruction = 0x0f
)

type VM struct {
	data          []byte
	pc            int    // 指向下一条要执行的字节的位置
	stack         *Stack // 栈
	contractState *State
}

type Stack struct {
	data []any
	sp   int // 栈的指针，初始化应当为-1
}

func (s *Stack) Push(v any) error {
	if s.sp >= len(s.data)-1 {
		return fmt.Errorf("stack overflow")
	}
	s.sp++
	s.data[s.sp] = v
	return nil
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

func NewVm(data []byte, state *State) *VM {
	return &VM{
		data:          data,
		stack:         NewStack(1024),
		pc:            0,
		contractState: state,
	}
}

func (vm *VM) Run() error {
	for vm.pc < len(vm.data) {
		instr := vm.data[vm.pc]
		if err := vm.Exec(Instruction(instr)); err != nil {
			return err // 如果 Exec 出错，立即返回
		}
	}
	return nil

}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {

	//case InstrStore:
	//	// 假设栈顶是 value，次顶是 key
	//	value := vm.stack.Pop()
	//	fmt.Println(value)
	//	key := vm.stack.Pop().([]byte)
	//	fmt.Println(key)
	//	var serializedValue []byte
	//	switch v := value.(type) {
	//	case int:
	//		serializedValue = serializeInt64(int64(v))
	//	default:
	//		panic("TODO: unknown type")
	//	}
	//	if err := vm.contractState.Put(key, serializedValue); err != nil {
	//		return err
	//	}
	//	vm.pc++

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
	default:
		return fmt.Errorf("invalid instruction: 0x%x at pc=%d", instr, vm.pc)
	}

	return nil
}

func serializeInt64(value int64) []byte {
	buf := make([]byte, 8)

	binary.LittleEndian.PutUint64(buf, uint64(value))

	return buf
}

func deserializeInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}
