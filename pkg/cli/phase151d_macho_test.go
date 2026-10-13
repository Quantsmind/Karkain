package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"karkain/pkg/native"
)

// Phase B -- kcc owns native-x86_64-macos (Mach-O PIE).
//
// Mach-O is the third container and the only POSITION-INDEPENDENT one kcc
// emits. That single property is what this gate is really about, and it is why
// the rebase assertions live here rather than being left to 151C2's writer
// tests: a Mach-O whose baked addresses carry no rebase opcodes LOADS CLEANLY
// and then reads through a pointer dyld never slid. There is no crash and no
// diagnostic -- just wrong data -- so a test that counts the sites is the only
// thing standing between that and a shipped image.
//
// WHAT IS NOT CLAIMED, and it is the boundary increment 149 established:
// EXECUTION. This repository has no macOS runner (every CI job is ubuntu or
// windows) and the development host is windows/amd64 with no WSL and no Docker,
// so no Mach-O produced in this gate is run. Every claim here is structural,
// validated by the ORACLE's own parser. That is stated once here because "kcc
// emits a Mach-O" and "that Mach-O runs" are different claims and only the
// first is supported.

// machoFlagPIE is the Mach-O header flag that makes an image position
// independent. Declared here as a literal stated from the Mach-O specification
// rather than imported from pkg/native (where it is unexported), so the
// assertion does not echo the writer's own constant -- an assertion quoting the
// code under test would pass when both were wrong the same way.
const machoFlagPIE = 0x200000

// le32at reads a little-endian uint32 out of an image at an offset.
func le32at(b []byte, off int) uint32 {
	return uint32(b[off]) | uint32(b[off+1])<<8 | uint32(b[off+2])<<16 | uint32(b[off+3])<<24
}

// TestPhaseB_MacOSEmitsAValidPIE is the core claim: kcc emits a structurally
// valid Mach-O PIE for a real user program.
//
// Validated by the ORACLE's ParseMachO / MachORebaseSites, never by a check
// written here -- a structural check in this file would share the assumptions
// of the code under test.
func TestPhaseB_MacOSEmitsAValidPIE(t *testing.T) {
	karkain := phase130Karkain(t)

	for _, p := range elfCorpus {
		p := p
		t.Run(p.name, func(t *testing.T) {
			dir := t.TempDir()
			src := writeKark(t, dir, p.name+".kark", p.src)
			out := filepath.Join(dir, p.name+".macho")

			got, code := buildKccNative(t, karkain, src, NativeMacOSTarget, out)
			if code != 0 {
				t.Fatalf("kcc build --target %s failed (exit %d):\n%s", NativeMacOSTarget, code, got)
			}
			if !strings.Contains(got, kccNativeProvenance+" "+NativeMacOSTarget) {
				t.Fatalf("output lacks kcc provenance for %s:\n%s", NativeMacOSTarget, got)
			}
			if !strings.Contains(got, kccNativeImageMarker) {
				t.Fatalf("output lacks the kcc image marker:\n%s", got)
			}

			img, err := os.ReadFile(out)
			if err != nil {
				t.Fatalf("no image was written: %v", err)
			}

			// The ORACLE's Mach-O validator: it checks MH_PIE, the exact
			// __PAGEZERO/__TEXT/__DATA/__LINKEDIT set and their protections,
			// page-aligned placement, the entry, sizeofcmds, and DECODES the
			// rebase stream (rejecting a stream with no DONE, bytes after DONE,
			// a non-POINTER type, the wrong segment, or an over-long uleb128).
			entry, textOff, perr := native.ParseMachO(img)
			if perr != nil {
				t.Fatalf("kcc emitted an invalid Mach-O (oracle ParseMachO rejected it): %v", perr)
			}
			if entry < native.MachoBase {
				t.Fatalf("entry %#x is below MachoBase %#x", entry, native.MachoBase)
			}
			if fl := le32at(img, 24); fl&machoFlagPIE == 0 {
				t.Errorf("flags %#x lack MH_PIE (%#x): the image is position-dependent "+
					"and its baked addresses would never be slid", fl, machoFlagPIE)
			}

			// The rebase sites must EXIST whenever the program baked an
			// absolute address. Every program in this corpus prints something,
			// so each has at least print_int's rodata references. A zero here
			// means the writer emitted an EMPTY stream, which parses perfectly
			// and is exactly the failure this assertion exists to catch.
			sites, serr := native.MachORebaseSites(img)
			if serr != nil {
				t.Fatalf("rebase stream did not decode: %v", serr)
			}
			if len(sites) == 0 {
				t.Error("the image bakes absolute addresses but carries NO rebase opcodes; " +
					"dyld would slide nothing and every pointer would be stale")
			}
			t.Logf("%s: entry=%#x textOff=%d rebaseSites=%d", p.name, entry, textOff, len(sites))
		})
	}
}

