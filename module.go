package sysgo

import (

)

// If the initializer errors, it should just close the passed in channel
type InitializerFunc func(chan<- ProceduralBlockEvent)

type Module struct {
	Name string
	// Ports []Port
	Wires []*Wire
	Registers []*Register
	Initializers []InitializerFunc
	SensitivityClauses []*SensitivityClause
	SubModules []*Module
}

func (A *Module) getNumBlocks() (numInits, numSenseClauses uint32) {
	numInits = uint32(len(A.Initializers))
	numSenseClauses = uint32(len(A.SensitivityClauses))

	for _, sm := range A.SubModules {
		n, s := sm.getNumBlocks()
		numInits += n
		numSenseClauses += s
	}

	return
}

func (A *Module) getAllInitializers() (iFuncs []InitializerFunc) {
	iFuncs = append([]InitializerFunc(nil), A.Initializers...)
	for _, sm := range A.SubModules {
		iFuncs = append(iFuncs, sm.getAllInitializers()...)
	}
	return
}
