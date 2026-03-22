"""MkDocs hook that generates llms.txt, llms-full.txt, and raw .md copies.

Runs on_post_build to write files into the site output directory.
Auto-discovers pages from the nav config - adding a new doc page requires
no changes here.
"""

import logging
import os
import re
import shutil
from pathlib import Path

log = logging.getLogger("mkdocs.hooks.generate_llms_txt")

HEADER = """\
# Rad

> A scripting language designed to make writing CLI tools delightful. \
Familiar, Python-like syntax with CLI essentials built-in."""

# Nav sections whose pages get full content inlined in llms-full.txt.
FULL_CONTENT_SECTIONS = {"Reference"}

# Pages to skip entirely (not useful for LLM consumption).
SKIP_PAGES = {"index.md"}


def on_post_build(config):
    site_dir = Path(config["site_dir"])
    docs_dir = Path(config["docs_dir"])
    site_url = config.get("site_url", "").rstrip("/") + "/"
    nav = config["nav"]

    pages = _parse_nav(nav)
    pages = [(s, f, t) for s, f, t in pages if f not in SKIP_PAGES]

    page_data = []
    for section, filepath, explicit_title in pages:
        source = docs_dir / filepath
        if not source.exists():
            log.warning("llms-txt: skipping %s (file not found)", filepath)
            continue
        raw = source.read_text(encoding="utf-8")
        title = _resolve_title(raw, explicit_title, filepath)
        h2s = _extract_h2s(raw)
        content = _strip_front_matter(raw)
        url = site_url + filepath
        page_data.append({
            "section": section,
            "filepath": filepath,
            "title": title,
            "h2s": h2s,
            "content": content,
            "url": url,
        })

    toc = _build_toc(page_data)

    # llms.txt - concise index
    llms_txt = f"{HEADER}\n\n{toc}\n"
    (site_dir / "llms.txt").write_text(llms_txt, encoding="utf-8")

    # llms-full.txt - TOC + full reference content
    full_sections = []
    for p in page_data:
        if p["section"] not in FULL_CONTENT_SECTIONS:
            continue
        body = p["content"].strip()
        # Avoid duplicating H1 if content already starts with one
        if body.startswith("# "):
            full_sections.append(body)
        else:
            full_sections.append(f"# {p['title']}\n\n{body}")

    full_content = "\n\n---\n\n".join(full_sections)
    llms_full_txt = f"{HEADER}\n\n{toc}\n\n---\n\n{full_content}\n"
    (site_dir / "llms-full.txt").write_text(llms_full_txt, encoding="utf-8")

    # Copy raw .md files so links resolve
    for p in page_data:
        src = docs_dir / p["filepath"]
        dest = site_dir / p["filepath"]
        dest.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, dest)

    log.info(
        "llms-txt: generated llms.txt, llms-full.txt, and %d .md files",
        len(page_data),
    )


def _parse_nav(nav, section=None):
    """Walk the nav config and return [(section, filepath, explicit_title)]."""
    pages = []
    for entry in nav:
        if isinstance(entry, str):
            pages.append((section, _clean_path(entry), None))
        elif isinstance(entry, dict):
            for key, value in entry.items():
                if isinstance(value, list):
                    # Section with children
                    pages.extend(_parse_nav(value, section=key))
                elif isinstance(value, str):
                    # Page with explicit title
                    pages.append((section, _clean_path(value), key))
    return pages


def _clean_path(path):
    """Strip leading ./ from nav paths."""
    return path.lstrip("./")


def _resolve_title(raw_content, explicit_title, filepath):
    """Determine page title: explicit nav title > front matter > H1 > filename."""
    if explicit_title:
        return explicit_title

    # Check front matter title
    fm = _extract_front_matter(raw_content)
    if fm:
        match = re.search(r"^title:\s*(.+)$", fm, re.MULTILINE)
        if match:
            return match.group(1).strip().strip("\"'")

    # Check first H1
    match = re.search(r"^# (.+)$", raw_content, re.MULTILINE)
    if match:
        return match.group(1).strip()

    # Fallback to filename
    return Path(filepath).stem.replace("-", " ").title()


def _extract_h2s(raw_content):
    """Extract all H2 heading texts."""
    return re.findall(r"^## (.+)$", raw_content, re.MULTILINE)


def _extract_front_matter(raw_content):
    """Return front matter string if present, else None."""
    if not raw_content.startswith("---"):
        return None
    end = raw_content.find("---", 3)
    if end == -1:
        return None
    return raw_content[3:end].strip()


def _strip_front_matter(raw_content):
    """Remove YAML front matter if present."""
    if not raw_content.startswith("---"):
        return raw_content
    end = raw_content.find("---", 3)
    if end == -1:
        return raw_content
    return raw_content[end + 3:].lstrip("\n")


def _build_toc(page_data):
    """Build the table of contents shared by llms.txt and llms-full.txt."""
    sections = []
    current_section = None
    current_lines = []

    for p in page_data:
        section = p["section"] or "Other"
        if section != current_section:
            if current_lines:
                sections.append((current_section, current_lines))
            current_section = section
            current_lines = []
        suffix = f": {', '.join(p['h2s'])}" if p["h2s"] else ""
        current_lines.append(f"- [{p['title']}]({p['url']}){suffix}")

    if current_lines:
        sections.append((current_section, current_lines))

    parts = []
    for name, lines in sections:
        parts.append(f"## {name}\n\n" + "\n".join(lines))

    return "\n\n".join(parts)
