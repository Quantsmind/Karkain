package parser

import (
	"strings"
	"testing"

	"karkain/pkg/lexer"
)

// parseForErrors runs the full lex+parse pipeline and returns the parser's
// recorded error messages. It exists so malformed-input regression tests can
// assert on the diagnostic set directly (parseSource discards p.Errors).
func parseForErrors(src string) []string {
	l := lexer.New(src)
	p := New(l)
	p.ParseProgram()
	return p.Errors
}

// hasErrContaining reports whether any recorded error contains sub.
func hasErrContaining(errs []string, sub string) bool {
	for _, e := range errs {
		if strings.Contains(e, sub) {
			return true
		}
	}
	return false
}

// These tests pin the systemic correctness rule the malformed-input audit
// established: a program the compiler ACCEPTS must be accepted according to
// Karkain's actual grammar, and an INVALID program must never build silently.
// Every rejected case below is asserted to carry a source location (a "line N"
// prefix), so a bare rejection without a span fails the gate.

// TestMalformedInput_RecoveryDoesNotEatFollowingStatements is the anti-
// corruption gate for the `let x` fix. The original defect was not merely a
// missing diagnostic: the declaration parser consumed one token assuming it
// was '=', so `let x` ate the following `print` keyword and that statement was
// gone from the AST entirely. A diagnostic-only assertion would still pass if
// recovery consumed the statement, so this asserts the AST shape directly.
func TestMalformedInput_RecoveryDoesNotEatFollowingStatements(t *testing.T) {
	src := "func main() {\n    let x\n    print(1);\n}\n"
	p := New(lexer.New(src))
	prog := p.ParseProgram()
	if len(p.Errors) == 0 {
		t.Fatal("expected a diagnostic for a declaration with no '='")
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 top-level statement, got %d", len(prog.Statements))
	}
	fd, ok := prog.Statements[0].(*FuncDecl)
	if !ok {
		t.Fatalf("statement 0 is %T, want *FuncDecl", prog.Statements[0])
	}
	if len(fd.Body) != 2 {
		t.Fatalf("expected 2 statements in the body, got %d: the parser ate a statement", len(fd.Body))
	}
	if _, ok := fd.Body[0].(*VarDeclStmt); !ok {
		t.Errorf("body[0] is %T, want *VarDeclStmt", fd.Body[0])
	}
	if _, ok := fd.Body[1].(*PrintStmt); !ok {
		t.Errorf("body[1] is %T, want *PrintStmt -- the following statement was consumed by the failed declaration", fd.Body[1])
	}
}

// Valid controls: these MUST parse with zero errors. In particular, a
// standalone integer expression statement is a legal `exprStmt` (SPEC §3.1
// lists exprStmt as a statement form; §4.2 lists literals as primary
// expressions), so `print(...);112343` is a well-formed program, not a defect.
func TestMalformedInput_ValidControlsParseClean(t *testing.T) {
	cases := map[string]string{
		"hello": "func main() {\n    print(\"Hello, Karkain!\");\n}\n",
		// The originally-reported case: a legal trailing expression statement.
		"trailing_int_exprstmt":    "func main() {\n    print(\"Hello, Karkain!\");112343\n}\n",
		"multi_statement":          "func main() {\n    print(\"one\");\n    let x = 2;\n    print(x);\n}\n",
		"unary_and_borrow":         "func main() {\n    let a = -5;\n    let b = !true;\n    let c = 1 + 2 * 3;\n    print(a);\n    print(b);\n    print(c);\n}\n",
		"if_while_with_conditions": "func main() {\n    let n = 3;\n    if (n > 1) { print(\"gt\"); }\n    while (n > 0) { n = n - 1; }\n    print(n);\n}\n",
		"simd_builtin":             "func main() {\n    let v = @simd_add(1, 2);\n    print(v);\n}\n",
		// Adjacent-to-the-fix controls. Each sits one token away from a newly
		// rejected shape, so a guard that is too eager fails here.
		"assignment_with_value":       "func main() {\n    let x = 1;\n    x = 2;\n    print(x);\n}\n",
		"assignment_in_expression":    "func main() {\n    let x = 1;\n    let y = (x = 2);\n    print(y);\n}\n",
		"comparison_not_assignment":   "func main() {\n    let a = 1;\n    let b = 2;\n    print(a == b);\n}\n",
		"print_space_form":            "func main() {\n    print 7;\n}\n",
		"print_nested_call":           "func main() {\n    print(len([1, 2, 3]));\n}\n",
		"declaration_with_init":       "func main() {\n    let x = 1;\n    print(x);\n}\n",
		"declaration_typed":           "func main() {\n    let n: int = 7;\n    print(n);\n}\n",
		"declaration_var_init":        "func main() {\n    var c = 0;\n    c = c + 1;\n    print(c);\n}\n",
		"declaration_const_init":      "func main() {\n    const k = 5;\n    print(k);\n}\n",
		"declaration_negative_init":   "func main() {\n    let x = -5;\n    print(x);\n}\n",
		"index_assignment_with_value": "func main() {\n    let a = [1, 2, 3];\n    a[0] = 9;\n    print(a[0]);\n}\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if errs := parseForErrors(src); len(errs) != 0 {
				t.Fatalf("valid program must parse cleanly, got %d error(s): %v", len(errs), errs)
			}
		})
	}
}

