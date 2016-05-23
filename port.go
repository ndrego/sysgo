package sysgo

import (
)

type Port struct {
	Name string
	driver DriverInterface
	receiver ReceiverInterface
}

func (A *Port) GetValue() ValueInterface {
	return A.driver.GetValue()
}

func (A *Port) GetLastValue() ValueInterface {
	return A.driver.GetLastValue()
}
