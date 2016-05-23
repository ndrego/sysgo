package sysgo

import (
)

type Wire struct {
	Name string
	drivers []DriverInterface
	receivers []ReceiverInterface
	currentValue *Value1
	lastValue *Value1
}

func (A *Wire) computeValue() {
	var v *Value1
	if len(A.drivers) > 0 {
		v = A.drivers[0].GetValue().(*Value1)
		for i := 1; i < len(A.drivers); i++ {
			v.combine(A.drivers[i].GetValue())
		}
	} else {
		v = NewValue(1).(*Value1)
		v.SetBit(0, Undefined)
	}

	A.SetValue(v)
}

func (A *Wire) propagate() {
	A.computeValue()
	for _, r := range A.receivers {
		r.SetValue(A.currentValue)
	}
}

func (A *Wire) GetValue() ValueInterface {
	return A.currentValue
}

func (A *Wire) GetLastValue() ValueInterface {
	return A.lastValue
}

// Mainly used for continous or forced assignments
func (A *Wire) SetValue(v ValueInterface) {
	switch v := v.(type) {
	case *Value1:
		A.lastValue = A.currentValue
		A.currentValue = v
	case *Value64, *ValueBig:
		A.lastValue = A.currentValue
		A.currentValue = NewValue(1).(*Value1)
		A.currentValue.SetBit(0, v.GetBit(0))
	}
}
