package sema

import (
	"fmt"

	"karkain/pkg/parser"
)

// Struct field checking.
//
// A Karkain record is a runtime map keyed by field name, so the compiler has
// never needed a struct layout to build one. That is exactly why an unknown
// field used to be silently accepted: `Point{zzz: 1}` stored an entry nobody
// can ever read, and `p.zzz` read a key that was never written. The program
// compiled, exited 0, and printed a wrong answer -- the worst failure mode a
// compiler can have, because nothing anywhere reports it.
//
// This pass consults the declared `type X struct { ... }` fields and rejects
// two shapes:
//
//	Point{zzz: 1}      unknown field name in a record literal
//	p.zzz              member read of a field the record type does not declare
//
// Deliberate limits, each one a boundary rather than an oversight:
//
//   - It is diagnostics-only. It never rewrites the AST, so it cannot change
//     emitted output and cannot affect codegen determinism.
//   - A member read is only checked when the base identifier is bound to a
//     record literal of a declared type within the same function body. Any
//     other base (a parameter, a call result, a map, a module qualifier, an
//     enum variant) is left alone, so the pass can only ever report a field
//     name that is positively known to be wrong.
//   - Missing fields and duplicate fields are NOT rejected. A record may be
//     built up across statements, and requiring completeness is a semantic
//     decision beyond this pass's remit.
//
// Scope note: this runs on the Go reference engine's check path (`--engine go`
// build/run/check) and in the LSP. The self-hosted engine's own check path does
// not consult this pass, so `karkain check` under the default kcc engine does
// not yet report these diagnostics. That is a known, recorded strictness
// difference, not a silent one.

// structFields returns the declared field names of a record type, or nil when
// the name is not a declared record type.
func (r *Resolver) structFields(typeName string) map[string]bool {
	return r.structs[typeName]
}

// collectStructTypes records every `type X struct { ... }` declaration and its
// field names. Called once from NewResolver after collectDefs.
func (r *Resolver) collectStructTypes(stmts []parser.Node) {
	for _, s := range stmts {
		sd, ok := s.(*parser.StructDeclStmt)
		if !ok || sd.Name == "" {
			continue
		}
		fields := make(map[string]bool, len(sd.Fields))
		for _, f := range sd.Fields {
			if f.Name != "" {
				fields[f.Name] = true
			}
		}
		if _, seen := r.structs[sd.Name]; !seen {
			r.structs[sd.Name] = fields
			order := make([]string, 0, len(sd.Fields))
			for _, f := range sd.Fields {
				if f.Name != "" {
					order = append(order, f.Name)
				}
			}
			r.structFieldOrder[sd.Name] = order
		}
	}
}

// checkStructFields validates record literal field names and member reads for
// every function body in the compile unit.
func (r *Resolver) checkStructFields(stmts []parser.Node) {
	for _, s := range stmts {
		switch n := s.(type) {
		case *parser.FuncDecl:
			r.checkStructFieldsIn(n.Body)
		case *parser.LambdaExpr:
			r.checkStructFieldsIn(n.Body)
		}
	}
}

