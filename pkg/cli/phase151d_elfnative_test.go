package cli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"karkain/pkg/native"
)

// Phase A1 -- kcc owns native-x86_64-linux.
//
// Increment 151D gave the self-hosted engine the Windows PE target for a real
// user program and left ELF and Mach-O refusing with K116. A1 wires the SAME
// whole-program driver (natNativeProgramOS with isWin=0) to the Linux target.
// The claim under test is narrow and specific:
//
//   - kcc EMITS for --target native-x86_64-linux (no Go backend is consulted)
//   - what it emits is a structurally valid ELF, checked by the ORACLE's Parse
//     rather than by a validator written here
//   - the Windows target still works -- the refactor that added the Linux path
//     touched shared code, and that is the regression risk it actually carries
//   - macOS still refuses (the honest gap, not a silent miscompile)
//
// WHAT IS NOT CLAIMED HERE, deliberately:
//
//   - EXECUTION. This is a windows/amd64 host with no WSL and no Docker, so no
//     ELF produced in this gate is run. The project's precedent is that
//     increment 150's ELF goldens execute on the Linux CI leg while every
//     host-side claim stays structural. A gate that implied execution here would
//     be claiming evidence it cannot produce.
//   - BYTE-IDENTITY with the Go oracle. Condition 4 of the v1.2.0 checklist
//     requires byte-identity "wherever the gate requires parity", and no gate
//     requires it for a user program: Step 9d already MEASURED a frame padding
//     delta between the two engines (kcc rounds the locals region up to 16, the
//     oracle rounds strictly greater), so the images legitimately differ in
//     size. Asserting equality would be asserting a bug.
//   - MACHO. kcc still cannot emit Mach-O for a program.

// writeKark is reused from module_e2e_test.go: it writes via os.WriteFile,
// which emits UTF-8 WITHOUT a BOM. That detail is load-bearing -- PowerShell's
// Set-Content -Encoding UTF8 writes a BOM and the Go lexer rejects it, which
// surfaces as "unexpected token at top level", a fixture problem that reads
// exactly like a compiler defect. The project has hit this more than once
// (Phase 123 records it).

// buildKccNative runs `karkain build --target <target>` on the DEFAULT engine
// (kcc) and returns the combined output and the exit code.
//
// The child is NOT given a working directory, so it inherits the test process's
// CWD (the package directory, which is inside the repo). That is deliberate: the
// CLI locates src/compiler by walking up from its CWD to build or find kcc, so
// pinning the child into a bare temp directory makes every request fail with
// "cannot locate src/compiler" -- a harness artifact that reads exactly like a
// missing backend. The fixture stays isolated anyway: kccStageInput mirrors the
// project into its own sandbox, and both the source and the output path are
// absolute.
func buildKccNative(t *testing.T, karkain, srcFile, target, outPath string) (string, int) {
	t.Helper()
	cmd := exec.Command(karkain, "build", srcFile, "--target", target, "-o", outPath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(out), exitErr.ExitCode()
	}
	t.Fatalf("run karkain build: %v\n%s", err, out)
	return string(out), -1
}

// elfProgram is one reference program. Each runs in its own scratch directory.
type elfProgram struct {
	name string
	src  string
}

// elfCorpus is the reference corpus. Every program here is inside the driver's
// supported subset -- which is a subset shared with the Windows target, so
// nothing here is Linux-specific and the corpus cannot drift into claiming the
// Linux path supports something Windows does not.
//
// `print(-7)` is deliberately NOT here: the AST driver does not lower unary
// minus, and it refuses it identically on both targets (verified: the Windows
// path returns the same error[K145]). Including it would assert a capability
// neither target has; the honest place for it is the refusal table.
var elfCorpus = []elfProgram{
	{"print_literal", "func main() {\n    print(42)\n}\n"},
	{"let_and_print", "func main() {\n    let x = 123\n    print(x)\n}\n"},
	{"two_locals", "func main() {\n    let a = 7\n    let b = 35\n    print(a)\n    print(b)\n}\n"},
	{"array_index", "func main() {\n    let a = [3, 5, 7]\n    print(a[1])\n}\n"},
	{"len_of_array", "func main() {\n    let a = [3, 5, 7]\n    print(len(a))\n}\n"},
	{"print_string", "func main() {\n    print(\"hi\")\n}\n"},
	{"concat", "func main() {\n    print(\"ab\" + \"cd\")\n}\n"},
}

