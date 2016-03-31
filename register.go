package sysgo

import (
	"fmt"
)

type Register struct {
	Name string
	currentValue Value
	nextValue Value
	modified bool
}

func (A *Register) SetValue(v Value) error {
	if A.modified {
		return fmt.Errorf("Setting register %s multiple times in same event.", A.Name)
	}

	A.nextValue = v
	A.modified = true

	return nil
}

func (A *Register) GetValue() Value {
	return A.currentValue
}

func NewRegister(name string) (r *Register) {
	r = new(Register)
	r.Name = name
	r.currentValue = Undefined
	r.nextValue = Undefined
	r.modified = false

	return
}
