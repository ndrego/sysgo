package sysgo

import (
)

type Wire struct {
	Name string
	drivers []DriverInterface
	receivers []ReceiverInterface
	currentValue Value
	lastValue Value
}

func (A *Wire) computeValue() {
	var v Value
	if len(A.drivers) > 0 {
		v = A.drivers[0].GetValue()
		for i := 1; i < len(A.drivers); i++ {
			v = CombineValue(v, A.drivers[i].GetValue())
		}
	} else {
		v = Undefined
	}

	A.lastValue = A.currentValue
	A.currentValue = v
}

func (A *Wire) propagate() {
	A.computeValue()
	for _, r := range A.receivers {
		r.SetValue(A.currentValue)
	}
}

func (A *Wire) GetValue() Value {
	return A.currentValue
}

func (A *Wire) GetLastValue() Value {
	return A.lastValue
}