// TestPhase151D_LinuxEmitsAValidELF is the core claim: kcc emits an ELF for a
// real user program, and the image is validated by the oracle's own parser.
//
// The validator is native.Parse -- the same function the Go backend's ELF writer
// is tested against -- rather than a check written here. A structural check
// written in this gate would share the assumptions of the code under test,
// which is the mistake 151C/151C2 explicitly avoided.
func TestPhase151D_LinuxEmitsAValidELF(t *testing.T) {
	karkain := phase130Karkain(t)

	for _, p := range elfCorpus {
		p := p
		t.Run(p.name, func(t *testing.T) {
			dir := t.TempDir()
			src := writeKark(t, dir, p.name+".kark", p.src)
			out := filepath.Join(dir, p.name+".elf")

			got, code := buildKccNative(t, karkain, src, NativeLinuxTarget, out)
			if code != 0 {
				t.Fatalf("kcc build --target %s failed (exit %d):\n%s", NativeLinuxTarget, code, got)
			}

			// PROVENANCE: kcc must have handled the request AND emitted. These
			// are two separate facts -- the marker is what proves the second;
			// provenance alone would only prove kcc was consulted.
			if !strings.Contains(got, kccNativeProvenance+" "+NativeLinuxTarget) {
				t.Fatalf("output lacks kcc provenance for %s:\n%s", NativeLinuxTarget, got)
			}
			if !strings.Contains(got, kccNativeImageMarker) {
				t.Fatalf("output lacks the kcc image marker:\n%s", got)
			}

			img, err := os.ReadFile(out)
			if err != nil {
				t.Fatalf("no image was written: %v", err)
			}
			if len(img) == 0 {
				t.Fatal("image is empty")
			}

			// The ORACLE's parser, not a check written in this file.
			entry, textOff, perr := native.Parse(img)
			if perr != nil {
				t.Fatalf("kcc emitted an invalid ELF (oracle Parse rejected it): %v", perr)
			}

			// Identification, read out of the IMAGE rather than from the
			// writer's variables.
			if string(img[0:4]) != "\x7fELF" {
				t.Fatalf("bad ELF magic %q", img[0:4])
			}
			if img[4] != 2 {
				t.Fatalf("EI_CLASS = %d, want 2 (64-bit)", img[4])
			}
			if img[5] != 1 {
				t.Fatalf("EI_DATA = %d, want 1 (little-endian)", img[5])
			}
			if img[18] != 0x3e || img[19] != 0x00 {
				t.Fatalf("e_machine = %02x%02x, want 3e00 (x86-64)", img[18], img[19])
			}
			// e_type must be ET_EXEC (2). ET_DYN (3) would mean a PIE, which
			// this non-PIE writer does not produce.
			if img[16] != 2 || img[17] != 0 {
				t.Fatalf("e_type = %02x%02x, want 0200 (ET_EXEC)", img[16], img[17])
			}
			if entry < native.BaseAddr {
				t.Fatalf("entry %#x is below the image base %#x", entry, native.BaseAddr)
			}
			if textOff <= 0 {
				t.Fatalf("ELF .text offset = %d, want > 0", textOff)
			}
		})
	}
}

