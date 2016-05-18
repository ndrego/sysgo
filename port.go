package sysgo

import (
)

type Port struct {
	Name string
	driver DriverInterface
	receiver ReceiverInterface
}

func (A *Port) GetValue() LogicState {
	return A.driver.GetValue()
}

func (A *Port) GetLastValue() LogicState {
	return A.driver.GetLastValue()
}
