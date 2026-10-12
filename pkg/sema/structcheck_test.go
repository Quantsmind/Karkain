package sema

import (
	"strings"
	"testing"

	"karkain/pkg/lexer"
	"karkain/pkg/parser"
)

func resolveErrorsFor(t *testing.T, src string) []ResolveError {
	t.Helper()
	p := parser.New(lexer.New(src))
	prog := p.ParseProgram()
	if len(p.Errors) > 0 {
		t.Fatalf("source failed to parse, cannot test resolver: %v", p.Errors)
	}
	return NewResolver(prog, SourceMap{}).Resolve()
}

// hasErrorContaining reports whether any diagnostic mentions sub.
func hasErrorContaining(errs []ResolveError, sub string) bool {
	for _, e := range errs {
		if strings.Contains(e.Msg, sub) {
			return true
		}
	}
	return false
}

// TestStructFieldChecking_RejectsUnknownFields is the regression gate for the
// silent-acceptance defect: before this pass, every program below was accepted,
// compiled, exited 0 and produced a wrong answer, because a record is a runtime
// map and an unknown key is simply a key nobody reads.
func TestStructFieldChecking_RejectsUnknownFields(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			"unknown field in literal",
			"type Point struct { x int; y int }\nfunc main() { let p = Point{zzz: 1}; print(p.x) }\n",
			"unknown field 'zzz'",
		},
		{
			"typo beside a valid field",
			"type Point struct { x int; y int }\nfunc main() { let p = Point{x: 1, yz: 2}; print(p.x) }\n",
			"unknown field 'yz'",
		},
		{
			"unknown member read on a bound record",
			"type Point struct { x int; y int }\nfunc main() { let p = Point{x: 1, y: 2}; print(p.zzz) }\n",
			"unknown field 'zzz' on Point value 'p'",
		},
		{
			"duplicate field in literal",
			"type Point struct { x int; y int }\nfunc main() { let p = Point{x: 1, x: 2}; print(p.x) }\n",
			"duplicate field 'x'",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := resolveErrorsFor(t, c.src)
			if !hasErrorContaining(errs, c.want) {
				t.Fatalf("expected a diagnostic containing %q, got none (errors: %+v)", c.want, errs)
			}
		})
	}
}

// TestStructFieldChecking_AcceptsValidPrograms is the anti-vacuity half: the
// pass must not reject programs that were already correct. Each case is a shape
// that is easy to get wrong -- a field name that shadows nothing, a record
// passed to a function, a record under a module qualifier, a generic record,
// and a member read whose base is NOT a locally bound record (all of which are
// deliberately out of scope and must stay silent).
func TestStructFieldChecking_AcceptsValidPrograms(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"all fields present",
			"type Point struct { x int; y int }\nfunc main() { let p = Point{x: 1, y: 2}; print(p.x + p.y) }\n"},
		{"record passed to a function",
			"type Point struct { x int; y int }\nfunc shift(p Point) int { return p.x }\nfunc main() { let q = Point{x: 1, y: 2}; print(shift(q)) }\n"},
		{"member read on a parameter is out of scope",
			"type Point struct { x int; y int }\nfunc get(p Point) int { return p.zzz }\nfunc main() { print(1) }\n"},
		{"member read on a map variable is out of scope",
			"func main() { let m = {\"a\": 1}; print(m.whatever) }\n"},
		{"record built inside a conditional",
			"type Point struct { x int; y int }\nfunc main() { if true { let p = Point{x: 1, y: 2}; print(p.x) } }\n"},
		{"record built inside a loop",
			"type Point struct { x int; y int }\nfunc main() { while true { let p = Point{x: 1, y: 2}; print(p.x); break } }\n"},
		{"generic record literal",
			"type Box struct { v int }\nfunc main() { let b = Box{v: 7}; print(b.v) }\n"},
		{"reassigned variable then valid member read",
			"type Point struct { x int; y int }\nfunc main() { let p = Point{x: 1, y: 2}; p = Point{x: 3, y: 4}; print(p.y) }\n"},
		{"unknown non-struct type name in a literal is not this pass's business",
			"func main() { let m = Whatever{zzz: 1}; print(1) }\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if errs := resolveErrorsFor(t, c.src); len(errs) != 0 {
				t.Fatalf("expected no diagnostics, got %d: %v", len(errs), errs)
			}
		})
	}
}

// TestStructFieldChecking_ListsDeclaredFields pins that the diagnostic names the
// record's real fields, so the message is actionable rather than just negative.
// Without this a check could report "unknown field" and still be useless.
func TestStructFieldChecking_ListsDeclaredFields(t *testing.T) {
	errs := resolveErrorsFor(t,
		"type Point struct { x int; y int }\nfunc main() { let p = Point{zzz: 1}; print(p.x) }\n")
	if len(errs) == 0 {
		t.Fatal("expected a diagnostic")
	}
	msg := errs[0].Msg
	if !strings.Contains(msg, "declared fields: x, y") {
		t.Fatalf("diagnostic should list the declared fields in order, got %q", msg)
	}
	if !strings.Contains(msg, "Point") {
		t.Fatalf("diagnostic should name the record type, got %q", msg)
	}
}

// TestStructFieldChecking_UndeclaredTypeIsSilent guards the boundary: a literal
// naming a type that is not declared as a record is left to whatever reports
// undefined types, so this pass cannot double-report or invent a verdict.
func TestStructFieldChecking_UndeclaredTypeIsSilent(t *testing.T) {
	errs := resolveErrorsFor(t, "func main() { let p = Nope{zzz: 1}; print(p.x) }\n")
	if hasErrorContaining(errs, "unknown field") {
		t.Fatalf("this pass must not judge an undeclared type, got %+v", errs)
	}
}
