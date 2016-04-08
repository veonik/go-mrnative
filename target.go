package mrnative

import "log"

// A targetType defines what type of target a struct is.
type targetType uint8

const (
	targetMapper targetType = iota
	targetReducer
)

func (t targetType) String() string {
	switch t {
	case targetMapper:
		return "Mapper"
	case targetReducer:
		return "Reducer"
	}
	return "Unknown"
}

// A Target is a declaration we intend to generate a bridge for.
type Target struct {
	typ targetType
	pkg *Package

	decl   *Struct
	ctor   *Func
	method *Method
	ctx    *Interface

	keyIn    *Param
	valueIn  *Param
	keyOut   *Param
	valueOut *Param
}

func NewTarget(pkg *Package, decl *Struct, typ targetType) *Target {
	tgt := &Target{
		typ:  typ,
		pkg:  pkg,
		decl: decl,
	}
	ctorName := "New" + decl.name
	for _, fn := range pkg.functions {
		if fn.name == ctorName {
			tgt.ctor = fn
			break
		}
	}
	if tgt.ctor == nil {
		log.Fatalf("Unable to locate constructor function %s.%s", pkg.name, ctorName)
	}
	var methName string
	if typ == targetMapper {
		methName = "Map"
	} else {
		methName = "Reduce"
	}
	for _, m := range decl.methods {
		if m.name == methName {
			tgt.method = m
			// TODO: Ensure params
		}
	}
	if tgt.method == nil {
		log.Fatalf("Unable to locate \"%s\" method on struct %s.%s", methName, pkg.name, decl.name)
	}
	ctxParam := tgt.method.params[len(tgt.method.params)-1]
	for _, i := range pkg.interfaces {
		if i.name == ctxParam.typ {
			tgt.ctx = i
		}
	}
	if tgt.ctx == nil {
		log.Fatalf("Unable to locate interface %s.%s", pkg.name, ctxParam.typ)
	}
	var ctxWrite *Method
	var ctxNext *Method
	for _, m := range tgt.ctx.methods {
		if m.name == "Write" {
			ctxWrite = m
		} else if typ == targetReducer && m.name == "Next" {
			ctxNext = m
		}
	}
	if ctxWrite == nil {
		log.Fatalf("Unable to locate \"Write\" method on interface %s.%s", pkg.name, tgt.ctx.name)
	}
	if typ == targetReducer && ctxNext == nil {
		log.Fatalf("Unable to locate \"Next\" method on interface %s.%s", pkg.name, tgt.ctx.name)
	}
	tgt.keyIn = tgt.method.params[0]
	if typ == targetReducer {
		tgt.valueIn = ctxNext.returns[0]
	} else {
		tgt.valueIn = tgt.method.params[1]
	}
	tgt.keyOut = ctxWrite.params[0]
	tgt.valueOut = ctxWrite.params[1]
	return tgt
}

// IsMapper returns true if this Target is a Mapper.
func (t *Target) IsMapper() bool {
	return t.typ == targetMapper
}

// IsReducer returns true if this Target is a Reducer.
func (t *Target) IsReducer() bool {
	return t.typ == targetReducer
}
