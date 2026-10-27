#!/usr/bin/env python3
"""Check every text in form.json against Annex 1 (model questionnaire) of the
2025 compilers' manual KS-01-25-065 (PDF).

usage: python3 -I verify-labels.py form.json KS-01-25-065-EN-N.pdf

The PDF is converted with `pdftotext -raw`. Only Annex 1 is used: from the
line "MODEL QUESTIONNAIRE VERSION 1.5" to the first line "Validation rules"
(start of Annex 2), which includes Annex 1's glossary. The explanatory notes
(Section 2.3) and the data transmission tables are NOT part of the corpus.
Page footers ("A1" / "Model questionnaire" / "European business(es)
statistics compiler... NNN") and the tick-box glyph U+F0A8 are removed.
Lines are joined with one space; a line ending in "letter-" is joined with no
space and the hyphen is kept as a soft-hyphen marker. When matching, that
marker may be skipped (layout hyphenation "tech-/nologies") or stand for a
real hyphen ("self-/driving"); nowhere else is anything tolerated. Runs of
whitespace are collapsed; no quote, dash or case normalisation.
Each label, hint and choice label of form.json is split at line breaks and
every line must occur. Failures are printed; exit 1.
"""
import json
import re
import subprocess
import sys

SHY = "­"
form_path, pdf_path = sys.argv[1], sys.argv[2]
raw = subprocess.run(["pdftotext", "-raw", pdf_path, "-"], check=True, capture_output=True, text=True).stdout
lines = [ln.replace("", " ").strip() for ln in raw.replace("\f", "\n").split("\n")]
start = lines.index("MODEL QUESTIONNAIRE VERSION 1.5")
stop = next(i for i in range(start, len(lines)) if lines[i] == "Validation rules")
annex = lines[start:stop]

FOOT = re.compile(r"^European business(es)? statistics compiler.{0,4}s? manual\s*for ICT usage and e-commerce in enterprises \d+$")
kept = []
for i, ln in enumerate(annex):
    if ln == "Model questionnaire" or FOOT.match(ln):
        continue
    if i + 1 < len(annex) and annex[i + 1] == "Model questionnaire" and ln.endswith("A1"):
        ln = ln[:-2].rstrip()
    if ln:
        kept.append(ln)

text = ""
for ln in kept:
    if re.search(r"[A-Za-z]-$", text):
        text = text[:-1] + SHY + ln
    else:
        text += (" " if text else "") + ln
norm = lambda s: re.sub(r"[ \t\n]+", " ", s).strip()
source = norm(text)


def pattern(line):
    parts = []
    for ch in line:
        parts.append("[-" + SHY + "]" if ch == "-" else re.escape(ch))
    return re.compile(("(?:" + SHY + ")?").join(parts))


form = json.load(open(form_path, encoding="utf-8"))
texts = []
for r in form["survey"]:
    for key in ("label", "hint", "guidance_hint"):
        if r.get(key):
            texts.append((f"survey {r.get('name')} {key}", r[key]))
for ch in form["choices"]:
    texts.append((f"choice {ch['list_name']}/{ch['name']}", ch["label"]))

fails = checked = 0
for where, t in texts:
    for line in t.split("\n"):
        line = norm(line)
        if not line:
            continue
        checked += 1
        if not pattern(line).search(source):
            fails += 1
            print(f"NOT IN ANNEX 1: {where}: {line!r}")
print(f"checked {checked} text lines from {len(texts)} texts: {fails} not found")
sys.exit(1 if fails else 0)
