rsl="\
---
Demo of --SHELL flag
---
args:
  name string # Name of the person
  age int = 30 # Age of the person"

eval "$(./main --SHELL --STDIN "$0" "$@" <<< "$rsl")"

echo "Name: $name"
echo "Age: $age"