// Invalid inputs: each MUST be rejected with a located diagnostic. These are
// the reproduced silent-accept defects (all previously built and ran with
// exit 0 on both the Go and kcc engines before the fix).
func TestMalformedInput_InvalidProgramsRejected(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		wantSub string // a substring the diagnostic must contain
	}{
		{
			name:    "bare_at_between_statements",
			src:     "func main() {\n    print(\"a\"); @@@\n    print(\"b\");\n}\n",
			wantSub: "@",
		},
		{
			name:    "junk_expr_between_statements",
			src:     "func main() {\n    print(\"a\");\n    234%&**\n    print(\"b\");\n}\n",
			wantSub: "expected expression after",
		},
		{
			name:    "at_inside_expression",
			src:     "func main() {\n    let x = 1 @ 2;\n    print(x);\n}\n",
			wantSub: "@",
		},
		{
			name:    "after_close_brace",
			src:     "func main() {\n    print(\"Hello\");\n}\n@@@ garbage 999\n",
			wantSub: "top level",
		},
		{
			name:    "before_first_decl",
			src:     "@@@ leading garbage\nfunc main() {\n    print(\"Hello\");\n}\n",
			wantSub: "@",
		},
		{
			name:    "missing_closing_brace",
			src:     "func main() {\n    print(\"Hello\");\n",
			wantSub: "'}'",
		},
		{
			name:    "missing_print_paren",
			src:     "func main() {\n    print(\"Hello\"\n}\n",
			wantSub: "')'",
		},
		{
			name:    "bare_while",
			src:     "func main() {\n    while\n    print(\"x\");\n}\n",
			wantSub: "condition",
		},
		{
			name:    "bare_if",
			src:     "func main() {\n    if\n    print(\"x\");\n}\n",
			wantSub: "condition",
		},
		{
			name:    "unterminated_string",
			src:     "func main() {\n    print(\"Hello\n}\n",
			wantSub: "unterminated string",
		},
		{
			name:    "dangling_binary_op",
			src:     "func main() {\n    let x = 1 +\n    print(x);\n}\n",
			wantSub: "expected expression after '+'",
		},
		{
			name:    "borrow_no_operand",
			src:     "func main() {\n    let x = &\n    print(x);\n}\n",
			wantSub: "'&'",
		},
		// A dangling assignment. `x =` built a BinaryExpr with a nil right side,
		// and genExpr fabricates make_int(0) for a nil node, so the malformed
		// source compiled to `x = 0`. The binary-operator guard did not cover
		// the assignment path.
		{
			name:    "dangling_assignment",
			src:     "func main() {\n    let x = 1;\n    x =\n    print(x);\n}\n",
			wantSub: "expected expression after '='",
		},
		{
			name:    "dangling_assignment_at_eof",
			src:     "func main() {\n    let x = 1;\n    x =\n}\n",
			wantSub: "expected expression after '='",
		},
		// `print` requires one operand. `print()` built a PrintStmt with a nil
		// value, which codegen turned into print_value(make_int(0)) -- the
		// program printed 0 and exited 0.
		{
			name:    "print_no_argument",
			src:     "func main() {\n    print()\n}\n",
			wantSub: "print requires an expression",
		},
		{
			name:    "print_no_argument_before_statement",
			src:     "func main() {\n    print();\n    let y = 2;\n    print(y);\n}\n",
			wantSub: "print requires an expression",
		},
		// A declaration with '=' but no initializer left Value nil, so the
		// binding silently became 0.
		{
			name:    "declaration_missing_initializer",
			src:     "func main() {\n    let x =\n    print(1);\n}\n",
			wantSub: "expected expression after '='",
		},
		{
			name:    "declaration_missing_initializer_at_eof",
			src:     "func main() {\n    let x =\n}\n",
			wantSub: "expected expression after '='",
		},
		// The declaration parser consumed one token unconditionally, assuming
		// it was '='. `let x` followed by a print statement therefore ate the
		// `print` keyword as the '=' and parsed `(1)` as the initializer: the
		// print statement disappeared entirely and the program built with a
		// body that was never written. This asserts both the diagnostic AND
		// that the following statement survived, because a diagnostic alone
		// would still pass if the recovery silently ate it.
		{
			name:    "declaration_without_equals_keeps_next_statement",
			src:     "func main() {\n    let x\n    print(1);\n}\n",
			wantSub: "expected '=' after 'let x'",
		},
		// Index assignment is a THIRD site that builds a '=' BinaryExpr
		// (parseIdentStatement's postfix branch). A guard there with no test
		// would be unfalsifiable, which a mutation run demonstrated: removing
		// it left every other case green.
		{
			name:    "index_assignment_missing_value",
			src:     "func main() {\n    let a = [1, 2, 3];\n    a[0] =\n    print(a[0]);\n}\n",
			wantSub: "expected expression after '=' operator",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := parseForErrors(tc.src)
			if len(errs) == 0 {
				t.Fatalf("invalid program must be rejected, got 0 errors (silent accept regression)")
			}
			if !hasErrContaining(errs, tc.wantSub) {
				t.Fatalf("expected a diagnostic containing %q, got: %v", tc.wantSub, errs)
			}
			// Every diagnostic must carry a source location.
			for _, e := range errs {
				if !strings.Contains(e, "line ") {
					t.Errorf("diagnostic %q lacks a source location", e)
				}
			}
		})
	}
}

