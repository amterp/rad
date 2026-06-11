# Rad Evolution Documents (REDs)

REDs record significant decisions and proposals for Rad - what we decided, why, and what we
rejected. A single RED spans its whole life, starting as a proposal and maturing into the
permanent record of the decision.

Start with [RED-1](0001-red-process.md), which defines the process itself. New REDs copy
[`template.md`](template.md) and follow the [style guide](STYLE.md) for voice and prose.

> This index is maintained by hand for now. The frontmatter is structured so it can later be
> generated automatically.
>
> Letter IDs (RED-A, RED-B, RED-C) are **provisional**: their chronological integer slot isn't
> settled yet. RED-A and RED-B are backfilled records pending the backfill of earlier decisions;
> RED-C is a current decision parked on a provisional ID until the one-time integer renumber.
> They'll be renumbered into their proper slots later.

## Index

| ID                                  | Title                               | Kind     | Status      |
|-------------------------------------|-------------------------------------|----------|-------------|
| [RED-1](0001-red-process.md)        | The RED process                     | Process  | Accepted    |
| [RED-2](0002-why-rad.md)            | Why Rad                             | Language | Implemented |
| [RED-3](0003-rad-block.md)          | The rad block                       | Language | Implemented |
| [RED-4](0004-declarative-args.md)   | Declarative argument parsing        | Language | Implemented |
| [RED-5](0005-removing-null.md)      | Removing null from the language     | Language | Superseded  |
| [RED-6](0006-bash-embedding.md)     | Embedding Rad in Bash               | Language | Implemented |
| [RED-7](0007-json-extraction-engine.md) | The JSON extraction engine      | Architecture | Implemented |
| [RED-8](0008-interactivity.md)      | Interactivity as a first-class capability | Language | Implemented |
| [RED-A](A-request-display-split.md) | Separate request and display blocks | Language | Superseded  |
| [RED-B](B-rad-block-unification.md) | Unify the rad block keywords        | Language | Implemented |
| [RED-C](C-radish.md)                | Own our interactivity layer (radish) | Architecture | Draft       |
