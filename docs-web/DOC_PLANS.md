# Documentation Plans

- Introduction
  - JSON Queries
  - Non-JSON Queries (general scripting)
- Advanced?
- Rad By Example
- Reference
  - Functions
  - Keywords
  - Rad block functionality
  - Global flags

---

## 2025-10-18 Doc Update Push

- [x] getting started
- [x] basics
- [x] args
- [x] functions
  - [x] NEW custom functions
  - [x] NEW lambdas
- [x] strings-advanced
  - [x] NEW thousands commas
- [x] rad blocks
  - [x] NEW lambdas for `map` 
- [x] NEW type annotations
- [x] NEW error handling (`catch`, `??`)
  - [x] error propagation model
- [x] shell commands !!
  - [x] shell error handling
- [ ] resources
- [ ] global flags
- [ ] json-paths-advanced
- [ ] defer/errdefer
- [ ] NEW misc
  - [ ] NEW macros
  - [ ] NEW del
  - [ ] NEW stdin

- [ ] NEW stashes (unsure if in guide or reference. Advanced Guide?)
- [ ] examples
   - [ ] example: hm
   - [ ] example: dot

## 2025-05-29 Doc Update Push (abandoned)

- [ ] update existing
    - [x] getting started
    - [x] basics
    - [x] args
    - [ ] rad blocks
    - [ ] functions
    - [ ] strings
    - [ ] resources
    - [ ] shell commands
    - [ ] global flags
    - [ ] defer/errdefer
- [ ] new
    - [ ] example: hm
    - [ ] example: dot
    - [ ] guide: stashes
    - [ ] guide: misc - macros


## Doc Todo

? change all external links to open in new tabs, don't go off rad docs site
? consider adding 'exercises'?

- elevator pitch / motivation / why
  - [ ] take 1
- getting started
  - [x] take 1
  - installation troubleshooting
- basics
  - [x] take 1
- args
  - [x] take 1
- rad blocks
  - [x] take 1
- functions
  - [x] take 1
- string interpolation & formatting, multiline strings, raw
  - [x] take 1
- resources, picking
  - [x] take 1
- shell cmds
  - [x] take 1
- global flags
  - [x] take 1
  - mock
- defer/errdefer
  - [x] take 1
- misc advanced
  - [ ] take 1
  - colors
  - varpath assignment?
  - shell embedding
  - json extraction algo details! visual aids? the idea of representing json as a tree?

- style guide? tips?
- something which goes through rsl snippets and runs them to ensure they compile
- parallel 'Rad by Example' guide?

---

## Updates to make

- --SRC, --SRC-TREE
- 'new' command
- --CONFIRM-SHELL
- 'confirm' shell modifier
- 'input' function
- if pre-stmt
- arg range constraint
- list comprehensions exprs dont need to return 1+ values
