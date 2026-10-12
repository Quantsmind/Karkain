# Karkain Language Completeness — Architectural Target

> **Status: architecture target specification.** This document defines *what
> capabilities Karkain must possess* before it can reasonably be described as a
> complete general-purpose programming language and ecosystem.
>
> **This is NOT a release roadmap and NOT a plan.** It does not schedule work,
> assign phases, or change any existing increment. The execution schedule remains
> `docs/audit/KARKAIN-VERSION-PLAN.md`. Where this document and the roadmap
> appear to disagree, the roadmap governs *what happens next*; this document
> governs *whether the result is a complete language*.
>
> **No capability is claimed here that is not evidenced elsewhere.** Every status
> in this document is transcribed from repository evidence, never inferred from
> ambition.

---

## 1. Purpose and the three-layer distinction

This document answers one question: **what must exist before "complete general-purpose
programming language and ecosystem" is a defensible description of Karkain?**

Three document layers are preserved and must not be merged:

| Layer | Document | Responsibility |
|---|---|---|
| **Language Guide** | `docs/source/language/` | Teaches Karkain. Progression-oriented. |
| **Language Specification** | `SPEC.md` (repo root) | Normatively defines Karkain behaviour. |
| **Architecture / Audit** | `docs/architecture/`, `docs/audit/` | Explains implementation, evidence, and the target. |

**Compiler / runtime / development docs** (`docs/audit/`, `docs/source/compiler/`)
explain *implementation and extension*. They are implementation documentation and are
not a substitute for either the guide or the specification.

### 1.1 Relationship chain

```text
Completeness requirement          (this document, L0–L10)
        ↓
Existing GAP                     (authoritative GAP inventory, §4)
        ↓
Dependency                       (what must exist first)
        ↓
Implementation slice             (a bounded unit of work)
        ↓
Existing Phase / Increment       (roadmap ownership, where documented)
        ↓
Release version                  (KARKAIN-VERSION-PLAN.md)
        ↓
Closure evidence                 (gate + audit note + CI)
```

A link is only as strong as its weakest end. An increment existing does **not**
close a requirement; only verified closure evidence does.

---

## 2. The completeness model (L0–L10)

Eleven levels. Each states the requirement, not the current state — current state is
in §4.

### L0 — Compiler foundation

A compiler that can reliably turn source into an executable.

- Lexer, parser, AST
- Semantic analysis / name resolution
- IR and optimization
- Code generation
- Executable and image generation
- Bootstrap foundations
- Two engines (Go reference front end; self-hosted `kcc`)

**Complete when:** a program compiles on both engines with no unexplained engine
divergence, and the compiler can compile itself to a bit-reproducible image.

### L1 — Computational language

The minimum needed to express any computation.

- Variables, bindings, assignment
- Arithmetic, comparison, logical operators
- Conditional and iterative control flow
- Functions, parameters, returns
- Primitive values and literals
- Basic execution semantics (evaluation order, scoping)

**Complete when:** the language can express arbitrary computation with defined
semantics for every operator, and no operator silently mis-evaluates.

### L2 — Type and data system

A coherent, usable data model:

| Element | Requirement |
|---|---|
| Integers / numeric types | Signed, unsigned, fixed-width, and (where claimed) arbitrary precision — all on **both** engines |
| Floating point | Present, with defined formatting and conversion behaviour |
| Booleans | Real boolean type with defined truthiness rules |
| Strings | UTF-8 byte semantics defined, including case handling verdict |
| Arrays | Element-typed, with defined bounds semantics |
| Maps | Defined key domain and key-equality semantics |
| Records / structs | Nominal identity and checked field types — **not** runtime maps |
| Enums | Tagged unions with declared payloads |
| Function values | First-class, callable through a value |
| Generic abstractions | Parameterised functions and types |
| Type inference | For unannotated bindings |
| Conversions | Explicit and defined |
| Equality | Defined for every comparable type |
| Constraints / traits | An abstraction mechanism to declare and check against |

**Complete when:** a type lattice exists as a first-class object in the checker,
structs have real layout and checked fields, and generic substitution is not purely
textual.

### L3 — Memory and safety model

The required model for allocation and access safety:

| Concern | Requirement |
|---|---|
| Allocation | A defined allocation surface |
| Ownership | Who is responsible for a value |
| Moves | Move vs copy distinction |
| References | A reference type distinct from the borrow operator |
| Borrowing | Shared / mutable borrowing rules |
| Lifetimes | Statically tracked where claimed |
| Aliasing | Rules preventing unsound aliasing |
| Deallocation / reclamation | A defined reclamation strategy |
| Memory safety | A stated guarantee — static, dynamic, or neither, **stated explicitly** |

**Complete when:** the language states one memory-safety guarantee and the
implementation enforces it. *"We check at runtime"* is a legitimate answer; leaving
it undefined is not.

### L4 — Runtime and failure semantics

Consistent, defined behaviour for every failure class:

| Class | Requirement |
|---|---|
| Bounds failures | Defined diagnostic, stream and exit behaviour |
| Invalid conversions | Defined |
| Division / arithmetic failures | Defined, including overflow verdict |
| Invalid encoding | Defined (hex, base64, UTF-8) |
| Allocation failure | Defined |
| Filesystem / I/O failures | Defined |
| Network failures | Defined |
| Process / system failures | Defined |
| Assertions | Defined failure behaviour |
| Recoverable errors | A real `Result`/`Option` surface that works |
| Fatal runtime errors | One diagnostic format with source location |
| Diagnostics | Consistent wording across engines |

**Complete when:** every row has a defined semantic *on both engines*, and
divergence is a reported defect rather than a documented surprise.

### L5 — Standard library

Practical domains a general-purpose language needs:

strings · collections · bytes · mathematics · encoding/decoding · filesystem ·
I/O · system/process · time · cryptography · networking · database access

**Complete when:** each domain has an importable, tested, both-engine module.
A `.kark` wrapper over a C builtin is a legitimate architecture, but a module that
only exists on one engine does not satisfy the domain.

### L6 — Concurrency

| Concern | Requirement |
|---|---|
| Tasks / threads | A spawnable unit |
| Channels | Defined send/receive semantics |
| Synchronization | Locks and/or atomics **exposed to the language** |
| Atomics | Language-level atomic primitives |
| Memory ordering | Defined and enforced |
| Safe sharing | `Send`/`Sync`-equivalent typing |
| Concurrency failure semantics | Defined (deadlock, closed channel, data race) |

**Depends on L3.** Without an ownership or borrow model there is no sound way to
share mutable state, so safe-sharing typing is unreachable before the memory model.

### L7 — Developer ecosystem

build · run · test · format · lint · diagnostics · debug · documentation ·
source mapping · profiling / tracing · IDE / LSP integration

**Complete when:** a developer can build, test, debug and profile a Karkain program
without leaving Karkain-adjacent tooling.

### L8 — Application ecosystem

Capability needed to build real applications:

CLI applications · web applications · HTTP services · serialization ·
persistence · databases · messaging · security · service/application frameworks ·
enterprise application support

**Long-term direction:** the ecosystem should increasingly be implemented *in
Karkain itself* rather than permanently depending on external-language frameworks.

### L9 — Distribution and platform ecosystem

package management · dependency resolution · lockfiles · build configuration ·
cross-compilation · platform support · release artifacts · installers · WASM ·
containers · native distribution

Roadmap ownership for platform/package milestones comes from
`KARKAIN-VERSION-PLAN.md` and `KARKAIN-INDEPENDENCE-ROADMAP.md`; this document does
not assign any.

### L10 — Self-sustaining / self-hosting direction

The long-term state where Karkain implements its own ecosystem and tooling in
Karkain. Four distinct terms, which are frequently conflated:

| Term | Meaning |
|---|---|
| **Compiler implementation language** | The language the compiler is written in |
| **User-facing ecosystem** | What Karkain programs can consume |
| **Self-hosted compiler** | The compiler can compile *itself* |
| **Bootstrap chain** | The exact stage sequence that produces a compiler binary |

**Karkain is not self-hosting in the L10 sense today.** `kcc` is written in Karkain
and produces C23, and a three-stage bootstrap reaches `stage2 == stage3`
bit-reproducibility — but bootstrap still links through a C compiler, and the Go
front end remains the reference oracle. The bootstrap chain is therefore
*C-owned*, not self-sustaining.

---

## 3. Status model

Two **independent** dimensions. Never merge them.

### 3.1 Feature status

| Status | Meaning |
|---|---|
| **NOT STARTED** | No design or implementation exists |
| **DESIGNED** | A decision or spec exists; no implementation |
| **IMPLEMENTING** | Partial implementation exists |
| **IMPLEMENTED** | Works, with tests |
| **VERIFIED** | Implemented and independently gated with evidence |
| **COMPLETE** | VERIFIED **plus** usable semantics, documentation, and closure evidence appropriate to the feature |

**`COMPLETE` is the only status that satisfies a completeness requirement.**
A feature that is `IMPLEMENTED` on one engine and missing on the other is not
complete. A feature with no documented semantics is not complete, however well it
works.

### 3.2 Release status

| Status | Meaning |
|---|---|
| **PLANNED** | Named in a roadmap; not scheduled |
| **BASELINED** | Has a baseline note; scope frozen |
| **IN PROGRESS** | Implementation started |
| **QA** | Implementation complete; undergoing release verification |
| **RELEASED** | Published under a version tag |

**No percentage completion is used anywhere in this document.** Karkain does not
express progress as a fraction; it expresses it as gate results and closure evidence.
A version is never closed as "green except X".

---

## 4. GAP mapping

### 4.1 Authoritative sources and ID namespacing

Two inventories exist and they **collide on ID prefixes**. They are namespaced here:

| Prefix | Source | Character |
|---|---|---|
| **`MM-`** | `docs/audit/KARKAIN-1.1.0-MATURITY-MATRIX.md` (dated 2026-10-01) | **Authoritative.** Evidence-based; re-validates its citations; declares `[1.1.0]` vs `[HEAD]` baselines and four evidence conflicts up front. 136 rows in sections 3.1-3.15 |
| **`GA-`** | `docs/audit/KARKAIN_GAP_ANALYSIS.md` | **Secondary and explicitly stale.** Its own header says findings predate substantial work, that some may be stale/fixed/intentional, and that it "does not authorize work". 37 rows |

`MM-A1` and `GA-A1` are **different findings**. The collision is real and is reported
as an ambiguity in section 9.

**Nothing is invented here.** Where an inventory records no phase or version
ownership, this document writes `UNOWNED / NOT BASELINED`. Where an increment only
partially addresses a gap it writes `PARTIAL`. "An increment exists" is never
recorded as "complete".

**Priority candidates P1-P8** in the matrix section 12 are *ordering*, not
ownership. P1 and P7 are recorded **closed**; P2-P6 and P8 remain open and unowned.

### 4.2 L0 — Compiler foundation

| GAP | Capability | Dependencies | Increment / Version | Status | Closure evidence |
|---|---|---|---|---|---|
| MM-H1 | Lexer | - | UNOWNED / NOT BASELINED | Mature (documented) | `pkg/lexer` suite |
| MM-H2 | Parser (recursive descent, recovery) | H1 | UNOWNED / NOT BASELINED | Mature (documented) | `pkg/parser`, Phase 105/123 |
| MM-H3 | AST (arena-allocated) | H1 | UNOWNED / NOT BASELINED | Mature (documented) | documented |
| MM-H4 | Semantic analysis | A13 | UNOWNED / NOT BASELINED | **PARTIAL** (resolution + annotation matching) | Phase 99/123 |
| MM-H5 | SSA IR | - | UNOWNED / NOT BASELINED | Partial use; closure-bearing functions take legacy path | `pkg/ir/ssa`, Phase 141 |
| MM-H6 | HIR | H5 | UNOWNED / NOT BASELINED | **STRUCTURAL ONLY - unwired** | unit tests only |
| MM-H7 | Optimizer | H5 | UNOWNED / NOT BASELINED | **PARTIAL** (2 of 6 passes on-path) | Phase 141 |
| MM-H8 | Code generation (C) | - | UNOWNED / NOT BASELINED | Mature (documented) | Phase 114 corpus, both engines |
| MM-H9 | Code generation (self-hosted) | H4, A5, A10 | **151** / v1.2.0 (IN PROGRESS) | Implemented, **divergent** on A5 and A10 | 61-file corpus byte-parity |
| MM-H10 | Code generation (C-free machine code) | H1-H3 | **150** COMPLETE / v1.2.0 | Implemented at HEAD; partial at v1.1.0 | Phase 145-150, 151B-151C3 |
| MM-H11 | Linker integration | G11 | **152** / v1.2.0 (IN PROGRESS) | **PARTIAL** (writers exist, no general linker) | Phase 111 |
| MM-H12 | Register allocation (native only) | H10 | **150C** COMPLETE / v1.2.0 | Implemented | Phase 150C |
| MM-H13 | Incremental compilation | H8 | UNOWNED / NOT BASELINED | Implemented (whole-assembly; native split refused with measured reason) | Phase 105/134 |
| MM-H14 | Bootstrap / self-hosting | H9, H10 | **154** / v1.2.0 (NOT BASELINED) | **PARTIAL** (`stage2==stage3` gated; Go oracle + C linker still required) | `TestBootstrap_BitwiseIdentity` |
| MM-H15 | Diagnostics | - | UNOWNED / NOT BASELINED | Mature (documented) | Phase 83/105/123 |
| MM-D1 | Modules / imports | - | UNOWNED / NOT BASELINED | Implemented (restored 2026-10-02; omitted from the first pass of this section) | Phase 122/144 |

