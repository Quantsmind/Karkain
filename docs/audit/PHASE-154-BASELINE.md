# Phase 154 BASELINE — I-6 Self-bootstrap (no-Go toolchain, C-free default path)

**Increment:** 154 (version **1.2.0**, "Sovereignty I — Off C, Off Go")
**Independence item:** **I-6** — toolchain independence (no Go anywhere on the build path; default user path C-free)
**Status at authoring:** baseline only. **No implementation.** Every claim below is in-tree evidence with a file/line, not a roadmap restatement.

Order is dependency-forced: **150 → 151 → 152 → 154** (`KARKAIN-VERSION-PLAN.md:157`). Increment 154 is the next queued increment after 152. `150` = native value model + native OS targets; `151` = kcc native parity (including 151D — kcc owns `native-x86_64-windows` and executes its own images); `152` = native stdlib + no-C closure (I-5); **154 = I-6**. `165` is free-floating CI/packaging work (Docker image), not a compiler dependency.

---

## 1. Objective

Shrink the compile-time surface of the Karkain toolchain: the default user path must not need the Go toolchain or a C compiler, and the self-hosted kcc must be able to build itself end-to-end. Concretely:

1. **Retire the `KARKAIN_ENGINE=go` stage-1 pin** in `pkg/bootstrap` — the `forceGoEngine` wrapper that forces the Go engine onto every compiler child the pipeline launches.
2. **Default engine of compiler children = the self-hosted kcc** (the default since Phase 97). The self-hosted compiler must compile and link itself: `kcc builds kcc`.
3. **Archive the Go tree** (`cmd/karkain`, `pkg/cli` Go drivers) as `reference/` at the repository root — off the user path, kept **only** as a differential oracle. It is never deleted.
4. **Gate:** with the Go toolchain absent from the build host, the self-bootstrap completes all stages yet and `stage2 == stage3` (SHA) bit-identical, with every compiler child running the default (kcc) engine.
5. **Full QA battery green in the no-Go state**: units, Phase-114 corpus on both engines, conformance, probes, every gate file, Sphinx `-W` + linkcheck, `verify-examples`, `install`/`verify-install`, `rc-journey`.

---

## 2. Current measured state (evidence from in-tree files, not assumption)

| # | Capability | State | Evidence |
|---|---|---|---|
| 1 | Stage-1 seed (`go build -o bin/karkain-stage1 ./cmd/karkain`) | present, required once per fresh tree | `pkg/bootstrap/bootstrap.go:202-206` |
| 2 | `KARKAIN_ENGINE=go` forced onto **every** compiler child across all stages | `forceGoEngine` in `runCmdOutputLabel` | `pkg/bootstrap/bootstrap.go:125, 149-151` |
| 3 | Stages 2/3 Go-free closure | proven with `go` shadowed (**Phase 143**, quarantined) — but the go pin still forces the Go engine on the compiler children | `pkg/bootstrap/phase143_seed_test.go:128-207` |
| 4 | kcc default engine | self-hosted kcc (Phase 97 onward) | CLI / `cmd/karkain/main.go` dispatch |
| 5 | Go tree location | `cmd/karkain`, `pkg/cli` source trees, tracked in the repo | repository root |
| 6 | Go-tree archival | none yet | `reference/` added by 154 |
| 7 | kcc-native targets | Go engine: `native-x86_64-windows` (151D, executed); none for a real user program on ELF/Mach-O (151A) | `pkg/cli/kcc_native.go` |

**Consequence worth stating plainly:** with the pin retired, the default engine of compiler children is kcc. The self-hosted kcc is therefore the only engine that must be able to emit and link its own compiler, and the Go tree is only a differential oracle. No language or target surface is removed by this increment.
## 3. Exit criteria (verbatim from `KARKAIN-VERSION-PLAN.md:212`, with the gate that proves each)

| # | Exit criterion | Proving gate |
|---|---|---|
| 1 | `kcc builds kcc`; **`stage2 == stage3` (SHA) with the Go toolchain absent from the build host** | `TestPhase154_SeedClosureNoGo` (pkg/bootstrap) |
| 2 | `KARKAIN_ENGINE=go` **stage-1 pin retired** (no compiler child forced the Go engine) | `TestPhase154_NoGoEnginePin` (pkg/bootstrap) |
## 5. Non-functional requirements

| ID | Requirement | Measured now | Target |
|---|---|---|---|
| NFR-15 | Toolchain independence (no Go; no C on the default path) | stages 2–3 go-free proven (143 gate); C still required for native links | no Go anywhere; no C compiler on the **default** path | 152, 154 |
| NFR-2 | Reproducible builds | stage-1 byte-reproducible (SHA `0c01de43…`) across CWDs; tamper divergence detected | same for the no-Go bootstrap + every packaged artifact | 152, 165, 166–168 |
| NFR-1 | Engine determinism (byte-identical Go↔kcc) | 61 corpus goldens + 64 conformance tests byte-identical | zero divergence on every Stable row, incl. native | 151 |
| NFR-3 | Memory footprint (4 GB host) | kcc check 544.5 MB · `go build` 229.9 MB · gcc big-TU 484 MB · guard threshold 1536 MiB | full no-Go self-build ≤ 2.5 GB | 162 |