// TestFunctionReturnTypeAnnotations covers the optional return-type
// annotation that may follow a parameter list. `bool` and `float64` lex as
// keyword tokens rather than identifiers; the parser consumed only Ident and
// Fn, so those two left the type in place, the following token was then taken
// as the body's '{', and the body parse derailed with "expected '}' to close
// function body". Every function declaring such a return type failed to
// parse, including one with an empty body.
//
// Each case must parse with ZERO errors. They are all valid programs.
func TestFunctionReturnTypeAnnotations(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"bool_return", "func f() bool { return true }\n"},
		{"bool_return_false", "func f() bool { return false }\n"},
		{"bool_return_empty_body", "func f() bool { }\n"},
		{"bool_return_with_param", "func f(a int, b int) bool { return a }\n"},
		{"float64_return", "func f() float64 { return 1.5 }\n"},
		{"float64_return_empty_body", "func f() float64 { }\n"},
		{"bigint_return", "func f() bigint { return 1 }\n"},
		{"bigfloat_return", "func f() bigfloat { return 1.5 }\n"},
		{"int_return", "func f() int { return 1 }\n"},
		{"string_return", "func f() string { return \"x\" }\n"},
		{"no_return_type", "func f() { print(1) }\n"},
		{"module_shape_from_pkg_module", "func array_contains(arr int[], val int) bool { return false }\n"},
		{"two_bool_functions", "func f() bool { return true }\nfunc g() bool { return false }\nfunc main() { print(f()) }\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if errs := parseForErrors(c.src); len(errs) != 0 {
				t.Fatalf("valid function must parse cleanly, got %d error(s): %v", len(errs), errs)
			}
		})
	}
}