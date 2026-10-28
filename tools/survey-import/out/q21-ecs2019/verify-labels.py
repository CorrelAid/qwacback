#!/usr/bin/env python3
"""Check every text in form.json against the ECS 2019 management questionnaire PDF.

usage: python3 -I verify-labels.py form.json ecs2019_mm_questionnaire.pdf

The PDF is converted twice: `pdftotext -raw` (content order, so the
two-column grids come out item by item, but some words are glued) and
`pdftotext -layout` (spaces kept, columns side by side); a line passes if it
occurs in either. The tick-box glyph U+2610 is removed. Lines are
joined with one space; a line ending in "letter-" is joined with no space and
the hyphen is kept as a marker that may be skipped (layout hyphenation) or read
as a real hyphen ("decision-/making"). Runs of whitespace are collapsed; quotes,
dashes, arrows and case are NOT normalised.

Each label, hint, guidance_hint and choice label is split at line breaks and
every line must occur in the PDF text. A line that is found only when all
whitespace is ignored is reported as SPACE-ONLY (the raw text layer glues some
words, e.g. "andretain"); check those against `pdftotext -layout`. Lines found
in neither way are NOT FOUND. Exit 1 if anything is NOT FOUND.
"""
import json
import re
import subprocess
import sys

SHY = "­"
form_path, pdf_path = sys.argv[1], sys.argv[2]
norm = lambda s: re.sub(r"\s+", " ", s).strip()


def corpus(mode):
    out = subprocess.run(["pdftotext", mode, pdf_path, "-"], check=True, capture_output=True, text=True).stdout
    lines = [ln.replace("☐", " ").strip() for ln in out.replace("\f", "\n").split("\n")]
    text = ""
    for ln in (l for l in lines if l):
        if re.search(r"[A-Za-z]-$", text):
            text = text[:-1] + SHY + ln
        else:
            text += (" " if text else "") + ln
    return norm(text)


# -raw keeps the two-column grids item by item but glues some words;
# -layout keeps the spaces but puts two columns side by side.
sources = [corpus("-raw"), corpus("-layout")]
source_nows = re.sub(r"\s+", "", sources[0])


def pattern(line, sep=""):
    parts = ["[-" + SHY + "]" if ch == "-" else re.escape(ch) for ch in line]
    return re.compile(("(?:" + SHY + ")?" + sep).join(parts))


form = json.load(open(form_path, encoding="utf-8"))
texts = []
for r in form["survey"]:
    for key in ("label", "hint", "guidance_hint"):
        if r.get(key):
            texts.append((f"survey {r.get('name')} {key}", r[key]))
for ch in form["choices"]:
    texts.append((f"choice {ch['list_name']}/{ch['name']}", ch["label"]))

fails = spaceonly = checked = 0
seen = set()
for where, t in texts:
    for line in t.split("\n"):
        line = norm(line)
        if not line:
            continue
        checked += 1
        if line in seen:
            continue
        seen.add(line)
        if any(pattern(line).search(src) for src in sources):
            continue
        if pattern(re.sub(r"\s+", "", line)).search(source_nows):
            spaceonly += 1
            print(f"SPACE-ONLY: {where}: {line!r}")
            continue
        fails += 1
        print(f"NOT FOUND: {where}: {line!r}")
print(f"checked {checked} text lines ({len(seen)} distinct) from {len(texts)} texts: "
      f"{fails} not found, {spaceonly} matched only ignoring whitespace")
sys.exit(1 if fails else 0)
