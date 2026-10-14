package cli

import (
	"crypto/sha256"
	"testing"
)

// Phase D2 -- kcc and the Go oracle emit the SAME BYTES for the same program.
//
// This is the first gate in the project that asserts byte-identity of a kcc
// user-program image against the Go oracle's, and it is the direct evidence for
// v1.2.0 closure condition 4 ("native images are byte-identical between the Go
// compiler and kcc wherever the gate requires parity").
//
// WHY THIS IS A DIFFERENTIAL AND NOT A GOLDEN: a golden pins the CURRENT bytes,
// and for most of increment 151 the current bytes were WRONG -- the images
// differed by a constant 39 bytes for reasons nobody had measured. A golden
// written during that period would have frozen the divergence and called it
// expected. Comparing the two engines means the assertion can only pass if BOTH
// agree, so a regression in either is caught and neither can be "pinned" as the
// truth.
//
// WHAT IS COVERED AND WHAT IS NOT, stated here so the gap is not inferred from
// a passing run:
//
//   - COVERED: SEVEN of the EIGHT program shapes below are byte-identical --
//     literals, locals, multiple locals, string literals, the ALLOCATING case
//     (string concatenation, whose arena is floored the way the oracle floors
//     it), a returned value, and `len()` of an array. D3 closed the allocating
//     case; D4 closed the two smallest of the three remaining gaps.
//   - NOT COVERED: ONE shape, `array_index`, which needs the Go backend's
//     bounds-checked array HELPER routines and the Phase-152-B0 runtime-error
//     reporter (neither of which exists on the kcc target). It is asserted to
//     DIFFER by TestPhaseD2_KnownNonParityIsStillNonParity below, which records
//     the measured numbers and why it needs its own baselined increment.
//
// Note the shape count moved as work landed. This header was written at D2 when
// four shapes passed and one was expected to differ; D3 added two cases (one to
// each list) and D4 added two more (both moved into this list when the
// non-parity gate failed). A passing run therefore says "these seven agree" and
// NOT "the subset is complete" -- the second list is where the remainder lives.
//
// EXECUTION: only the Windows PE can execute on this host, and it is checked
// separately by TestPhaseB_OtherTargetsAreUnchanged. The ELF bytes compared here
// are never run on the dev host (no WSL, no Docker); the Linux CI leg is where
// an ELF produced by these exact bytes executes.
func TestPhaseD2_KccAndGoEmitIdenticalImages(t *testing.T) {
	karkain := phase130Karkain(t)

	cases := []struct {
		name string
		src  string
	}{
		{"print_literal", "func main() {\n    print(42)\n}\n"},
		{"local", "func main() {\n    let x = 123\n    print(x)\n}\n"},
		{"two_locals", "func main() {\n    let a = 7\n    let b = 35\n    print(a)\n    print(b)\n}\n"},
		{"string_literal", "func main() {\n    print(\"hi\")\n}\n"},
		// Phase D3 added the allocating shape: the arena bound is floored the
		// way the oracle floors it, and the arena image is part of the image,
		// so getting it wrong moved every baked arena address. It was asserted
		// to DIFFER by TestPhaseD2_KnownNonParityIsStillNonParity, which failed
		// when it became identical -- which is the signal that this row belongs
		// HERE.
		{"concat", "func main() {\n    print(\"ab\" + \"cd\")\n}\n"},
		// Phase D4 added the two shapes that were the SMALLEST remaining gaps.
		//
		// return_value was a missing 5-byte `jmp rel32`: the oracle routes every
		// return through one `$ret` label so the frame teardown is emitted once
		// per function, and kcc inlined the epilogue with no jump at all.
		//
		// len_of_array was a missing REX.B: the array element index went into RCX
		// with `natMovRegImm32`, which emits a bare `B9+rd` with no REX prefix --
		// correct for rax..rdi, but for RCX that IS `mov ecx, imm32`, so the
		// index landed in the low half of the register the scaled store indexes
		// with.
		//
		// Both were asserted to DIFFER by TestPhaseD2_KnownNonParityIsStillNonParity
		// below, and both moved out of it when that gate failed -- which is the
		// signal that the row belongs HERE.
		{"return_value", "func main() {\n    let z = 7\n    return z\n}\n"},
		{"len_of_array", "func main() {\n    let a = [3, 5, 7]\n    print(len(a))\n}\n"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			// The Go engine is pinned through the environment rather than a
			// flag, because the fallback this gate is about happens in the
			// Go-side dispatch and pinning the env is what makes it
			// observable at all.
			kccRes := nativeBuildWithEngine(t, karkain, c.src, NativeLinuxTarget, "kcc", c.name+"-kcc.elf")
			if kccRes.ExitCode != 0 || len(kccRes.Image) == 0 {
				t.Fatalf("kcc build produced no image (exit %d):\n%s", kccRes.ExitCode, kccRes.Output)
			}
			goRes := nativeBuildWithEngine(t, karkain, c.src, NativeLinuxTarget, "go", c.name+"-go.elf")
			if goRes.ExitCode != 0 || len(goRes.Image) == 0 {
				t.Fatalf("go build produced no image (exit %d):\n%s", goRes.ExitCode, goRes.Output)
			}
			kccBytes, goBytes := kccRes.Image, goRes.Image

			kh := sha256.Sum256(kccBytes)
			gh := sha256.Sum256(goBytes)
			if kh != gh {
				t.Fatalf("kcc and the Go oracle emitted DIFFERENT images for the same "+
					"program and target.\n  kcc: %d bytes sha %x\n  go : %d bytes sha %x\n"+
					"  closure condition 4 requires byte-identity; if the difference is "+
					"expected at this point, the program belongs in the not-yet-parity "+
					"list in this file's header rather than here.", len(kccBytes), kh,
					len(goBytes), gh)
			}
			t.Logf("%s: byte-identical, %d bytes, sha %x", c.name, len(kccBytes), kh[:6])
		})
	}
}