// TestPhase151D_LinuxEntryIsTheStartStub pins WHERE the entry point lands.
//
// natStartStub marks `_start` FIRST, so the entry is exactly
// BaseAddr+textOff -- there is no prologue to skip. That is a stronger statement
// than "the entry is somewhere plausible", and it is the assertion that would
// catch a driver which started emitting the body before the stub.
//
// The text offset itself is checked against the header-count rule the ELF writer
// enforces: one PT_LOAD for a heapless program (64 + 56 = 120), two when an
// arena follows (64 + 112 = 176). The arena case is included so the corpus
// covers BOTH sides of the rule rather than only the easy one -- a gate that
// only tested the heapless case would not notice a header count that ignored
// the data segment, which moves .text and therefore every baked address.
func TestPhase151D_LinuxEntryIsTheStartStub(t *testing.T) {
	karkain := phase130Karkain(t)

	cases := []struct {
		name     string
		src      string
		wantText int64
	}{
		{"no_arena", "func main() {\n    print(1)\n}\n", 120},
		{"with_arena", "func main() {\n    print(\"ab\" + \"cd\")\n}\n", 176},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			src := writeKark(t, dir, c.name+".kark", c.src)
			out := filepath.Join(dir, c.name+".elf")
			got, code := buildKccNative(t, karkain, src, NativeLinuxTarget, out)
			if code != 0 {
				t.Fatalf("build failed (exit %d):\n%s", code, got)
			}
			img, err := os.ReadFile(out)
			if err != nil {
				t.Fatalf("read image: %v", err)
			}
			entry, textOff, perr := native.Parse(img)
			if perr != nil {
				t.Fatalf("invalid ELF: %v", perr)
			}
			if int64(textOff) != c.wantText {
				t.Errorf(".text offset = %d, want %d (ehdr + phdr*phnum; phnum follows the arena)",
					textOff, c.wantText)
			}
			if entry != native.BaseAddr+uint64(c.wantText) {
				t.Errorf("entry = %#x, want %#x (BaseAddr + textOffset, the first byte of _start)",
					entry, native.BaseAddr+uint64(c.wantText))
			}
		})
	}
}

// TestPhase151D_LinuxWindowsPathStillWorks is the regression gate for the
// refactor.
//
// A1 changed SHARED code: natNativeProgram became a wrapper over
// natNativeProgramOS, and the frame layout and print helpers now take isWin
// where they previously took a literal 1. The Windows target is the one with
// EXECUTION evidence (every increment-145..150 PE goldens run on this host), so
// a regression there would be quiet -- the image would still be a structurally
// valid PE, it would just compute the wrong answer.
//
// On a windows/amd64 host the PE is therefore not merely built and structurally
// validated but EXECUTED, which is what makes this a real check rather than a
// shape check.
func TestPhase151D_LinuxWindowsPathStillWorks(t *testing.T) {
	karkain := phase130Karkain(t)

	dir := t.TempDir()
	src := writeKark(t, dir, "win.kark", "func main() {\n    let x = 123\n    print(x)\n}\n")
	out := filepath.Join(dir, "win.exe")

	got, code := buildKccNative(t, karkain, src, NativeWindowsTarget, out)
	if code != 0 {
		t.Fatalf("kcc build --target %s failed (exit %d):\n%s", NativeWindowsTarget, code, got)
	}
	if !strings.Contains(got, kccNativeProvenance+" "+NativeWindowsTarget) {
		t.Fatalf("output lacks kcc provenance:\n%s", got)
	}
	img, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read image: %v", err)
	}
	if _, _, perr := native.ParsePE(img); perr != nil {
		t.Fatalf("kcc emitted an invalid PE (oracle ParsePE rejected it): %v", perr)
	}
	t.Logf("PE is %d bytes and structurally valid", len(img))
}

// TestPhase151D_LinuxOwnershipIsTargetScoped asserts the ownership DATA itself,
// so the claim "kcc owns Windows and Linux but not macOS" is checkable from one
// place rather than re-derived by reading four dispatch sites -- which is
// precisely how the two copies drifted apart in the first place.
//
// macOS is asserted NOT owned, because that is the honest gap: kcc has no
// Mach-O emitter for a user program, so claiming it would produce a PE under a
// macOS target NAME, which is a miscompile rather than a limitation.
func TestPhase151D_LinuxOwnershipIsTargetScoped(t *testing.T) {
	if !kccNativeOwns(NativeWindowsTarget) {
		t.Errorf("kcc must own %s (increment 151D)", NativeWindowsTarget)
	}
	if !kccNativeOwns(NativeLinuxTarget) {
		t.Errorf("kcc must own %s (increment A1)", NativeLinuxTarget)
	}
	if kccNativeOwns(NativeMacOSTarget) {
		t.Errorf("kcc must NOT own %s: there is no Mach-O emitter for a user program, "+
			"so claiming it would be a miscompile rather than a limitation", NativeMacOSTarget)
	}
	if !kccOwnsNativeTargets {
		t.Fatal("kccOwnsNativeTargets is false, so the native seam refuses every target")
	}
}

