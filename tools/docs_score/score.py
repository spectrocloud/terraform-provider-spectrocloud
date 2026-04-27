#!/usr/bin/env python3
"""
Deterministic documentation quality scorer for terraform-provider-spectrocloud.

Walks docs/resources/ and docs/data-sources/, parses each tfplugindocs-generated
page, and awards defect points. Lower total is better.

See autoresearch_docs/program.md for the metric definition. This script IS the
ground truth — if the program.md and this file disagree, this file wins and
the program.md should be updated.

Usage:
    python3 tools/docs_score/score.py           # prints summary, writes score.json
    python3 tools/docs_score/score.py --top 10  # also prints worst-10 pages

Output file: tools/docs_score/score.json
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from dataclasses import dataclass, field
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
DOCS_DIRS = [
    REPO_ROOT / "docs" / "resources",
    REPO_ROOT / "docs" / "data-sources",
]
OUT_FILE = Path(__file__).resolve().parent / "score.json"

# Defect weights. Keep these as named constants so diffs in program.md vs
# behavior are easy to audit.
W_ATTR_EMPTY_DESC = 10
W_ATTR_STUB_DESC = 5
W_ATTR_NO_TERMINAL_PUNCT = 1
W_ATTR_NAME_ECHO = 4
W_ATTR_LIST_NO_ELEMENT_SEMANTICS = 2

W_PAGE_NO_EXAMPLE = 15
W_PAGE_EMPTY_EXAMPLE = 10
W_PAGE_BROKEN_FENCES = 8
W_PAGE_EMPTY_FRONTMATTER_DESC = 5
W_PAGE_EMPTY_TOP_DESC = 8

STUB_LEN = 15  # descriptions strictly shorter than this are "stubs"

# Lines inside ## Schema that describe attributes look like:
#   - `name` (Type) description text
# Nested block references look like:
#   - `timeouts` (Block, Optional) (see [below for nested schema](...))
# We skip the "see [below..." pattern because the real description lives on
# the nested schema section.
ATTR_LINE_RE = re.compile(
    r"^\s*-\s+`(?P<name>[a-zA-Z0-9_]+)`\s+\((?P<meta>[^)]+)\)\s*(?P<rest>.*?)\s*$"
)
NESTED_REF_RE = re.compile(r"^\s*\(see \[below for nested schema\]")

# "List of ..." / "Set of ..." / "Map of ..." in the (Type) metadata.
LIST_TYPE_RE = re.compile(r"^(List|Set|Map)\s+of\s+", re.IGNORECASE)

# Frontmatter extraction.
FRONTMATTER_RE = re.compile(r"^---\n(.*?)\n---\n", re.DOTALL)


@dataclass
class PageReport:
    path: str
    score: int = 0
    defects: list[str] = field(default_factory=list)
    attrs_total: int = 0
    attrs_defective: int = 0


def extract_frontmatter(text: str) -> tuple[str, str]:
    """Returns (frontmatter_body, rest_of_doc). If no frontmatter, ('', text)."""
    m = FRONTMATTER_RE.match(text)
    if not m:
        return "", text
    return m.group(1), text[m.end():]


def parse_frontmatter_description(fm: str) -> str:
    """Extract the `description: |-` block from frontmatter."""
    if "description:" not in fm:
        return ""
    # Handle both `description: |-\n  body` and `description: body` forms.
    lines = fm.splitlines()
    out: list[str] = []
    in_block = False
    block_indent: int | None = None
    for ln in lines:
        if ln.strip().startswith("description:"):
            after = ln.split("description:", 1)[1].strip()
            if after in ("|", "|-", "|+"):
                in_block = True
                continue
            if after:
                # single-line description
                return after.strip().strip('"')
            in_block = True
            continue
        if in_block:
            if not ln.strip():
                out.append("")
                continue
            indent = len(ln) - len(ln.lstrip())
            if block_indent is None:
                block_indent = indent
            if indent < (block_indent or 0):
                break
            out.append(ln[block_indent:])
    return "\n".join(out).strip()


def find_top_description(body: str) -> str:
    """The paragraph between `# ... (Resource|Data Source)` heading and the
    first `## ...` section. May be empty for poorly-described pages."""
    lines = body.splitlines()
    start = None
    for i, ln in enumerate(lines):
        if ln.startswith("# ") and ("(Resource)" in ln or "(Data Source)" in ln):
            start = i + 1
            break
    if start is None:
        return ""
    collected: list[str] = []
    for ln in lines[start:]:
        if ln.startswith("## "):
            break
        collected.append(ln)
    return "\n".join(collected).strip()


def analyze_example_section(body: str) -> tuple[bool, bool, bool]:
    """Return (has_example_heading, has_non_empty_fence, fences_balanced_in_section).

    We slice from `## Example Usage` to the next `## ` heading. Inside that
    slice we count ``` fences. Odd = broken. We also check that at least one
    fenced block has non-whitespace content.
    """
    lines = body.splitlines()
    in_section = False
    section_lines: list[str] = []
    for ln in lines:
        if ln.startswith("## "):
            if ln.strip().lower().startswith("## example usage"):
                in_section = True
                continue
            if in_section:
                break
        if in_section:
            section_lines.append(ln)
    if not in_section and "## Example Usage" not in body:
        return False, False, True

    # Count opening fences vs closing; track whether any block has content.
    fences = [i for i, ln in enumerate(section_lines) if ln.lstrip().startswith("```")]
    fences_balanced = len(fences) % 2 == 0
    has_non_empty = False
    for a, b in zip(fences[0::2], fences[1::2]):
        content = "\n".join(section_lines[a + 1:b]).strip()
        if content:
            has_non_empty = True
            break
    return True, has_non_empty, fences_balanced