### 4.3 L1 — Computational language

| GAP | Capability | Dependencies | Increment / Version | Status | Closure evidence |
|---|---|---|---|---|---|
| MM-A1 | 64-bit signed `int` | - | UNOWNED / NOT BASELINED | Implemented; width is a C-preamble property, not a type fact | Both-engine corpus |
| MM-A2 | `float64` | - | UNOWNED / NOT BASELINED | Implemented | Corpus, both engines |
| MM-A3 | `bool` | - | UNOWNED / NOT BASELINED | Implemented | Corpus |
| MM-A4 | `string` | - | **158** / v1.3.0 (PLANNED) | Implemented; **case handling ASCII-only** (NFR-16 debt) | Corpus + Phase 132 freeze |
| MM-A8 | Arrays | A1 | **158** / v1.3.0 (PLANNED) | Implemented; heterogeneous, no static element type | Corpus, both engines |
| MM-A14 | Name resolution / scope | - | UNOWNED / NOT BASELINED | Implemented | Corpus, both engines |
| MM-A16 | Compile-time diagnostics | - | UNOWNED / NOT BASELINED | Implemented (spans, JSON contract) | Phase 83/105 |
| MM-N3 | SIMD lane types | - | UNOWNED / NOT BASELINED | Implemented and gated | Phase 106 |
| MM-A15 | Module visibility | D1 | UNOWNED / NOT BASELINED | Implemented | Phase 103 |
| MM-D2 | Namespaces | D1 | UNOWNED / NOT BASELINED | Implemented; kcc namespaces are flattened, not hierarchical (restored 2026-10-02) | Phase 103 |
| MM-D3 | Multi-file programs | D1 | UNOWNED / NOT BASELINED | Implemented; kcc performs the sibling join itself (restored 2026-10-02) | Phase 122 |

### 4.4 L2 — Type and data system

| GAP | Capability | Dependencies | Increment / Version | Status | Closure evidence |
|---|---|---|---|---|---|
| MM-A5 | `bigint` / `bigfloat` | MM-P3 (value-model parity) | UNOWNED / NOT BASELINED | **PARTIAL - Go engine only.** Verified divergence: the default engine rejects them | divergence reproduced in the matrix |
| MM-A6 | Unsigned / 128-bit integers | A1 | UNOWNED / NOT BASELINED | **MISSING** (annotations parse and are inert) | verified inert |
| MM-A7 | Type annotations as semantics | A6 | UNOWNED / NOT BASELINED | **PARTIAL** (stored as strings, never consulted) | verified inert |
| MM-A9 | Records / structs | A13 (type object) | UNOWNED / NOT BASELINED | **PARTIAL - runtime maps.** No layout, no field checking, no nominal typing | `Point{zzz:1}` is not rejected |
| MM-A10 | Maps | A1 | UNOWNED / NOT BASELINED | Implemented (Go) / **PARTIAL** (self-hosted). Priority candidate **P1 - recorded CLOSED**: typed key equality via `values_equal`, no `strcmp` coercion | `pkg/cli/p1_map_key_test.go`; `KCC-MAP-KEY-EQUALITY-FIX.md` |
| MM-A11 | First-class `fn` values | A9 | UNOWNED / NOT BASELINED | **PARTIAL** (bare function references broken on both engines) | `examples/closures/`, Phase 133 |
| MM-A12 | Enums + `match` | A9 | UNOWNED / NOT BASELINED | **PARTIAL** (tag-only patterns; **no payload destructuring**; no exhaustiveness) | corpus `14_adt_match`, Phase 123 |
| MM-A13 | Type checking / inference | - | UNOWNED / NOT BASELINED | **PARTIAL** (no lattice, no unification, no inference for unannotated `let`) | `pkg/sema` suites |
| MM-C1 | Generic functions | A13 | UNOWNED / NOT BASELINED | **PARTIAL - textual substitution only** (whole-string, `*T`, `[]T`) | Phase 146 gates |
| MM-C2 | Generic types | C1, A9 | UNOWNED / NOT BASELINED | **PARTIAL** (same textual limit) | Phase 146b |
| MM-C3 | Type parameters | C1 | UNOWNED / NOT BASELINED | Implemented (v1); **no inference** | Phase 146 |
| MM-C4 | Constraints / traits | C1 | UNOWNED / NOT BASELINED | **STRUCTURAL ONLY** - nothing declares against them | Phase 146d |
| MM-C5 | Compile-time evaluation | - | UNOWNED / NOT BASELINED | **MISSING** (no `const`, no const-fn) | - |
| MM-C6 | Macros | C1 | UNOWNED / NOT BASELINED | **PARTIAL** (expansion-based, not hygienic) | Phase 82 |
| MM-C7 | Reflection | A13 | UNOWNED / NOT BASELINED | **MISSING** | - |

> **Inventory correction (Step 2).** `MM-A15` (module visibility) was listed in this
> section by the first pass. Step 2 maps it to **L1**, because a name-access rule is a
> scoping / execution-semantics question rather than a type-system one; it now appears
> in section 4.3. `MM-D6` was likewise first recorded only in section 4.11, which is
> where it correctly belongs. See Step 2 section 2.1.

### 4.5 L3 — Memory and safety model

| GAP | Capability | Dependencies | Increment / Version | Status | Closure evidence |
|---|---|---|---|---|---|
| MM-B1 | Ownership tracking | B2 | UNOWNED / NOT BASELINED | **PARTIAL - lexical, not semantic** (scope-stack identity; no heap borrow, no lifetimes) | `pkg/sema` suites |
| MM-B2 | Borrow checker | B1 | UNOWNED / NOT BASELINED | **PARTIAL** (no aliasing rules, no use-after-free detection, no `&mut`) | Phase 51 |
| MM-B3 | Escape analysis | - | UNOWNED / NOT BASELINED | **STRUCTURAL ONLY - dead** (computed and discarded; no codegen reads it) | `pkg/parser/escape_test.go` |
| MM-B4 | Heap allocation | - | UNOWNED / NOT BASELINED | Implemented (runtime) | Phase 101/102 |
| MM-B5 | Deallocation | B4 (and ownership) | UNOWNED / NOT BASELINED | **MISSING at language level** (no GC, no refcount, no `free`) | - |
| MM-B6 | Move semantics | B1 | UNOWNED / NOT BASELINED | **MISSING** | - |
| MM-B7 | References / `&` | B1 | UNOWNED / NOT BASELINED | **PARTIAL** (`&` is the Phase 41 borrow operator in prefix position; infix `&` is bitwise) | Phase 51 |
| MM-B8 | Aliasing rules | B1 | UNOWNED / NOT BASELINED | **MISSING** | - |
| MM-B9 | Safety guarantees | B2 | UNOWNED / NOT BASELINED | **PARTIAL** - safety is runtime-only, not proven | Phase 100 |
| MM-M4 | Memory safety | B5 | UNOWNED / NOT BASELINED | **MISSING** (generated C uses raw pointers; user allocations never reclaimed) | - |

### 4.6 L4 — Runtime and failure semantics

| GAP | Capability | Dependencies | Increment / Version | Status | Closure evidence |
|---|---|---|---|---|---|
| MM-E1 | `?` propagation | A13, E2 | **151** (LH-1) / v1.2.0 COMPLETE | Implemented on the C path, both engines; **refused on the C-free native target** (no Result to propagate to) | LH-1 gate, 11/11 |
| MM-E2 | Result / Option values | A13 | UNOWNED / NOT BASELINED | **STRUCTURAL ONLY** per the matrix (C enum tags exist). Superseded in practice: LH-1 added expression-level constructors, so the matrix row predates that | LH-1 gate; `docs/source/language/errors.rst` |
| MM-E3 | Exceptions / panic | - | UNOWNED / NOT BASELINED | **MISSING** (deliberately: `Result` is the chosen direction) | - |
| MM-E4 | Runtime error reporting | - | **152-B0** / v1.2.0 (IN PROGRESS) | Implemented on both engines; **native path newly added** | `pkg/native/phase152b0_runtime_error_test.go`; Phase 100/101 |
| MM-E5 | Compile-time error recovery | - | UNOWNED / NOT BASELINED | Implemented (project-wide multi-error preflight) | Phase 105 |
| MM-E6 | Compile-fail corpus | - | UNOWNED / NOT BASELINED | Implemented (`examples/type_errors/`, `examples/runtime_errors/`) | Phase 123 |
| MM-M1 | Bounds checking | - | UNOWNED / NOT BASELINED | Implemented (runtime), **not a static guarantee** | Phase 100 |
| MM-M2 | Division-by-zero / bad index | - | UNOWNED / NOT BASELINED | Implemented; now also on the native path | Phase 100; 152-B0 |
| MM-M3 | Integer overflow | - | UNOWNED / NOT BASELINED | **MISSING** (`<<` wraps; no checked `+ - *`) | - |
| MM-M5 | Null / bounds on native backend | E4 | **152-B0** / v1.2.0 (IN PROGRESS) | **PARTIAL** - array bounds now diagnosed; **string-slice semantics unresolved** (C23 clamps, native traps) | `PHASE-152-BASELINE.md` |
| MM-M7 | Input validation | - | UNOWNED / NOT BASELINED | **PARTIAL** (not systematic) | - |
| MM-M11 | Compiler crash handling | - | UNOWNED / NOT BASELINED | **PARTIAL** - fallback covers malformed IR, not wrong-code survival | `codegen.go:356-361` |
| MM-G1 | Layered runtime | - | UNOWNED / NOT BASELINED | Implemented - freestanding / core / concurrency (restored 2026-10-02) | Phase 101/102/107 |
| MM-G2 | Startup model | - | UNOWNED / NOT BASELINED | Implemented - generated `main`; native `_start` and PEB bootstrap (restored 2026-10-02) | Phase 149 |
| MM-G3 | Value representation | - | UNOWNED / NOT BASELINED | Implemented - tagged `Value` union, mirrored by kcc (restored 2026-10-02) | corpus |
| MM-G4 | C interoperability | - | UNOWNED / NOT BASELINED | **PARTIAL** - `import "C"` unchecked, unverified at link time (restored 2026-10-02) | Phase 89-103 |
| MM-G5 | System calls | - | UNOWNED / NOT BASELINED | **PARTIAL (runtime only)** - raw syscalls; **not exposed as a language surface** (restored 2026-10-02) | Phase 149 |

### 4.7 L5 — Standard library

Domains are satisfied only by an **importable, tested, both-engine** module. A
`.kark` wrapper over a C builtin is a legitimate architecture; a module that exists
on only one engine does not satisfy the domain.

