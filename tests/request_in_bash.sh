rsl="\
---
Test
---
args:
  name string
  age int

Name = json[].commit.author.name
Email = json[].commit.author.email
Date = json[].commit.author.date

url = 'https://api.github.com/repos/torvalds/linux/commits?per_page=2'

rad url:
    fields Date, Name, Email
"

./main --STDIN "$0" "$@" <<< "$rsl"
