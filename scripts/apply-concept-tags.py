#!/usr/bin/env python3
"""Copy the concept tags from seed_data/*.xml into a running qwacback.

Tags are the <concept> elements after the first one on a <var> or <varGrp>
(#19). A fresh database gets them from the seed import; an existing one
(production) was seeded before the tags were added, so this script writes
them onto the matching records: study by title, then question by name.

Dry run by default: it prints what would change. With --apply it writes,
authenticated as a PocketBase superuser.

    QWACBACK_URL=https://qwacback.correlaid.org \\
    PB_ADMIN_EMAIL=… PB_ADMIN_PASSWORD=… \\
    scripts/apply-concept-tags.py --apply

Needs a qwacback that has the `tags` field (migration concept_tags); the
script stops before writing if the field is missing, since PocketBase would
silently ignore it.
"""

import argparse
import glob
import json
import os
import sys
import urllib.parse
import urllib.request
import xml.etree.ElementTree as ET

NS = {"d": "ddi:codebook:2_5"}
XML_LANG = "{http://www.w3.org/XML/1998/namespace}lang"


def seed_tags(path):
    """(study title, [(collection, name, tags)]) for one seed file."""
    root = ET.parse(path).getroot()
    title = root.findtext("d:stdyDscr/d:citation/d:titlStmt/d:titl", namespaces=NS)
    out = []
    for tag, collection in (("varGrp", "variable_groups"), ("var", "variables")):
        for el in root.iter("{%s}%s" % (NS["d"], tag)):
            concepts = el.findall("d:concept", NS)
            tags = [
                {"lang": c.get(XML_LANG, ""), "text": (c.text or "").strip()}
                for c in concepts[1:]
                if (c.text or "").strip()
            ]
            if tags:
                tags = [{k: v for k, v in t.items() if v} for t in tags]
                out.append((collection, el.get("name"), tags))
    return title, out


def request(base, path, token=None, method="GET", body=None):
    req = urllib.request.Request(base.rstrip("/") + path, method=method)
    if token:
        req.add_header("Authorization", token)
    data = None
    if body is not None:
        data = json.dumps(body).encode()
        req.add_header("Content-Type", "application/json")
    with urllib.request.urlopen(req, data) as resp:
        return json.load(resp)


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--apply", action="store_true", help="write the tags (default: dry run)")
    args = ap.parse_args()
    base = os.environ.get("QWACBACK_URL", "http://127.0.0.1:8090")

    token = None
    if args.apply:
        email, password = os.environ.get("PB_ADMIN_EMAIL"), os.environ.get("PB_ADMIN_PASSWORD")
        if not email or not password:
            sys.exit("--apply needs PB_ADMIN_EMAIL and PB_ADMIN_PASSWORD")
        token = request(base, "/api/collections/_superusers/auth-with-password",
                        method="POST", body={"identity": email, "password": password})["token"]

    studies = request(base, "/api/collections/studies/records?perPage=500&fields=id,title")["items"]
    study_id = {s["title"]: s["id"] for s in studies}
    # /api/questions is public: its ids are the variable or group record ids.
    questions = request(base, "/api/questions?perPage=10000")
    by_study_name = {(q["study_id"], q["name"]): q for q in questions}

    changes, missing = [], []
    for path in sorted(glob.glob(os.path.join(os.path.dirname(__file__), "..", "seed_data", "*.xml"))):
        title, entries = seed_tags(path)
        sid = study_id.get(title)
        if not sid:
            print(f"skip {os.path.basename(path)}: no study titled {title!r}")
            continue
        for collection, name, tags in entries:
            q = by_study_name.get((sid, name))
            if not q:
                missing.append(f"{title}: {name}")
                continue
            changes.append((collection, q["id"], name, tags, q.get("tags") or []))

    todo = [c for c in changes if c[3] != c[4]]
    print(f"{len(changes)} tagged questions matched, {len(todo)} to update, {len(missing)} not found")
    for m in missing:
        print("  not found:", m)
    for collection, rid, name, tags, _ in todo:
        print(f"  {collection}/{rid} {name}: " + ", ".join(t["text"] for t in tags))

    if not args.apply:
        print("dry run; pass --apply to write")
        return

    # Refuse to write into a database without the field.
    for collection in {c[0] for c in todo}:
        rid = next(c[1] for c in todo if c[0] == collection)
        rec = request(base, f"/api/collections/{collection}/records/{rid}", token)
        if "tags" not in rec:
            sys.exit(f"{collection} has no `tags` field: deploy the concept_tags migration first")

    for collection, rid, name, tags, _ in todo:
        request(base, f"/api/collections/{collection}/records/{rid}", token, "PATCH", {"tags": tags})
    print(f"updated {len(todo)} records")


if __name__ == "__main__":
    main()