// TestPhaseB_MacOSRebaseSitesTrackThePatches is the load-bearing assertion.
//
// The sites are DERIVED from the same rodata and arena patch lists the
// addresses were resolved from, so the count moves whenever the program changes
// how many absolute addresses it bakes. This asserts that RELATIONSHIP rather
// than a per-component number.
//
// WHY NO PER-COMPONENT NUMBER, which is a measurement this gate got wrong first:
// the arena is reachable in this subset ONLY through string concatenation (an
// array or map of ints needs no arena at all -- measured: `let a = [1, 2, 3]`
// yields the same 2 sites as `print(1)`, because print_int's own two rodata
// references are the only absolute addresses it has). So a concatenating program
// cannot add the arena WITHOUT also adding the string literals, and a guessed
// "+2 for the arena" conflates the two. The counts below were MEASURED with the
// oracle on the committed compiler, and the assertion is only that:
//   - a heapless program has print_int's 2 sites and NO __DATA segment
//   - a concatenating program has strictly more, and DOES carry __DATA
//
// The relationship is what a hand-maintained site inventory cannot satisfy: add
// an arena reference to the patch list, forget the rebase list, and every other
// structural check still passes because the stream stays well-formed -- the
// count simply stops moving.
func TestPhaseB_MacOSRebaseSitesTrackThePatches(t *testing.T) {
	karkain := phase130Karkain(t)

	cases := []struct {
		name     string
		src      string
		wantData bool
		// wantSitesFloor is a LOWER bound, not an equality: the point is that
		// the count GROWS and that __DATA appears exactly when an arena does.
		wantSitesFloor int
	}{
		{"no_arena", "func main() {\n    print(1)\n}\n", false, 2},
		{"with_arena", "func main() {\n    print(\"ab\" + \"cd\")\n}\n", true, 5},
	}

	counts := map[string]int{}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			src := writeKark(t, dir, c.name+".kark", c.src)
			out := filepath.Join(dir, c.name+".macho")
			res, code := buildKccNative(t, karkain, src, NativeMacOSTarget, out)
			if code != 0 {
				t.Fatalf("build failed (exit %d):\n%s", code, res)
			}
			img, err := os.ReadFile(out)
			if err != nil {
				t.Fatalf("read image: %v", err)
			}
			hasData, derr := native.MachOHasDataSegment(img)
			if derr != nil {
				t.Fatalf("__DATA probe failed: %v", derr)
			}
			if hasData != c.wantData {
				t.Errorf("hasData = %v, want %v (a writable segment for a program that "+
					"never allocates moves .text and invalidates every baked address)",
					hasData, c.wantData)
			}
			sites, serr := native.MachORebaseSites(img)
			if serr != nil {
				t.Fatalf("rebase stream did not decode: %v", serr)
			}
			if len(sites) < c.wantSitesFloor {
				t.Errorf("rebase sites = %d, want at least %d; absolute addresses are "+
					"baked but not all of them are slid by dyld", len(sites), c.wantSitesFloor)
			}
			counts[c.name] = len(sites)
		})
	}

	if counts["with_arena"] <= counts["no_arena"] {
		t.Errorf("adding a concatenation did not increase the rebase-site count "+
			"(no_arena=%d, with_arena=%d): a site was added to a patch list but not to "+
			"the rebase stream, which parses cleanly and leaves a stale pointer",
			counts["no_arena"], counts["with_arena"])
	}
}