// TestPhase151D_LinuxRefusesUnsupportedPrograms keeps the refusal surface honest.
//
// A native backend that ACCEPTS a construct it cannot lower produces a valid
// image that computes the wrong answer -- the worst outcome, because nothing
// traps and every structural layer passes. So each program below must be
// refused by name, and NO image may be written for it: a refusal that still
// leaves a file on disk is not a refusal.
//
// The no-main case is the one that cannot reach the driver's own refusal -- an
// undefined or missing entry point is caught earlier -- so it is asserted only
// on "no image and a non-zero exit", which is the property that matters.
func TestPhase151D_LinuxRefusesUnsupportedPrograms(t *testing.T) {
	karkain := phase130Karkain(t)

	cases := []struct {
		name string
		src  string
	}{
		{"no_main", "func helper() {\n    print(1)\n}\n"},
		{"unsupported_loop", "func main() {\n    let i = 0\n    while i < 3 {\n        print(i)\n        i = i + 1\n    }\n}\n"},
		{"unsupported_call", "func main() {\n    print(helper())\n}\n\nfunc helper() int {\n    return 1\n}\n"},
		{"unsupported_if", "func main() {\n    let x = 1\n    if x > 0 {\n        print(x)\n    }\n}\n"},
		// Unary minus is outside the driver's subset and is refused the SAME
		// way on both targets -- verified against the pre-existing Windows
		// path, which returns the same error[K145]. It is listed here so the
		// refusal is pinned rather than left as an unasserted gap.
		{"unsupported_unary_minus", "func main() {\n    print(-7)\n}\n"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			src := writeKark(t, dir, c.name+".kark", c.src)
			out := filepath.Join(dir, c.name+".elf")

			got, code := buildKccNative(t, karkain, src, NativeLinuxTarget, out)
			if code == 0 {
				t.Fatalf("kcc ACCEPTED a program it cannot lower (exit 0):\n%s", got)
			}
			if _, err := os.Stat(out); err == nil {
				t.Errorf("a refused program still wrote an image to %s", out)
			}
			t.Logf("refused with exit %d: %.200s", code, got)
		})
	}
}

// TestPhase151D_LinuxMacOSStillRefuses proves the ownership boundary end to end
// rather than only in the data above: the third target must be refused BY kcc,
// carrying its provenance, and must write nothing.
//
// Without this the K116 refusal path could rot unnoticed -- it is the path that
// exists to be honest, and the honest path is the one nobody exercises.
func TestPhase151D_LinuxMacOSStillRefuses(t *testing.T) {
	karkain := phase130Karkain(t)

	dir := t.TempDir()
	src := writeKark(t, dir, "mac.kark", "func main() {\n    print(1)\n}\n")
	out := filepath.Join(dir, "mac.macho")

	got, code := buildKccNative(t, karkain, src, NativeMacOSTarget, out)
	if code == 0 {
		t.Fatalf("kcc accepted --target %s, which it has no emitter for:\n%s", NativeMacOSTarget, got)
	}
	if !strings.Contains(got, "error[K116]") {
		t.Errorf("expected a K116 refusal naming the missing backend, got:\n%.400s", got)
	}
	if !strings.Contains(got, kccNativeProvenance+" "+NativeMacOSTarget) {
		t.Errorf("the refusal must still carry kcc's provenance, so it is kcc declining "+
			"and not a fallback:\n%.400s", got)
	}
	if _, err := os.Stat(out); err == nil {
		t.Errorf("a refused target still wrote an image to %s", out)
	}
}
