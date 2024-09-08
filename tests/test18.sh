args="\
\"\"\"
Demo of --BASH flag
\"\"\"
args:
  name string # Name of the person
  age int = 30 # Age of the person"

eval "$(./main --STDIN --BASH $* <<< "$args")"

echo "Name: $name"
echo "Age: $age"
