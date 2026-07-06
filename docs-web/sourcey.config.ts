import { defineConfig, godoc } from "sourcey";

export default defineConfig({
  name: "Rad Go API Reference",
  siteUrl: "https://amterp.dev",
  baseUrl: "/rad/sourcey-go-api",
  prettyUrls: "slash",
  repo: "https://github.com/amterp/rad",
  editBranch: "main",
  navigation: {
    tabs: [
      {
        tab: "Go API",
        slug: "go-api",
        source: godoc({
          module: "..",
          packages: ["./..."],
          mode: "live",
          includeTests: true,
          includeUnexported: false,
          hideUndocumented: false,
          exclude: [
            "github.com/amterp/rad/tools/docir",
            "github.com/amterp/rad/tools/gen-docs-embed",
            "github.com/amterp/rad/tools/gen-errors-page",
            "github.com/amterp/rad/tools/gen-funcs-go",
            "github.com/amterp/rad/tools/gen-funcs-page",
            "github.com/amterp/rad/tools/gen-funcs-sigs"
          ],
          sourceBasePath: ""
        })
      }
    ]
  }
});