## 6. Scope and boundaries

## 7. Implementation notes (filled during the increment, before any gate is written)

1. `pkg/bootstrap/bootstrap.go`:
   - Introduce a package-level pin knob (`goEnginePin`, default `true`) and a `RetireGoEnginePin()`/`GoEnginePinRetired()` pair; **`runCmdOutputLabel` applies `forceGoEngine` only while the pin is active**. Verify by mutation: re-enabling the pin unconditionally fails `TestPhase154_NoGoEnginePin`.
   - `RunStage1` keeps the `go build` seed step (the seed build is the 143 ceremony, not 154's work), but stages 2 and 3 must launch the compiler children with the **default** engine — i.e., with the pin retired and `KARKAIN_ENGINE` unset, so the self-hosted kcc drives the transpile.
2. `pkg/cli/kcc_native.go` / dispatch: assert, at CLI/CI level, that the default engine for a user program is the self-hosted kcc (the 150–151 paths are the proof of concept).
3. Archive the Go tree into `reference/` (root of the repository), `.gitignore`-covered, with a `reference/README.md` naming it the differential oracle (Go CLI + Go-engine C) — a differential, never a replacement.
4. New gate `pkg/bootstrap/phase154_selfbootstrap_test.go` (see §4) + CI step under the `Test` job.
5. Regression audit: full `go test ./...`, `go vet`, `pkg/native`, `pkg/parser`, `pkg/lexer`, `pkg/sema`, `pkg/codegen`, Phase 114 corpus both legs, Phase 122 KIRContinuity, Phase 123/124/135 gates, Sphinx `-W`, `verify-examples`; version identity (VERSION, both Go banner sites, three kcc banners, generated headers, LSP, every pinning test).
6. Audit note written; commit `develop` → merge `main` → push; per the branch rule.

**Amendment discipline:** a version is not closed as "green except X". Any new idea in this increment goes to the next version or an explicit carry-over list.

**In scope for 154:** pin retirement; default-engine wiring; `reference/` archival; the `TestPhase154_*` gate; the CI step; the phase-154 audit note; a no-Go bootstrap regression suite in `pkg/bootstrap`; the phase-154 version-identity and regression audit note.

**Out of scope for 154 (carried or owned elsewhere):** deleting the Go tree (keeps `reference/`); the C path (`--target c23`) — always stays; the standalone whole-toolchain bootstrap without Go (154 proves **stages 2–3** go-free, not a zero-Go build of the whole repository); 152's no-C stdlib closure (done by 152-A/B0/B1/B2); 165 (Docker image); the kcc ELF/Mach-O user-program emitters (151A).

**Stated, not papered over:** without a C compiler on the default path the native target is limited to the kcc Windows emitter (151D). The Linux/macOS native paths still need the value-model completion (151A) before their default-path proof can be claimed. The Go tree is the differential oracle for both engines for the lifetime of the repo.

| 3 | Go tree **archived as `reference/`** differential oracle, off-path, never deleted | `TestPhase154_ReferenceArchive` (pkg/bootstrap) |
| 4 | **Full QA battery green in the no-Go state** (units, Phase-114 both engines, conformance, probes, every gate, Sphinx `-W` + linkcheck, verify-examples, install/verify-install, rc-journey) | version-close ceremony (owner-verified; the gate step below runs on capable hosts) |
| 5 | `--target c23` and all other targets unchanged; the C path stays a first-class option | existing `--target c23` gates; 154 gate values asserted per §4 |

## 4. Gates (each is a test file **plus** its CI entry — AGENTS.md "gate is a test file plus a CI entry")

- `pkg/bootstrap/phase154_selfbootstrap_test.go`
  - `TestPhase154_PinRetired` — asserts the go pin is retired (`goEnginePin == false`), that a compiler child's environment carries no `KARKAIN_ENGINE=go`, and that a transplanted stage-2 build therefore runs with the default kcc engine.
  - `TestPhase154_NoGoEnginePin` — asserts the pin retirement is what makes the closure go-free: with `go` unresolvable (shadowed) and the pin retired, stages 2 and 3 run and `stage2 == stage3` (SHA) bit-identical.
  - `TestPhase154_ReferenceArchive` — asserts `reference/` exists, is gitignored, documents its role, and that the archived tree is the only Go tree on the user path.
  - `TestPhase154_NoGoQACheck` — bounded smoke of the no-Go state (unit subsets + a native link round trip) so CI reports the 154 gate itself, not a skip.
- CI step (named explicitly, not incidentally picked up): `.github/workflows/ci.yml` — "Run Increment 154 self-bootstrap gate" → `go test ./pkg/bootstrap/ -run 'TestPhase154_' -count=1 -timeout 60m`.
- **No-Go full QA battery** (criterion 4) is verified by the owner at the version-close ceremony — the same arrangement as the 152-A no-C closure gate on the Linux host, and the B2 seed-closure job on capable runners.

