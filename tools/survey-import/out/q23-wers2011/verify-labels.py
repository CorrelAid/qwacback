#!/usr/bin/env python3
"""Check every text in form.json against the WERS6 Management Questionnaire
(BIS/11/803, PDF).

usage: python3 -I verify-labels.py form.json 11-803-...-management-questionnaire.pdf

The PDF is converted with `pdftotext -raw`. The corpus is only the parts that
were transcribed: Section EPQ from its heading up to the line
"ZFULLM" (pp. 4-5), and Section C from its heading up to the heading of
Section D (pp. 30-37). Page footers ("Page N of 111") are removed. Lines are
joined with one space; a line ending in "letter-" is joined with no space and
the hyphen is kept as a soft-hyphen marker, which may be skipped or read as a
hyphen when matching. Runs of whitespace are collapsed; there is no quote,
dash or case normalisation.
Each label, hint, guidance_hint, constraint_message and choice label is split
at line breaks and every line must occur: texts of the EPQ group in the EPQ
part, all other texts (Section C and all choice labels) in Section C.
Failures are printed; exit 1.
"""
import json
import re
import subprocess
import sys

SHY = "­"
form_path, pdf_path = sys.argv[1], sys.argv[2]
raw = subprocess.run(["pdftotext", "-raw", pdf_path, "-"], check=True, capture_output=True, text=True).stdout
lines = [re.sub(r"Page \d+ of 111", "", ln).strip() for ln in raw.replace("\f", "\n").split("\n")]


def part(first, stop):
    a = lines.index(first)
    b = next(i for i in range(a + 1, len(lines)) if lines[i] == stop)
    return lines[a:b]


def joined(ls):
    text = ""
    for ln in ls:
        if not ln:
            continue
        if re.search(r"[A-Za-z]-$", text):
            text = text[:-1] + SHY + ln
        else:
            text += (" " if text else "") + ln
    return norm(text)


norm = lambda s: re.sub(r"[ \t\n]+", " ", s).strip()
epq = joined(part("SECTION EPQ EMPLOYEE PROFILE QUESTIONNAIRE", "ZFULLM"))
secc = joined(part("SECTION C RECRUITMENT, TRAINING AND ORGANISATION OF WORK", "SECTION D CONSULTATION AND COMMUNICATION"))


def pattern(line):
    parts = ["[-" + SHY + "]" if ch == "-" else re.escape(ch) for ch in line]
    return re.compile(("(?:" + SHY + ")?").join(parts))


form = json.load(open(form_path, encoding="utf-8"))
texts = []
section = None
for r in form["survey"]:
    if r["type"] == "begin_group" and r["name"].startswith("section_"):
        section = r["name"]
    for key in ("label", "hint", "guidance_hint", "constraint_message"):
        if r.get(key):
            texts.append((f"survey {r.get('name')} {key}", r[key], section))
for ch in form["choices"]:
    texts.append((f"choice {ch['list_name']}/{ch['name']}", ch["label"], "section_c"))

fails = checked = 0
for where, t, sec in texts:
    corpus = epq if sec == "section_epq" else secc
    for line in t.split("\n"):
        line = norm(line)
        if not line:
            continue
        checked += 1
        if not pattern(line).search(corpus):
            fails += 1
            print(f"NOT IN {sec}: {where}: {line!r}")
print(f"checked {checked} text lines from {len(texts)} texts: {fails} not found")
sys.exit(1 if fails else 0)
