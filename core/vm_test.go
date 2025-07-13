package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVM_Run(t *testing.T) {

	code := []byte{
		byte(InstrPushInt), 63, // push 63
		byte(InstrPushInt), 62, // push 62
		byte(InstrSub),
		byte(InstrPushInt), 2,
		byte(InstrAdd),
	}
	vm := NewVm(code)
	vm.Run()
	res, err := vm.stack.Top()
	assert.Nil(t, err)
	assert.Equal(t, 3, res)

}
