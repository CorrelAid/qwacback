#!/usr/bin/env python3
"""Check every text in form.json against the 2022 AES manual v1.3 (docx).

usage: python3 -I verify-labels.py form.json manual.docx

The manual is converted with pandoc to HTML; paragraphs and table cells are
collected as plain text. Each label, hint and choice label of form.json is
split at line breaks, and every line must occur character for character in
the manual text (only runs of whitespace are collapsed to one space; no
quote, dash or case normalisation). Lines that fail are printed; exit 1.
"""
import html.parser
import json
import re
import subprocess
import sys

form_path, docx_path = sys.argv[1], sys.argv[2]
html_text = subprocess.run(
    ["pandoc", "-t", "html", "--wrap=none", docx_path], check=True, capture_output=True, text=True
).stdout


class Collector(html.parser.HTMLParser):
    BLOCK = {"p", "td", "th", "li", "h1", "h2", "h3", "h4", "h5", "h6"}

    def __init__(self):
        super().__init__()
        self.parts, self.buf = [], None

    def handle_starttag(self, tag, attrs):
        if tag in self.BLOCK:
            self.flush()
            self.buf = []
        elif tag == "br" and self.buf is not None:
            self.buf.append("\n")

    def handle_endtag(self, tag):
        if tag in self.BLOCK:
            self.flush()

    def handle_data(self, data):
        if self.buf is not None:
            self.buf.append(data)

    def flush(self):
        if self.buf:
            self.parts.append("".join(self.buf))
        self.buf = None


c = Collector()
c.feed(html_text)
c.flush()
norm = lambda s: re.sub(r"\s+", " ", s).strip()
manual = "\n".join(norm(p) for p in c.parts if p.strip())

form = json.load(open(form_path, encoding="utf-8"))
texts = []
for r in form["survey"]:
    for key in ("label", "hint"):
        if r.get(key):
            texts.append((f"survey {r.get('name')} {key}", r[key]))
for ch in form["choices"]:
    texts.append((f"choice {ch['list_name']}/{ch['name']}", ch["label"]))

fails = checked = 0
for where, text in texts:
    for line in text.split("\n"):
        line = norm(line)
        if not line:
            continue
        checked += 1
        if line not in manual:
            fails += 1
            print(f"NOT IN MANUAL: {where}: {line!r}")
print(f"checked {checked} text lines from {len(texts)} texts: {fails} not found")
sys.exit(1 if fails else 0)
