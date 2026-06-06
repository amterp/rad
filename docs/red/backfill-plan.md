# RED Backfill Plan

> **Working doc, not a RED.** The catalog of past Rad decisions that are candidates to backfill into
> [REDs](README.md), in roughly chronological order. In a future session, go through these one by one
> and decide in the moment whether each is worth a RED (or should fold into another). As each is
> written, update its row and the [index](README.md).
>
> Dates / shipped versions are evidence; titles, groupings, and IDs are a sketch for the per-RED
> interview, not settled scope. No "why" is asserted here - that's recovered when each RED is written.
> The `Evidence` pointers are the *starting trail* for archaeology, not exhaustive; the
> `red-archaeologist` Phase 1 should still dig fresh.

**Intended workflow.** Work through the list in **strict chronological order** with the
`red-archaeologist` - top to bottom of the master list, one item at a time, finishing each before
starting the next. **Do not jump around** - not by tier, not by theme (see [Write order](#write-order)
below). The chronology *is* the sequence: a decision is backfilled at its place in time, never pulled
forward because it relates to something earlier. For each item, decide in the moment whether it's worth
a RED: if **yes**, write it and mark the row `✅ done` (`RED-N`); if **no**, leave the row in place but
mark it `⏭ skip` with a one-line reason (don't delete - the record of "considered and declined" is the
point). Repeat until the list is all `✅`/`⏭`.

**Conventions.** **Tier** = priority, *not* write order (chronology is the order - see [Write
order](#write-order)): **1** = foundational, almost certainly worth a RED and worth the most depth; **2**
= notable; **3** = small, often a fold-candidate or skip. **Kind** = Language / Architecture / Tooling / Process. **IDs** are provisional
fractional letters (between `A` and `B` → `AA`, etc.); a single integer renumber happens at the end.
Bias is to **split** into discrete chronological decisions, with a few thematic exceptions noted inline.
`RED-2` (genesis) also absorbs small genesis-level facets (see its entry).

Legend: (unmarked) = to-do · ✅ done · ⏭ skip (considered, not worth a RED) · ⊘ reserved · ↦ supersession
link. Original name was **RSL** (Rad Scripting
Language); renamed to Rad in 2025-05 (`AG`). Commit messages in this repo explain *why* - mine them.

---

## Master list

### Genesis & v0.1–v0.2 (2024-08 to 2024-09)

- **`2` Why build a language + founding principles** · Language · v0.1.0 · **T1** · ✅ (RED-2)
  Why a purpose-built CLI language over Bash / general-purpose lang + library; the founding principles.
  *Also absorbs genesis-level facets too small to stand alone:* interpreter (tree-walk) over compiler;
  EBNF-first design (Crafting-Interpreters-influenced) + the AST codegen pipeline; the Go implementation
  choice (rationale silent in commits - interview); flat variable scoping as a deliberate scripting
  choice; the pre-1.0 "breaking changes in minor versions" versioning policy (advertised, no single commit).
  *Evidence:* `0dd21e62` (first commit, 2024-08-01); `93ec082f` (begin interpreter); `66054cc2` (AST
  codegen); `477b9c30` (flat scoping stated explicitly); `docs/thinking/principles.md`, `docs/ebnf.md`;
  `README.md`; RED-3 Context section.

- **`3` The rad block** · Language · v0.1.0 · **T1** · ✅ (RED-3)

- **`3A` Declarative argument parsing (args block)** · Language · v0.1.0 · **T1** · ✅ (RED-4)
  The `args:` block: auto-generated help, types, constraints. Language-level model only; the constraint
  *sub-language* is `AAE` and the parsing *engine* arc (Cobra→pflag→Ra) is `AI`.
  *Evidence:* `24313d97` "Begin EBNF: argBlock"; `4f1fe67f` (delete Cobra → pflag, 2024-11-19, explains
  the misfit: too custom, double-execute, binary size); `core/args.go`.

- **`3B` Removing null from the language** · Language · v0.1.0 · **T2** · ↦ superseded by `AEA`
  Null deliberately removed early - optional args handled by a "set" flag instead. Reversed in v0.5.32;
  write as a supersession pair with `AEA`.
  *Evidence:* `9ac0bc74` "Remove nulls from RSL" (rich in-commit rationale).

- **`3C` Bash-embedding ergonomics** · Language · v0.1–v0.2 · **T2**
  Thematic RED (accepted split-exception): several early choices driven by one force - Rad scripts should
  embed cleanly in Bash. The `---` header (vs `"""`, to avoid Bash-var escaping), single-quote strings,
  the `--STDIN`/`--SHELL` "borrow Rad's arg-parsing from Bash via `eval`" mode, and stderr-routed output.
  *Evidence:* `56da0669` (`---` separator), `d9f03fff` (single quotes), `eb88805b` (`--STDIN`/`--SHELL`,
  richest commit in the era), `0608d5f3` (Printer funnel → stderr).

- **`3D` JSON extraction engine (trie + merge algorithm)** · Architecture · v0.2.6→v0.4.7 · **T1**
  The computational core of the rad block: defined fields → trie → traversal → row/column emission. The
  v3 rewrite is one of the most reasoned commits in the repo (trie-node = one JSON scope; a four-case
  row/column/singleton merge). Distinct from the block *syntax* (`3`).
  *Evidence:* `24cdc989` (first trie), `18a923ee` (`*` wildcard), `c1062ee6` (v3 rewrite);
  `docs/thinking/json_syntax.md`; `core/json_*.go`.

- **`3E` `pick` / interactive choice model** · Language · v0.2.9 · **T3**
  First interactive builtin; programmatic pre-filter + interactive fallback; the charmbracelet/huh
  dependency cost was flagged. Family grew (`pick_kv`, `multipick`, `pick_from_resource`).
  *Evidence:* `e140973b` (implement pick).

### v0.3–v0.4 (2024-09 to 2025-01)

- **`3F` Data model: mixed-type arrays + maps (JSON parity)** · Language · v0.3.0 · **T1**
  Reworking arrays from typed (`int[]`) to mixed/nested, plus maps, so Rad can represent any JSON. The
  lasting split: typed collections only in `args:`; in-script collections dynamically typed (cut ~740 lines).
  *Evidence:* `e5b57558` (rework arrays), `23688565` (remove non-mixed), `2eb746a4` (maps);
  `docs/thinking/json_syntax.md`.

- **`A` request / display split** · Language · v0.3.1 · **T2** · ✅ (RED-A, Superseded by `B`)

- **`AA` Shell command integration model** · Language · v0.4 · **T1**
  Embedding shell while dodging Bash pitfalls: `$`/`$!`, `fail:`/`recover:`, critical commands, and the
  (error_code, stdout, stderr) capture ordering. Redesigned later in `AL`. Evidence reads like a mini-RED
  incl. rejected alternatives (`must`/`require`; `failure:` → `fail:`/`recover:`).
  *Evidence:* `2bb6e918`, `dcea8538`, `009c60d0` (capture ordering); `docs/thinking/shell_cmds.md`.

- **`AAC` String system: attributed strings (RslString)** · Language · v0.4.13 · **T1**
  Strings as structs carrying per-segment style attributes (vs baked-in escape codes), so
  `len()`/slicing/equality stay correct on styled text. Foundation of all color/styling. Raises the live
  `red("Alice") == "Alice"` equality question. Folds in: inline-expr interpolation + format specs,
  raw/multiline strings, thousands separator.
  *Evidence:* `f7a0c14d` (replace string with RslString), `802cba0d` (color funcs); `2b0f88f6`/`4803db58`
  (raw/multiline); `docs/thinking/string_thonk.md`.

- **`AAD` Early error model: error-map → prefix-`catch`** · Language · v0.4.21→v0.5.43 · **T2** · ↦ superseded by `AL`
  Rad's first two error-handling iterations before the `catch` suffix: (1) Go-style multi-return error
  maps with `RADxxxxx` codes you opt into checking; (2) a first-class `RadError` type + prefix `catch`
  expression. Worth preserving *why* each was tried and dropped.
  *Evidence:* `a534cbf1` (error-map for parse_int/float), `52969742` (error philosophy in `.dot` commit),
  `f01081b0` (RadError + prefix catch), `f4696b36` (destructuring).

- **`AAE` Arg constraint system** · Language · v0.4.25→v0.5.14 · **T2**
  The constraint sub-language and its progression: enum/regex as freestanding declaration lines (the enum
  commit argues explicitly *against* inline constraints) → range with interval notation `[0,100)` →
  relational `requires`/`excludes`/`mutually excludes` (first constraint *between* args; `is_defined()`
  string-ref limitation flagged).
  *Evidence:* `033d04b9` (enum), `306f3a4d` (regex), `12a9ca71` (range), `9d466157` (relational).

### Tree-sitter era (v0.5, 2025-02 to 2025-09)

- **`AB` Adopting tree-sitter (the Great Deletion)** · Architecture · v0.5.0 · **T1**
  Deleting the handwritten lexer/parser for a separate `tree-sitter-rad` grammar; confines CGo to `rts/`,
  enables the LSP. Root cause of the distribution constraints (`AIA`). Folds in the module shuffle (rts
  inlined, then rad/rts/lsp-server merged into one module after the workspace experiment failed).
  *Evidence:* `c42d44b1` "The Great Deletion™", `b47c4229`, `486a1107`; `47b257ba`, `8d8ad4ea` (module
  moves); memory `gotreesitter_migration`.

- **`ABB` UFCS (uniform function call syntax)** · Language · v0.5.17 · **T2**
  `a.foo(b)` ≡ `foo(a, b)`. A distinct, earlier decision than lambdas (don't bury in `AE`). The commit
  notes the usual method-ambiguity objection doesn't apply (Rad has no object methods); a `|` pipe
  alternative was weighed.
  *Evidence:* `7a549ad6`; `docs/thinking/ufcs.md`.

- **`AC` Creating the LSP / editor tooling** · Tooling · v0.4.34 · **T2**
  Origin decision: standing up radls + the VS Code extension; the bet that editor support matters. The
  2026 full-feature build-out is `BBA`; the radls-unpackaged-to-PATH move folds here.
  *Evidence:* `515708f6` "Initialize rsl lsp", `9c4df752` (VS Code starter), `d79186f` (PATH unpackaging);
  `docs/thinking/lsp.md`.

- **`AD` Stashes (persistent script state)** · Language · v0.5.24 · **T1**
  Durable per-script state/config keyed by a stable script identity. Distinctive to Rad; uses the `@macro`
  header (`ADA`) for identity.
  *Evidence:* `5f9f496a` "Implement new concept: 'stash'", `377e1cd6` (gen-id); `docs/thinking/stashes.md`.

- **`ADA` File-header `@macro` metadata syntax** · Language · v0.5.37 · **T2**
  A novel mechanism for static, pre-runtime metadata in the `---` header (`@script_id`,
  `@enable_args_block`, …); enables stash identity and passthrough scripts. Design iterated (double-header
  rejected for TextMate-unfriendliness; a disable→enable naming flip).
  *Evidence:* `8cbd26dc`, `bfde63f0`, `f627488f`; `docs/thinking/macros.md`.

- **`AE` User-defined functions & lambdas (`fn`)** · Language · v0.5.28 · **T2**
  `fn` lambdas + named functions and the functional `map`/`filter` style. Folds in: built-ins as
  first-class referenceable values.
  *Evidence:* `710ddfec` (fn lambdas), `f5816d3b` (map), `f9f11c76` (built-ins first-class);
  `docs/thinking/custom_functions.md`.

- **`AEA` Null re-addition (JSON parity)** · Language · v0.5.32 · **T1** · ↦ supersedes `3B`
  "I've been trying to avoid this since the start, but there's no way" - re-added for JSON type parity and
  relational-constraint usability. Also lands undefined-optional-args-resolve-to-null and `or` falsy
  coalescing (which deferred `??`). The other half of the null story.
  *Evidence:* `bd55d4ff` (add null type), `7caf64cc` (nulls falsy), `7c812865` (undefined args null),
  `374c05ab` (falsy coalescing).

- **`AF` `rad check` (static-check command)** · Tooling · v0.5.37 · **T3**
  Standalone static checker; foundation for later diagnostics and `--from-logs`.
  *Evidence:* `2f6a5b6d`; `docs/thinking/rad_check.md`.

- **`AG` The RSL → Rad rename (project identity)** · Process · ~v0.5.38 · **T2**
  Renaming project/language/extension (`.rsl`→`.rad`)/grammar-repo/module. The later lsp-server→radls +
  github.io→amterp.dev rebrand is a smaller echo (fold here or in `AIA`).
  *Evidence:* `4d208ee2`, `8d069ed9` + `.git-blame-ignore-revs`; `e0b90b94` (VSCode), `d30921ae`
  (grammar repo), `cb973de7` (module, 2025-08-03), `3df330f` (radls/amterp.dev).

- **`AH` Type system: runtime foundations** · Language · v0.5.43 · **T1**
  Runtime function type checking + the builtin signature model - Rad's first real typing. Folds in the
  numeric model (int64; `int/int`=float; numeric funcs type-preserving; `num` removal). Sequel: `BB`.
  *Evidence:* `4f7974f1` (runtime fn type checking), `0ce84ff5` (signatures); `b17b274a` (int64),
  `743d6fca` (int/int=float), `f183799f` (num removal), `18209b7` (type-preserving funcs);
  `docs/type_system.md`.

- **`AI` Extracting Ra (the args library)** · Architecture · v0.5.59 · **T2**
  pflag → standalone **Ra**, Rad's own arg-parsing library. Marquee instance of the forking pattern (`AHA`).
  *Evidence:* `043d1047` "Replace pflag with Ra" (PR #38, 2025-08); `docs/thinking/arg_library.md`.

- **`AJ` REPL** · Tooling · v0.5.59 · **T3** · *Evidence:* PR #43.

### Maturity (v0.6–v0.8, 2025-09 to 2026-01)

- **`AKB` stdin / Unix-pipe composability (`read_stdin`)** · Language · v0.6.10 · **T2**
  `read_stdin()`/`has_stdin()` for pipe-composable scripts; `read_stdin()` returns `str?|error` to
  distinguish nothing-piped / piped-but-empty / read-error; streaming deferred. Distinct from the v0.1
  `--STDIN` shell-export mode.
  *Evidence:* `cd880f7b`.

- **`AK` Config (global `config.toml`)** · Tooling · v0.6.11 · **T3**
  Global `~/.rad/config.toml`. (The invocation-logging / migration story it enabled is `AKA`.)
  *Evidence:* `a068c49b`.

- **`AL` Shell + error-handling overhaul (`catch` suffix)** · Language · v0.6.12 · **T1** · ↦ supersedes `AAD`
  Breaking redesign: `$!`/`unsafe`/`fail:`/`recover:` → unified `catch` suffix, critical-by-default. Third
  stage of the error-handling arc. Also lands named shell-output assignment (magic `stdout`/`stderr`/`code`
  names; name-inference chosen over destructuring).
  *Evidence:* `5c0253c`, `7ef3f46`; `docs/thinking/shell_cmds.md` (2025-10 entries),
  `docs/thinking/error_handling.md`.

- **`AM` Commands (subcommands) in Rad** · Language · v0.6.16 · **T2**
  First-class subcommands within a Rad script (multi-command CLIs). Folds in command-name hyphenation
  (kubectl-style).
  *Evidence:* `cef96618` "Implement commands in Rad", `c8373a2` (hyphenation); `docs/thinking/commands.md`.

- **`AN` For-loop redesign (`with loop`)** · Language · v0.7.0 · **T2**
  Breaking: implicit index-by-extra-variable → explicit `with loop` context (+ `loop.src` immutable
  snapshot). Rad-block lambda context (`ctx.idx/src/field`) shipped alongside.
  *Evidence:* `e98f7fc1` (full rationale in message), `21976ec1` (lambda context);
  `docs-web/docs/migrations/v0.7.md`.

- **`AO` Cross-platform / Windows first-class** · Architecture · v0.7.1 · **T3**
  Platform-abstraction layer + path normalization; "all platforms equal" CI reframe.
  *Evidence:* `5fdf8fd8`, `83c839e`, `668a509`; `core/common/platform.go`.

- **`AP` Remove `get_default` for `??`** · Language · v0.8.0 · **T2**
  Dropping `get_default` for the `??` fallback operator. First step of the null/error-operator arc
  (continues in `AQ`); required making out-of-bounds list/string indexing a catchable panic as a
  prerequisite.
  *Evidence:* `7bbad61` (remove get_default), `7350604` (`??` for indexing);
  `docs-web/docs/migrations/v0.8.md`.

### v0.9 & modern (2026-03 onward; some unreleased)

- **`AQ` `??` null-coalescing + `catch` operator** · Language · v0.9.0 · **T2**
  `??` broadened to fire on null *and* error; inline `catch` for error-only. Key principle: "data issue vs
  logic bug" decides what's catchable (null indexing yes, type errors no).
  *Evidence:* `4fa150c` (broaden `??`), `a7878a4` (catch operator); `docs-web/docs/migrations/v0.9.md`.

- **`AR` Strict `+` (no implicit coercion)** · Language · v0.9.0 · **T2**
  `+` no longer coerces int/float/bool to string; pushes users to interpolation/`str()`. Cross-ref `AH`/`BB`.
  *Evidence:* `1ec9e5c`, `7372764` (inline migration hint); v0.9 migration guide.

- **`ARA` Stop normalizing user data + `split_lines`** · Language · v0.9.0 · **T2**
  Principled, breaking: Rad normalizes script *source* internally but preserves *user data* exactly (no
  silent line-ending normalization); `split_lines()` added for explicit handling.
  *Evidence:* `14a4b2c0`.

- **`AS` Diagnostics & error-code system (`rad explain`)** · Architecture · v0.9.0 · **T1**
  Unified diagnostic type system, stable `RADxxxxx` codes, Rust-style renderer, `rad explain`, error docs,
  stack traces, did-you-mean. The rendering/codes layer; the upgrade-survival layer atop it is `AKA`.
  *Evidence:* `e2e06d4` (renderer), `7458ca7` (diagnostic types), `2f599d6` (rad explain), `338ffe3`
  (stack traces), `95ce0ff`/`2a09807` (did-you-mean) - all 2026-01-31; `core/error_docs/`, `rts/rl/errors.go`.

- **`AT` CST→AST migration** · Architecture · v0.9.0 · **T1**
  `core/` evaluates a Go-native AST instead of the tree-sitter CST; CGo sealed into `rts/`. The converter
  commit enumerates ~8 design choices. Sequel to `AB`.
  *Evidence:* `c518d3e` (AST nodes), `ec3185e` (converter), `31eb894` (interpreter), `14c797d` (checker),
  `b820903` (LSP), `2d3b92d` (seal CST), `a258ce7` (phases 5-8).

- **`B` Rad block unification** · Language · v0.9.0 · **T2** · ✅ (RED-B)
  Folds in: `filter` modifier + permanent-vs-temporary mutation semantics (`022267b`).

- **`BA` docs/funcs as source-of-truth (codegen)** · Tooling · ~v0.10 · **T2**
  `docs/funcs/<name>.md` as the single source generating the checker signature, LSP hover, and reference
  page; design goal "drift is structurally impossible," gated by `make verify-generated`.
  *Evidence:* `4c74002f`, `b2ce132d`; `docs/funcs/README.md`.

- **`BB` Type system: static checker (Pyright-model)** · Language/Arch · ~v0.10 / unreleased · **T1**
  Full bidirectional checker, much larger than "deepening": a binder separate from the checker (stable
  Symbol identity, named after Pyright), flow narrowing (if/elif/switch/loop join + exhaustiveness),
  Dynamic-vs-Any, Never bottom type, ErrorType poison, static user-fn typing, structural matching. *May
  warrant splitting (binder / narrowing / inference); verify what's actually released.* Drop-`list[T]`-for-`T[]` folds in.
  *Evidence:* `9c788fe9` (binder, names Pyright), `11b67a5a`, `78dc8c9a` (user-fn typing), `53da9c01`
  (narrowing, Pyright-grade message), `c5a3a701` (switch exhaustiveness), `4363b5c9`/`80cfc241`/`640f2137`
  (Dynamic/Never/ErrorType), `b3dba6b` (drop `list[T]`).

- **`BBA` Modern radls: full LSP + concurrent document model** · Arch/Tooling · unreleased · **T1**
  The 2026-05 sprint from diagnostics-only stub to a real language server (hover, completion, goto-def,
  references, document symbols, semantic tokens, rename, code actions) on a concurrent document model
  (immutable `DocumentVersion` snapshots via atomic pointer, debounce, position-encoding negotiation).
  Distinct from `AC`; landed alongside `BB`. Wire-level `lstesting` harness → `BC`.
  *Evidence:* `6241558e` (hover), `73965325` (rename), `e4b20229` (completion), `cd70d467` (immutable
  snapshots), `65a230eb` (position encoding) - mostly unreleased; `radls/lstesting/`.

### Cross-cutting / infrastructure (span multiple releases)

- **`AHA` The `amterp/*` library-forking pattern** · Architecture · 2024-09→ongoing · **T2**
  Recurring philosophy: when Rad leans heavily on an unmaintained library, fork and own it as `amterp/*`.
  ~7 libraries; ongoing maintenance is the trade-off. Relates to `AI`.
  *Evidence:* `4d58183c` (go-tbl ← tablewriter), `f4fe8143` (jsoncolor), `df97a509` (color), `043d1047`
  (ra ← pflag); `go.mod` (also flexid, go-delta, tree-sitter-rad).

- **`AIA` Distribution infrastructure (CGo-shaped)** · Tooling · v0.5.55→v0.10 · **T2**
  A multi-year arc all shaped by the tree-sitter **CGo** dep: module rename for `go install`, single-module
  consolidation (after the workspace experiment failed), goreleaser-cross + Docker for cross-compilation,
  `netgo`/`osusergo` for static Linux builds, Homebrew tap → Homebrew Core. Root cause: `AB`.
  *Evidence:* `cb973de7` (module rename), `8d8ad4ea` (consolidation), `33f30e4f` (goreleaser), `622026c`
  (static DNS), `5411d992` (Homebrew Core); `.goreleaser.yml`.

- **`AKA` Migration & upgrade-safety system** · Architecture/Process · v0.6.11→v0.9 · **T1**
  The deliberate system that makes frequent breaking changes survivable: invocation logging (designed for
  *eager* breakage discovery), `rad check --from-logs`, the three-layer help pattern (inline diagnostic
  hint → `rad explain RADxxxxx` → migration guide URL), and logging-on-by-default. Codified in AGENTS.md
  as the expected process for every breaking change. Distinct from `AS` (how errors render).
  *Evidence:* `c5ee8e77` (invocation logging), `468b4551` (`--from-logs`), `7372764` (three-layer hint),
  `b7b95e2` (default-on); `docs-web/docs/migrations/index.md`; CLAUDE.md "Breaking Changes & Migration Diagnostics".

- **`BC` Snapshot testing methodology** · Process · 2024→2026 · **T1**
  The choice to build a *custom* multi-section snapshot format (not an off-the-shelf Go lib), evolved in
  waves and now load-bearing across interpreter/parser/LSP/checker. Flagged by user memory
  `feedback_prefer_snapshot_tests`.
  *Evidence:* `80773eeb` (Go run-and-assert DSL), `5f4ee177` (error snapshots + doc-coverage gate),
  `7658ed37`/`9d57d246`/`e5cb3233` (unified CST/AST + `-update`), `52791c9c` (LSP wire-level), `68a78e18`
  (Check pipeline); `core/testing/snapshots/`, `rts/test/st_snapshots/`, `radls/lstesting/`.

- **`1` The RED process** · Process · — · **T1** · ✅ (RED-1)

**Tier-1 to-write ≈ 16** (the realistic "all the big decisions" target). Tier 2 ≈ 16, Tier 3 ≈ 6.

---

## Fold-ins (sections within a parent RED, not standalone)

Real decisions, but better as a section of another RED (evidence lives under the parent above):

- Numeric model (int64, `int/int=float`, type-preserving funcs, `num` removal) → `AH`.
- String features (inline-expr interpolation + format specs, raw/multiline strings, thousands separator) → `AAC`.
- Built-ins as first-class values → `AE`. · UFCS-vs-pipe alternative → `ABB`.
- Module structure (rts inlining; 3 modules → 1) → `AB` / `AIA`.
- Rad-block modifiers (filter, mutation semantics, lambda context) → `B` / `3`.
- Command hyphenation → `AM`. · `loop.src` → `AN`. · all-platforms-equal CI → `AO`.
- v0.9 small breaks (`parse_epoch` unit rename `2a09c43`; `get_stash_dir`→`get_stash_path` `b8ed277`;
  `trim_prefix/suffix` literal-vs-charset fix `0f02eb1` + `trim_left/right` `a7225df`) → notes in relevant
  func REDs. *(trim fix is borderline-standalone if the naming-philosophy angle is wanted.)*
- drop `list[T]` for `T[]` (`b3dba6b`) → `BB`. · lsp-server→radls + amterp.dev rebrand (`3df330f`) → `AG` / `AIA`.
- Small language-shape calls (named/keyword-only args `881cac6b`, var-args `*name` syntax `eff08bf0`,
  `defer`/`errdefer` `8895ba5a`/`6fe0e8ac`, switch/`yield` redesign `1643d259`, `not` vs `!` `8645a860`,
  `++`/`--` as statements `7171ec92`) → a "language shape" section of `RED-2`, or tiny REDs later.

## Considered & below the bar (not RED-worthy)

Recorded so we don't re-litigate: Apache-2.0 license `215093a0` (rationale silent); `--MOCK-RESPONSE`
`254ded49` / `Requester`; docs-strategy bits (thinking-docs practice, SYNTAX.md `959e40b0`, MkDocs
`5d2c6164`, llms.txt `212e690d`); AI dev policy `7154cb36`; CI gates (PR benchmarking `046a3c22`,
govulncheck-over-Dependabot `5caffe8a` - the latter has nice rationale if ever wanted); HTTP url-encoding
`d91350b`, underscore→hyphen arg names `e6407aee`, function hoisting `e535795d`, `confirm`/`pass`,
`in`/`not in`, `del`, per-method `http_*` `a87a3d41`, TLS-bypass `322e9a2`.

## Write order

**The chronology is the order.** Walk the master list top to bottom and write (or skip) each item
*where it sits in time*. Do **not** reorder - not by tier, not by theme. Tier decides *whether* an item
earns a RED and how deep to go, never *when* it's written. The master list is already in (roughly)
chronological order, so its sequence is the backfill sequence; when in doubt, the `released` version is
the tiebreaker. Finish one item's archaeology → interview → write/skip → mark-the-row before touching
the next.

Concretely: after `2` (genesis, v0.1.0), the next unwritten item is **`3A`** (declarative argument
parsing, v0.1.0) - *not* `AB` (tree-sitter), which is v0.5.0 and many decisions away. Don't let a
later, "foundational-feeling" architecture decision jump the queue.

**Relationships span the timeline - note them, don't reorder for them.** Several decisions form arcs
across many releases. You write each member at its own slot; when you reach the later one, its
`supersedes`/`related` frontmatter links tie it back. That linkage is the mechanism - not writing the
arc together. The main arcs to be aware of:

- Parsing / architecture: `AB` (tree-sitter, v0.5.0) → `AT` (CST→AST, v0.9.0).
- Arg parsing: `3A` (args block, v0.1) → `AAE` (constraints, v0.4→v0.5) → `AI` (Ra, v0.5.59).
- Shell + errors: `AA` (shell, v0.4) → `AAD` (early error model) → `AL` (`catch`, v0.6.12).
- Null: `3B` (removal, v0.1) → `AEA` (re-addition, v0.5.32).
- Rad-block substrate: `3D` (JSON engine) + `3F` (data model) + `AAC` (strings).
- Type system: `AH` (runtime typing, v0.5.43) → `BB` (static checker, ~v0.10).

Renumber the provisional letter IDs to integers once the chronology is fully filled in.
