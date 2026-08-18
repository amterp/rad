---
title: Release Notes
---

# Release Notes

All Rad releases. Newest first.

---

## [v0.12.0](https://github.com/amterp/rad/releases/tag/v0.12.0) - 2026-08-18

v0.12 is the biggest release yet for shell commands, and the largest breaking one so far: eight breaking changes, half of which can change what a script does without raising an error. **Read the [migration guide](https://amterp.dev/rad/migrations/v0.12/) before upgrading**, and run `rad check --from-logs all` to find the scripts that need attention.

**✨ Highlights**

**Shell commands are quoted, and readable inline.** Values you interpolate are now quoted, so a filename with a space - or a string from somewhere you don't control - arrives as one argument instead of turning into shell syntax. A command can be read where it stands:

```rad
branch = $`git branch --show-current`.stdout.trim() catch "unknown"

if not $`which docker`.ok:
    exit(1)
```

Captures now bind `(stdout, stderr, code)`, so `out = $cmd` gives you the output rather than the exit code. And a command written as a list skips the shell entirely, so it behaves the same whichever shell is installed.

**A REPL you can work in.** `rad repl` gives you multi-line blocks, auto-indent, an editable line with history, and errors that hand you back the prompt instead of ending the session. Plus `:vars`, `:docs <topic>`, `:load <file>` and `:help`. The old `--repl` flag is gone.

**Commands nest, and one can be the default.** A command block can contain command blocks, so `tool remote add` works the way `git remote add` does, with shared args accepted anywhere along the path. Mark one command per level `default` and it runs when the user names none.

**Scripts that prompt can run without a terminal.** rad reads the script up front and refuses to start until every reachable prompt has an answer, given with `--reply` (keyed by line) or `--reply-na` for branches a run won't take. Nothing executes until they are all answered, so a CI or agent run is never left half-finished. rad also looks for a terminal on `/dev/tty` now, so `echo x | rad script.rad` prompts normally.

**`split()` and `replace()` match literally.** A regex is opt-in with `regex=true`. Most calls were passing literal text and quietly getting regex behavior - a `.` matching every character rather than a dot.

**Error output fits your terminal.** Diagnostics wrap to the width you actually have, and `rad check` renders through the same code as runtime errors. `rad check` also lost a batch of false positives: hints across a 206-script corpus dropped from 15 to 4.

**🧰 New builtins**

- `get_os()` - `"macos"`, `"linux"` or `"windows"`
- `mkdir()` - recursive and idempotent, like `mkdir -p`
- `base_name()`, `dir_name()`, `join_paths()` - path helpers that work on paths that don't exist yet
- `format_epoch()` - turns an epoch into a formatted date string, the counterpart to `parse_date()`
- `len` bounds on how many values a list or variadic arg takes: `pair len [2,2]`

And plenty of bug fixes, see the list of commits below.

Don't hesitate to ask questions or raise issues!

- Mark 4 cards done after cherry-picking sprint work ([3229dd24](https://github.com/amterp/rad/commit/3229dd24))
- feat!: answer a script's prompts when there's no terminal ([00cf40bb](https://github.com/amterp/rad/commit/00cf40bb))
- feat!: list args take several positional values ([b3297b57](https://github.com/amterp/rad/commit/b3297b57))
- feat!: shell captures bind stdout first, exit code last ([070dda5c](https://github.com/amterp/rad/commit/070dda5c))
- feat!: shell commands can be read inline ([a75d948d](https://github.com/amterp/rad/commit/a75d948d))
- feat!: shell commands quote what you interpolate into them ([27280ecc](https://github.com/amterp/rad/commit/27280ecc))
- feat!: split() and replace() match literally, regex= opts in ([6411a257](https://github.com/amterp/rad/commit/6411a257))
- feat(repl): make the REPL a command instead of a global flag ([caa05f9d](https://github.com/amterp/rad/commit/caa05f9d))
- feat: a failed shell command has its own error code ([a6686f13](https://github.com/amterp/rad/commit/a6686f13))
- feat: a fatal error can unwind instead of ending the process ([fb4b52d1](https://github.com/amterp/rad/commit/fb4b52d1))
- feat: add get_os, mkdir, path helpers, and format_epoch builtins ([f69ebced](https://github.com/amterp/rad/commit/f69ebced))
- feat: dedicated error for '#' used as a comment ([5267bcc1](https://github.com/amterp/rad/commit/5267bcc1))
- feat: hint at variadic form when list args overflow positionals ([bcbbb04c](https://github.com/amterp/rad/commit/bcbbb04c))
- feat: let a command run when the user names none ([ee56a96d](https://github.com/amterp/rad/commit/ee56a96d))
- feat: let commands nest, like `git remote add` ([52dfca8b](https://github.com/amterp/rad/commit/52dfca8b))
- feat: make the REPL a session you can work in ([0f876122](https://github.com/amterp/rad/commit/0f876122))
- feat: prompt errors warn that a retry starts the script over ([b5fc1f4f](https://github.com/amterp/rad/commit/b5fc1f4f))
- feat: rad check explains what the shell syntax change did to your script ([c457ddca](https://github.com/amterp/rad/commit/c457ddca))
- feat: rad shows where a repeated prompt is called from ([2ee6fc2a](https://github.com/amterp/rad/commit/2ee6fc2a))
- feat: settle what a brace means inside a string ([fdfa395c](https://github.com/amterp/rad/commit/fdfa395c))
- fix!: constraints on a list arg are enforced, not ignored ([a29621b4](https://github.com/amterp/rad/commit/a29621b4))
- fix(check): narrow through short-circuits and `in` guards ([5ac6891c](https://github.com/amterp/rad/commit/5ac6891c))
- fix(check): stop false positives on provably-safe code ([6c7aa76b](https://github.com/amterp/rad/commit/6c7aa76b))
- fix(ci): benchmark runner survives the v0.12 shell quoting change ([678e4890](https://github.com/amterp/rad/commit/678e4890))
- fix(ci): regenerate functions.txt, force LF on embedded/fixture paths ([edd61692](https://github.com/amterp/rad/commit/edd61692))
- fix(ci): trust the homebrew tap before bumping the formula ([e0a28aec](https://github.com/amterp/rad/commit/e0a28aec))
- fix: ?? and catch fire on an error you already hold ([5d2bb6c2](https://github.com/amterp/rad/commit/5d2bb6c2))
- fix: a command that can't be started reports 127, not a Rad bug ([cce70428](https://github.com/amterp/rad/commit/cce70428))
- fix: a negative fallback no longer needs parentheses ([0a7ea3df](https://github.com/amterp/rad/commit/0a7ea3df))
- fix: a pick that settles itself no longer discards your answer ([a9e68b38](https://github.com/amterp/rad/commit/a9e68b38))
- fix: a re-raised error names where it re-fired ([108bf7cb](https://github.com/amterp/rad/commit/108bf7cb))
- fix: compound assign writes the binding it reads ([10137ab3](https://github.com/amterp/rad/commit/10137ab3))
- fix: deleting during list iteration no longer duplicates an element ([d31f5f53](https://github.com/amterp/rad/commit/d31f5f53))
- fix: fit error output to the terminal ([70dbaa11](https://github.com/amterp/rad/commit/70dbaa11))
- fix: interactive prompts encode dash answers with the =-form ([497628fd](https://github.com/amterp/rad/commit/497628fd))
- fix: list args show their type in --help again ([ebe95d0b](https://github.com/amterp/rad/commit/ebe95d0b))
- fix: on Windows, prompts draw to the console, not its input buffer ([f8644d4d](https://github.com/amterp/rad/commit/f8644d4d))
- fix: rad fmt no longer declines files with a catch block ([70da8e8c](https://github.com/amterp/rad/commit/70da8e8c))
- fix: rad no longer invents answers in its --reply suggestion ([736ccaba](https://github.com/amterp/rad/commit/736ccaba))
- fix: releasing survives scratch tags and notices a dirty tree ([89aef63a](https://github.com/amterp/rad/commit/89aef63a))
- fix: running out of --reply answers no longer names one cause ([65e05f40](https://github.com/amterp/rad/commit/65e05f40))
- fix: shrink ten diagnostics to the budget their own guide sets ([9a89c6ed](https://github.com/amterp/rad/commit/9a89c6ed))
- fix: the benchmark harness writes its README without a shell ([b82b8da5](https://github.com/amterp/rad/commit/b82b8da5))
- refactor: drop the arg branches nothing could reach ([a973e9ae](https://github.com/amterp/rad/commit/a973e9ae))
- refactor: the shell executor works from a spec, not a statement ([df4f07cf](https://github.com/amterp/rad/commit/df4f07cf))

---

## [v0.11.0](https://github.com/amterp/rad/releases/tag/v0.11.0) - 2026-07-04

- chore(check): tighten resolve types and binder issue plumbing ([176e3e91](https://github.com/amterp/rad/commit/176e3e91))
- docs(check): fix stale Hint note in typed_local snapshot ([7649595b](https://github.com/amterp/rad/commit/7649595b))
- docs(errors): sweep error_docs comments + add validation test ([62581d9d](https://github.com/amterp/rad/commit/62581d9d))
- docs(fmt): document full-convergence vision and roadmap ([e9409a0b](https://github.com/amterp/rad/commit/e9409a0b))
- docs(type-system): document scoping and name resolution ([956b6197](https://github.com/amterp/rad/commit/956b6197))
- feat!: reject inapplicable global flags on built-in commands ([613e2b82](https://github.com/amterp/rad/commit/613e2b82))
- feat(check): T? suggestion in every nullable type position ([d9a8608e](https://github.com/amterp/rad/commit/d9a8608e))
- feat(check): add Scope/Symbol/Resolved name-resolution layer ([9c788fe9](https://github.com/amterp/rad/commit/9c788fe9))
- feat(check): add bidirectional type-checker skeleton ([11b67a5a](https://github.com/amterp/rad/commit/11b67a5a))
- feat(check): add binder-emitted issues + duplicate-param diagnostic ([af4d5e17](https://github.com/amterp/rad/commit/af4d5e17))
- feat(check): add narrowing primitives + null predicate ([38187694](https://github.com/amterp/rad/commit/38187694))
- feat(check): bind for / while / list-comp scopes ([aa0e16f3](https://github.com/amterp/rad/commit/aa0e16f3))
- feat(check): bind function and lambda parameter scopes ([d324dc98](https://github.com/amterp/rad/commit/d324dc98))
- feat(check): bind switch / defer / cmd_block scopes ([b7ce7aeb](https://github.com/amterp/rad/commit/b7ce7aeb))
- feat(check): catch block narrows target to error ([d437dfe5](https://github.com/amterp/rad/commit/d437dfe5))
- feat(check): catch indexed assignment of wrong element type ([a7bcfc82](https://github.com/amterp/rad/commit/a7bcfc82))
- feat(check): closure rule drops narrowing on reassigned captures ([f276968b](https://github.com/amterp/rad/commit/f276968b))
- feat(check): collapse uninhabited tuples in union joins ([876ce872](https://github.com/amterp/rad/commit/876ce872))
- feat(check): detect unreachable switch cases (RAD40012) ([a57904b9](https://github.com/amterp/rad/commit/a57904b9))
- feat(check): diagnose lambda return body vs declared mismatch ([de973052](https://github.com/amterp/rad/commit/de973052))
- feat(check): enforce args-block and cmd-block types statically ([6a0fe53d](https://github.com/amterp/rad/commit/6a0fe53d))
- feat(check): faithful typing for fallible calls and literals ([e33fd269](https://github.com/amterp/rad/commit/e33fd269))
- feat(check): for/while loop narrowing ([45e2a214](https://github.com/amterp/rad/commit/45e2a214))
- feat(check): gate provable type mismatches as errors ([e83aa78d](https://github.com/amterp/rad/commit/e83aa78d))
- feat(check): infer hoisted fn return types with SCC ordering ([d3cb2958](https://github.com/amterp/rad/commit/d3cb2958))
- feat(check): infer lambda return type for structural matching ([997def0e](https://github.com/amterp/rad/commit/997def0e))
- feat(check): infer tuple return type for multi-value returns ([451af1d3](https://github.com/amterp/rad/commit/451af1d3))
- feat(check): not / and / or narrowing predicates ([ebe52425](https://github.com/amterp/rad/commit/ebe52425))
- feat(check): operator overload table for static expression types ([684a7e4b](https://github.com/amterp/rad/commit/684a7e4b))
- feat(check): per-arg type-check on builtin calls ([2f95f536](https://github.com/amterp/rad/commit/2f95f536))
- feat(check): plumb suggestions through static diagnostics ([3e2537d1](https://github.com/amterp/rad/commit/3e2537d1))
- feat(check): pre-plant lambda signature for recursive references ([85b5d874](https://github.com/amterp/rad/commit/85b5d874))
- feat(check): promote provable type diagnostics to Error ([070e9176](https://github.com/amterp/rad/commit/070e9176))
- feat(check): reassignment widening for while-loop bodies ([9de2d1c2](https://github.com/amterp/rad/commit/9de2d1c2))
- feat(check): refine catch/?? result types ([b2b0d6e8](https://github.com/amterp/rad/commit/b2b0d6e8))
- feat(check): reject typed re-declaration (RAD40011) ([e04715e5](https://github.com/amterp/rad/commit/e04715e5))
- feat(check): snapshot harness runs full Check() pipeline ([68a78e18](https://github.com/amterp/rad/commit/68a78e18))
- feat(check): static arity-check builtin calls ([7f4231de](https://github.com/amterp/rad/commit/7f4231de))
- feat(check): static call typing for user-defined functions ([78dc8c9a](https://github.com/amterp/rad/commit/78dc8c9a))
- feat(check): static synthesis for list and map literals ([cddc15ef](https://github.com/amterp/rad/commit/cddc15ef))
- feat(check): static undefined-identifier diagnostic with did-you-mean ([ec36cf96](https://github.com/amterp/rad/commit/ec36cf96))
- feat(check): structural matching for fn-value arguments ([f1317156](https://github.com/amterp/rad/commit/f1317156))
- feat(check): suggest '?' when 'null' appears in a type union ([0d712643](https://github.com/amterp/rad/commit/0d712643))
- feat(check): switch narrowing + exhaustiveness diagnostic ([c5a3a701](https://github.com/amterp/rad/commit/c5a3a701))
- feat(check): tighten ErrorType cascade prevention ([ce1aca96](https://github.com/amterp/rad/commit/ce1aca96))
- feat(check): truthy null-strip and in-literal-list narrowing ([33067e64](https://github.com/amterp/rad/commit/33067e64))
- feat(check): type for-loop vars from union iterables and map k/v pairs ([068f1d0e](https://github.com/amterp/rad/commit/068f1d0e))
- feat(check): type_of, string-enum, in/not-in predicates ([c1ee6e97](https://github.com/amterp/rad/commit/c1ee6e97))
- feat(check): typed local declarations end-to-end ([41d317ba](https://github.com/amterp/rad/commit/41d317ba))
- feat(check): walk lambda bodies under captured frame ([af58a63c](https://github.com/amterp/rad/commit/af58a63c))
- feat(check): wire narrowing through if/elif/else ([53da9c01](https://github.com/amterp/rad/commit/53da9c01))
- feat(check,radls): diagnose unknown typed-map keys + key completion ([69b4a848](https://github.com/amterp/rad/commit/69b4a848))
- feat(errors): retire dead parser codes + invest in three high-value ones ([e7af68c3](https://github.com/amterp/rad/commit/e7af68c3))
- feat(fmt): add 'rad fmt' canonical formatter ([4ade0bd3](https://github.com/amterp/rad/commit/4ade0bd3))
- feat(fmt): turn formatting rules into a documented catalog ([a16e46ad](https://github.com/amterp/rad/commit/a16e46ad))
- feat(lsp): resolve command callback references ([60fdaa50](https://github.com/amterp/rad/commit/60fdaa50))
- feat(radls): UFCS-aware completion ranking ([bf1a4164](https://github.com/amterp/rad/commit/bf1a4164))
- feat(radls): add textDocument/definition ([3153632b](https://github.com/amterp/rad/commit/3153632b))
- feat(radls): add textDocument/documentSymbol ([b4108828](https://github.com/amterp/rad/commit/b4108828))
- feat(radls): add textDocument/hover ([6241558e](https://github.com/amterp/rad/commit/6241558e))
- feat(radls): add textDocument/references ([df4e1c15](https://github.com/amterp/rad/commit/df4e1c15))
- feat(radls): add textDocument/semanticTokens/full ([b977872f](https://github.com/amterp/rad/commit/b977872f))
- feat(radls): code actions surface diagnostic quick fixes ([8cb11d10](https://github.com/amterp/rad/commit/8cb11d10))
- feat(radls): debounce diagnostics publishing per URI ([29446ba5](https://github.com/amterp/rad/commit/29446ba5))
- feat(radls): filter internal builtins + parameter token for args ([8aaaeb91](https://github.com/amterp/rad/commit/8aaaeb91))
- feat(radls): immutable DocumentVersion snapshots per document ([cd70d467](https://github.com/amterp/rad/commit/cd70d467))
- feat(radls): introduce FileID indirection for documents ([492c0dd3](https://github.com/amterp/rad/commit/492c0dd3))
- feat(radls): negotiate LSP position encoding via LineIndex ([65a230eb](https://github.com/amterp/rad/commit/65a230eb))
- feat(radls): plumb context.Context and wire $/cancelRequest ([2e0b2326](https://github.com/amterp/rad/commit/2e0b2326))
- feat(radls): real completion - scope-walk + type-aware ([e4b20229](https://github.com/amterp/rad/commit/e4b20229))
- feat(radls): semantic tokens for type annotations ([f9ee5c2b](https://github.com/amterp/rad/commit/f9ee5c2b))
- feat(radls): structured docs in builtin hover + funcs doc infrastructure ([b2ce132d](https://github.com/amterp/rad/commit/b2ce132d))
- feat(radls): textDocument/rename ([73965325](https://github.com/amterp/rad/commit/73965325))
- feat(types): add Dynamic (implicit any) distinct from explicit Any ([4363b5c9](https://github.com/amterp/rad/commit/4363b5c9))
- feat(types): add ErrorType poison for static-check cascade prevention ([640f2137](https://github.com/amterp/rad/commit/640f2137))
- feat(types): add IsAssignableFrom to TypingT interface ([35208d96](https://github.com/amterp/rad/commit/35208d96))
- feat(types): add Never bottom type for narrowing exhaustiveness ([80cfc241](https://github.com/amterp/rad/commit/80cfc241))
- feat(types): parens type groups + inline-colon lambdas + drop list[T] ([b3dba6bd](https://github.com/amterp/rad/commit/b3dba6bd))
- feat(typing): introduce TypingNullT for sound null literal handling ([d4ade013](https://github.com/amterp/rad/commit/d4ade013))
- feat: add get_pid() builtin ([a446ae82](https://github.com/amterp/rad/commit/a446ae82))
- feat: add signal_trap and signal_ignore builtins ([dc047f60](https://github.com/amterp/rad/commit/dc047f60))
- feat: add to_json() builtin ([5f80b7a2](https://github.com/amterp/rad/commit/5f80b7a2))
- feat: add weekday field to time maps ([05a233e7](https://github.com/amterp/rad/commit/05a233e7))
- feat: back pick() with radish instead of huh ([1834da2f](https://github.com/amterp/rad/commit/1834da2f))
- feat: clean and navigate the rad docs corpus ([a2b47a3d](https://github.com/amterp/rad/commit/a2b47a3d))
- feat: colorize the -i transcript and equivalent invocation ([8be4d28e](https://github.com/amterp/rad/commit/8be4d28e))
- feat: docs/funcs/ as source-of-truth for builtins + diagnostic sweep ([4c74002f](https://github.com/amterp/rad/commit/4c74002f))
- feat: format the args block in rad fmt ([841d4e29](https://github.com/amterp/rad/commit/841d4e29))
- feat: interactive arg-entry mode (--interactive / -i) ([f3055f81](https://github.com/amterp/rad/commit/f3055f81))
- feat: make signal force-exit testable and shutdown deterministic ([8b35833f](https://github.com/amterp/rad/commit/8b35833f))
- feat: parse_duration accepts 'w' (weeks) suffix ([3a54604e](https://github.com/amterp/rad/commit/3a54604e))
- feat: readable -i transcripts and grouped bool multi-pick ([6c3040d9](https://github.com/amterp/rad/commit/6c3040d9))
- feat: replace huh with radish for input/confirm/multipick ([ed21d6ee](https://github.com/amterp/rad/commit/ed21d6ee))
- feat: run defer blocks on SIGINT/SIGTERM exit ([b4ad6b0f](https://github.com/amterp/rad/commit/b4ad6b0f))
- feat: serve documentation from the binary via rad docs ([475af15a](https://github.com/amterp/rad/commit/475af15a))
- fix(benchmark): migrate off removed $! shell syntax ([dae8dfa0](https://github.com/amterp/rad/commit/dae8dfa0))
- fix(check): accept string literals matching a StrEnum param ([126df537](https://github.com/amterp/rad/commit/126df537))
- fix(check): bind scopes only at function boundaries ([4b298120](https://github.com/amterp/rad/commit/4b298120))
- fix(check): consult builtin set for cmd-callback resolution ([8eadc6da](https://github.com/amterp/rad/commit/8eadc6da))
- fix(check): deep divergence scan for branch joins ([a928822b](https://github.com/amterp/rad/commit/a928822b))
- fix(check): give dynamic union arms bare-dynamic leniency in operators ([dc7bc20a](https://github.com/amterp/rad/commit/dc7bc20a))
- fix(check): isolate nested function bodies from outer narrowing ([35d6990a](https://github.com/amterp/rad/commit/35d6990a))
- fix(check): keep uncertain type mismatches non-blocking ([95890d84](https://github.com/amterp/rad/commit/95890d84))
- fix(check): narrowByTypeOf preserves null arm of Optional ([957271a0](https://github.com/amterp/rad/commit/957271a0))
- fix(check): null soundness in unions + promote bare-null mismatch ([fca8db56](https://github.com/amterp/rad/commit/fca8db56))
- fix(check): stop false positives on valid scripts (#129-131, #133) ([9d986d86](https://github.com/amterp/rad/commit/9d986d86))
- fix(check): structural union subsumption in joinFrames ([fa4872f2](https://github.com/amterp/rad/commit/fa4872f2))
- fix(check): switchDiverges read ExprTypes, don't re-enter synth ([57eb3212](https://github.com/amterp/rad/commit/57eb3212))
- fix(check): while loop carries body frame mutations forward ([cb769aca](https://github.com/amterp/rad/commit/cb769aca))
- fix(check): widen while-loop narrowing on nested reassignments ([7e9ca521](https://github.com/amterp/rad/commit/7e9ca521))
- fix(check, radls): name-precision DeclSpans for params/loop vars ([12a4c7bc](https://github.com/amterp/rad/commit/12a4c7bc))
- fix(check, radls): point fn DeclSpan + SelectionRange at name ([39c90b68](https://github.com/amterp/rad/commit/39c90b68))
- fix(radls): Mux.Notify - drop dead branch, omit nil params ([d85f6302](https://github.com/amterp/rad/commit/d85f6302))
- fix(radls): correct inverted nil check in newError ([d3233ff8](https://github.com/amterp/rad/commit/d3233ff8))
- fix(radls): debouncer race - superseded callback no longer clobbers successor ([b62b89d5](https://github.com/amterp/rad/commit/b62b89d5))
- fix(radls): dispatch quick fixes on error code, not message text ([5a0fac03](https://github.com/amterp/rad/commit/5a0fac03))
- fix(radls): emit semantic token at fn name definition site ([d90a3dee](https://github.com/amterp/rad/commit/d90a3dee))
- fix(radls): file-scope completion respects cursor position ([a0035300](https://github.com/amterp/rad/commit/a0035300))
- fix(radls): find references / goto-def from fn / args decl sites ([d174a1eb](https://github.com/amterp/rad/commit/d174a1eb))
- fix(radls): hover - nil for unresolved, drop redundant builtin tag ([9c587ee1](https://github.com/amterp/rad/commit/9c587ee1))
- fix(radls): refcount DocumentVersion to free tree-sitter C memory ([f7edd7ef](https://github.com/amterp/rad/commit/f7edd7ef))
- fix(radls): release per-request context cancel funcs ([be4c7da0](https://github.com/amterp/rad/commit/be4c7da0))
- fix(radls): scope-proximity completion order via SortText ([5d0f42b4](https://github.com/amterp/rad/commit/5d0f42b4))
- fix(radls): sort document symbols in source order ([4747f1fd](https://github.com/amterp/rad/commit/4747f1fd))
- fix(radls): suppress info-only code actions from the lightbulb ([3be5ad76](https://github.com/amterp/rad/commit/3be5ad76))
- fix(shell): don't crash when a confirm prompt is aborted ([fb92f6f5](https://github.com/amterp/rad/commit/fb92f6f5))
- fix(signatures): complete type_of return enum ([990e9261](https://github.com/amterp/rad/commit/990e9261))
- fix(types): close silent runtime enforcement gaps ([991e30a7](https://github.com/amterp/rad/commit/991e30a7))
- fix: catch handlers losing caller scope on nested error propagation ([a60a9b7e](https://github.com/amterp/rad/commit/a60a9b7e))
- fix: drop empty Examples markers from generated function docs ([f607b008](https://github.com/amterp/rad/commit/f607b008))
- fix: drop stray "=" artifact in multipick docs ([200d43ba](https://github.com/amterp/rad/commit/200d43ba))
- fix: expand leading ~ in path builtins ([226277ff](https://github.com/amterp/rad/commit/226277ff))
- fix: explain under-indented multi-line string errors ([8416607d](https://github.com/amterp/rad/commit/8416607d))
- fix: harden -i argv synthesis and align walk with ra semantics ([5ae73437](https://github.com/amterp/rad/commit/5ae73437))
- fix: keep go:embed'd doc corpora LF on Windows checkouts ([2a83f61a](https://github.com/amterp/rad/commit/2a83f61a))
- fix: make --shell reserve stdout for eval-able output on help ([6f9e554f](https://github.com/amterp/rad/commit/6f9e554f))
- fix: normalize docs/funcs source to LF to match its embedded copy ([bd6b1925](https://github.com/amterp/rad/commit/bd6b1925))
- fix: preserve for-loop with-clause in rad fmt ([d2393a0d](https://github.com/amterp/rad/commit/d2393a0d))
- fix: report out-of-range int literals instead of panicking ([8b0f39ce](https://github.com/amterp/rad/commit/8b0f39ce))
- fix: unify any[], dynamic[], and list into one type ([686be9d4](https://github.com/amterp/rad/commit/686be9d4))
- fix: write state and stash files atomically ([c7ba354a](https://github.com/amterp/rad/commit/c7ba354a))
- refactor(check): emit fn/arg shadowing from the binder ([72160d0a](https://github.com/amterp/rad/commit/72160d0a))
- refactor(check): route cmd-callback hint through resolved view ([414f3e2d](https://github.com/amterp/rad/commit/414f3e2d))
- refactor(radls): bind encoding into DocumentVersion ([6fe1639d](https://github.com/amterp/rad/commit/6fe1639d))
- refactor(radls): consolidate AST walk + span helpers ([d460e822](https://github.com/amterp/rad/commit/d460e822))
- refactor(radls): handlers grab snapshots at the request boundary ([32d7f371](https://github.com/amterp/rad/commit/32d7f371))
- refactor(rts): RadTree holds language not parser - close cgo trap ([03dcc61b](https://github.com/amterp/rad/commit/03dcc61b))
- refactor: adopt radish's cleaner pick result contract ([8ae4efd5](https://github.com/amterp/rad/commit/8ae4efd5))
- style(check): fix shell-escape artifacts in comments ([fe92034f](https://github.com/amterp/rad/commit/fe92034f))
- style: reformat code ([3319e1e7](https://github.com/amterp/rad/commit/3319e1e7))
- test(check): comprehensive snapshot coverage for narrowing ([c6082908](https://github.com/amterp/rad/commit/c6082908))
- test(check): snapshot harness for narrowing ([677b79ac](https://github.com/amterp/rad/commit/677b79ac))
- test(radls): differential harness for incremental vs from-scratch ([535a5647](https://github.com/amterp/rad/commit/535a5647))
- test(radls): extend lstesting harness with 5 new action types ([816efe54](https://github.com/amterp/rad/commit/816efe54))
- test(radls): port documentSymbol cases to wire-level snapshot ([2952d3e8](https://github.com/amterp/rad/commit/2952d3e8))
- test(radls): port findReferences cases to wire-level snapshot ([ae57e989](https://github.com/amterp/rad/commit/ae57e989))
- test(radls): port goto-definition cases to wire-level snapshot ([ca0310e6](https://github.com/amterp/rad/commit/ca0310e6))
- test(radls): port hover behavior cases to wire-level snapshot ([27ad1b5c](https://github.com/amterp/rad/commit/27ad1b5c))
- test(radls): port semanticTokens cases to wire-level snapshot ([b3ee77c4](https://github.com/amterp/rad/commit/b3ee77c4))
- test(radls): trim completion + code-action Go tests ([e7b31a2d](https://github.com/amterp/rad/commit/e7b31a2d))
- ux(radls, check): contextual rename errors + did-you-mean grammar + hover stutter ([552f59dc](https://github.com/amterp/rad/commit/552f59dc))

---

## [v0.10.1](https://github.com/amterp/rad/releases/tag/v0.10.1) - 2026-05-07

- feat(shell): auto-detect shell on Windows ([cd04129e](https://github.com/amterp/rad/commit/cd04129e))

---

## [v0.10.0](https://github.com/amterp/rad/releases/tag/v0.10.0) - 2026-04-19

- ci: replace Dependabot with govulncheck for vulnerability scanning ([5caffe8a](https://github.com/amterp/rad/commit/5caffe8a))
- feat: add 'transpose [expr]' directive to rad blocks ([e454be5b](https://github.com/amterp/rad/commit/e454be5b))
- feat: add dim() text styling function ([fb6669b2](https://github.com/amterp/rad/commit/fb6669b2))
- feat: add shell autocompletion for rad CLI and scripts ([ff29f1a4](https://github.com/amterp/rad/commit/ff29f1a4))
- feat: add strikethrough() text styling function ([d8d6c03d](https://github.com/amterp/rad/commit/d8d6c03d))
- feat: improve transpose behavior, support color modifies ([b745339b](https://github.com/amterp/rad/commit/b745339b))
- fix: reject style attrs in rad block color modifiers ([e7c7a307](https://github.com/amterp/rad/commit/e7c7a307))
- style: fix up formatting ([48da2ca2](https://github.com/amterp/rad/commit/48da2ca2))

---

## [v0.9.2](https://github.com/amterp/rad/releases/tag/v0.9.2) - 2026-04-04

- feat: add end-to-end snapshot testing for radls ([52791c9](https://github.com/amterp/rad/commit/52791c9))
- feat: add parse_date() built-in function ([3aaff8f](https://github.com/amterp/rad/commit/3aaff8f))
- fix: correctly capture defining scope in escaping closures ([5840c0f](https://github.com/amterp/rad/commit/5840c0f))
- fix: emit clean error when void value used in expression ([a191108](https://github.com/amterp/rad/commit/a191108))
- fix: preserve float precision in alignment formatting ([b7b910c](https://github.com/amterp/rad/commit/b7b910c))
- fix: prevent cross-type map key collisions ([c209834](https://github.com/amterp/rad/commit/c209834))
- fix: produce clean error for null used as map key ([8dd2a14](https://github.com/amterp/rad/commit/8dd2a14))
- fix: return error when int() is called with Inf or NaN ([f633747](https://github.com/amterp/rad/commit/f633747))
- fix: support map in list for the in operator ([8c59128](https://github.com/amterp/rad/commit/8c59128))
- fix: support string comparison operators (< > <= >=) ([9a05633](https://github.com/amterp/rad/commit/9a05633))
- fix: swap operands for list-in-list 'in' operator ([3d80d0f](https://github.com/amterp/rad/commit/3d80d0f))
- style: reformat code ([47c1afe](https://github.com/amterp/rad/commit/47c1afe))

---

## [v0.9.1](https://github.com/amterp/rad/releases/tag/v0.9.1) - 2026-04-02

- Add llms.txt and llms-full.txt generation to docs site ([212e690](https://github.com/amterp/rad/commit/212e690))
- Expand args docs with parsing model and edge cases ([5d82de7](https://github.com/amterp/rad/commit/5d82de7))
- Release VSCode extension v0.4.0 ([7205d79](https://github.com/amterp/rad/commit/7205d79))
- Release VSCode extension v0.5.0 ([682c36c](https://github.com/amterp/rad/commit/682c36c))
- fix: use pipe delimiter in sed to avoid clash with URL slashes ([9e1794e](https://github.com/amterp/rad/commit/9e1794e))
- refactor: rename lsp-server/ to radls/, update URLs to amterp.dev ([3df330f](https://github.com/amterp/rad/commit/3df330f))

---

## [v0.9.0](https://github.com/amterp/rad/releases/tag/v0.9.0) - 2026-03-11

v0.9 is here!

Please note this version contains a higher number of breaking changes than usual.

I am trying to get them out of the way sooner rather than later, to minimize the overall pain.

I've done my best to make migrating really clear, including having error detection and clear messages that try and detect when you're doing something "the old ways".

See https://amterp.dev/rad/migrations/v0.9/ for details about the breaking changes.

- feat!: broaden ?? to null-coalesce, catchable null indexing ([4fa150c](https://github.com/amterp/rad/commit/4fa150c))
- feat!: make + operator strict about type matching ([1ec9e5c](https://github.com/amterp/rad/commit/1ec9e5c))
- feat!: shorten parse_epoch unit parameter names ([2a09c43](https://github.com/amterp/rad/commit/2a09c43))
- feat!: stop normalizing user data, add split_lines() ([14a4b2c](https://github.com/amterp/rad/commit/14a4b2c))
- feat!: unify request/display into single rad keyword ([3d09b27](https://github.com/amterp/rad/commit/3d09b27))
- feat(funcs): make numeric functions type-preserving ([18209b7](https://github.com/amterp/rad/commit/18209b7))
- feat: accept Enter as confirmation in confirm() ([52655ab](https://github.com/amterp/rad/commit/52655ab))
- feat: add "did you mean?" suggestions for undefined vars ([2a09807](https://github.com/amterp/rad/commit/2a09807))
- feat: add 'rad explain list' to show available error codes ([4b6b4c4](https://github.com/amterp/rad/commit/4b6b4c4))
- feat: add --ast-tree and --cst-tree debug flags ([f3c3ba8](https://github.com/amterp/rad/commit/f3c3ba8))
- feat: add CST-to-AST converter ([ec3185e](https://github.com/amterp/rad/commit/ec3185e))
- feat: add Rust-style diagnostic renderer ([e2e06d4](https://github.com/amterp/rad/commit/e2e06d4))
- feat: add TLS certificate verification bypass ([322e9a2](https://github.com/amterp/rad/commit/322e9a2))
- feat: add call stack tracking and stack traces ([338ffe3](https://github.com/amterp/rad/commit/338ffe3))
- feat: add catch inline operator for error-only catching ([a7878a4](https://github.com/amterp/rad/commit/a7878a4))
- feat: add error testing framework and documentation ([5f4ee17](https://github.com/amterp/rad/commit/5f4ee17))
- feat: add fill character and zero-pad to format specs ([5d5a851](https://github.com/amterp/rad/commit/5d5a851))
- feat: add heuristics for ERROR node detection ([9b4d2e8](https://github.com/amterp/rad/commit/9b4d2e8))
- feat: add index_of() built-in for strings and lists ([b33748f](https://github.com/amterp/rad/commit/b33748f))
- feat: add inline migration hint for + operator type error ([7372764](https://github.com/amterp/rad/commit/7372764))
- feat: add limit param to split() ([539e566](https://github.com/amterp/rad/commit/539e566))
- feat: add parse_duration and convert_duration functions ([cef9648](https://github.com/amterp/rad/commit/cef9648))
- feat: add rad explain command for error documentation ([2f599d6](https://github.com/amterp/rad/commit/2f599d6))
- feat: add semantic grammar checks for control flow ([790c615](https://github.com/amterp/rad/commit/790c615))
- feat: add source span tracking to RadValue ([f71e0dd](https://github.com/amterp/rad/commit/f71e0dd))
- feat: add symbol table for tracking variable definitions ([95ce0ff](https://github.com/amterp/rad/commit/95ce0ff))
- feat: add trim_left and trim_right functions ([a7225df](https://github.com/amterp/rad/commit/a7225df))
- feat: add unified diagnostic type system ([7458ca7](https://github.com/amterp/rad/commit/7458ca7))
- feat: define AST node types, interfaces, and enums ([c518d3e](https://github.com/amterp/rad/commit/c518d3e))
- feat: enable invocation logging by default ([b7b95e2](https://github.com/amterp/rad/commit/b7b95e2))
- feat: expand MISSING node messages with parent context ([930212f](https://github.com/amterp/rad/commit/930212f))
- feat: extract shared duration parser with day support ([76983e2](https://github.com/amterp/rad/commit/76983e2))
- feat: render markdown error docs with terminal styling ([15e4995](https://github.com/amterp/rad/commit/15e4995))
- feat: show [command] instead of [subcommand] in help ([8861a11](https://github.com/amterp/rad/commit/8861a11))
- feat: show per-severity diagnostic counts in check --from-logs ([c412221](https://github.com/amterp/rad/commit/c412221))
- fix!: align trim_prefix/trim_suffix semantics with their names ([0f02eb1](https://github.com/amterp/rad/commit/0f02eb1))
- fix(deps): resolve Dependabot security alerts ([4c19a81](https://github.com/amterp/rad/commit/4c19a81))
- fix: AST dump span padding now walks full tree ([3f514c9](https://github.com/amterp/rad/commit/3f514c9))
- fix: Levenshtein distance operates on bytes instead of runes ([e6ac0a1](https://github.com/amterp/rad/commit/e6ac0a1))
- fix: TruthyFalsy crashes on RadError values ([6837175](https://github.com/amterp/rad/commit/6837175))
- fix: add missing error codes to validation errors ([bb9637d](https://github.com/amterp/rad/commit/bb9637d))
- fix: bugs found in code review of AST foundation ([3c4e5ad](https://github.com/amterp/rad/commit/3c4e5ad))
- fix: bump ra to v0.5.1 ([1a97b79](https://github.com/amterp/rad/commit/1a97b79))
- fix: clone target node in compound assign/incr-decr desugaring ([12c49b6](https://github.com/amterp/rad/commit/12c49b6))
- fix: derive gutter width from displayed lines, not labels ([21c218a](https://github.com/amterp/rad/commit/21c218a))
- fix: diagnostic panics and lambda file name ([b245ca0](https://github.com/amterp/rad/commit/b245ca0))
- fix: diagnostic renderer byte-slices UTF-8 in truncation ([ea38df9](https://github.com/amterp/rad/commit/ea38df9))
- fix: flaky truncate tests due to shared TerminalIsUtf8 state ([d7e964a](https://github.com/amterp/rad/commit/d7e964a))
- fix: include response headers in http_* return value ([902e00f](https://github.com/amterp/rad/commit/902e00f))
- fix: minor comment adjustments ([3870211](https://github.com/amterp/rad/commit/3870211))
- fix: nil node panic in error emission functions ([d81c49a](https://github.com/amterp/rad/commit/d81c49a))
- fix: nil node safety and same-line label coloring ([8f2dac0](https://github.com/amterp/rad/commit/8f2dac0))
- fix: panic on short non-JSON response body in RequestJson ([f3d2a4c](https://github.com/amterp/rad/commit/f3d2a4c))
- fix: preserve text attributes in trim, reverse, slice ([39bed25](https://github.com/amterp/rad/commit/39bed25))
- fix: prevent hang on non-regular files like /dev/stdin ([66dcf7b](https://github.com/amterp/rad/commit/66dcf7b))
- fix: release notes workflow inserting into front matter ([6093f5f](https://github.com/amterp/rad/commit/6093f5f))
- fix: remove invalid --latest flag from gh release view ([84382dc](https://github.com/amterp/rad/commit/84382dc))
- fix: shell catch nil pointer dereference ([abce025](https://github.com/amterp/rad/commit/abce025))
- fix: split_lines strips trailing empty element ([9f3dbb2](https://github.com/amterp/rad/commit/9f3dbb2))
- fix: string slicing now correctly handles multi-byte characters ([a0af29c](https://github.com/amterp/rad/commit/a0af29c))
- fix: truncate() now handles multi-byte characters correctly ([f62848d](https://github.com/amterp/rad/commit/f62848d))
- fix: unify equality semantics across == and Equals() ([27dc974](https://github.com/amterp/rad/commit/27dc974))
- fix: update go mod, add more details to AGENTS.md re: migrations ([da5efea](https://github.com/amterp/rad/commit/da5efea))
- fix: use pure Go DNS resolver for Linux static builds ([622026c](https://github.com/amterp/rad/commit/622026c))
- perf: fix IndexAt() to extract runes from segment directly ([d082015](https://github.com/amterp/rad/commit/d082015))
- perf: fix ToRuneList() O(n^2) complexity ([5fbaed2](https://github.com/amterp/rad/commit/5fbaed2))
- perf: short-circuit Plain() for single-segment strings ([ea6c560](https://github.com/amterp/rad/commit/ea6c560))
- perf: use strings.Builder in Plain() and Reverse() ([273beb2](https://github.com/amterp/rad/commit/273beb2))
- refactor: clarify diagnostic type hierarchy ([3136be7](https://github.com/amterp/rad/commit/3136be7))
- refactor: complete AST migration phases 5-8 ([a258ce7](https://github.com/amterp/rad/commit/a258ce7))
- refactor: migrate all errorf calls to new diagnostic system ([a0a26f3](https://github.com/amterp/rad/commit/a0a26f3))
- refactor: migrate checker from CST to AST ([14c797d](https://github.com/amterp/rad/commit/14c797d))
- refactor: migrate checker to AST, remove dead code ([9fcb675](https://github.com/amterp/rad/commit/9fcb675))
- refactor: migrate interpreter from CST to AST ([31eb894](https://github.com/amterp/rad/commit/31eb894))
- refactor: migrate metadata extraction to AST ([531a112](https://github.com/amterp/rad/commit/531a112))
- refactor: move Span type from core/ to rts/rl/ ([47acb87](https://github.com/amterp/rad/commit/47acb87))
- refactor: remove dead code from interpreter ([7f93506](https://github.com/amterp/rad/commit/7f93506))
- refactor: remove unused SymbolTable, keep Levenshtein ([de5cdc7](https://github.com/amterp/rad/commit/de5cdc7))
- refactor: seal CST from core/, add Children() to AST ([2d3b92d](https://github.com/amterp/rad/commit/2d3b92d))
- refactor: seal CST leaks from public interfaces ([0dc6c61](https://github.com/amterp/rad/commit/0dc6c61))
- refactor: simplify evalCatchingPanic, document error contract ([9ddc1bb](https://github.com/amterp/rad/commit/9ddc1bb))
- refactor: unify dump tests into Go-based CST snapshot system ([7658ed3](https://github.com/amterp/rad/commit/7658ed3))
- refactor: unify error output to Rust-style format ([ccb3bec](https://github.com/amterp/rad/commit/ccb3bec))
- refactor: wire LSP to use AST-aware checker ([b820903](https://github.com/amterp/rad/commit/b820903))
- rename get_stash_dir to get_stash_path ([b8ed277](https://github.com/amterp/rad/commit/b8ed277))
- revert: remove RadValue span tracking (perf fix) ([162eb5a](https://github.com/amterp/rad/commit/162eb5a))
- style: minor adjustments ([83a7ac9](https://github.com/amterp/rad/commit/83a7ac9))
- style: reformat ([5328143](https://github.com/amterp/rad/commit/5328143))
- tests: improve snapshot failure diff output ([9f0e414](https://github.com/amterp/rad/commit/9f0e414))
- tests: migrate common input->output assert tests to snapshot tests ([9d57d24](https://github.com/amterp/rad/commit/9d57d24))

---

## [v0.8.1](https://github.com/amterp/rad/releases/tag/v0.8.1) - 2026-03-08

- fix: use pure Go DNS resolver for Linux static builds ([622026ca](https://github.com/amterp/rad/commit/622026ca))

---

## [v0.8.0](https://github.com/amterp/rad/releases/tag/v0.8.0) - 2026-01-29

- feat!: remove get_default function in favor of ?? operator ([7bbad61](https://github.com/amterp/rad/commit/7bbad61))
- feat(docs): redesign website with "Sunset Terminal" theme ([79b0103](https://github.com/amterp/rad/commit/79b0103))
- feat: add ?? fallback support for list and string indexing ([7350604](https://github.com/amterp/rad/commit/7350604))
- feat: add list support to reverse function ([d13e6d7](https://github.com/amterp/rad/commit/d13e6d7))
- feat: hyphenate command names for CLI invocation ([c8373a2](https://github.com/amterp/rad/commit/c8373a2))
- fix: use workflow_run trigger for release notes workflow ([7c450bf](https://github.com/amterp/rad/commit/7c450bf))

---

## [v0.7.1](https://github.com/amterp/rad/releases/tag/v0.7.1) - 2026-01-26

- Release VSCode extension v0.3.0 ([9443d85](https://github.com/amterp/rad/commit/9443d85))
- ci: add automated docs deployment and release notes ([883a593](https://github.com/amterp/rad/commit/883a593))
- ci: add cross-platform testing for macOS and Windows ([e085011](https://github.com/amterp/rad/commit/e085011))
- ci: fix Windows build output filename in cross-platform tests ([4474ee3](https://github.com/amterp/rad/commit/4474ee3))
- ci: treat all platforms as equal first-class citizens ([668a509](https://github.com/amterp/rad/commit/668a509))
- feat: add platform abstraction for Windows compatibility ([5fdf8fd](https://github.com/amterp/rad/commit/5fdf8fd))
- feat: add underscore to v0.7 for-loop migration hint detection ([cf4af1a](https://github.com/amterp/rad/commit/cf4af1a))
- feat: complete platform normalization for Windows compatibility ([83c839e](https://github.com/amterp/rad/commit/83c839e))
- fix: improve gen_fid collision resistance ([0ee37ff](https://github.com/amterp/rad/commit/0ee37ff))

---

## [v0.7.0](https://github.com/amterp/rad/releases/tag/v0.7.0) - 2026-01-16

- docs(SYNTAX.md): add script commands ([9e4e595](https://github.com/amterp/rad/commit/9e4e595))
- feat!: redesign for-loop syntax with explicit context access ([e98f7fc](https://github.com/amterp/rad/commit/e98f7fc))
  * See [v0.7 migration guide](https://amterp.dev/rad/migrations/v0.7/) for more information.
- feat: add context support to rad block map/filter lambdas ([21976ec](https://github.com/amterp/rad/commit/21976ec))


---

## [v0.6.27](https://github.com/amterp/rad/releases/tag/v0.6.27) - 2026-01-11

- Release VSCode extension v0.2.0 ([80c4102](https://github.com/amterp/rad/commit/80c4102))
- fix: multiline string interpolation followed by content ([5365366](https://github.com/amterp/rad/commit/5365366))

---

## [v0.6.26](https://github.com/amterp/rad/releases/tag/v0.6.26) - 2026-01-02

- ci: optimize ci benchmarks by interleaving & reducing runs ([3a4a7f4](https://github.com/amterp/rad/commit/3a4a7f4))
- feat(pick): implement 'prefer_exact' named arg ([a3092c7](https://github.com/amterp/rad/commit/a3092c7))
- feat: improve syntax error messages with specific diagnostics ([82d9472](https://github.com/amterp/rad/commit/82d9472))
- feat: unpackage radls from vs code extension, use PATH instead ([d79186f](https://github.com/amterp/rad/commit/d79186f))

---

## [v0.6.25](https://github.com/amterp/rad/releases/tag/v0.6.25) - 2025-12-16

- Release VSCode extension v0.1.13 ([c492a49](https://github.com/amterp/rad/commit/c492a49))
- docs(thinking): Rewrite imports.md and add some quick counter thoughts ([f50a960](https://github.com/amterp/rad/commit/f50a960))
- feat: add additional get_path fields e.g. modified_millis ([f536f12](https://github.com/amterp/rad/commit/f536f12))
- feat: allow parse_epoch to accept floats ([c15def5](https://github.com/amterp/rad/commit/c15def5))
- fix: avoid Go 'MISSING' formatting malformed print in shell cmds ([6a69141](https://github.com/amterp/rad/commit/6a69141))

---

## [v0.6.24](https://github.com/amterp/rad/releases/tag/v0.6.24) - 2025-12-14

- feat: add flat_map function ([0f72798](https://github.com/amterp/rad/commit/0f72798))

---

## [v0.6.23](https://github.com/amterp/rad/releases/tag/v0.6.23) - 2025-12-10

- feat: allow min/max to accept var args of numbers ([fb6582d](https://github.com/amterp/rad/commit/fb6582d))


---

## [v0.6.21](https://github.com/amterp/rad/releases/tag/v0.6.21) - 2025-12-03

- ci: use full brew hash ([4dfb9ed](https://github.com/amterp/rad/commit/4dfb9ed))

---

## [v0.6.20](https://github.com/amterp/rad/releases/tag/v0.6.20) - 2025-12-03

- ci: attempt brew CI fix by pulling latest commit ([775afad](https://github.com/amterp/rad/commit/775afad))

---

## [v0.6.19](https://github.com/amterp/rad/releases/tag/v0.6.19) - 2025-12-02

- ci: attempt to fix homebrew release ([da006f4](https://github.com/amterp/rad/commit/da006f4))

---

## [v0.6.18](https://github.com/amterp/rad/releases/tag/v0.6.18) - 2025-11-29

- feat: implement 'filter' rad block field modifier syntax ([022267b](https://github.com/amterp/rad/commit/022267b))
- fix: un-indent command descriptions ([5f1480e](https://github.com/amterp/rad/commit/5f1480e))

---

## [v0.6.17](https://github.com/amterp/rad/releases/tag/v0.6.17) - 2025-11-20

- Release VSCode extension v0.1.12 ([d70dbbe](https://github.com/amterp/rad/commit/d70dbbe))
- feat(checker): add warning against undefined cmd callback references ([868d6f8](https://github.com/amterp/rad/commit/868d6f8))
- feat: add function 'multipick' ([8c96666](https://github.com/amterp/rad/commit/8c96666))
- fix: delete pointless multipick tests ([6a0ebba](https://github.com/amterp/rad/commit/6a0ebba))

---

## [v0.6.16](https://github.com/amterp/rad/releases/tag/v0.6.16) - 2025-11-19

- Implement commands in Rad ([cef9661](https://github.com/amterp/rad/commit/cef9661))

---

## [v0.6.15](https://github.com/amterp/rad/releases/tag/v0.6.15) - 2025-11-09

- Release VSCode extension v0.1.11 ([1c6bb65](https://github.com/amterp/rad/commit/1c6bb65))
- feat: recognize hoisted functions for unknown functions check ([e535795](https://github.com/amterp/rad/commit/e535795))
- fix: allow --version --src and --src-tree on invalid scripts ([4220c18](https://github.com/amterp/rad/commit/4220c18))
- fix: avoid passing arg-less prints through printf ([7e16b1a](https://github.com/amterp/rad/commit/7e16b1a))
- fix: fix x/crypto dependency ([5bb2ea7](https://github.com/amterp/rad/commit/5bb2ea7))

---

## [v0.6.14](https://github.com/amterp/rad/releases/tag/v0.6.14) - 2025-11-04

- feat(check): add error check against hoisted functions shadowing args ([ff839e6](https://github.com/amterp/rad/commit/ff839e6))
- feat: allow non-str types in colorize function ([ba71c19](https://github.com/amterp/rad/commit/ba71c19))

---

## [v0.6.13](https://github.com/amterp/rad/releases/tag/v0.6.13) - 2025-11-03

- Release VSCode extension v0.1.10 ([e27b21c](https://github.com/amterp/rad/commit/e27b21c))
- ci: fix test-runner.rad syntax ([18c9bad](https://github.com/amterp/rad/commit/18c9bad))
- docs(guide): add complete example to getting started ([c4df48a](https://github.com/amterp/rad/commit/c4df48a))
- docs(guide): add error-handling.md ([4a7f9f5](https://github.com/amterp/rad/commit/4a7f9f5))
- docs(guide): add type-annotations.md ([20fe398](https://github.com/amterp/rad/commit/20fe398))
- docs(guide): shorten shell-commands.md header ([d704c25](https://github.com/amterp/rad/commit/d704c25))
- docs(guide): update args.md ([ab615c6](https://github.com/amterp/rad/commit/ab615c6))
- docs(guide): update basics.md ([0a38d12](https://github.com/amterp/rad/commit/0a38d12))
- docs(guide): update functions.md ([67b5fdc](https://github.com/amterp/rad/commit/67b5fdc))
- docs(guide): update getting started ([7974f4c](https://github.com/amterp/rad/commit/7974f4c))
- docs(guide): update rad-blocks.md ([5c58d83](https://github.com/amterp/rad/commit/5c58d83))
- docs(guide): update shell-commands.md ([863c3c1](https://github.com/amterp/rad/commit/863c3c1))
- docs(guide): update strings-advanced.md ([2776ef6](https://github.com/amterp/rad/commit/2776ef6))
- feat: revamp http url encoding behavior ([d91350b](https://github.com/amterp/rad/commit/d91350b))
- fix(fomatting): accept thousands_separator before alignment/padding ([c18c971](https://github.com/amterp/rad/commit/c18c971))
- fix(shell): fix shell modifier keywords ([e448f72](https://github.com/amterp/rad/commit/e448f72))
- fix: check built-in function signatures for parsing errors ([17b14fe](https://github.com/amterp/rad/commit/17b14fe))
- fix: use Print, not Printf for --src ([8579dac](https://github.com/amterp/rad/commit/8579dac))
- tests: fix broken tests from earlier commit ([0dba710](https://github.com/amterp/rad/commit/0dba710))

---

## [v0.6.12](https://github.com/amterp/rad/releases/tag/v0.6.12) - 2025-10-16

## Breaking!

This release changes shell and error handling syntax, expect breaks.

```rad
// OLD: Critical command
$!`make build`

// NEW: Same behavior (critical by default)
$`make build`

// OLD: Unsafe command
unsafe $`command_that_might_fail`

// NEW: Use catch block
$`command_that_might_fail` catch:
    pass

// OLD: Recover block
$`curl {url}`
recover:
    print_err("Failed but continuing")

// NEW: Use catch block
$`curl {url}` catch:
    print_err("Failed but continuing")

// OLD: Fail block
$`curl {url}`
fail:
    print_err("Failed, exiting")
    exit(1)

// NEW: Use catch block with exit
$`curl {url}` catch:
    print_err("Failed, exiting")
    exit(1)
```

- Release VSCode extension v0.1.9 ([46e3eed](https://github.com/amterp/rad/commit/46e3eed))
- feat: implement new error handling syntax (catch suffix) ([5c0253c](https://github.com/amterp/rad/commit/5c0253c))
- feat: implement new shell syntax + catch suffix syntax ([7ef3f46](https://github.com/amterp/rad/commit/7ef3f46))
- feat: update syntax dump tests for new shell & error handling syntax ([78a4857](https://github.com/amterp/rad/commit/78a4857))
- feat: update textmate grammar for new shell + error handling syntax ([981fdfb](https://github.com/amterp/rad/commit/981fdfb))


---

## [v0.6.11](https://github.com/amterp/rad/releases/tag/v0.6.11) - 2025-10-12

- feat: add --from-logs flag to check cmd ([468b455](https://github.com/amterp/rad/commit/468b455))
- feat: add global config loading ([a068c49](https://github.com/amterp/rad/commit/a068c49))
- feat: add invocation logging & rolling ([c5ee8e7](https://github.com/amterp/rad/commit/c5ee8e7))
- refactor: move Run up in runner.go ([2257aa7](https://github.com/amterp/rad/commit/2257aa7))

---

## [v0.6.10](https://github.com/amterp/rad/releases/tag/v0.6.10) - 2025-10-06

- Release VSCode extension v0.1.8 ([5b94a4e](https://github.com/amterp/rad/commit/5b94a4e))
- feat: add stdin support for Unix-style piping ([cd880f7](https://github.com/amterp/rad/commit/cd880f7))

---

## [v0.6.9](https://github.com/amterp/rad/releases/tag/v0.6.9) - 2025-10-05

- feat: allow scientific notation for int defaults ([8fd724a](https://github.com/amterp/rad/commit/8fd724a))
- feat: validate script immediately after parsing ([127f484](https://github.com/amterp/rad/commit/127f484))
- fix: correctly allow unknown flags to be absorbed in var args ([34ee9be](https://github.com/amterp/rad/commit/34ee9be))
- fix: resolve shell command pipe race condition ([319a956](https://github.com/amterp/rad/commit/319a956))

---

## [v0.6.8](https://github.com/amterp/rad/releases/tag/v0.6.8) - 2025-09-30

- Release VSCode extension v0.1.7 ([4b6613e](https://github.com/amterp/rad/commit/4b6613e))
- ci: implement 5 more PR benchmarks ([f8310f0](https://github.com/amterp/rad/commit/f8310f0))
- feat: allow int short count arg input style ([05b7ee2](https://github.com/amterp/rad/commit/05b7ee2))

---

## [v0.6.7](https://github.com/amterp/rad/releases/tag/v0.6.7) - 2025-09-28

- Bump version to v0.6.7 ([379d785](https://github.com/amterp/rad/commit/379d785))
- ci: add PR benchmarking ([046a3c2](https://github.com/amterp/rad/commit/046a3c2))
- ci: add PR check automation ([9efd9b4](https://github.com/amterp/rad/commit/9efd9b4))
- ci: add benchmark-scripts ([7f6349c](https://github.com/amterp/rad/commit/7f6349c))
- feat: support formatting numbers with commas separating thousands ([9e2bc45](https://github.com/amterp/rad/commit/9e2bc45))
- tests: fix local TZ-dependent time tests ([8e2dd59](https://github.com/amterp/rad/commit/8e2dd59))

---

## [v0.6.6](https://github.com/amterp/rad/releases/tag/v0.6.6) - 2025-09-27

- ci: iterate goreleaser ([f3a331a](https://github.com/amterp/rad/commit/f3a331a))

---

## [v0.6.5](https://github.com/amterp/rad/releases/tag/v0.6.5) - 2025-09-27

- ci: iterate goreleaser ([284797b](https://github.com/amterp/rad/commit/284797b))

---

## [v0.6.2](https://github.com/amterp/rad/releases/tag/v0.6.2) - 2025-09-21


---

## [v0.6.1](https://github.com/amterp/rad/releases/tag/v0.6.1) - 2025-09-21


---

## [v0.6.0](https://github.com/amterp/rad/releases/tag/v0.6.0) - 2025-09-21

* fix: returning from fn in for loop by @amterp in https://github.com/amterp/rad/pull/44
* feat(args): support var args by @amterp in https://github.com/amterp/rad/pull/45

---

## [v0.5.59](https://github.com/amterp/rad/releases/tag/v0.5.59) - 2025-09-14

* Replace pflag with Ra by @amterp in https://github.com/amterp/rad/pull/38
* feat: add function 'matches' for regex matching by @amterp in https://github.com/amterp/rad/pull/39
* fix: convert arg constraint targets to external name format by @amterp in https://github.com/amterp/rad/pull/40
* feat: add global flag '--rad-args-dump' by @amterp in https://github.com/amterp/rad/pull/41
* feat(repl): implement REPL by @amterp in https://github.com/amterp/rad/pull/43

---

# Historical Releases

Releases before automated GitHub releases. Notable changes only.

---

## v0.5 - 2025-02-12

- Replaced handwritten lexer/parser with tree sitter
- `defer` statements
- Further shell command support improvements, critical shell commands
- Emoji support
- Basic syntax highlighter
- Reworked JSON field extraction algo
- `errdefer`
- Reworked string character escaping
- Improved rad block sorting operation, added matching `sort` function
- Added more functions: `confirm`, `range`, `split` etc
- Removed Cobra
- Reworked strings (`RslString`, attributes, colors)
- http functions
- parsing functions e.g. parse_int, parse_float
- `.dot.syntax` for map key access
- Truthy/falsy logic
- Raw strings
- Multiline strings
- Arg constraints - enum, regex
- Modulo operator `%`
- `++`/`--` operators

---

## v0.4 - 2024-10-28

- `exit` function
- Allow output pass-through in `rad` blocks
- `rad` field modifiers: `truncate`, `color`
- Reworked arrays: all arrays now allow mixed types
- maps
- collection entry assignment
- `del`
- `in` and `not in`
- Json algo: allow capturing json nodes as maps
- Added list/string slicing
- Improved indexing, including negative indexing
- Added ternary expressions
- Added inline expressions for string interpolation, including formatting
- Implemented shell command invocation

---

## v0.3 - 2024-09-29

- Improved shell embedding
- Improved table-to-terminal size adjustment
- Good unit testing
- Compound assignments
- Allow mocking responses `--MOCK-RESPONSE`
- Json algo: add `*` wildcard capture
- `rad` sort statements
- Colorized headers
- Switch from `int` to `int64` representation of ints
- Add `pick` functions, including `pick_from_resource`
- Add list comprehensions
- `request` and `display` blocks

---

## v0.2 - 2024-09-09

- Added Apache License 2.0
- Arg defaults
- std functions: date functions, replace, join, upper/lower, etc
- 'Single quote' strings

---

## v0.1 - 2024-09-08

- Initial version
- Newest notable feature was `--STDIN` and output shell export commands.
