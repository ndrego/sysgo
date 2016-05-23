package sysgo

import (
	"fmt"
)

type SensitivityQualifier int

type SensitivitySimFunc func(*SimChanPair) (bool, error)

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

func NewSensitivity(q SensitivityQualifier, sig DriverInterface) (s *Sensitivity, e error) {
	e = nil
	s = new(Sensitivity)

	// Type assert that a sensitivity signal should only be a 1-bit item
	if _, ok := s.signal.GetValue().(*Value1); !ok {
		e = fmt.Errorf("Multi-bit wires can not be used for sensitivity signals")
	}
	
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
