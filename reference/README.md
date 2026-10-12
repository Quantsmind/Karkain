# reference/

Differential oracle for the Go toolchain: the canonical `karkain` CLI and the
Go-engine C codegen output, kept **off the user path** so the default engine of
the compiler pipeline is the self-hosted kcc.

**Role (154 / I-6):** with the `KARKAIN_ENGINE=go` stage-1 pin retired, every
compiler child launched by the bootstrap runs the **default** engine — the
self-hosted kcc — and `kcc builds kcc`. The Go tree is therefore not part of the
user-visible toolchain; it is retained here as the differential oracle for the
Go engine (and for the Go-side diagnostics), so both engines can be compared
byte-for-byte, and for never deleting the Go path (the C path `--target c23`
remains first-class).

**Contents:** an archived copy of the Go CLI source (`cmd/karkain`,
`pkg/cli`) and the Go-engine C that the Go engine emits for `src/compiler`
(and the `reference/` tree is the only place on the user path where the Go
toolchain exists).

**Safety:** never delete the Go path. If the self-hosted compiler becomes
unusable, restore this archive from the tagged release and re-verify
byte-for-byte before trusting the self-hosted compiler again.

**Gate:** `pkg/bootstrap/phase154_selfbootstrap_test.go`
`TestPhase154_ReferenceArchive` asserts this directory plus its
`.gitignore` coverage.
