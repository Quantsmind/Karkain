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

// Valid controls: these MUST parse with zero errors. In particular, a
// standalone integer expression statement is a legal `exprStmt` (SPEC §3.1
// lists exprStmt as a statement form; §4.2 lists literals as primary
// expressions), so `print(...);112343` is a well-formed program, not a defect.
func TestMalformedInput_ValidControlsParseClean(t *testing.T) {
	cases := map[string]string{
		"hello": "func main() {\n    print(\"Hello, Karkain!\");\n}\n",
		// The originally-reported case: a legal trailing expression statement.
		"trailing_int_exprstmt": "func main() {\n    print(\"Hello, Karkain!\");112343\n}\n",
		"multi_statement": "func main() {\n    print(\"one\");\n    let x = 2;\n    print(x);\n}\n",
		"unary_and_borrow": "func main() {\n    let a = -5;\n    let b = !true;\n    let c = 1 + 2 * 3;\n    print(a);\n    print(b);\n    print(c);\n}\n",
		"if_while_with_conditions": "func main() {\n    let n = 3;\n    if (n > 1) { print(\"gt\"); }\n    while (n > 0) { n = n - 1; }\n    print(n);\n}\n",
		"simd_builtin": "func main() {\n    let v = @simd_add(1, 2);\n    print(v);\n}\n",
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
