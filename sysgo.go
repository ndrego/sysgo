package sysgo

import (

)

type DriverInterface interface {
	GetValue() LogicState
	GetLastValue() LogicState
}

type ReceiverInterface interface {
	SetValue(LogicState)
}