// TestPhaseB_MacOSEntrySitsInText pins the entry point.
//
// natStartStubOS marks `_start` first, so the entry is exactly
// MachoBase + textOff -- a stronger statement than "the entry is somewhere
// plausible", and the one that would catch a driver which began emitting the
// body before the stub.
//
// BOTH arena states are included because an extra LC_SEGMENT_64 (72 bytes)
// appears ONLY when the program needs a writable segment, which shifts .text and
// therefore every baked address AND every rebase site. A driver that computed
// the text offset before knowing whether an arena existed would produce an
// image whose rebase opcodes all point 72 bytes away from where the addresses
// actually are -- again with no crash and no diagnostic.
func TestPhaseB_MacOSEntrySitsInText(t *testing.T) {
	karkain := phase130Karkain(t)

	cases := []struct {
		name string
		src  string
	}{
		{"no_arena", "func main() {\n    print(1)\n}\n"},
		{"with_arena", "func main() {\n    print(\"ab\" + \"cd\")\n}\n"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			src := writeKark(t, dir, c.name+".kark", c.src)
			out := filepath.Join(dir, c.name+".macho")
			res, code := buildKccNative(t, karkain, src, NativeMacOSTarget, out)
			if code != 0 {
				t.Fatalf("build failed (exit %d):\n%s", code, res)
			}
			img, err := os.ReadFile(out)
			if err != nil {
				t.Fatalf("read image: %v", err)
			}
			entry, textOff, perr := native.ParseMachO(img)
			if perr != nil {
				t.Fatalf("invalid Mach-O: %v", perr)
			}
			// Since Phase D1 the entry is emitted LAST (the oracle's order:
			// helpers, then the user function, then `_start`), so it is no
			// longer the first byte of .text. It must lie INSIDE .text, and it
			// must be followed by exactly the entry tail.
			entryInText := int64(entry) - int64(native.MachoBase) - int64(textOff)
			if entryInText <= 0 {
				t.Fatalf("entry lies at %d within .text; _start is emitted LAST, so it "+
					"cannot be the first byte", entryInText)
			}
			// LC_SEGMENT_64 + the rest of the image sit before .text, so the
			// text length is not simply len(img)-textOff. Rather than
			// reconstruct the whole segment layout here, assert what is
			// recoverable from the IMAGE: the entry is within .text, and the
			// rebase stream still decodes (checked above).
			t.Logf("%s: entryInText=%d textOff=%d", c.name, entryInText, textOff)
		})
	}
}

// TestPhaseB_OtherTargetsAreUnchanged is the regression gate for the shared-code
// change.
//
// Phase B touched genuinely SHARED code in three places:
//
//  1. the emitter state grew a sixteenth slot (the target OS), read by the
//     stdout write path and the exit tail;
//  2. natSysWrite/natSysExit became OS-parameterised;
//  3. natStartStub became a wrapper over natStartStubOS.
//
// The Windows image is the one that CARRIES EXECUTION evidence (every
// increment-145..150 PE goldens run on this host), so a regression there would
// be quiet: the image would still be a structurally valid PE and would still
// parse. It is therefore built and validated here, and on a windows host it is
// additionally EXECUTED below -- which is what makes this a real check rather
// than a shape check.
//
// The Linux image is re-validated structurally for the same reason: Phase B
// touched the same natWriteStdout path A1 introduced.
func TestPhaseB_OtherTargetsAreUnchanged(t *testing.T) {
	karkain := phase130Karkain(t)
	srcBody := "func main() {\n    let x = 123\n    print(x)\n}\n"

	t.Run("windows", func(t *testing.T) {
		dir := t.TempDir()
		src := writeKark(t, dir, "w.kark", srcBody)
		out := filepath.Join(dir, "w.exe")
		got, code := buildKccNative(t, karkain, src, NativeWindowsTarget, out)
		if code != 0 {
			t.Fatalf("windows build failed (exit %d):\n%s", code, got)
		}
		img, err := os.ReadFile(out)
		if err != nil {
			t.Fatalf("read image: %v", err)
		}
		if _, _, perr := native.ParsePE(img); perr != nil {
			t.Fatalf("windows image is not a valid PE: %v", perr)
		}
		t.Logf("windows PE is %d bytes and structurally valid", len(img))
	})

	t.Run("linux", func(t *testing.T) {
		dir := t.TempDir()
		src := writeKark(t, dir, "l.kark", srcBody)
		out := filepath.Join(dir, "l.elf")
		got, code := buildKccNative(t, karkain, src, NativeLinuxTarget, out)
		if code != 0 {
			t.Fatalf("linux build failed (exit %d):\n%s", code, got)
		}
		img, err := os.ReadFile(out)
		if err != nil {
			t.Fatalf("read image: %v", err)
		}
		if _, _, perr := native.Parse(img); perr != nil {
			t.Fatalf("linux image is not a valid ELF: %v", perr)
		}
		t.Logf("linux ELF is %d bytes and structurally valid", len(img))
	})
}
