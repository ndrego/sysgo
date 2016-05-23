package sysgo

import (

)

type DriverInterface interface {
	GetValue() ValueInterface
	GetLastValue() ValueInterface
}

type ReceiverInterface interface {
	SetValue(ValueInterface)
}
