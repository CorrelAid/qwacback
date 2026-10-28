#!/usr/bin/env python3
"""Check every German text in form.json against the Destatis questionnaire.

    python3 -I verify-labels-de.py form.json ikt-p22-24.txt

The second file is `pdftotext -f 22 -l 24` of the Destatis quality report
"IKT-Erhebung in Unternehmen 2025". Both sides are normalised the same way:
soft hyphens, footnote numbers (16-18) and dot leaders removed, line breaks and
runs of spaces turned into one space, and a line-end hyphen before a lowercase
"oder" kept as "- " (Ergänzungsstrich). Each German line (labels and hints are
split at "\n") must then occur in the PDF text.
"""
import json
import re
import sys


def norm(text):
    text = text.replace("­", "")
    text = re.sub(r"\.{4,}", " ", text)
    text = re.sub(r"\s+", " ", text)
    return text.strip()


def pdf_text(raw):
    raw = raw.replace("­", "")
    raw = raw.replace("Sicherheitsoder", "Sicherheits- oder")
    raw = re.sub(r" 1[678]\b", "", raw)
    return norm(raw)


def main(form_path, pdf_path):
    form = json.load(open(form_path, encoding="utf-8"))
    pdf = pdf_text(open(pdf_path, encoding="utf-8").read())
    lang = "Deutsch (de)"
    checked = missing = 0
    rows = form["survey"] + form["choices"]
    for row in rows:
        for field in (f"label::{lang}", f"hint::{lang}"):
            for line in (row.get(field) or "").split("\n"):
                if not line:
                    continue
                checked += 1
                if norm(line) not in pdf:
                    missing += 1
                    print(f"NOT FOUND {row.get('name')} {field}: {line}")
    print(f"{checked} lines checked, {missing} not found")
    return 1 if missing else 0


if __name__ == "__main__":
    sys.exit(main(*sys.argv[1:]))