def iter_schema_attribute_lines(body: str):
    """Yield attribute lines from all ## Schema and ### Nested Schema sections.

    tfplugindocs emits sections like ### Required / ### Optional / ### Read-Only
    under ## Schema, plus ### Nested Schema for `foo` with Required:/Optional:/
    Read-Only: subheaders. We treat any `- `name` (Meta) rest` line inside
    these sections as an attribute line.

    We skip attribute lines under a `Read-Only:` subheader *inside* a
    `### Nested Schema for ...` section because tfplugindocs strips the
    Description field from those, making the defect uncorrectable at the
    schema level. Top-level `### Read-Only` under `## Schema` renders the
    Description correctly (e.g. kubeconfig), so we do NOT skip those.
    """
    lines = body.splitlines()
    in_schema = False
    in_nested = False
    in_nested_read_only = False
    in_nested_timeouts = False
    for ln in lines:
        if ln.startswith("## "):
            in_schema = ln.strip().lower().startswith("## schema")
            in_nested = False
            in_nested_read_only = False
            in_nested_timeouts = False
            continue
        if ln.startswith("### Nested Schema"):
            in_schema = True
            in_nested = True
            in_nested_read_only = False
            in_nested_timeouts = "`timeouts`" in ln.lower() or "timeouts`" in ln.lower()
            continue
        if ln.startswith("### "):
            in_nested = False
            in_nested_read_only = False
            in_nested_timeouts = False
        if not in_schema:
            continue
        stripped = ln.strip()
        if in_nested and stripped.endswith(":") and stripped.rstrip(":").lower() in {
            "required", "optional", "read-only",
        }:
            in_nested_read_only = stripped.lower() == "read-only:"
            continue
        # Skip nested Read-Only attrs (tfplugindocs strips their descriptions)
        # and skip `timeouts` nested blocks (SDK-generated, not author-editable).
        if in_nested_read_only or in_nested_timeouts:
            continue
        m = ATTR_LINE_RE.match(ln)
        if m:
            yield m.group("name"), m.group("meta"), m.group("rest")


