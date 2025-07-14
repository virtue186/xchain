package core

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVM_Run(t *testing.T) {

	code := []byte{
		byte(InstrPushByte), 'f', // push 63
		byte(InstrPushByte), 'o', // push 62
		byte(InstrPushByte), 'o',
		byte(InstrPushInt), 3,
		byte(InstrPack),
		byte(InstrPushInt), 2,
		byte(InstrPushInt), 3,
		byte(InstrAdd),
		byte(InstrStore),
	}
	vm := NewVm(code, NewState())
	vm.Run()
	fmt.Println(vm.contractState)
	value, err := vm.contractState.Get([]byte("foo"))
	newvalue := deserializeInt64(value)
	assert.Nil(t, err)
	assert.Equal(t, newvalue, int64(5))

}