| GAP | Capability | Dependencies | Increment / Version | Status | Closure evidence |
|---|---|---|---|---|---|
| MM-K-string | `std.string` | - | **158** / v1.3.0 (PLANNED) | Implemented, frozen; EXTEND backlog recorded | Phase 109/132 |
| MM-K-collections | `std.collections` | - | **158** / v1.3.0 (PLANNED) | Implemented, frozen; EXTEND backlog recorded | Phase 109 |
| MM-K-bytes | Byte-level surface | MM-A5 | UNOWNED / NOT BASELINED | **PARTIAL** - exists only as `*_bytes` C builtins behind `std.encoding` | Phase 109, NIST/RFC vectors |
| MM-K-math | `std.math` | - | UNOWNED / NOT BASELINED | **NOT IMPORTABLE** - source-present behind the Phase-109 boundary | Phase 87 probes |
| MM-K-encoding | `std.encoding` | - | **158** / v1.3.0 (PLANNED) | Implemented; **thin wrapper over C builtins**; JSON moves out to a new `std.json` | Phase 109 |
| MM-K-crypto | `std.crypto` | A6 (bitwise) | **158** / v1.3.0 (PLANNED) | Implemented; **thin wrapper over C builtins** | Phase 109, FIPS 180 |
| MM-K-io | `std.io` | G10 | **156** MERGE / v1.3.0 (PLANNED) | Implemented; merge verdict with `std.system` recorded | Phase 109 |
| MM-K-system | `std.system` | G10 | **156** / v1.3.0 (PLANNED) | **STRUCTURAL ONLY** - no OS API surface behind it | - |
| MM-K-time | Time | G10 | UNOWNED / NOT BASELINED | **MISSING** - no clock, duration or calendar surface recorded in either inventory | - |
| MM-K-fs | Filesystem | G6 | **156** / v1.3.0 (PLANNED) | Implemented on the C path; **absent on the C-free native target** | Phase 109; 152-B1 scope |
| MM-K-process | Processes / env / dynamic libs | G10 | UNOWNED / NOT BASELINED | **MISSING** | - |
| MM-K-net | `std.net` | G7 | **158** / **161** / v1.3.0 (PLANNED) | Implemented; thin wrapper; forces `-lws2_32` on Windows; **absent on native** | Phase 125A; 152-A capability boundary |
| MM-K-db | `std.db` | - | UNOWNED / NOT BASELINED | Implemented, **non-embedded** SQL-text engine; no real transactions | Phase 125A |
| MM-K-testing | `std.testing` | - | **163** / v1.4.0 (PLANNED) | Implemented; EXTEND backlog recorded | KTF-001 |
| MM-K-numerics | `std.numerics` | - | **159** / v1.3.0 (PLANNED) | Implemented; EXTEND backlog recorded | Phase 126 |
| MM-K-generics | `std.generics` | C1 | UNOWNED / NOT BASELINED | Implemented (exercises monomorphization) | Phase 146b |
| MM-K-core | `std.core` | - | UNOWNED / NOT BASELINED | **STRUCTURAL ONLY** - outside the 10 frozen modules | - |
| MM-K-async | `std.async` | F1 | UNOWNED / NOT BASELINED | **STRUCTURAL ONLY** - non-canonical syntax, no builtin backing | - |
| MM-K-gpu | `std.gpu` | N1 | UNOWNED / NOT BASELINED | Experimental | Phase 124 |
| MM-D4 | Library consumption | D1 | UNOWNED / NOT BASELINED | Implemented - `import std.x`, 15 modules (restored 2026-10-02) | Phase 109/131/132 |
| MM-N5 | Autodiff | - | **159** / v1.3.0 (PLANNED) | Implemented (reverse-mode over existing IR) | `pkg/sema/autodiff.go` |

### 4.8 L6 — Concurrency

**Depends on L3.** Without an ownership or borrow model there is no sound way to
share mutable state, so safe-sharing typing is unreachable before the memory model.

| GAP | Capability | Dependencies | Increment / Version | Status | Closure evidence |
|---|---|---|---|---|---|
| MM-F1 | `spawn` / `join` / `wait_all` | - | UNOWNED / NOT BASELINED | Implemented (both engines) | Phase 107 |
| MM-F2 | Channels | - | **161** / v1.3.0 (PLANNED) | Implemented (bounded/unbounded) | Phase 107 |
| MM-F3 | Actors | - | **161** / v1.3.0 (PLANNED) | Implemented | Phase 107 |
| MM-F4 | Scheduler | - | UNOWNED / NOT BASELINED | Implemented (runtime); work-stealing | Phase 107, >1M msg/s |
| MM-F5 | Atomics exposed to the language | F1 | **161** / v1.3.0 (PLANNED) | **MISSING** - `__atomic` helpers exist only inside the C runtime | - |
| MM-F6 | Data-race guarantees | B2, F5 | **161** / v1.3.0 (PLANNED) | **MISSING** - no `Send`/`Sync` typing; shared mutable state unconstrained | - |
| MM-F7 | Parallel execution | F4 | UNOWNED / NOT BASELINED | Partial (runtime only, not a language contract) | Phase 107 |

### 4.9 L7 — Developer ecosystem

| GAP | Capability | Dependencies | Increment / Version | Status | Closure evidence |
|---|---|---|---|---|---|
| MM-L1 | Formatter | - | UNOWNED / NOT BASELINED | Implemented (token-level, idempotent; no reflow) | Phase 82 |
| MM-L2 | LSP | - | UNOWNED / NOT BASELINED | Implemented (no rename/refactor; no incremental cache) | Phase 136, 30 tests |
| MM-L3 | Debugger | - | **153** / v1.4.0 (PLANNED) | **PARTIAL** - drives gdb in batch; **not a DAP server**; Go engine only | Phase 140 |
| MM-L4 | Test runner | - | UNOWNED / NOT BASELINED | Implemented | Phase 96, 61-file corpus |
| MM-L5 | Doc generation | - | UNOWNED / NOT BASELINED | Implemented; language reference is thin vs 140 audit reports | CI docs job |
| MM-L6 | IDE integration | - | UNOWNED / NOT BASELINED | Implemented (VS Code extension) | `TestVSCodeExtension` |
| MM-L7 | REPL | - | UNOWNED / NOT BASELINED | **MISSING** | - |
| MM-L8 | Linter | - | UNOWNED / NOT BASELINED | Implemented | Phase 86 |
| MM-L9 | Profiler | - | UNOWNED / NOT BASELINED | Implemented (counts only `malloc`/`free` text in generated code) | Phase 110 |
| MM-O1 | Language specification | - | UNOWNED / NOT BASELINED | **PARTIAL** - present and sectioned but behaves as a **feature ledger, not a specification** (no formal grammar); overstates the type system | `SPEC.md` |
| MM-O2 | Architecture documentation | - | UNOWNED / NOT BASELINED | Partial (140 audit reports; documents history, not a current reference) | `docs/audit/` |
| MM-O3 | Standard-library docs | - | UNOWNED / NOT BASELINED | Implemented (repaired after drift) | `docs/source/stdlib/` |
| MM-O4 | Examples | - | UNOWNED / NOT BASELINED | Implemented (220 `.kark`, 42 categories) | Phase 114 corpus is a 61-file subset |
| MM-O5 | Tutorials | - | UNOWNED / NOT BASELINED | **MISSING** - no tutorial path in `docs/source` | - |
| MM-O6 | API reference | - | UNOWNED / NOT BASELINED | **PARTIAL** - hand-maintained, not generated from source | `docs/source/reference/stable-api.rst` |
| MM-O7 | Release documentation | - | UNOWNED / NOT BASELINED | Implemented | `docs/release/`, `SECURITY.md` |
| MM-O8 | Conformance documentation | - | UNOWNED / NOT BASELINED | Implemented (12 files / 64 tests, both engines) | conformance suite |
| MM-O9 | Spec-vs-code conformance tooling | O1 | UNOWNED / NOT BASELINED | **MISSING** - root cause of the matrix's Conflict #3 | - |

### 4.10 L8 — Application ecosystem

| GAP | Capability | Dependencies | Increment / Version | Status | Closure evidence |
|---|---|---|---|---|---|
| MM-G8 | HTTP | G7 | **158** / v1.3.0 (PLANNED) | Implemented; client-oriented, **no TLS** | Phase 125A |
| MM-G9 | Database | - | UNOWNED / NOT BASELINED | Implemented (non-embedded); no real transactions | Phase 125A |
| MM-G6 | Filesystem / I/O | G10 | **156** / v1.3.0 (PLANNED) | Implemented (C path only) | Phase 109 |
| MM-G10 | Processes / env / dynamic libs | - | UNOWNED / NOT BASELINED | **MISSING** - no spawn, environment or `dlopen` surface | - |
| MM-G7 | Networking | G10 | **158** / **161** / v1.3.0 (PLANNED) | Implemented (C path only); Windows link contract on every build | Phase 125A |
| MM-O5b | Application framework | L5, L7 | UNOWNED / NOT BASELINED | **MISSING** - no service/application framework in either inventory | - |
| MM-N6 | Distributed computing | - | UNOWNED / NOT BASELINED | **MISSING** | - |

### 4.11 L9 — Distribution and platform ecosystem

Roadmap ownership for platform/package milestones is taken from
`KARKAIN-VERSION-PLAN.md` and `KARKAIN-INDEPENDENCE-ROADMAP.md`. This document
assigns none.

| GAP | Capability | Dependencies | Increment / Version | Status | Closure evidence |
|---|---|---|---|---|---|
| MM-J1 | Dependency resolution | - | UNOWNED / NOT BASELINED | Implemented | Phase 97/135/144 |
| MM-J2 | Lock file | J1 | UNOWNED / NOT BASELINED | Implemented | Phase 105/135 |
| MM-D5 | Dependency handling | J1 | UNOWNED / NOT BASELINED | Implemented (restored 2026-10-02) | Phase 97/135/144 |
| MM-M9 | Reproducible builds (see note) | H13, H14 | **154** / v1.2.0 (NOT BASELINED) | **Partial** - **duplicate capability record of `MM-J11`** (restored 2026-10-02) | `phase131_repro_test.go` |
| MM-J3 | Package manifests | J1 | UNOWNED / NOT BASELINED | Implemented (`karkain.toml`) | Phase 144 |
| MM-J4 | Workspace support | J3 | UNOWNED / NOT BASELINED | Implemented | Phase 144 |
| MM-J5 | Local registry | J1 | UNOWNED / NOT BASELINED | Implemented (directory registry, SHA-256) | Phase 135 |
| MM-J6 | Remote registry | J5 | UNOWNED / NOT BASELINED | **PARTIAL** - real HTTP client; **default endpoint reachability unverified**; git fetch is an explicit "not available" error | not network-probed |
| MM-J7 | Integrity | - | UNOWNED / NOT BASELINED | Implemented; **does not manage link-time native deps** | Phase 135 |
| MM-D6 | API stability / versioning | D5 | UNOWNED / NOT BASELINED | **MISSING** - a SemVer policy document exists but **no enforcement mechanism** was found in `pkg/`; policy is documentation only | `semver-policy.rst` (policy only) |
| MM-J8 | Cross compilation | I1 | UNOWNED / NOT BASELINED | Implemented; depends on externally installed cross-linkers | Phase 111 |
| MM-J9 | Installation / distribution | J1 | **165** / v1.2.0 (PLANNED) | Implemented; requires the release job to publish archives | Phase 142/143 |
| MM-J10 | Build profiles | H7 | UNOWNED / NOT BASELINED | **MISSING** - no user-visible optimization control | - |
| MM-J11 | Reproducible builds | H13, H14 | **154** / v1.2.0 (NOT BASELINED) | **PARTIAL** - incremental cache and bootstrap gated; ordinary `build` determinism not gated | `phase131_repro_test.go` |

> **`MM-M9` is a duplicate capability record of `MM-J11`.** The authoritative matrix
> states reproducible builds twice — once as `J11` (build/package) and once as `M9`
> (security/reliability) — with the same substance. Both are mapped in Step 2 and both
> land at **L9**, but **L9 coverage must count them once**. It is not a separate
> capability and no new gap is created.
| MM-G11 | Static linking | J7 | **152** / v1.2.0 (IN PROGRESS) | **PARTIAL** - link-time external deps unmanaged by the toolchain | observed CI failure class |
| MM-I1 | x86_64 Linux (C path) | - | UNOWNED / NOT BASELINED | Implemented + tested | Phase 114; Linux CI |
| MM-I2 | x86_64 Windows (C path) | - | UNOWNED / NOT BASELINED | Implemented + tested (`-lws2_32` contract) | Phase 88-90 |
| MM-I3 | x86_64 macOS (C path) | I1 | UNOWNED / NOT BASELINED | Implemented, **untested** - no macOS runner | - |
| MM-I4 | aarch64 (Linux/Windows/macOS) | I1 | UNOWNED / NOT BASELINED | **PARTIAL** - CI builds only; no execution evidence | CI build matrix |
| MM-I5 | riscv64 Linux | I1 | UNOWNED / NOT BASELINED | **STRUCTURAL ONLY** - never executed | - |
| MM-I6 | WebAssembly / WASI | H8 | UNOWNED / NOT BASELINED | Implemented; **Go engine only, no kcc parity** | Phase 108, wasmtime-executed |
| MM-I7 | Native ELF (C-free) | H10, H14 | **150** COMPLETE, **151** in progress, **152** / v1.2.0 | Implemented; **kcc cannot emit it yet** | Phase 148; Linux CI executes |
| MM-I8 | Native PE (C-free) | H10, H14 | **150** COMPLETE, **151** in progress, **152** / v1.2.0 | Implemented; **kcc cannot emit it yet** | 10 PE-execution tests, windows/amd64 |
| MM-I9 | Native Mach-O PIE (C-free) | H10 | **150D** COMPLETE / v1.2.0 | **STRUCTURAL ONLY** - no Intel-mac runner; nothing executes | structural validation |
| MM-I10 | GPU / NPU / SIMD / quantum targets | - | UNOWNED / NOT BASELINED | Experimental - catalog and emitters only | Phase 124 catalog |
| MM-N1 | GPU backend | I10 | UNOWNED / NOT BASELINED | Experimental (real WGSL emit; compile-only) | Phase 124 |
| MM-N2 | NPU adapters | I10 | UNOWNED / NOT BASELINED | Experimental (always falls back to CPU reference) | Phase 78/98 |
| MM-N4 | Quantum | I10 | UNOWNED / NOT BASELINED | **STRUCTURAL ONLY** - modelling only | - |
| MM-N7 | Embedded / bare-metal targets | - | UNOWNED / NOT BASELINED | **MISSING** | - |
| MM-N8 | Interpreter / JIT | - | UNOWNED / NOT BASELINED | **STRUCTURAL ONLY** - not wired into the compile path | `pkg/jit` |

