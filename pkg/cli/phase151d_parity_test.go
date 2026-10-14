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
//   - COVERED: every program in the supported subset whose image involves only
//     the integer/string value model -- literals, locals, multiple locals, and
//     string literals. Four of the four program shapes below are byte-identical.
//   - NOT COVERED: programs that ALLOCATE (string concatenation, maps, push),
//     because the arena allocator is emitted by a different path whose bytes
//     have not been reconciled. `c_concat` is asserted to DIFFER, so the gate
//     fails loudly if that ever silently regresses to "parity" by accident and
//     also documents the remaining work.
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

// TestPhaseD2_KnownNonParityIsStillNonParity pins the REMAINING difference so it
// cannot be forgotten.
//
// A concatenating program allocates, and the arena allocator's bytes are not yet
// reconciled with the oracle. Today that program produces a different image. This
// test asserts they DIFFER. That looks backwards until you read what it is for:
//
//   - if it ever became byte-identical, that is real progress and this test
//     FAILS, telling us to move the program into the parity list above;
//   - if a future change made kcc emit a VALID but WRONG arena, the parity list
//     would not catch it either -- so the structural gates still apply.
//
// In other words it converts "we know this differs" from a comment nobody
// re-reads into an assertion that has to be updated on purpose. A missing
// assertion is how a known gap quietly becomes an unknown one.
func TestPhaseD2_KnownNonParityIsStillNonParity(t *testing.T) {
	karkain := phase130Karkain(t)

	const src = "func main() {\n    print(\"ab\" + \"cd\")\n}\n"
	kccRes := nativeBuildWithEngine(t, karkain, src, NativeLinuxTarget, "kcc", "concat-kcc.elf")
	if kccRes.ExitCode != 0 || len(kccRes.Image) == 0 {
		t.Fatalf("kcc build produced no image (exit %d):\n%s", kccRes.ExitCode, kccRes.Output)
	}
	goRes := nativeBuildWithEngine(t, karkain, src, NativeLinuxTarget, "go", "concat-go.elf")
	if goRes.ExitCode != 0 || len(goRes.Image) == 0 {
		t.Fatalf("go build produced no image (exit %d):\n%s", goRes.ExitCode, goRes.Output)
	}
	kccBytes, goBytes := kccRes.Image, goRes.Image

	kh := sha256.Sum256(kccBytes)
	gh := sha256.Sum256(goBytes)
	if kh == gh {
		t.Errorf("the allocating program is now byte-identical (sha %x). That is "+
			"progress: move it out of the known-non-parity list and into "+
			"TestPhaseD2_KccAndGoEmitIdenticalImages so it is covered by the parity "+
			"assertion.", kh)
	}
	t.Logf("known non-parity (arena allocator not yet reconciled): kcc=%d go=%d",
		len(kccBytes), len(goBytes))
}
