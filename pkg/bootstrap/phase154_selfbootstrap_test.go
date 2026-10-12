// Package bootstrap TestPhase154 ---- Increment 154 (v1.2.0 "Sovereignty I"):
// retires the KARKAIN_ENGINE=go stage-1 pin so compiler children run the
// self-hosted kcc by default ("kcc builds kcc").
package bootstrap

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// findProjectRoot exists in this package (see bootstrap.go); declared here
// for the white-box gate.

// skipIfInsufficientMemory skips the test honestly when the bootstrap memory
// guard (K127) fires; a skipped gate is never a failure.
func skipIfInsufficientMemory(t *testing.T, stage int) {
	if err := CheckBootstrapMemory(stage); err != nil {
		if errors.Is(err, errInsufficientMemory) {
			t.Skipf("insufficient free RAM for bootstrap stage %d: %v", stage, err)
		}
		t.Fatalf("memory probe failed: %v", err)
	}
}

// TestPhase154_ControlPinActive is the control for the retirement: with the
// go pin STILL active, the pipeline must still build all three stages and
// produce stage2 == stage3. It proves the retirement is a real opt-in
// behavior change (the engine of compiler children), not a no-op.
func TestPhase154_ControlPinActive(t *testing.T) {
	projectRoot := findProjectRoot(t)
	cleanupBinaries(t, projectRoot)
	defer cleanupBinaries(t, projectRoot)

	s1, err := RunStage1(projectRoot)
	if err != nil {
		t.Fatalf("control stage 1 seed build failed: %v", err)
	}
	skipIfInsufficientMemory(t, 2)
	s2, err := RunStage2(projectRoot, s1.Binary)
	if err != nil {
		t.Fatalf("control stage 2 failed: %v", err)
	}
	skipIfInsufficientMemory(t, 3)
	s3, err := RunStage3(projectRoot, s2.Binary)
	if err != nil {
		t.Fatalf("control stage 3 failed: %v", err)
	}
	identical, err := VerifyIdentity(s2, s3)
	if err != nil {
		t.Fatalf("control VerifyIdentity error: %v", err)
	}
	if !identical {
		t.Fatalf("control stage2==stage3 FAILED: %x != %x", s2.SHA256, s3.SHA256)
	}
	// The control child env must still record the Go pin, demonstrating the
	// retirement is observable and the pipeline was actually wired through
	// the old path before 154.
	for _, kv := range CompilerChildEnv() {
		if kv == "KARKAIN_ENGINE=go" {
			return
		}
	}
	t.Error("control: compiler child env must carry KARKAIN_ENGINE=go while the pin is active")
}

// TestPhase154_PinRetired is the 154 main assertion: with the pin retired,
// the seed stage is still built by the Go toolchain (the 143 ceremony), but
// every compiler child from stage 2 onward runs the self-hosted kcc engine
// with no KARKAIN_ENGINE=go in its environment.
func TestPhase154_PinRetired(t *testing.T) {
	projectRoot := findProjectRoot(t)
	cleanupBinaries(t, projectRoot)
	defer cleanupBinaries(t, projectRoot)

	s1, err := RunStage1(projectRoot)
	if err != nil {
		t.Fatalf("stage 1 seed build failed: %v", err)
	}
	t.Cleanup(func() { cleanupBinaries(t, projectRoot) })

	skipIfInsufficientMemory(t, 2)
	SetGoEnginePin(true) // ensure the control state is clean
	RetireGoEnginePin()
	compilerChildEnv = nil // no compiler child may have run since retirement
	defer SetGoEnginePin(true)

	s2, err := RunStage2(projectRoot, s1.Binary)
	if err != nil {
		t.Fatalf("stage 2 with pin retired failed: %v", err)
	}
	s3, err := RunStage3(projectRoot, s2.Binary)
	if err != nil {
		t.Fatalf("stage 3 with pin retired failed: %v", err)
	}
	identical, err := VerifyIdentity(s2, s3)
	if err != nil {
		t.Fatalf("VerifyIdentity error: %v", err)
	}
	if !identical {
		t.Fatalf("stage2==stage3 FAILED under the retired pin: %x != %x", s2.SHA256, s3.SHA256)
	}
}

// TestPhase154_NoGoEnginePin asserts the direct contract of the increment:
// once retired, a compiler child's environment never carries
// KARKAIN_ENGINE=go, so the self-hosted kcc is the default engine of the
// pipeline.
func TestPhase154_NoGoEnginePin(t *testing.T) {
	projectRoot := findProjectRoot(t)
	cleanupBinaries(t, projectRoot)
	defer cleanupBinaries(t, projectRoot)

	s1, err := RunStage1(projectRoot)
	if err != nil {
		t.Fatalf("stage 1 seed build failed: %v", err)
	}
	defer cleanupBinaries(t, projectRoot)

	skipIfInsufficientMemory(t, 2)
	RetireGoEnginePin()
	compilerChildEnv = nil // no compiler child may have run since retirement
	defer SetGoEnginePin(true)

	// A transplanted stage-2 build must already run without the pin.
	s2, err := RunStage2(projectRoot, s1.Binary)
	if err != nil {
		t.Fatalf("stage 2 without the go pin failed: %v", err)
	}
	s3, err := RunStage3(projectRoot, s2.Binary)
	if err != nil {
		t.Fatalf("stage 3 without the go pin failed: %v", err)
	}
	identical, err := VerifyIdentity(s2, s3)
	if err != nil {
		t.Fatalf("VerifyIdentity error: %v", err)
	}
	if !identical {
		t.Fatalf("stage2==stage3 FAILED: %x != %x", s2.SHA256, s3.SHA256)
	}
	for _, kv := range CompilerChildEnv() {
		if kv == "KARKAIN_ENGINE=go" {
			t.Fatalf("compiler child still carries the Go pin: %q", kv)
		}
	}
}

// TestPhase154_ReferenceArchive asserts the Go tree is archived, not
// deleted: reference/ holds the Go CLI + Go-engine C as a differential
// oracle, off the user path.
func TestPhase154_ReferenceArchive(t *testing.T) {
	projectRoot := findProjectRoot(t)

	if _, err := os.Stat(filepath.Join(projectRoot, "reference")); err != nil {
		t.Fatalf("reference/ missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "reference", "README.md")); err != nil {
		t.Fatalf("reference/README.md missing: %v", err)
	}
	gitignore, err := os.ReadFile(filepath.Join(projectRoot, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !bytes.Contains(gitignore, []byte("reference/")) {
		t.Error(".gitignore does not cover reference/")
	}
}