### 4.12 L10 — Self-sustaining / self-hosting direction

| GAP | Capability | Dependencies | Increment / Version | Status | Closure evidence |
|---|---|---|---|---|---|
| MM-H14b | Self-hosted compile without the Go toolchain | H9, H10, H14 | **154** / v1.2.0 (NOT BASELINED) | **PARTIAL / NOT STARTED** | none yet |
| MM-P3 | Value-model parity on the self-hosted engine | MM-P1 (closed) | **151** / v1.2.0 (IN PROGRESS) | **PARTIAL** - `bigint` is the hard part (`mpz_t` in kcc's `Value`) | `PHASE-151A-BASELINE.md` |
| MM-P6 | kcc-owned machine-code emission | MM-P3 | **151** / v1.2.0 (IN PROGRESS) | **NOT STARTED** - guarded by `kccOwnsNativeTargets == false`; encoder and all three containers already byte-identical | `PHASE-151-BASELINE.md` |
| MM-M12 | Fuzzing | - | UNOWNED / NOT BASELINED | **MISSING** - no fuzz targets | - |
| MM-M10 | Dependency authenticity | J7 | **158** / v1.3.0 (PLANNED) | **PARTIAL** - SHA-256 integrity, **no signature verification or advisory mechanism** | Phase 135 |
| MM-M6 | FFI isolation | - | UNOWNED / NOT BASELINED | **MISSING** - `import "C"` unchecked and unverified at link time | - |
| MM-M8 | Determinism | - | UNOWNED / NOT BASELINED | Implemented (codegen deterministic; KIR byte-stable) | documented |

### 4.13 Secondary inventory (`GA-`) - stale, unowned, retained for traceability

`KARKAIN_GAP_ANALYSIS.md` is recorded here **without adopting it as status**. Its own
header disclaims currency and authorisation. Each row is listed so none is silently
dropped, with the matrix's present-day position where one exists. **None has phase
or version ownership** - the document pre-dates the current roadmap.

| GAP | Stale finding | Superseded by | Present-day position |
|---|---|---|---|
| GA-C1 | Parser OOM in call-arg loops | MM-H2 | Not re-verified in the matrix; **no owner** |
| GA-C2 | No whole-program type checker | MM-A13, MM-H4 | Still open - resolver plus a conservative kcc checker, not a type checker |
| GA-C3 | No symbol table / undefined detection | MM-A14 | **Fixed** - name resolution exists |
| GA-C4 | Borrow checker: moves never expire; error line 0 | MM-B2 | Not re-verified; **no owner** |
| GA-C5 | Error handling non-functional; `?` a no-op | MM-E1, MM-E2 | **Fixed on the C path** (LH-1); a Result/Option surface now exists |
| GA-C6 | BUG-1/2/3/7/8 codegen correctness | MM-H8, MM-H9 | Superseded; corpus parity now gate-pinned |
| GA-A1 | Memory-model decision not formally adopted | MM-B1..B9 | **Still open and still unowned** - the largest structural decision outstanding |
| GA-A2 | Generics strategy undecided | MM-C1, MM-C2 | **Partially decided** - monomorphization shipped (Phase 146); textual-substitution limit remains |
| GA-A3 | Error-handling philosophy | MM-E2, MM-E3 | **Decided** - `Result`/`Option`, no exceptions |
| GA-A4 | Module/visibility model | MM-A15, MM-D1 | **Fixed** (Phase 103) |
| GA-A5 | Unified execution model undecided | MM-N1..N10 | Partially realised as separate accelerator targets; no unifying decision |
| GA-A6 | Capability security / effect system | MM-M6 | **Still undecided**; effect system has no representation in the matrix |
| GA-I1..I10 | Generics/FFI/sema/concurrency/GPU wiring gaps | MM-C1..C7, MM-G4, MM-F5 | Largely superseded by Phase 103/104/107/125A/146 |
| GA-T1..T5 | Parser tests, E2E, goldens, perf gates, fuzzing | MM-H13, MM-M12 | **Substantially fixed** (gates exist); fuzzing (GA-T5 = MM-M12) still missing |
| GA-D1..D4 | Spec coverage, memory-model spec, ABI spec, conformance matrix | MM-O1, MM-O9, MM-B1 | Still open: `SPEC.md` is a feature ledger; no spec-vs-code check; `C-ABI.md` exists but is not enforced |
| GA-F1..F6 | Unified execution, capability security, effect system, SIMD, AI-native, distributed | MM-N1..N8 | **Future ideas.** Not scheduled, not owned; N3 (SIMD) is already implemented |

---

## 5. Roadmap relationship

### 5.1 The roadmap is the execution schedule

`KARKAIN-VERSION-PLAN.md` and `KARKAIN-INDEPENDENCE-ROADMAP.md` determine **what is
implemented and when**. They own increment numbers, baselines, dependencies between
increments, version closure conditions and the QA battery. This document changes
none of that.

### 5.2 The completeness model is the target specification

L0-L10 define **what capabilities Karkain ultimately needs**. Together:

- **Completeness does not replace the roadmap.** The roadmap is the schedule; this
  document is the specification of the destination.
- **The roadmap must eventually cover the completeness requirements.** Every level
  in section 4 must end up owned by some increment, or it stays visible as
  `UNOWNED / NOT BASELINED`.
- **An increment closes only the capabilities actually implemented and verified.**
  Increment 152 owns no-C linking; it does not own the memory model.
- **A planned increment is not automatically complete.** v1.3.0 and v1.4.0 rows
  above are `PLANNED`, not baselined, and their capabilities are not delivered.
- **Unowned gaps must remain visible.** A large fraction of section 4 is currently
  unowned. That is recorded, not hidden.
- **Future version decisions should be made from dependencies and closure
  evidence**, not from the wish to tick off a level.

### 5.3 Reading the two documents together

| Question | Answered by |
|---|---|
| What must exist for a complete language? | This document (L0-L10) |
| What is missing right now? | Section 4, from the authoritative matrix |
| What is being built next? | `KARKAIN-VERSION-PLAN.md` |
| What did a given increment actually close? | Its baseline note and final report |
| What evidence was produced? | Gate files, audit notes, CI steps |

---

## 6. Version decision rule

```text
Completeness requirement
        ↓
GAP identified                       (with repository evidence, not intuition)
        ↓
Dependency analysis                  (what must exist first)
        ↓
Existing increment identified        (or found not to exist)
        ↓
Implementation slice defined         (bounded; one capability or one defect class)
        ↓
Increment baselined                  (scope frozen in a baseline note)
        ↓
Focused verification                 (a gate that cannot pass by skipping)
        ↓
Increment COMPLETE                   (feature status COMPLETE, not merely IMPLEMENTED)
        ↓
Release closure                      (all version closure conditions, none waived)
```

**If a capability cannot fit an existing increment without violating its documented
scope, it is marked `UNOWNED / CARRY-OVER`.** It is never folded into a neighbouring
increment to make a level look closer to complete. `AGENTS.md` carries the
frozen-increment rule that governs how corrective work is evidenced against a
completed increment.

**This document creates no phases and assigns no new ownership.** Where a row above
reads `UNOWNED / NOT BASELINED`, that is a finding for a future scheduling decision,
not a pending assignment.

---

## 7. Current-release treatment: v1.2.0 as documented

v1.2.0 is **Sovereignty I - Off C, Off Go** (increments 150, 151, 152, 154, plus
free-floating 165). Its scope is **not rewritten here**, and three distinctions are
preserved exactly as the repository states them:

### 7.1 What v1.2.0 does target

- **Go-free native execution / toolchain behaviour.** Increment 150 delivered the
  C-free x86-64 backend with ELF, PE and Mach-O writers; the 152 increment queue
  targets linking and running `hello`, `stdlib_v2` and one `std.net` program with
  `gcc`, `clang` and `cl` **absent from `PATH`**. Increment 154 targets
  `stage2 == stage3` with no Go toolchain on the build host.
- **`--target c23` remains available and unchanged.** Removing the C path is
  explicitly *not* a v1.2.0 goal.

### 7.2 The remaining bootstrap / C toolchain boundary

After 154 the toolchain is **Go-free but not C-free**: `kcc` emits C23, so bootstrap
still links through a C compiler. The version plan states this explicitly - it "does
not make building the compiler C-free", and removing that last borrow "needs `kcc`
to compile and link itself natively, which has no increment, baseline or gate yet"
(`KARKAIN-ARCHITECTURE-ROADMAP.md` section 6). That capability is therefore
**recorded here as unowned**, and is **not** counted against v1.2.0 and **not**
silently dropped.

### 7.3 Future removal of remaining dependencies

Dependency removal proceeds in the documented order `150 -> 151 -> 152 -> 154`.
Anything beyond it - the C-free compiler link, the memory model, first-class types -
belongs to a later version and, where no increment names it, is `UNOWNED`.

### 7.4 v1.2.0 does not close language completeness

**v1.2.0 is a sovereignty release, not a completeness release.** At the time of
writing, increment 150 is COMPLETE, 151 is in progress (151A entry stub and syscall
tails still open), and 152 is in progress (152-A complete; 152-B0 complete but with
stderr delivery unproven on this host; 152-B1 outstanding). **None of that changes
any level's status.** Specifically:

| Level | v1.2.0 effect |
|---|---|
| L0 | Advances it (150 complete, 151/152/154 partial) |
| L1, L2, L3, L4, L5, L6, L7, L8, L9, L10 | **No level is closed by v1.2.0** |

---

## 8. Architectural conclusion

Karkain's goal is **not** merely Turing completeness. Being able to express any
computation is L1, and Karkain has that. "Complete general-purpose language and
ecosystem" is a much larger target, and it is the sum of:

```text
Language                        (L1)
+ Type/Data Model               (L2)
+ Memory/Safety                 (L3)
+ Runtime                       (L4)
+ Standard Library              (L5)
+ Concurrency                   (L6)
+ Developer Tooling             (L7)
+ Application Ecosystem         (L8)
+ Distribution                  (L9)
+ Self-Sustaining Ecosystem     (L10)
```

These four things must never be conflated:

| # | Category | What it is today |
|---|---|---|
| 1 | **Architectural target** | L0-L10 as defined here. The destination. Not a schedule. |
| 2 | **Currently implemented capabilities** | What the gates prove today: a working compiler, two engines, a real value model, a real standard library, real concurrency, real tooling - and a documented set of divergences between the two engines (MM-A5, MM-A10 origin, MM-H9). |
| 3 | **Currently baselined roadmap work** | 150 COMPLETE; 151 and 152 in progress with baseline notes; 154 and 165 named but not baselined. v1.3.0/v1.4.0 increments named but PLANNED. |
| 4 | **Future / unowned gaps** | Most of section 4. The memory model (GA-A1, MM-B1..B9), a first-class type object (matrix candidate P2), string-slice semantics, the C-free compiler link, wasm kcc parity, process/system APIs, time, traits, a REPL, fuzzing, spec-vs-code conformance. |

**The honest summary: Karkain is a working, tested, genuinely capable language
today, with a real two-engine compiler and a broad standard library. It is not a
complete general-purpose language, and the distance is concentrated in exactly four
places - the memory model (L3), the type system (L2), recoverable-error semantics
(L4), and the self-sustaining toolchain (L10).** None of those is closed by v1.2.0,
and this document does not pretend otherwise.

Progress against this target is measured by gates and closure evidence, never by
percentage.

---

## 9. Ambiguities requiring an architectural decision

Recorded, not resolved here. Each needs an owner.

| # | Ambiguity | Why it blocks | Where |
|---|---|---|---|
| 1 | **ID collision between the two inventories.** `MM-A1` and `GA-A1` are unrelated findings | Any future tool reading these documents by ID would resolve the wrong row. `GA-` needs a distinct prefix or retirement | section 4.1 |
| 2 | **String-slice bounds semantics.** C23's `karkain_slice` clamps; the native backend traps | A language-semantics decision, not an implementation one. Blocks a clean L4 story | `PHASE-152-BASELINE.md` |
| 3 | **Memory model.** Ownership/borrowing direction is evident but never formally adopted | Blocks B5, B8, F6 and therefore parts of L3 and L6 | `MEMORY_MODEL_DECISION.md`, GA-A1 |
| 4 | **A first-class type object.** Annotations are strings; `int` width is a C-preamble property; structs are maps | Blocks A9 field checking, generics beyond textual substitution, constraints, exhaustiveness, and a coherent `?` design | matrix candidate P2 |
| 5 | **The C-free compiler link.** `kcc` emits C23, so bootstrap needs a C compiler | No increment, baseline or gate owns it | `KARKAIN-ARCHITECTURE-ROADMAP.md` section 6 |
| 6 | **`SPEC.md` as specification vs feature ledger.** It states features the default engine cannot compile (matrix Conflict #3) | Blocks MM-O9 and every downstream status claim | matrix section 1.3 |
| 7 | **Time is absent from both inventories.** No clock, duration or calendar surface is recorded anywhere | L5 completeness cannot be assessed for a domain nobody has documented | section 4.7 |

---

## Step 2 — GAP → Completeness-Level Mapping

### 2.1 Mapping rules

Step 2 answers exactly one question:

> **What part of a complete general-purpose Karkain does each authoritative GAP belong to?**

It does **not** answer which phase or version owns the GAP (Step 3), what must happen before what (Step 4), which work has no owner (Step 5), or how future versions are decided (Step 6).

#### Authoritative inventory

| Source | Rows | Count |
|---|---|---|
| `MM-*` area rows | `docs/audit/KARKAIN-1.1.0-MATURITY-MATRIX.md` §3.1–§3.15 | **136** |
| `K-*` standard-library module rows | same, §3.11 | **16** |
| | **Total authoritative GAPs** | **152** |

Counts were extracted mechanically from the matrix tables, not estimated. An earlier working assumption of 19 `K-*` rows was **wrong** and is corrected here: the matrix lists 16 stdlib modules. Three further records (`MM-K-bytes`, `MM-K-time`, `MM-K-fs`) were synthesised by Step 1 to name domains that have no module; they are retained below as **Domain Coverage Records** and are **not** counted.

`GA-*` rows from `KARKAIN_GAP_ANALYSIS.md` remain historical and are **not** mapped here. A `GA-*` finding is relevant only where the authoritative matrix independently supports the corresponding `MM-*` row.

#### Primary-level rule

A GAP's primary level is the L-level whose **requirement list names the capability's defining characteristic**. Where two levels name it, the more specific one is primary.

Worked example: `MM-B5` (deallocation) → **L3**, because reclamation belongs primarily to the memory and safety model, even though the allocator that would need reclaiming is runtime code.

Classification is by **capability**, never by the phase a GAP currently appears in. A GAP that happens to sit inside increment 152 is still mapped to whatever level its capability belongs to.

#### Secondary-level rule

A secondary level is recorded **only** where a genuine architectural relationship exists — that is, where the capability has a real obligation at another level that is not merely incidental. "This capability is useful elsewhere" is not a reason.

Every secondary assignment carries its own rationale in section 2.2. **18** of the 152 GAPs carry a secondary level.

Two level collisions required a tie-break rule, stated here because a purely mechanical rule would otherwise produce a wrong answer:

| Situation | Rule applied |
|---|---|
| A `K-*` **library module** record and a `G*` **runtime capability** record describe the same area (networking, database, file I/O) | The `K-*` module record is **L5**. The `G*` capability record is **L4/L5** if it is runtime substrate, or **L8** if it is an application-domain capability (HTTP services, database-backed applications). Every capability therefore has exactly one home. |
| The same capability is recorded twice in the matrix (`MM-M9` / `MM-J11`) | Both are mapped for traceability and explicitly marked as one capability. Coverage counts it once. |

#### Dependency rule

Dependencies are transcribed **only** where the authoritative matrix records them in its `Dependencies` column. Where none is recorded the entry reads *"No precise GAP dependency recorded."* No dependency is inferred from intuition, co-location in a section, or adjacency in a diagram.

Two dependency kinds are distinguished and never conflated:

| Kind | Meaning | Recorded as |
|---|---|---|
| **GAP dependency** | A direct dependency between two identified GAPs | The matrix `Dependencies` column, verbatim |
| **Architectural prerequisite** | An unresolved design decision constraining several GAPs | The named decision — **never** a fabricated GAP ID |

The architectural prerequisites already identified are the **memory-model direction**, the **first-class type-object model**, and **string-slice semantics**.

#### Inventory correction carried out by Step 2

Step 1's section 4 omitted **11** authoritative rows. All eleven are restored into their appropriate existing sections, marked `(restored 2026-10-02)`:

`MM-D1`, `MM-D2`, `MM-D3`, `MM-D4`, `MM-D5`, `MM-G1`, `MM-G2`, `MM-G3`, `MM-G4`, `MM-G5`, `MM-M9`.

The D-series and G-series omissions were the material ones: they are substrate rows other GAPs depend on, and omitting them made the dependency picture look cleaner than the evidence supports.

#### Relationship to section 4

Section 4 remains the **status and evidence** record; this section is the **level mapping**. Where they differ, **section 2.2 is authoritative for the level**, and the divergence is recorded rather than silently applied. Seven rows were refined:

| GAP | Section 4 placement | Step 2 primary | Reason |
|---|---|---|---|
| `MM-A14` | L1 | **L0** | L0's own requirement list reads "semantic analysis / name resolution" |
| `MM-A15` | L2 | **L1** | A name-access rule is a scoping / execution-semantics question, not a type-lattice one |
| `MM-A16` | L1 | **L4** | Split from `MM-H15`; L4 lists "diagnostics: consistent wording across engines" |
| `MM-M8` | L10 | **L0** | Codegen determinism and KIR byte-stability are compiler infrastructure |
| `MM-M12` | L10 | **L7** | Fuzzing is a verification technique; L10 is a self-hosting *state* |
| `MM-N3` | L1 | **L2** | Lane-vector *types* are a type-system capability (L9 secondary) |
| `MM-A5` | L2 | **L2** | Unchanged; L10 added as a secondary, not a move |

---

### 2.2 Authoritative GAP mapping

All **152** authoritative GAPs, each appearing exactly once. "No precise dependency recorded" means the authoritative matrix records none.

#### A — Language and type foundation

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-A1 | 64-bit signed `int` | **L1** | L2 | No precise dependency recorded | Implemented; width is a C-preamble property | A primitive value needed to express computation is L1; the recorded gap (no `int` type object in the checker) is a type-representation concern, hence L2 secondary |
| MM-A2 | `float64` | **L1** | — | No precise dependency recorded | Implemented | Primitive value; L1 names "primitive values and literals" |
| MM-A3 | `bool` | **L1** | — | No precise dependency recorded | Implemented | Primitive value with defined truthiness |
| MM-A4 | `string` | **L1** | L2 | No precise dependency recorded | Implemented; case handling ASCII-only | The value is an L1 primitive; its recorded gap (undefined case-handling verdict) is an L2 string-semantics requirement |
| MM-A5 | `bigint` / `bigfloat` | **L2** | L10 | MM-A13 | **PARTIAL — Go engine only**, verified divergence | Arbitrary-precision numerics is an L2 type-family requirement; carrying it on the default engine is a self-hosting obligation |
| MM-A6 | Unsigned / 128-bit integers | **L2** | — | MM-A1 | **MISSING** — annotations parse and are inert | A missing numeric type family is a type-system requirement |
| MM-A7 | Type annotations as semantics | **L2** | — | MM-A6 | **PARTIAL** — stored as strings, never consulted | Annotations failing to carry semantics is a type-representation gap |
| MM-A8 | Arrays | **L1** | L2 | MM-A1 | Implemented; heterogeneous, no static element type | An indexable aggregate is an L1 computational need; element typing and bounds semantics are L2 |
| MM-A9 | Records / structs | **L2** | — | Architectural prerequisite: first-class type object (matrix records "Type system") | **PARTIAL — runtime maps**; no layout, no field checking | L2 requires nominal identity and checked field types; the recorded gap is precisely that absence |
| MM-A10 | Maps | **L2** | — | MM-A1 | Implemented (Go) / divergent (kcc); candidate **P1 recorded CLOSED** | L2 requires a defined key domain and key-equality semantics |
| MM-A11 | First-class `fn` values | **L2** | — | MM-A9 | **PARTIAL** — bare function references broken on both engines | L2 names function values callable through a value |
| MM-A12 | Enums + `match` | **L2** | — | MM-A9 | **PARTIAL** — tag-only; no payload destructuring, no exhaustiveness | Tagged unions with declared payloads are an L2 requirement |
| MM-A13 | Type checking / inference | **L2** | — | No precise dependency recorded | **PARTIAL** — no lattice, unification, or inference | The core L2 requirement: a type lattice as a first-class checker object |
| MM-A14 | Name resolution / scope | **L0** | L1 | No precise dependency recorded | Implemented | L0 explicitly names "semantic analysis / name resolution"; scoping as *execution semantics* is the L1 obligation |
| MM-A15 | Module visibility | **L1** | L2 | MM-D1 | Implemented | A name-access rule is part of scoping semantics, not of the type lattice |
| MM-A16 | Compile-time diagnostics | **L4** | L7 | No precise dependency recorded | Implemented — spans, JSON contract | Compile-time rejection is a failure class, and L4 requires consistent wording across engines. Distinct from `MM-H15`, which is diagnostics *infrastructure* (L0) |

#### B — Memory model

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-B1 | Ownership tracking | **L3** | — | MM-B2 | **PARTIAL — lexical, not semantic** | Ownership is a defining L3 concern |
| MM-B2 | Borrow checker | **L3** | — | MM-B1 | **PARTIAL** — no aliasing rules, no use-after-free detection, no `&mut` | Shared/mutable borrowing rules are the core L3 concern |
| MM-B3 | Escape analysis | **L3** | — | No precise dependency recorded | **STRUCTURAL ONLY — dead** | Determines stack-vs-heap placement, an L3 allocation/lifetime question |
| MM-B4 | Heap allocation | **L3** | — | No precise dependency recorded | Implemented (runtime) | A defined allocation surface is an explicit L3 requirement |
| MM-B5 | Deallocation | **L3** | — | MM-B4 | **MISSING at language level** | A defined reclamation strategy is an explicit L3 requirement; the matrix notes it needs ownership to know when an allocation is dead |
| MM-B6 | Move semantics | **L3** | — | MM-B1 | **MISSING** | The move/copy distinction is a defining L3 concern |
| MM-B7 | References / `&` | **L3** | — | MM-B1 | **PARTIAL** — `&` is the Phase 41 borrow operator in prefix position; infix `&` is bitwise | A reference type distinct from the borrow operator is an L3 requirement |
| MM-B8 | Aliasing rules | **L3** | — | MM-B1 | **MISSING** | Rules preventing unsound aliasing are an L3 requirement |
| MM-B9 | Safety guarantees | **L3** | — | MM-B2 | **PARTIAL** — safety is runtime-only, not proven | L3 requires one stated memory-safety guarantee and its enforcement |
| MM-M4 | Memory safety | **L3** | — | MM-B5 | **MISSING** — generated C uses raw pointers; user allocations never reclaimed | The end-to-end memory-safety guarantee is an L3 requirement |

#### C — Generics and metaprogramming

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-C1 | Generic functions | **L2** | — | MM-A13 | **PARTIAL — textual substitution only** | L2 requires generic substitution not be purely textual |
| MM-C2 | Generic types | **L2** | — | MM-C1, MM-A9 | **PARTIAL** — same textual limit | Parameterised types are an L2 requirement |
| MM-C3 | Type parameters | **L2** | — | MM-C1 | Implemented (v1); **no inference** | L2 requires inference for unannotated bindings; the explicit-type-arg design is the recorded compromise |
| MM-C4 | Constraints / traits | **L2** | — | MM-C1 | **STRUCTURAL ONLY** — nothing declares against them | L2 requires an abstraction mechanism to declare and check against |
| MM-C5 | Compile-time evaluation | **L2** | — | No precise dependency recorded | **MISSING** — no `const`, no const-fn | Compile-time evaluation is a type/generic-system capability, not a runtime one |
| MM-C6 | Macros | **L2** | — | MM-C1 | **PARTIAL** — expansion-based, not hygienic | Metaprogramming operates on the type/AST model, so it sits in the same level as the rest of the compile-time language |
| MM-C7 | Reflection | **L2** | — | MM-A13 | **MISSING** — no runtime type query surface | Runtime type queries require a first-class type object, which is an L2 requirement |

#### D — Modules and libraries

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-D1 | Modules / imports | **L0** | — | No precise dependency recorded | Implemented — kcc owns flat assembly itself | Module assembly is a compiler pipeline responsibility (`resolveSources`, `kccStageInput`, `assembleProject`); the language-level name-binding aspect is L1 |
| MM-D2 | Namespaces | **L1** | — | MM-D1 | Implemented; kcc namespaces flattened, not hierarchical | Name qualification is an execution-semantics question; the flattening is a compiler lowering detail |
| MM-D3 | Multi-file programs | **L1** | — | MM-D1 | Implemented — sibling join performed by both engines | Composing a computation from several source units is an L1 language need |
| MM-D4 | Library consumption | **L5** | — | MM-D1 | Implemented — `import std.x`, 15 modules | Consuming an importable, tested library is exactly L5's stated requirement |
| MM-D5 | Dependency handling | **L9** | — | MM-J1 | Implemented | Resolving, fetching and locking third-party dependencies is distribution, which is L9 |
| MM-D6 | API stability / versioning | **L9** | — | MM-D5 | **MISSING** — a SemVer policy document exists but no enforcement mechanism was found | Versioning policy governs released artifacts, i.e. distribution |

#### E — Error handling

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-E1 | `?` propagation | **L4** | L2 | MM-A13 | Partial — lowers on the C path; **refused on the C-free native target** | Recoverable-error propagation is a defining L4 requirement; designing it coherently needs a type object |
| MM-E2 | Result / Option values | **L4** | L2 | MM-A13 | **STRUCTURAL ONLY** per the matrix; superseded in practice by LH-1 expression-level constructors | A real `Result`/`Option` surface that works is an explicit L4 requirement |
| MM-E3 | Exceptions / panic | **L4** | — | No precise dependency recorded | **MISSING** — deliberately; `Result` is the chosen direction | Exceptions are a failure class; the matrix records the choice not to have them |
| MM-E4 | Runtime error reporting | **L4** | — | No precise dependency recorded | Implemented on both engines; native path added by 152-B0 | One diagnostic format with source location is an explicit L4 requirement |
| MM-E5 | Compile-time error recovery | **L4** | — | No precise dependency recorded | Implemented — project-wide multi-error preflight | Recovering from and reporting multiple failures is an L4 requirement |
| MM-E6 | Compile-fail corpus | **L4** | — | No precise dependency recorded | Implemented | The negative corpus is the evidence that L4 failure semantics are defined |

#### F — Concurrency

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-F1 | `spawn` / `join` / `wait_all` | **L6** | — | No precise dependency recorded | Implemented (both engines) | A spawnable unit is an L6 requirement |
| MM-F2 | Channels | **L6** | — | No precise dependency recorded | Implemented — bounded/unbounded | Defined send/receive semantics are an L6 requirement |
| MM-F3 | Actors | **L6** | — | No precise dependency recorded | Implemented | An L6 concurrency abstraction |
| MM-F4 | Scheduler | **L6** | — | No precise dependency recorded | Implemented (runtime); work-stealing | The execution facility behind L6 tasks |
| MM-F5 | Atomics exposed to the language | **L6** | — | MM-F1 | **MISSING** — `__atomic` helpers exist only inside the C runtime | Language-level atomic primitives are an explicit L6 requirement |
| MM-F6 | Data-race guarantees | **L6** | — | MM-B2, MM-F5 | **MISSING** — no `Send`/`Sync` typing | Safe-sharing typing is an explicit L6 requirement, and L6 is stated to depend on L3 |
| MM-F7 | Parallel execution | **L6** | — | MM-F4 | Partial — runtime only, not a language contract | Parallel execution belongs to the concurrency level regardless of where it is realised |

#### G — Runtime, system and FFI

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-G1 | Layered runtime | **L4** | — | No precise dependency recorded | Implemented — freestanding / core / concurrency | Runtime architecture is the substrate L4 failure semantics execute on |
| MM-G2 | Startup model | **L4** | — | No precise dependency recorded | Implemented — generated `main`; native `_start` and PEB bootstrap | Process/system startup is a runtime concern; its absence produces undefined failure behaviour |
| MM-G3 | Value representation | **L4** | L2 | No precise dependency recorded | Implemented — tagged `Value` union mirrored by kcc | The runtime value representation is runtime substrate; that a *type lattice* must also be representable is the L2 obligation |
| MM-G4 | C interoperability | **L4** | L7 | No precise dependency recorded | **PARTIAL** — `import "C"` unchecked and unverified at link time | An unchecked foreign boundary is a failure-class and safety concern; tooling must surface it |
| MM-G5 | System calls | **L4** | L5 | No precise dependency recorded | **PARTIAL (runtime only)** — raw syscalls; **not exposed as a language surface** | The recorded evidence concerns syscall/runtime substrate (L4); a future stdlib-facing syscall API would be an L5 obligation |
| MM-G6 | File I/O | **L5** | L4 | No precise dependency recorded | Implemented on the C path; absent on the C-free native target | A filesystem module is an L5 domain; the runtime primitives behind it are L4 |
| MM-G7 | Networking | **L5** | L4 | No precise dependency recorded | Implemented; `-lws2_32` contract on every generated-C link | A networking module is an L5 domain; sockets are runtime substrate |
| MM-G8 | HTTP | **L8** | — | MM-G7 | Implemented; client-oriented, **no TLS** | HTTP services are named explicitly in L8's application-ecosystem list |
| MM-G9 | Database | **L8** | — | No precise dependency recorded | Implemented, non-embedded; no real transactions | Databases and persistence are named explicitly in L8's application-ecosystem list |
| MM-G10 | Processes / env vars / dynamic libs | **L5** | — | No precise dependency recorded | **MISSING** | A system/process stdlib domain that no module currently backs; distribution is the secondary obligation once packages ship native code |
| MM-G11 | Static linking | **L9** | — | MM-J7 | **PARTIAL** — link-time external deps unmanaged by the toolchain | Controlling what an artifact links against is a distribution/build concern |

#### H — Compiler architecture

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-H1 | Lexer | **L0** | — | No precise dependency recorded | Mature | Lexer is an L0 stage |
| MM-H2 | Parser | **L0** | — | MM-H1 | Mature | Parser is an L0 stage |
| MM-H3 | AST | **L0** | — | MM-H1 | Mature | AST is an L0 stage |
| MM-H4 | Semantic analysis | **L0** | — | No precise dependency recorded | **PARTIAL** — resolution plus annotation matching | Semantic analysis / name resolution is named in L0 |
| MM-H5 | SSA IR | **L0** | — | No precise dependency recorded | Implemented, partial use; closure-bearing functions take the legacy path | IR is named in L0 |
| MM-H6 | HIR | **L0** | — | MM-H5 | **STRUCTURAL ONLY — unwired** | An IR representation is an L0 concern |
| MM-H7 | Optimizer | **L0** | — | MM-H5 | **PARTIAL** — 2 of 6 passes on the real path | Optimization is named in L0 |
| MM-H8 | Code generation (C) | **L0** | — | No precise dependency recorded | Mature | Code generation is named in L0 |
| MM-H9 | Code generation (self-hosted) | **L0** | — | MM-H4 | Implemented, **divergent** on A5 and A10 | The second engine is an explicit L0 requirement |
| MM-H10 | Code generation (C-free machine code) | **L0** | — | MM-H1–H3 | Implemented at HEAD; partial at v1.1.0 | Machine-code emission is an L0 stage |
| MM-H11 | Linker integration | **L0** | — | MM-G11 | **PARTIAL** — writers exist, no general linker | Linking code-generation output is an L0 stage; the artifacts it produces are an L9 concern |
| MM-H12 | Register allocation (native only) | **L0** | — | MM-H10 | Implemented | A code-generation sub-stage |
| MM-H13 | Incremental compilation | **L0** | — | MM-H8 | Implemented — whole-assembly cache; native split refused with a measured reason | An L0 compilation capability; reproducible artifacts are the secondary obligation |
| MM-H14 | Bootstrap / self-hosting | **L0** | — | MM-H9, MM-H10 | **PARTIAL** — `stage2==stage3` gated; Go oracle and C linker still required | Bootstrap foundations are named in L0; the self-sustaining endpoint is L10 |
| MM-H15 | Diagnostics | **L0** | — | No precise dependency recorded | Mature | Diagnostic infrastructure is compiler machinery; its presentation is a tooling concern |

#### I — Targets and platforms

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-I1 | x86_64 Linux (C path) | **L9** | — | No precise dependency recorded | Implemented and tested | Platform support is named in L9 |
| MM-I2 | x86_64 Windows (C path) | **L9** | — | No precise dependency recorded | Implemented and tested (`-lws2_32` contract) | Platform support is named in L9 |
| MM-I3 | x86_64 macOS (C path) | **L9** | — | MM-I1 | Implemented, **untested** — no macOS runner | Platform support is named in L9 |
| MM-I4 | aarch64 (Linux/Windows/macOS) | **L9** | — | MM-I1 | **PARTIAL** — CI builds only, no execution evidence | Platform support is named in L9 |
| MM-I5 | riscv64 Linux | **L9** | — | MM-I1 | **STRUCTURAL ONLY** — never executed | Platform support is named in L9 |
| MM-I6 | WebAssembly / WASI | **L9** | — | MM-H8 | Implemented; **Go engine only, no kcc parity** | WASM is named explicitly in L9 |
| MM-I7 | Native ELF (C-free) | **L9** | — | MM-H10, MM-H14 | Implemented; **kcc cannot emit it yet** | Native distribution is named in L9 |
| MM-I8 | Native PE (C-free) | **L9** | — | MM-H10, MM-H14 | Implemented; **kcc cannot emit it yet** | Native distribution is named in L9 |
| MM-I9 | Native Mach-O PIE (C-free) | **L9** | — | MM-H10 | **STRUCTURAL ONLY** — no Intel-mac runner; nothing executes | Native distribution is named in L9 |
| MM-I10 | GPU / NPU / SIMD / quantum targets | **L9** | — | No precise dependency recorded | Experimental — catalog and emitters only | Accelerator targets are a platform/distribution surface |

#### J — Build, package and distribution

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-J1 | Dependency resolution | **L9** | — | No precise dependency recorded | Implemented | Package management is named in L9 |
| MM-J2 | Lock file | **L9** | — | MM-J1 | Implemented | Lockfiles are named in L9 |
| MM-J3 | Package manifests | **L9** | — | MM-J1 | Implemented (`karkain.toml`) | Package manifests and build configuration are named in L9 |
| MM-J4 | Workspace support | **L9** | — | MM-J3 | Implemented | Workspace dependency grouping is a distribution concern |
| MM-J5 | Local registry | **L9** | — | MM-J1 | Implemented — directory registry, SHA-256 | Package management is named in L9 |
| MM-J6 | Remote registry | **L9** | — | MM-J5 | **PARTIAL** — real HTTP client; default endpoint reachability unverified; git fetch is an explicit "not available" error | A remote package source is a distribution concern |
| MM-J7 | Integrity | **L9** | — | No precise dependency recorded | Implemented; **does not manage link-time native deps** | Content integrity for distributed artifacts |
| MM-J8 | Cross compilation | **L9** | — | MM-I1 | Implemented; depends on externally installed cross-linkers | Cross-compilation is named in L9 |
| MM-J9 | Installation / distribution | **L9** | — | MM-J1 | Implemented; requires the release job to publish archives | Installers and release artifacts are named in L9 |
| MM-J10 | Build profiles | **L9** | — | MM-H7 | **MISSING** — no user-visible optimization control | Build configuration is named in L9 |
| MM-J11 | Reproducible builds | **L9** | — | MM-H13, MM-H14 | **PARTIAL** — incremental cache and bootstrap gated; ordinary `build` determinism not gated | Reproducibility of distributed artifacts is a distribution property |

#### K — Standard library modules

All 16 rows in this table are the authoritative `K-*` records from matrix §3.11. Per the tie-break rule in section 2.1, a **module** record is L5 even where a `G*` runtime-capability record covers the same area.

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-K-string | `std.string` | **L5** | — | No precise dependency recorded | Implemented, frozen; EXTEND backlog recorded | An importable, tested, both-engine string module |
| MM-K-collections | `std.collections` | **L5** | — | No precise dependency recorded | Implemented, frozen; EXTEND backlog recorded | An importable, tested, both-engine collections module |
| MM-K-io | `std.io` | **L5** | — | Architectural prerequisite: none recorded; matrix notes `std.io` is gated native-only | Implemented; merge verdict with `std.system` recorded | The filesystem/I-O domain is named in L5 |
| MM-K-encoding | `std.encoding` | **L5** | — | No precise dependency recorded | Implemented; **thin wrapper over C builtins**; JSON to move to a new `std.json` | Encoding/decoding is named in L5 |
| MM-K-crypto | `std.crypto` | **L5** | — | Matrix records the dependency on unsigned/bitwise support (MM-A6 class) | Implemented; **thin wrapper over C builtins** | Cryptography is named in L5 |
| MM-K-numerics | `std.numerics` | **L5** | — | No precise dependency recorded | Implemented; EXTEND backlog recorded | Mathematics is named in L5 |
| MM-K-math | `std.math` | **L5** | — | No precise dependency recorded | **NOT IMPORTABLE** — source-present behind the Phase-109 boundary | Mathematics is named in L5; the recorded gap is that the domain is not actually satisfied |
| MM-K-testing | `std.testing` | **L5** | — | No precise dependency recorded | Implemented; EXTEND backlog recorded | A test-support module is part of the library surface |
| MM-K-generics | `std.generics` | **L5** | — | MM-C1 | Implemented — exercises monomorphization; Go only (kcc boundary pinned) | A generic container library is an L5 module resting on the L2 generics capability |
| MM-K-net | `std.net` | **L5** | — | Architectural prerequisite: none recorded; matrix records `MM-G7` as the capability | Implemented; thin wrapper; forces `-lws2_32` on Windows; **absent on native** | The networking domain is named in L5; the `G7` capability record is mapped separately |
| MM-K-http | `std.http` | **L5** | — | MM-G7 | Implemented; canonical `.kark` codec and loopback client | An importable HTTP module is an L5 library record; the application-domain obligation is `MM-G8` at L8 |
| MM-K-db | `std.db` | **L5** | — | No precise dependency recorded | Implemented, **non-embedded** SQL-text engine; no real transactions | A database module is an L5 library record; the application-domain obligation is `MM-G9` at L8 |
| MM-K-core | `std.core` | **L5** | — | No precise dependency recorded | **STRUCTURAL ONLY** — outside the 10 frozen modules | A core library module is an L5 record; its recorded state is that it is not a satisfied domain |
| MM-K-system | `std.system` | **L5** | — | No precise dependency recorded | **STRUCTURAL ONLY** — no OS API surface behind it | A system/process module is an L5 record; the missing surface is `MM-G10` |
| MM-K-async | `std.async` | **L5** | — | No precise dependency recorded | **STRUCTURAL ONLY** — non-canonical syntax, no builtin backing | An async library module is an L5 record; concurrency obligations proper are at L6 (`MM-F1`–`F7`) |
| MM-K-gpu | `std.gpu` | **L5** | — | MM-N1 | Experimental | A GPU library module is an L5 record; the backend capability is `MM-N1` at L9 |

#### L — Developer tooling

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-L1 | Formatter | **L7** | — | No precise dependency recorded | Implemented — token-level, idempotent; no reflow | Format is named in L7 |
| MM-L2 | LSP | **L7** | — | No precise dependency recorded | Implemented — no rename/refactor, no incremental cache | IDE / LSP integration is named in L7 |
| MM-L3 | Debugger | **L7** | — | No precise dependency recorded | **PARTIAL** — drives gdb in batch; **not a DAP server**; Go engine only | Debug is named in L7 |
| MM-L4 | Test runner | **L7** | — | No precise dependency recorded | Implemented — kcc owns discovery and driver synthesis | Test is named in L7 |
| MM-L5 | Doc generation | **L7** | — | No precise dependency recorded | Implemented; language reference is thin vs 140 audit reports | Documentation is named in L7 |
| MM-L6 | IDE integration | **L7** | — | No precise dependency recorded | Implemented — VS Code extension | IDE integration is named in L7 |
| MM-L7 | REPL | **L7** | — | No precise dependency recorded | **MISSING** | An interactive evaluation loop belongs with developer tooling |
| MM-L8 | Linter | **L7** | — | No precise dependency recorded | Implemented | Lint is named in L7 |
| MM-L9 | Profiler | **L7** | — | No precise dependency recorded | Implemented — counts only `malloc`/`free` text in generated code | Profiling / tracing is named in L7 |

#### M — Security and reliability

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-M1 | Bounds checking | **L4** | — | No precise dependency recorded | Implemented (runtime), **not a static guarantee** | Bounds failures are an explicit L4 failure class |
| MM-M2 | Division-by-zero / bad index | **L4** | — | No precise dependency recorded | Implemented; now also on the native path | Arithmetic failures are an explicit L4 failure class |
| MM-M3 | Integer overflow | **L4** | L2 | No precise dependency recorded | **MISSING** — `<<` wraps; no checked `+ - *` | L4 requires a defined overflow verdict; a checked-arithmetic type lattice is the L2 obligation behind it |
| MM-M5 | Null / bounds on native backend | **L4** | — | MM-E4 | **PARTIAL** — array bounds diagnosed; **string-slice semantics unresolved** | Backend-specific bounds behaviour is a failure-semantics question |
| MM-M6 | FFI isolation | **L4** | L3 | No precise dependency recorded | **MISSING** — `import "C"` unchecked and unverified at link time | An unenforced foreign boundary is a failure-class issue; the aliasing it permits is an L3 issue |
| MM-M7 | Input validation | **L4** | — | No precise dependency recorded | **PARTIAL** — not systematic | Invalid-conversion and validation behaviour are L4 failure classes |
| MM-M8 | Determinism | **L0** | L10 | No precise dependency recorded | Implemented — codegen deterministic, KIR byte-stable | Deterministic code generation and IR stability are compiler infrastructure; bit-reproducible bootstrap is what consumes it |
| MM-M9 | Reproducible builds | **L9** | — | MM-H13, MM-H14 | **PARTIAL** — **duplicate capability record of `MM-J11`**; counted once in L9 coverage | Reproducibility of distributed artifacts is a distribution property. This row and `MM-J11` describe **one** capability |
| MM-M10 | Dependency authenticity | **L9** | — | MM-J7 | **PARTIAL** — SHA-256 integrity, **no signature verification or advisory mechanism** | Authenticity of distributed dependencies is a supply-chain/distribution concern |
| MM-M11 | Compiler crash handling | **L0** | — | No precise dependency recorded | **PARTIAL** — fallback covers malformed IR, not wrong-code survival | Surviving malformed IR inside the compiler is a compiler-robustness concern |
| MM-M12 | Fuzzing | **L7** | — | No precise dependency recorded | **MISSING** — no fuzz targets | Fuzzing is a verification technique belonging with the test/lint toolchain, not with the self-hosting state L10 describes |

#### N — Advanced capabilities

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-N1 | GPU backend | **L9** | — | MM-I10 | Experimental — real WGSL emit, compile-only | An accelerator backend is a platform/target surface |
| MM-N2 | NPU adapters | **L9** | — | MM-I10 | Experimental — always falls back to CPU reference | An accelerator backend is a platform/target surface |
| MM-N3 | SIMD | **L2** | L9 | No precise dependency recorded | Implemented and gated | Lane-vector *types* extend the type system; their realisation as a target/backend is the L9 obligation |
| MM-N4 | Quantum | **L9** | — | MM-I10 | **STRUCTURAL ONLY** — modelling only | A quantum target is a platform surface |
| MM-N5 | Autodiff | **L5** | — | No precise dependency recorded | Implemented — reverse-mode over the existing IR | Automatic differentiation is a mathematics-domain library capability |
| MM-N6 | Distributed computing | **L8** | — | No precise dependency recorded | **MISSING** | Messaging and distributed application support are named in L8 |
| MM-N7 | Embedded / bare-metal targets | **L9** | — | No precise dependency recorded | **MISSING** | Embedded targets are a platform surface |
| MM-N8 | Interpreter / JIT | **L9** | — | No precise dependency recorded | **STRUCTURAL ONLY** — not wired into the compile path | An alternative execution/distribution surface |

#### O — Documentation and specification

| GAP ID | GAP description | Primary L-level | Secondary | Dependencies | Current status | Mapping rationale |
|---|---|---|---|---|---|---|
| MM-O1 | Language specification | **L7** | — | No precise dependency recorded | **PARTIAL** — behaves as a **feature ledger, not a specification** (no formal grammar); overstates the type system | Documentation is named in L7; the type facts it must state are L2 |
| MM-O2 | Architecture documentation | **L7** | — | No precise dependency recorded | Partial — 140 audit reports; documents history, not a current reference | Documentation is named in L7 |
| MM-O3 | Standard-library docs | **L7** | — | No precise dependency recorded | Implemented — repaired after drift | Documentation is named in L7 |
| MM-O4 | Examples | **L7** | — | No precise dependency recorded | Implemented — 220 `.kark`, 42 categories | Documentation is named in L7 |
| MM-O5 | Tutorials | **L7** | — | No precise dependency recorded | **MISSING** — no tutorial path in `docs/source` | Documentation is named in L7 |
| MM-O6 | API reference | **L7** | — | No precise dependency recorded | **PARTIAL** — hand-maintained, not generated from source | Documentation is named in L7 |
| MM-O7 | Release documentation | **L7** | — | No precise dependency recorded | Implemented | Documentation is named in L7 |
| MM-O8 | Conformance documentation | **L7** | — | No precise dependency recorded | Implemented — 12 files / 64 tests, both engines | Documentation is named in L7 |
| MM-O9 | Spec-vs-code conformance tooling | **L7** | — | MM-O1 | **MISSING** — root cause of the matrix's Conflict #3 | The automated check that keeps documentation and implementation honest is tooling |

---

### 2.3 Dependency observations

These are **observations**, not a dependency graph. The complete graph is Step 4's work.

**1. Level-level dependencies declared by the model itself.** L6 states that it depends on L3: *"Without an ownership or borrow model there is no sound way to share mutable state."* That is why `MM-F6` records `MM-B2` and `MM-F5` as GAP dependencies and is blocked behind the memory model.

**2. The type-representation cluster is the widest single blocker.** Per the matrix's §11.2 chain, everything downstream is blocked on one missing thing: a first-class type object. Affected: `MM-A9`, `MM-C1`, `MM-C2`, `MM-C4`, `MM-A12`, `MM-E1`, `MM-E2`, `MM-C7`, `MM-M3`. All map cleanly to L2, so the ambiguity constrains their **dependencies**, not their levels.

**3. The memory-model cluster.** Per the matrix §11.3, `MM-B5` cannot be solved before ownership exists — without ownership there is no way to know when an allocation is dead. This propagates to `MM-M4` and forward to `MM-F6`.

**4. The restored rows are substrate, not incidental.** `MM-D1` is depended on by `MM-A15`, `MM-D2`, `MM-D3`, `MM-D4` and `MM-D5`; `MM-G11` by `MM-H11`; `MM-J7` by `MM-G11` and `MM-M10`; `MM-H13`/`MM-H14` by `MM-J11` and `MM-M9`; `MM-I10` by `MM-N1`, `MM-N2` and `MM-N4`. Omitting them understated how much of the inventory hangs off a handful of foundation rows.

**5. Two cross-level chains worth naming without graphing.**
- *Determinism to self-hosting*: `MM-M8` (L0) is what `MM-H14` bootstrap byte-identity consumes. Determinism is a **precondition** of L10's closure claim, not part of it.
- *Library modules to runtime capability*: each `K-*` module at L5 rests on a `G*` runtime capability at L4/L5. `MM-K-net` needs `MM-G7`; `MM-K-system` exists precisely because `MM-G10` is missing.

**6. Dependencies were not inferred.** Where the matrix `Dependencies` column reads `-`, this mapping writes "No precise dependency recorded" rather than substituting a plausible neighbour. Several rows carry no recorded dependency where one architecturally exists — `MM-A9` records the non-GAP phrase "Type system" — and those are recorded as **architectural prerequisites**, never as invented GAP IDs.

---

### 2.4 Ambiguous mappings

All seven Step 1 ambiguities remain open. None is resolved here, and none is resolved by existing repository evidence.

| # | Ambiguity | GAPs whose mapping is affected | Effect on the level | Effect on dependencies / rationale |
|---|---|---|---|---|
| 1 | `MM-*` vs `GA-*` ID collision | All 136 `MM-*` and 16 `K-*` rows, by construction | **None** | Traceability only. A tool reading these documents by ID would resolve the wrong row; `GA-` needs a distinct prefix or retirement |
| 2 | String-slice bounds semantics (C23 clamps, native traps) | `MM-A4`, `MM-M5` | **None** — `A4` stays L1/L2, `M5` stays L4 | The string-semantics rationale for `MM-A4` and the bounds rationale for `MM-M5` cannot be finalised. `MM-M5`'s recorded status already says "string-slice semantics unresolved" |
| 3 | Memory-model direction never formally adopted | `MM-B1`–`MM-B9`, `MM-M4`, `MM-F6` | **None** — all L3, and `MM-F6` L6 | Dependencies recorded as the **architectural prerequisite** "memory-model direction", never a fabricated GAP. Matrix candidate P4 is the same decision |
| 4 | First-class type-object model | `MM-A9`, `MM-A12`, `MM-C1`, `MM-C2`, `MM-C4`, `MM-C7`, `MM-E1`, `MM-E2`, `MM-M3` | **None** — all L2, with `E1`/`E2` at L4 and `M3` at L4 | Recorded as the **architectural prerequisite** "first-class type-object model". Matrix candidate P2 is the same decision |
| 5 | C-free compiler-link boundary (`kcc` emits C23, so bootstrap needs a C compiler) | `MM-H14`, `MM-I7`, `MM-I8` | **None** — `H14` L0 (L10 secondary), `I7`/`I8` L9 | No increment, baseline or gate owns it. Constrains the **rationale** for calling those rows partial, not their level |
| 6 | `SPEC.md` as specification vs feature ledger | `MM-O1`, `MM-O9`, and every status claim derived from `SPEC.md` | **None** — both L7 | `MM-O1`'s rationale is explicitly qualified: the ledger's overstated type claims are matrix Conflict #3, exactly what `MM-O9` would catch |
| 7 | Missing time / date / duration capability | **No GAP is affected, because no GAP exists** | **None** | `MM-K-time` is a non-authoritative Domain Coverage Record. **No GAP ID is invented for it.** This is the one ambiguity with no GAP to attach to |

**Count: 7 ambiguities recorded, 0 resolved, 0 GAP mappings abandoned.** No level assignment was forced to a false certainty, and none had to be abandoned because an ambiguity blocked it.

---

### 2.5 L0–L10 coverage summary

**A GAP count is not a completeness measure.** A level may have many GAPs because it is genuinely broad, few because its capabilities are largely implemented, or none because nobody has documented the gaps it has. Only the *nature* of the recorded rows distinguishes these, which is why the final column carries the interpretation and the count alone does not.

| Level | Primary GAPs | Major domains represented | Coverage reading | Notes |
|---|---|---|---|---|
| **L0** | **19** | Lexer, parser, AST, semantic analysis, SSA/HIR, optimizer, C codegen, self-hosted codegen, C-free codegen, linking, register allocation, incremental compilation, bootstrap, diagnostics, determinism, module assembly, crash handling | Broad; mostly **implemented**, several explicitly partial | Largest compiler-side cluster. `MM-H9` divergence, unwired optimizer passes and the `MM-H14` bootstrap limits are the open rows |
| **L1** | **8** | Primitives (`int`, `float64`, `bool`, `string`), arrays, visibility, namespaces, multi-file composition | **Nearly complete on recorded evidence** — every row reads Implemented | Smallest primary cluster. The low count reflects genuine capability here, not missing documentation |
| **L2** | **16** | Numeric type families, annotations-as-semantics, records, maps, function values, enums, type checking, generics (functions, types, parameters, constraints, const-eval, macros, reflection), SIMD lane types | The **least complete** high-count level | `MM-A6`, `MM-C5`, `MM-C7` MISSING; `MM-A9` map-based; `MM-C4` structural. Matches the section 8 conclusion that the distance is concentrated in the type system |
| **L3** | **10** | Ownership, borrowing, escape analysis, allocation, deallocation, moves, references, aliasing, safety guarantees, end-to-end memory safety | Almost entirely **MISSING or structural** | The single largest structural hole. `MM-B5`, `B6`, `B8`, `M4` MISSING; `MM-B3` dead |
| **L4** | **18** | Compile-time diagnostics, `?`, Result/Option, exceptions, runtime error reporting, error recovery, fail corpus, runtime layering, startup, value representation, C interop, syscalls, bounds, div-by-zero, overflow, native bounds, FFI isolation, input validation | **Mixed and the most heterogeneous level** | Diagnostic and failure-reporting rows are implemented; the value-semantic rows (`E1`/`E2`) and validation rows (`M3`, `M6`, `M7`) are not. `G1`–`G5` were restored here by Step 2 |
| **L5** | **21** | 16 stdlib module records plus filesystem/IO, networking, processes-env-dylibs, autodiff | **Highest count and the most uneven** | `std.core`, `std.system`, `std.async` STRUCTURAL ONLY; `std.math` NOT IMPORTABLE; `std.encoding`/`std.crypto` are thin C wrappers. Most recorded gaps here are module-existence rather than capability gaps |
| **L6** | **7** | Tasks, channels, actors, scheduler, atomics, data-race guarantees, parallel execution | **Runtime complete, language surface absent** | `F1`–`F4`, `F7` implemented; `F5`/`F6` MISSING. The runtime is capable and the language cannot express it |
| **L7** | **19** | Formatter, LSP, debugger, test runner, doc generation, IDE, REPL, linter, profiler, all 9 documentation/spec rows, fuzzing | **Broadly implemented** with specific holes | `MM-L3` is not a DAP server; `MM-L7`, `MM-O5`, `MM-O9`, `MM-M12` MISSING; `MM-O1` is a ledger, not a specification |
| **L8** | **3** | HTTP, database, distributed computing | **Lowest non-empty count** | Two of three implemented but qualified (no TLS; non-embedded). `MM-N6` MISSING. `MM-G10` is the adjacent ecosystem capability |
| **L9** | **31** | 11 build/package rows, 10 target/platform rows, dependency handling, API versioning, static linking, reproducibility, dependency authenticity, 5 advanced-capability rows | Largest primary count; **mixed**, with several MISSING/structural rows | `MM-J6`, `MM-J10`, `MM-J11`, `MM-D6`, `MM-M10` open. `MM-M9` is counted **once** here as a duplicate of `MM-J11`, so the distinct-capability count is 30 |
| **L10** | **0** | *(no authoritative GAP row maps here)* | **No authoritative primary GAP row exists** | Self-hosting capability is recorded **inside** `MM-H9`, `MM-H10`, `MM-H14`, `MM-I7`, `MM-I8` at L0/L9, plus synthesized Step 1 records `H14b`/`P3`/`P6` that are not authoritative. **No GAP is promoted or fabricated to populate L10.** Two secondary-level entries point at it: `MM-A5` (L2→L10) and `MM-M8` (L0→L10) |

#### 2.5.1 Distinguishing the four reasons a level may look unrepresented

Per the Step 1 brief, a low GAP count must not be read as completeness. Each level's position is classified explicitly:

| Level | Classification |
|---|---|
| L0, L1, L7 | **Implemented capability with recorded residual gaps.** Not complete; no level is. |
| L2, L3, L5, L6, L9 | **Recorded gaps, unevenly.** Real capability exists and real holes are documented |
| L8 | **Implemented capability, narrow.** Only three rows, two qualified |
| **L10** | **Capability not represented in the current GAP inventory.** The self-hosting *state* has no authoritative row of its own; it is an attribute of L0/L9 rows plus synthesized non-authoritative records |
| **Time / date / duration** | **Capability not represented in the current GAP inventory, and not representable yet** — there is no module and no row to map. See the Domain Coverage Records below |

**No level is architecturally complete.** The largest clusters by count (L5, L9) are not the least complete; the smallest by count (L8) is not the most complete. Count is a measure of how much is *documented*, not of how much is *done*.

---

### Domain Coverage Records — Not Authoritative GAPs

> **These records are completeness-domain observations and are not included in the authoritative GAP count.**

Step 1 synthesised three records to name domains that the authoritative inventory does not represent as rows. They are retained because each records a real completeness observation, and they are kept **out** of the 152 because no authoritative GAP row and (for two of them) no stdlib module corresponds to them. They must not be read as gaps, and **no GAP ID is invented for them**.

| Record | Domain | Observation | Would-be level | Authoritative status |
|---|---|---|---|---|
| `MM-K-bytes` | Bytes as a first-class surface | Byte-level operations exist **only** as `*_bytes` C builtins reached through `std.encoding`; there is no `std.bytes` module and no independent row | L5 | **Not authoritative.** Real gap is already visible in `MM-K-encoding` (thin wrapper over C builtins) and `MM-A6` |
| `MM-K-fs` | Filesystem as a distinct domain | There is no `std.fs` module. The domain is carried by `stdlib/io` (`MM-K-io`) and the `MM-G6` runtime capability, both of which are authoritative rows | L5 | **Not authoritative.** Fully represented by existing rows; this record adds no new gap |
| `MM-K-time` | **Time / date / duration** | **The authoritative inventory contains no GAP row and no stdlib module for this domain.** `stdlib/` has no time module; `MM-G10` (processes/env) is MISSING but is not a time surface; neither the matrix nor `KARKAIN_GAP_ANALYSIS.md` records a clock, duration or calendar capability anywhere | L5 | **Not authoritative, and not representable.** This is the one domain where L5 completeness **cannot be assessed at all** — not "assessed and found missing", but *unassessable* because no row exists |

`MM-K-time` is recorded here rather than promoted because the honest statement is that **the capability is absent from the inventory**, not that a gap row was lost. Recording it as a GAP would assert an inventory completeness the repository does not have. It remains **Step 1 ambiguity #7**, open, and is deliberately the reason L5's time coverage is reported as *not representable* in section 2.5.

---

### 2.6 Step 2 conclusion

**What Step 2 established.**

1. **The authoritative inventory is 152 GAPs** — 136 `MM-*` area rows and 16 `K-*` stdlib module rows — extracted mechanically from the authoritative matrix, not estimated. A prior working assumption of 19 `K-*` rows was **wrong** and is corrected.
2. **Every one of the 152 has exactly one primary L-level**, assigned by capability rather than by phase, with a stated rationale. Distribution: L0 19, L1 8, L2 16, L3 10, L4 18, L5 21, L6 7, L7 19, L8 3, L9 31, L10 0.
3. **18 GAPs carry a justified secondary level.** Each carries its own architectural reason.
4. **Eleven authoritative rows omitted by Step 1 are restored** — `MM-D1`–`D5`, `MM-G1`–`G5`, `MM-M9` — into their appropriate existing sections. The omission was material because these are substrate rows other GAPs depend on.
5. **`MM-M9` is recorded as a duplicate capability record of `MM-J11`.** Both are mapped; L9 coverage counts the capability once (30 distinct, not 31).
6. **Seven architectural ambiguities remain recorded and unresolved**, with the specific GAPs each one affects and an explicit statement of what it does *not* affect. **No mapping was abandoned and no certainty was forced.**
7. **L10 has no authoritative primary GAP row.** That is reported as a finding, not repaired by promoting or inventing a row.
8. **Three synthesized records are quarantined** as Domain Coverage Records, with `MM-K-time` recorded as a domain the inventory cannot represent at all.
9. **Two level collisions were resolved by a stated tie-break rule**, so that every capability has exactly one home rather than appearing at two levels.

**What Step 2 deliberately did not establish.**

- **No phase or version ownership.** This document assigns nothing to v1.2.0 or any other version, and does not restate or revise the increments named in section 4's historical columns. That is **Step 3**.
- **No complete dependency graph.** Section 2.3 records observations only, and does not order work. That is **Step 4**.
- **No list of unowned work.** The `UNOWNED / NOT BASELINED` cells in section 4 are transcribed, not adjudicated. That is **Step 5**.
- **No future-version decision rule.** Section 6 already records the rule; Step 2 does not extend or apply it. That is **Step 6**.
- **No implementation, no scope change, no status promotion.** Nothing here moves a capability's status. Per section 3.1, `COMPLETE` is the only status satisfying a completeness requirement, and Step 2 changes no status.
- **No completeness claim.** Karkain remains a working, capable language with the distance concentrated in L2, L3, L4's value-semantic rows, and L10 — as section 8 already records. Step 2 redistributes that distance across levels; it does not shorten it.
