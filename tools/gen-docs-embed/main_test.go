package main

import "testing"

func TestResolveDocLink(t *testing.T) {
	slugs := map[string]bool{
		"guide/shell-commands": true,
		"guide/args":           true,
		"reference/functions":  true,
	}
	funcs := map[string]bool{"pick": true, "len": true}

	cases := []struct {
		name             string
		text, href, base string
		want             string
	}{
		{"relative up-and-over", "shell", "../guide/shell-commands.md", "guide/basics", "shell (rad docs guide/shell-commands)"},
		{"relative same-dir", "args", "./args.md", "guide/basics", "args (rad docs guide/args)"},
		{"function anchor", "`pick`", "../reference/functions.md#pick", "guide/resources", "`pick` (rad docs pick)"},
		{"functions page no anchor", "ref", "../reference/functions.md", "guide/x", "ref (rad docs reference/functions)"},
		{"non-function anchor falls back to page", "http", "../reference/functions.md#http-functions", "guide/x", "http (rad docs reference/functions)"},
		{"in-page anchor", "float", "#float", "guide/basics", "float"},
		{"external url keeps bare link", "site", "https://example.com", "guide/x", "site (https://example.com)"},
		{"image is not a topic", "img", "./pic.png", "guide/x", "img"},
		{"unknown page collapses to text", "missing", "../guide/missing.md", "guide/x", "missing"},
	}
	for _, c := range cases {
		if got := resolveDocLink(c.text, c.href, c.base, slugs, funcs); got != c.want {
			t.Errorf("%s: resolveDocLink(%q, %q, %q) = %q, want %q",
				c.name, c.text, c.href, c.base, got, c.want)
		}
	}
}
