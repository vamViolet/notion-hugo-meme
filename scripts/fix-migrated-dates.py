#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Self-heal the publish date of migrated articles.

The Notion→Hugo Action (rxrw/notion-blog) sets each synced article's
front-matter `date:` from the Notion page's `created_time`. That field is a
read-only system property auto-set to *now* on page creation and cannot be
overridden via the API. So articles migrated from the old Gridea blog (originally
dated 2021-2023) would all land with today's date and clump at the top of the
archives as "newest".

This step runs on every sync, AFTER the Action writes the .md files and AFTER
fix-tags, BEFORE the commit. It reads scripts/migrated-dates.json (title ->
original YYYY-MM-DD) and rewrites the `date:` (and `lastmod:`) front-matter
lines of any matching content/zh/**/*.md back to the original date, so the
articles sort and display with their real publication dates.

Idempotent: a file whose date already equals the mapped value is left untouched.
Removing an article's entry from the mapping stops patching it (the Action's
date then wins again). Title match is exact (after stripping surrounding
quotes/whitespace); the Notion page Name is the original title, and the Action
writes `title: "{{.Title}}"`, so this lines up.
"""
import glob
import json
import os
import re
import sys

MAPPING_FILE = os.path.join("scripts", "migrated-dates.json")
CONTENT_GLOB = "content/zh/**/*.md"


def load_mapping():
    with open(MAPPING_FILE, encoding="utf-8") as f:
        m = json.load(f)
    # drop empty/invalid dates defensively
    return {k: v for k, v in m.items() if v}


def front_matter_bounds(lines):
    """Return (start, end) indices of the front matter body, or None.
    start is the index after the opening ---, end is the index of the closing ---."""
    if not lines or lines[0].strip() != "---":
        return None
    for i in range(1, len(lines)):
        if lines[i].strip() == "---":
            return (1, i)
    return None


def title_of(lines, start, end):
    for i in range(start, end):
        m = re.match(r"^title\s*:\s*(.*)$", lines[i])
        if not m:
            continue
        val = m.group(1).strip()
        if (val.startswith('"') and val.endswith('"')) or (
            val.startswith("'") and val.endswith("'")
        ):
            val = val[1:-1]
        return val
    return None


def rewrite_date_line(line, new_date):
    """Replace the value of a `date:` / `lastmod:` line, preserving the key."""
    return re.sub(r"^(\s*(?:date|lastmod)\s*:\s*).*$", r"\g<1>" + new_date, line)


def main():
    mapping = load_mapping()
    if not mapping:
        print("migrated-dates.json empty or missing; nothing to patch.")
        return 0
    patched = 0
    skipped = 0
    for path in glob.glob(CONTENT_GLOB, recursive=True):
        with open(path, encoding="utf-8") as f:
            lines = f.readlines()
        bounds = front_matter_bounds(lines)
        if bounds is None:
            continue
        start, end = bounds
        title = title_of(lines, start, end)
        if title is None:
            continue
        new_date = mapping.get(title)
        if not new_date:
            continue  # not a migrated article
        changed = False
        date_now = lastmod_now = None
        for i in range(start, end):
            if re.match(r"^date\s*:", lines[i]):
                date_now = lines[i]
            elif re.match(r"^lastmod\s*:", lines[i]):
                lastmod_now = lines[i]
        for i in range(start, end):
            if re.match(r"^date\s*:", lines[i]) or re.match(r"^lastmod\s*:", lines[i]):
                candidate = rewrite_date_line(lines[i], new_date)
                if candidate != lines[i]:
                    lines[i] = candidate
                    changed = True
        if changed:
            with open(path, "w", encoding="utf-8") as f:
                f.writelines(lines)
            patched += 1
            print(f"patched {path}: date/lastmod -> {new_date}")
        else:
            skipped += 1
    print(f"Migrated dates: {patched} patched, {skipped} already correct.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