// TestPhaseD2_KnownNonParityIsStillNonParity pins the REMAINING differences so
// they cannot be forgotten.
//
// Phase D4 closed the two SMALLEST of the three (return_value and
// len_of_array moved into the parity list above when this gate failed). ONE shape
// is measurably still different, and it is now the largest gap in the increment
// by a wide margin:
//
//	array_index  kcc 496  vs go 1096   (was 481; D4's REX.B fix moved it 15)
//
// Asserting it DIFFER is the point. It converts "we know this differs" from a
// comment nobody re-reads into an assertion that has to be updated on purpose,
// and it fails loudly if a shape is accidentally added to the parity list while
// it is still wrong. A missing assertion is how a known gap quietly becomes an
// unknown one.
//
// WHY array_index is deferred rather than merely unfinished: it is not a
// byte-form difference. Both images were dumped and mapped region by region --
// karkain_main is nearly identical (Go `sub rsp,0x290`, kcc `sub rsp,0x288`),
// but Go emits 12 `ret`s where kcc emits 4, because the Go backend emits a set
// of bounds-checked ARRAY HELPER routines (three, ~200 bytes, each with its own
// `jge`/`jnz` guard) while kcc inlines its bounds check at the use site. Those
// helpers are what call the Phase-152-B0 runtime-error reporter, which exists
// for the Go backend only -- 152-B0 added 379 lines to pkg/native/program.go and
// touched no .kark file. On the kcc target an out-of-range index TRAPS with Int3
// instead of reporting "runtime error: array index out of range at <file>:<line>"
// and exiting 1.
//
// A port of the reporter was attempted and REVERTED. It emitted, and a valid
// index then printed correctly, but the out-of-range case crashed with
// 0xC0000005 where the Go backend exits 1 cleanly -- 152-B0 already documents
// that GetStdHandle(STD_ERROR_HANDLE) returns INVALID_HANDLE_VALUE on minimal
// PEs, which is a trap in exactly this area. Shipping a diagnostic that crashes
// instead of trapping would be strictly worse than the honest Int3, so the port
// was reverted rather than shipped half-working.
//
// So this needs a properly BASELINED increment, not a patch: the array helper
// set, a working PE stderr path against the INVALID_HANDLE_VALUE limitation,
// and source-basename plumbing so `file:line` is real rather than a placeholder.
func TestPhaseD2_KnownNonParityIsStillNonParity(t *testing.T) {
	karkain := phase130Karkain(t)

	cases := []struct {
		name string
		src  string
	}{
		{"array_index", "func main() {\n    let a = [3, 5, 7]\n    print(a[1])\n}\n"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			kccRes := nativeBuildWithEngine(t, karkain, c.src, NativeLinuxTarget, "kcc", c.name+"-kcc.elf")
			if kccRes.ExitCode != 0 || len(kccRes.Image) == 0 {
				t.Fatalf("kcc build produced no image (exit %d):\n%s", kccRes.ExitCode, kccRes.Output)
			}
			goRes := nativeBuildWithEngine(t, karkain, c.src, NativeLinuxTarget, "go", c.name+"-go.elf")
			if goRes.ExitCode != 0 || len(goRes.Image) == 0 {
				t.Fatalf("go build produced no image (exit %d):\n%s", goRes.ExitCode, goRes.Output)
			}
			kh := sha256.Sum256(kccRes.Image)
			gh := sha256.Sum256(goRes.Image)
			if kh == gh {
				t.Errorf("this shape is now byte-identical (sha %x). That is progress: "+
					"move it into TestPhaseD2_KccAndGoEmitIdenticalImages so it is covered "+
					"by the parity assertion rather than merely noted here.", kh)
			}
			t.Logf("known non-parity: kcc=%d go=%d", len(kccRes.Image), len(goRes.Image))
		})
	}
}
