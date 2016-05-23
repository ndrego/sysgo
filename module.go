package sysgo

import (
	_ "fmt"
	"sync"
)

// If the initializer errors, it should just close the passed in channel
type InitializerFunc func(*SimChanPair) (bool, error)

type Module struct {
	Name string
	// Ports []Port
	Wires []*Wire
	Registers []*Register
	Initializers []InitializerFunc
	SensitivityClauses []*SensitivityClause
	SubModules []*Module
	sim *Simulator
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

func (A *Module) run(wg, blockReadyWG *sync.WaitGroup) {
	defer wg.Done()
	
	A.sim = GetSimulator()
	cp := newSimChannelPair(module)
	A.sim.RegisterChannelPair(cp)
	defer A.sim.UnregisterChannelPair(cp.id)

	modWG := new(sync.WaitGroup)
	modWG.Add(len(A.Initializers) + len(A.SensitivityClauses) + len(A.SubModules))

	for _, sm := range A.SubModules {
		go sm.run(modWG, blockReadyWG)
	}

	for i, _ := range A.Initializers {
		go A.runInitializer(i, modWG, blockReadyWG)
	}

	for i, _ := range A.SensitivityClauses {
		go A.runSensitivityClause(i, modWG, blockReadyWG)
	}

EventLoop:
	for {
		e := cp.Recv(updateRegisters | propagateWireValues | simFinish)
		switch e {
		case updateRegisters:
			A.updateRegisters()
			cp.Send(registerUpdateComplete)
		case propagateWireValues:
			A.propagateWires()
			cp.Send(wirePropagateComplete)
		case simFinish:
			break EventLoop
		}
	}

	modWG.Wait()
}

func (A *Module) runInitializer(i int, modWG, blockReadyWG *sync.WaitGroup) {
	defer modWG.Done()

	cp := newSimChannelPair(initializer)
	A.sim.RegisterChannelPair(cp)
	defer A.sim.UnregisterChannelPair(cp.id)

	blockReadyWG.Done()

	cp.Recv(blockRun)
	finish, err := A.Initializers[i](cp)

	var event simInternalEvent
	if err != nil {
		// Probably want to print or somehow bubble-up the error?
		event = blockError
	} else if finish {
		event = simFinish
	} else {
		event = blockComplete
	}

	// Blocking send, so that we don't
	// unregister the channel before the message
	// is actually read.
	cp.Send(event)
}

func (A *Module) runSensitivityClause(i int, modWG, blockReadyWG *sync.WaitGroup) {
	defer modWG.Done()

	cp := newSimChannelPair(sensitivity)
	A.sim.RegisterChannelPair(cp)
	defer A.sim.UnregisterChannelPair(cp.id)

	sc := A.SensitivityClauses[i]

	blockReadyWG.Done()
	
	// Process events of concern to sensitivity clauses
EventLoop:
	for {
		e, ok := <- cp.recv
		if !ok {
			return
		} else {
			switch e {
			case simFinish:
				break EventLoop
			case blockRun:
				var ev simInternalEvent
				if A.evalSensitivity(sc.s) {
					finish, error := sc.sf(cp)
					if finish {
						ev = simFinish
					} else if error != nil {
						ev = blockError
					} else {
						ev = blockProgress
					}
				} else {
					ev = blockWait
				}
				cp.Send(ev)
			}
		}
	}
}

func (A *Module) updateRegisters() {
	for _, r := range A.Registers {
		r.updateValue()
	}
}

func (A *Module) propagateWires() {
	for _, w := range A.Wires {
		w.computeValue()
	}
}

func (A *Module) evalSensitivity(s []*Sensitivity) bool {
	// No sensitivity list means do it all the time.
	if len(s) == 0 {
		return true
	}
	
	for _, sense := range s {
		v, _ := sense.signal.GetValue().(*Value1)
		lv, _ := sense.signal.GetLastValue().(*Value1)
		b := v.GetBit(0)
		lb := lv.GetBit(0)
		switch sense.qualifier {
		case None, Poslevel:
			if b == Hi {
				return true
			}
		case Neglevel:
			if b == Lo {
				return true
			}
		case Posedge:
			if b == Hi && lb == Lo {
				return true
			}
		case Negedge:
			if b == Lo && lb == Hi {
				return true
			}
		}
		
	}
	return false
}