// checkStructFieldsIn walks one function body. bound tracks local names that
// hold a value of a declared record type, so `p.zzz` can be checked after
// `let p = Point{...}`.
func (r *Resolver) checkStructFieldsIn(body []parser.Node) {
	bound := make(map[string]string)
	var walk func(n parser.Node)
	walk = func(n parser.Node) {
		switch v := n.(type) {
		case nil:
			return
		case *parser.StructLiteral:
			r.checkStructLiteral(v)
			for _, f := range v.Fields {
				walk(f)
			}
		case *parser.VarDeclStmt:
			if v.Name != "" {
				if sl, ok := v.Value.(*parser.StructLiteral); ok {
					if r.structFields(sl.TypeName) != nil {
						bound[v.Name] = sl.TypeName
					} else {
						delete(bound, v.Name)
					}
				} else {
					delete(bound, v.Name)
				}
			}
			walk(v.Value)
		case *parser.DotExpr:
			r.checkStructMember(v, bound)
			walk(v.Left)
		case *parser.BinaryExpr:
			walk(v.Left)
			walk(v.Right)
		case *parser.UnaryExpr:
			walk(v.Operand)
		case *parser.CallExpr:
			for _, a := range v.Args {
				walk(a)
			}
		case *parser.IndirectCallExpr:
			walk(v.Target)
			for _, a := range v.Args {
				walk(a)
			}
		case *parser.IndexExpr:
			walk(v.Left)
			walk(v.Index)
		case *parser.SliceExpr:
			walk(v.Target)
			walk(v.Start)
			walk(v.End)
		case *parser.ExprStmt:
			walk(v.Expression)
		case *parser.PrintStmt:
			walk(v.Value)
		case *parser.ReturnStmt:
			walk(v.Value)
		case *parser.IfStmt:
			walk(v.Condition)
			for _, x := range v.Consequence {
				walk(x)
			}
			for _, x := range v.Alternative {
				walk(x)
			}
		case *parser.WhileStmt:
			walk(v.Condition)
			for _, x := range v.Body {
				walk(x)
			}
		case *parser.ForStmt:
			walk(v.Init)
			walk(v.Condition)
			walk(v.Post)
			for _, x := range v.Body {
				walk(x)
			}
		case *parser.ForInStmt:
			walk(v.Iter)
			for _, x := range v.Body {
				walk(x)
			}
		case *parser.MatchExpr:
			walk(v.Value)
			for _, a := range v.Arms {
				walk(a.Body)
			}
		case *parser.ArrayLiteral:
			for _, e := range v.Elements {
				walk(e)
			}
		case *parser.MapLiteral:
			for _, e := range v.Keys {
				walk(e)
			}
			for _, e := range v.Values {
				walk(e)
			}
		}
	}
	for _, s := range body {
		walk(s)
	}
}

// checkStructLiteral rejects a field name the record type does not declare.
func (r *Resolver) checkStructLiteral(lit *parser.StructLiteral) {
	fields := r.structFields(lit.TypeName)
	if fields == nil {
		return
	}
	seen := make(map[string]bool, len(lit.Fields))
	for _, f := range lit.Fields {
		be, ok := f.(*parser.BinaryExpr)
		if !ok {
			continue
		}
		key, ok := be.Left.(*parser.Identifier)
		if !ok || key.Name == "" {
			continue
		}
		if !fields[key.Name] {
			r.errors = append(r.errors, ResolveError{
				Line: lit.Line,
				Msg: fmt.Sprintf("unknown field '%s' in %s literal (declared fields: %s)",
					key.Name, lit.TypeName, r.sortedFieldNames(lit.TypeName)),
			})
			continue
		}
		if seen[key.Name] {
			r.errors = append(r.errors, ResolveError{
				Line: lit.Line,
				Msg:  fmt.Sprintf("duplicate field '%s' in %s literal", key.Name, lit.TypeName),
			})
		}
		seen[key.Name] = true
	}
}

// checkStructMember rejects `x.field` when x is known to hold a declared record
// type that has no such field.
func (r *Resolver) checkStructMember(dot *parser.DotExpr, bound map[string]string) {
	id, ok := dot.Left.(*parser.Identifier)
	if !ok || id.Name == "" {
		return
	}
	typeName, boundHere := bound[id.Name]
	if !boundHere {
		return
	}
	fields := r.structFields(typeName)
	if fields == nil || fields[dot.Right] {
		return
	}
	r.errors = append(r.errors, ResolveError{
		Line: dot.Line,
		Msg: fmt.Sprintf("unknown field '%s' on %s value '%s' (declared fields: %s)",
			dot.Right, typeName, id.Name, r.sortedFieldNames(typeName)),
	})
}

// sortedFieldNames renders a record type's declared fields in declaration
// order so the diagnostic lists them the way the type is written.
func (r *Resolver) sortedFieldNames(typeName string) string {
	order := r.structFieldOrder[typeName]
	if len(order) == 0 {
		return "(none)"
	}
	out := ""
	for i, n := range order {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}
