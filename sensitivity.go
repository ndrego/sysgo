package sysgo

import (

)

type SensitivityQualifier int

type SensitivitySimFunc func() error

const (
	None SensitivityQualifier = iota
	Posedge
	Negedge
	Poslevel
	Neglevel
)

type Sensitivity struct {
	signal DriverInterface
	qualifier SensitivityQualifier
}

type SensitivityClause struct {
	s []*Sensitivity // This list gets logically OR'ed
	sf SensitivitySimFunc
}

func NewSensitivity(q SensitivityQualifier, sig DriverInterface) (s *Sensitivity) {
	s = new(Sensitivity)

	s.signal = sig
	s.qualifier = q
	return
}

func NewSensitivityClause(f SensitivitySimFunc, senses ...*Sensitivity) (sc *SensitivityClause) {
	sc = new(SensitivityClause)

	sc.sf = f
	sc.s = senses

	return
}