def score_attribute(name: str, meta: str, rest: str) -> tuple[int, list[str]]:
    points = 0
    defects: list[str] = []
    rest_stripped = rest.strip()

    # A nested-block reference isn't a missing description, the real
    # description lives in the nested section. Skip.
    if NESTED_REF_RE.match(rest):
        return 0, []

    if not rest_stripped:
        points += W_ATTR_EMPTY_DESC
        defects.append(f"empty_desc:{name}")
        return points, defects

    if len(rest_stripped) < STUB_LEN:
        points += W_ATTR_STUB_DESC
        defects.append(f"stub_desc:{name}")

    # Accept ')' because tfplugindocs appends cross-refs like
    # "(see [below for nested schema](#...))" onto list/block attribute lines,
    # even when the underlying Description ends cleanly.
    if rest_stripped and rest_stripped[-1] not in ".!?`)":
        points += W_ATTR_NO_TERMINAL_PUNCT
        defects.append(f"no_punct:{name}")

    # Name-echo: description is just the field name with underscores->spaces,
    # optionally titled. e.g. name=cloud_account_id, desc="Cloud account id"
    normalized_name = name.replace("_", " ").strip().lower()
    normalized_rest = re.sub(r"[^a-z ]", "", rest_stripped.lower()).strip()
    if normalized_rest and normalized_rest == normalized_name:
        points += W_ATTR_NAME_ECHO
        defects.append(f"name_echo:{name}")

    # Lists/Sets/Maps without element semantics: description doesn't mention
    # what's inside. Heuristic: no word from {list, of, each, element, entry,
    # item, string, number, object, block, tag, uid, id, name}.
    if LIST_TYPE_RE.match(meta):
        hint_words = {
            "list", "set", "map", "of", "each", "element", "entry", "item",
            "string", "number", "object", "block", "tag", "uid", "id",
            "name", "key", "value", "pair",
        }
        tokens = set(re.findall(r"[a-zA-Z]+", rest_stripped.lower()))
        if not (tokens & hint_words):
            points += W_ATTR_LIST_NO_ELEMENT_SEMANTICS
            defects.append(f"list_no_element_semantics:{name}")

    return points, defects


def score_page(path: Path) -> PageReport:
    text = path.read_text()
    rep = PageReport(path=str(path.relative_to(REPO_ROOT)))

    fm, body = extract_frontmatter(text)
    fm_desc = parse_frontmatter_description(fm)
    if len(fm_desc) < 20:
        rep.score += W_PAGE_EMPTY_FRONTMATTER_DESC
        rep.defects.append("empty_frontmatter_desc")

    top_desc = find_top_description(body)
    if not top_desc or len(top_desc) < 20:
        rep.score += W_PAGE_EMPTY_TOP_DESC
        rep.defects.append("empty_top_desc")

    has_example, has_content, balanced = analyze_example_section(body)
    if not has_example:
        rep.score += W_PAGE_NO_EXAMPLE
        rep.defects.append("no_example_section")
    else:
        if not has_content:
            rep.score += W_PAGE_EMPTY_EXAMPLE
            rep.defects.append("empty_example_block")
        if not balanced:
            rep.score += W_PAGE_BROKEN_FENCES
            rep.defects.append("broken_example_fences")

    for name, meta, rest in iter_schema_attribute_lines(body):
        rep.attrs_total += 1
        pts, attr_defects = score_attribute(name, meta, rest)
        if pts:
            rep.attrs_defective += 1
            rep.score += pts
            rep.defects.extend(attr_defects)

    return rep


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--top", type=int, default=0,
                    help="Print the N worst-scoring pages")
    ap.add_argument("--json-only", action="store_true",
                    help="Suppress human-readable output; still writes score.json")
    args = ap.parse_args()

    pages: list[PageReport] = []
    for d in DOCS_DIRS:
        if not d.exists():
            continue
        for p in sorted(d.glob("*.md")):
            pages.append(score_page(p))

    total = sum(p.score for p in pages)
    failing = sum(1 for p in pages if p.score > 0)
    total_attrs = sum(p.attrs_total for p in pages)
    defective_attrs = sum(p.attrs_defective for p in pages)

    summary = {
        "total_score": total,
        "pages_total": len(pages),
        "pages_failing": failing,
        "attrs_total": total_attrs,
        "attrs_defective": defective_attrs,
        "pages": [
            {
                "path": p.path,
                "score": p.score,
                "defects": p.defects,
                "attrs_total": p.attrs_total,
                "attrs_defective": p.attrs_defective,
            }
            for p in pages
        ],
    }
    OUT_FILE.write_text(json.dumps(summary, indent=2) + "\n")

    if args.json_only:
        return 0

    print(f"total_score:     {total}")
    print(f"pages_total:     {len(pages)}")
    print(f"pages_failing:   {failing}")
    print(f"attrs_total:     {total_attrs}")
    print(f"attrs_defective: {defective_attrs}")
    print(f"wrote:           {OUT_FILE.relative_to(REPO_ROOT)}")

    if args.top > 0:
        print()
        print(f"top {args.top} worst pages:")
        worst = sorted(pages, key=lambda p: p.score, reverse=True)[:args.top]
        for p in worst:
            print(f"  {p.score:>5}  {p.path}  ({p.attrs_defective}/{p.attrs_total} attrs, "
                  f"{len(p.defects)} defects)")

    return 0


if __name__ == "__main__":
    sys.exit(main())
