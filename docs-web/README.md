# Rad Docs

Commands from the repo root.

## Local development

```sh
mkdocs serve -f ./docs-web/mkdocs.yml
```

## Deploy

Docs are automatically deployed via GitHub Actions:

- **On push to main**: Changes to `docs-web/` trigger automatic deployment
- **On release**: New releases automatically update the release notes and redeploy

Manual deployment (if needed):

```sh
mkdocs gh-deploy -f ./docs-web/mkdocs.yml --force
```
