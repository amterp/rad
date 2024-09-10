rsl="\
---
Test
---
args:
  name string
  age int
"

eval "$(./main --SHELL --STDIN "$0" "$@" <<< "$rsl")"
