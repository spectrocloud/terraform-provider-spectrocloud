#!/usr/bin/env python3
"""Generate a static, searchable HTML docs/examples site for the provider.

Standard library only — no pip install, no virtualenv. Run it, then open
docsite/output/index.html directly in a browser (works with no server, no
GitHub Pages, nothing but a clone of this repo):

    python3 docsite/generate_site.py

Output is regenerated in full from examples/ and docs/ every run, so the
site can never drift from what's actually in the repo. The generated
docsite/output/ directory IS committed (there's no build step for a
visitor to run) — re-run this script and commit the diff whenever
examples/ or docs/ changes.
"""
from __future__ import annotations

import html
import json
import re
import shutil
from pathlib import Path

DOCSITE_DIR = Path(__file__).resolve().parent
REPO_ROOT = DOCSITE_DIR.parent
EXAMPLES_DIR = REPO_ROOT / "examples"
DOCS_DIR = REPO_ROOT / "docs"
ASSETS_SRC = DOCSITE_DIR / "assets"
OUTPUT_DIR = DOCSITE_DIR / "output"

# examples/ subfolders that are NOT example "categories" -- resources and
# data-sources are per-resource snippets already shown inline on their
# Reference page, not standalone examples to list here.
NON_CATEGORY_DIRS = {"data-sources", "resources"}

# Preferred display order for known category folder names; anything else
# found under examples/ is appended alphabetically. This list is advisory
# only -- discover_categories() always reflects what's actually on disk, so
# a renamed/added/removed examples/ subfolder can't go stale here.
PREFERRED_CATEGORY_ORDER = [
    "e2e", "end-to-end-usecases", "tutorials", "import", "provider",
]

EXAMPLE_CATEGORIES: list[tuple[str, str]] = []


def discover_categories() -> list[tuple[str, str]]:
    present = {
        p.name for p in EXAMPLES_DIR.iterdir()
        if p.is_dir() and p.name not in NON_CATEGORY_DIRS
    }
    ordered = [name for name in PREFERRED_CATEGORY_ORDER if name in present]
    ordered += sorted(present - set(ordered))
    return [(name, humanize(name)) for name in ordered]

CODE_EXTENSIONS = {".tf", ".yaml", ".yml"}

ACRONYMS = {
    "aws", "gcp", "gke", "eks", "aks", "rbac", "maas", "pcg", "sso", "dns",
    "ippool", "capi", "vsphere", "vm", "vpc", "cli", "api", "ssh", "e2e",
}

GITHUB_BLOB = "https://github.com/spectrocloud/terraform-provider-spectrocloud/tree/main"

search_entries: list[dict] = []


# --------------------------------------------------------------------------
# small helpers
# --------------------------------------------------------------------------

def humanize(name: str) -> str:
    words = re.split(r"[_\-]+", name.replace("(", " ").replace(")", " "))
    parts = []
    for word in words:
        if not word:
            continue
        lower = word.lower()
        parts.append(lower.upper() if lower in ACRONYMS else lower.capitalize())
    return " ".join(parts) if parts else name


def slugify(name: str) -> str:
    slug = re.sub(r"[^a-z0-9]+", "-", name.lower())
    return slug.strip("-") or "page"


def first_paragraph(text: str, max_len: int = 180) -> str:
    """First non-heading paragraph, with wrapped lines joined into one sentence."""
    words: list[str] = []
    started = False
    for line in text.splitlines():
        stripped = line.strip()
        if not stripped:
            if started:
                break
            continue
        if stripped.startswith("#"):
            continue
        started = True
        words.append(stripped)
    para = " ".join(words)
    if len(para) > max_len:
        para = para[: max_len - 1].rstrip() + "…"
    return para


# --------------------------------------------------------------------------
# tiny markdown -> html converter (headers, bold/italic, inline code, links,
# fenced code blocks, lists, paragraphs) -- good enough for these READMEs
# --------------------------------------------------------------------------

INLINE_CODE_RE = re.compile(r"`([^`]+)`")
BOLD_RE = re.compile(r"\*\*([^*]+)\*\*")
ITALIC_RE = re.compile(r"(?<!\*)\*([^*]+)\*(?!\*)")
LINK_RE = re.compile(r"\[([^\]]+)\]\(([^)]+)\)")


def inline_md(text: str) -> str:
    text = html.escape(text)
    text = INLINE_CODE_RE.sub(lambda m: f"<code>{m.group(1)}</code>", text)
    text = BOLD_RE.sub(lambda m: f"<strong>{m.group(1)}</strong>", text)
    text = ITALIC_RE.sub(lambda m: f"<em>{m.group(1)}</em>", text)
    text = LINK_RE.sub(lambda m: f'<a href="{m.group(2)}">{m.group(1)}</a>', text)
    return text


def md_to_html(md_text: str) -> str:
    lines = md_text.splitlines()
    out: list[str] = []
    para: list[str] = []
    list_items: list[str] = []
    list_tag: str | None = None
    in_code = False
    code_lines: list[str] = []
    code_lang = ""

    def flush_para():
        if para:
            out.append(f"<p>{inline_md(' '.join(para))}</p>")
            para.clear()

    def flush_list():
        nonlocal list_tag
        if list_items:
            out.append(f"<{list_tag}>")
            out.extend(f"<li>{inline_md(item)}</li>" for item in list_items)
            out.append(f"</{list_tag}>")
            list_items.clear()
        list_tag = None

    for line in lines:
        if line.strip().startswith("```"):
            if in_code:
                escaped = html.escape("\n".join(code_lines))
                cls = f' class="language-{code_lang}"' if code_lang else ""
                out.append(f"<pre><code{cls}>{escaped}</code></pre>")
                code_lines = []
                in_code = False
            else:
                flush_para()
                flush_list()
                in_code = True
                code_lang = line.strip()[3:].strip()
            continue

        if in_code:
            code_lines.append(line)
            continue

        stripped = line.strip()

        header_match = re.match(r"^(#{1,4})\s+(.*)$", stripped)
        if header_match:
            flush_para()
            flush_list()
            level = len(header_match.group(1))
            out.append(f"<h{level}>{inline_md(header_match.group(2))}</h{level}>")
            continue

        bullet_match = re.match(r"^[-*]\s+(.*)$", stripped)
        number_match = re.match(r"^\d+\.\s+(.*)$", stripped)
        if bullet_match:
            flush_para()
            if list_tag != "ul":
                flush_list()
            list_tag = "ul"
            list_items.append(bullet_match.group(1))
            continue
        if number_match:
            flush_para()
            if list_tag != "ol":
                flush_list()
            list_tag = "ol"
            list_items.append(number_match.group(1))
            continue

        if not stripped:
            flush_para()
            flush_list()
            continue

        para.append(stripped)

    flush_para()
    flush_list()
    return "\n".join(out)


def code_lang(suffix: str) -> str:
    return "yaml" if suffix in (".yaml", ".yml") else "hcl"


def render_code_block(path: Path, label: str) -> str:
    escaped = html.escape(path.read_text().rstrip())
    return (
        f'<div class="file-label">{html.escape(label)}</div>'
        f'<pre><code class="language-{code_lang(path.suffix)}">{escaped}</code></pre>'
    )


# --------------------------------------------------------------------------
# page shell
# --------------------------------------------------------------------------

NAV_ITEMS = [
    ("Home", "index.html"),
    ("Getting Started", "getting-started.html"),
    ("Examples", "examples/index.html"),
    ("Reference", "reference/index.html"),
]


def render_page(
    *,
    out_path: Path,
    title: str,
    active: str,
    breadcrumb: list[tuple[str, str]],
    content_html: str,
    sidebar_html: str = "",
) -> None:
    depth = len(out_path.relative_to(OUTPUT_DIR).parts) - 1
    rel_root = "../" * depth

    nav_html = "\n".join(
        f'<a href="{rel_root}{href}" class="{"active" if label == active else ""}">{label}</a>'
        for label, href in NAV_ITEMS
    )

    crumbs_html = " / ".join(
        f'<a href="{rel_root}{href}">{html.escape(label)}</a>' if href else html.escape(label)
        for label, href in breadcrumb
    )

    sidebar_block = f'<aside class="sidebar">{sidebar_html}</aside>' if sidebar_html else ""

    doc = f"""<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{html.escape(title)} · Terraform Provider for Spectro Cloud</title>
<link rel="stylesheet" href="{rel_root}assets/style.css">
</head>
<body data-rel-root="{rel_root}">
<header class="site-header">
  <div class="site-title">Terraform Provider for Spectro Cloud</div>
  <nav class="site-nav">{nav_html}</nav>
  <div class="search-box">
    <input id="search-input" type="text" placeholder="Search examples &amp; reference..." autocomplete="off">
    <div id="search-results" class="search-results"></div>
  </div>
</header>
<div class="layout">
  {sidebar_block}
  <main class="content">
    <div class="breadcrumb">{crumbs_html}</div>
    {content_html}
  </main>
</div>
<footer class="site-footer">
  Generated from <code>examples/</code> and <code>docs/</code> by <code>docsite/generate_site.py</code>.
  <a href="https://github.com/spectrocloud/terraform-provider-spectrocloud">View on GitHub</a>
</footer>
<script src="{rel_root}search-index.js"></script>
<script src="{rel_root}assets/search.js" defer></script>
</body>
</html>
"""
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(doc)


def add_search_entry(title: str, url: Path, section: str, snippet: str) -> None:
    search_entries.append({
        "title": title,
        "url": url.relative_to(OUTPUT_DIR).as_posix(),
        "section": section,
        "snippet": snippet[:200],
    })


# --------------------------------------------------------------------------
# examples
# --------------------------------------------------------------------------

def is_wanted_code_file(p: Path) -> bool:
    # terraform.tfvars (real values, always gitignored/local-only) is never
    # embedded -- only the tracked, placeholder-only terraform.template.tfvars.
    if p.suffix == ".tfvars":
        return "template" in p.name
    return p.suffix in CODE_EXTENSIONS


def collect_code_files(directory: Path) -> list[Path]:
    files = [p for p in sorted(directory.iterdir()) if p.is_file() and is_wanted_code_file(p)]
    for sub in sorted(p for p in directory.iterdir() if p.is_dir()):
        files.extend(p for p in sorted(sub.iterdir()) if p.is_file() and is_wanted_code_file(p))
    return files


def example_sidebar(active_category: str) -> str:
    items = "\n".join(
        f'<li><a href="{"../" if active_category else ""}{slug}/index.html" '
        f'class="{"active" if slug == active_category else ""}">{title}</a></li>'
        for slug, title in EXAMPLE_CATEGORIES
    )
    return f'<h3>Examples</h3><ul>{items}</ul>'


def build_example_category(category: str, title: str, cat_dir: Path) -> None:
    out_cat = OUTPUT_DIR / "examples" / category
    subdirs = sorted(p for p in cat_dir.iterdir() if p.is_dir())
    examples = subdirs if subdirs else [cat_dir]

    cards = []
    for example_dir in examples:
        name = example_dir.name
        slug = slugify(name)
        page_title = humanize(name)
        readme = example_dir / "README.md"
        desc = first_paragraph(readme.read_text()) if readme.exists() and readme.stat().st_size > 0 else ""

        body_parts = []
        if readme.exists() and readme.stat().st_size > 0:
            body = readme.read_text().strip()
            body = re.sub(r"^#\s+.*\n+", "", body, count=1)
            body_parts.append(md_to_html(body))

        code_files = collect_code_files(example_dir)
        if code_files:
            body_parts.append("<h2>Configuration Files</h2>")
            for f in code_files:
                rel = f.relative_to(example_dir)
                body_parts.append(render_code_block(f, str(rel)))

        src_url = f"{GITHUB_BLOB}/examples/{category}/{example_dir.name}"
        body_parts.append(f'<p><a href="{src_url}">View source on GitHub</a></p>')

        content = f"<h1>{html.escape(page_title)}</h1>\n" + "\n".join(body_parts)
        out_path = out_cat / f"{slug}.html"
        render_page(
            out_path=out_path,
            title=page_title,
            active="Examples",
            breadcrumb=[("Home", "index.html"), ("Examples", "examples/index.html"), (title, f"examples/{category}/index.html"), (page_title, "")],
            content_html=content,
            sidebar_html=example_sidebar(category),
        )
        add_search_entry(page_title, out_path, f"Examples / {title}", desc)
        cards.append((page_title, f"{slug}.html", desc))

    card_items = "\n".join(
        f'<li><a href="{slug}">{html.escape(t)}</a>{f"<p>{html.escape(d)}</p>" if d else ""}</li>'
        for t, slug, d in cards
    )
    index_content = f"<h1>{html.escape(title)}</h1>\n<ul class=\"card-grid\">{card_items}</ul>"
    index_path = out_cat / "index.html"
    render_page(
        out_path=index_path,
        title=title,
        active="Examples",
        breadcrumb=[("Home", "index.html"), ("Examples", "examples/index.html"), (title, "")],
        content_html=index_content,
        sidebar_html=example_sidebar(category),
    )
    add_search_entry(title, index_path, "Examples", f"{len(cards)} examples")


def build_examples_index() -> None:
    rows = "\n".join(
        f'<li><a href="{slug}/index.html">{title}</a></li>' for slug, title in EXAMPLE_CATEGORIES
    )
    content = f"""<h1>Examples</h1>
<p>Generated directly from the repository's <a href="{GITHUB_BLOB}/examples">examples/</a> folder —
every code block is the real, current file content, not a hand-copied snippet.</p>
<ul class="card-grid">{rows}</ul>"""
    out_path = OUTPUT_DIR / "examples" / "index.html"
    render_page(
        out_path=out_path,
        title="Examples",
        active="Examples",
        breadcrumb=[("Home", "index.html"), ("Examples", "")],
        content_html=content,
        sidebar_html=example_sidebar(""),
    )
    add_search_entry("Examples", out_path, "Examples", "Browse all provider examples")


def build_examples() -> None:
    global EXAMPLE_CATEGORIES
    EXAMPLE_CATEGORIES = discover_categories()

    if (OUTPUT_DIR / "examples").exists():
        shutil.rmtree(OUTPUT_DIR / "examples")
    build_examples_index()
    for category, title in EXAMPLE_CATEGORIES:
        cat_dir = EXAMPLES_DIR / category
        if cat_dir.is_dir():
            build_example_category(category, title, cat_dir)


# --------------------------------------------------------------------------
# reference (resources / data-sources), synced from docs/
# --------------------------------------------------------------------------

FRONTMATTER_RE = re.compile(r"\A---\n(.*?)\n---\n", re.DOTALL)
DESC_RE = re.compile(r"description:\s*\|-\s*\n((?:\s+.*\n?)+)")
GENERIC_CODE_BLOCK_RE = re.compile(r"```(\w*)\n(.*?)```", re.DOTALL)


def registry_md_to_html(body: str) -> str:
    """Render tfplugindocs-generated markdown: strip frontmatter, keep code fences."""
    parts = []
    last = 0
    for m in GENERIC_CODE_BLOCK_RE.finditer(body):
        parts.append(md_to_html(body[last:m.start()]))
        lang = m.group(1) or "hcl"
        code = html.escape(m.group(2).rstrip())
        parts.append(f'<pre><code class="language-{lang}">{code}</code></pre>')
        last = m.end()
    parts.append(md_to_html(body[last:]))
    return "\n".join(parts)
    parts = []
    last = 0
    for m in GENERIC_CODE_BLOCK_RE.finditer(body):
        parts.append(md_to_html(body[last:m.start()]))
        lang = m.group(1) or "hcl"
        code = html.escape(m.group(2).rstrip())
        parts.append(f'<pre><code class="language-{lang}">{code}</code></pre>')
        last = m.end()
    parts.append(md_to_html(body[last:]))
    return "\n".join(parts)


def reference_sidebar(section: str) -> str:
    resources = sorted((DOCS_DIR / "resources").glob("*.md")) if (DOCS_DIR / "resources").is_dir() else []
    data_sources = sorted((DOCS_DIR / "data-sources").glob("*.md")) if (DOCS_DIR / "data-sources").is_dir() else []

    def links(files: list[Path], sub: str) -> str:
        return "\n".join(
            f'<li><a href="{"../" if section else ""}{sub}/{f.stem}.html">{humanize(f.stem)}</a></li>'
            for f in files
        )

    return (
        "<h3>Resources</h3><ul>" + links(resources, "resources") + "</ul>"
        "<h3>Data Sources</h3><ul>" + links(data_sources, "data-sources") + "</ul>"
    )


def build_reference_section(section: str, title: str, source: Path) -> list[tuple[str, str, str]]:
    out_dir = OUTPUT_DIR / "reference" / section
    entries = []
    for md_file in sorted(source.glob("*.md")):
        raw = md_file.read_text()
        fm_match = FRONTMATTER_RE.match(raw)
        desc = ""
        body = raw
        if fm_match:
            body = raw[fm_match.end():]
            desc_match = DESC_RE.search(fm_match.group(1))
            if desc_match:
                desc = " ".join(line.strip() for line in desc_match.group(1).splitlines()).strip()

        page_title = humanize(md_file.stem)
        body = re.sub(r"^#\s+.*\n+", "", body.strip(), count=1)
        content = f"<h1>{html.escape(page_title)}</h1>\n" + registry_md_to_html(body)
        out_path = out_dir / f"{md_file.stem}.html"
        render_page(
            out_path=out_path,
            title=page_title,
            active="Reference",
            breadcrumb=[("Home", "index.html"), ("Reference", "reference/index.html"), (title, f"reference/{section}/index.html"), (page_title, "")],
            content_html=content,
            sidebar_html=reference_sidebar(section),
        )
        add_search_entry(page_title, out_path, f"Reference / {title}", desc)
        entries.append((page_title, f"{md_file.stem}.html", desc))

    card_items = "\n".join(
        f'<li><a href="{slug}">{html.escape(t)}</a>{f"<p>{html.escape(d)}</p>" if d else ""}</li>'
        for t, slug, d in entries
    )
    index_content = f"<h1>{html.escape(title)}</h1>\n<ul class=\"card-grid\">{card_items}</ul>"
    index_path = out_dir / "index.html"
    render_page(
        out_path=index_path,
        title=title,
        active="Reference",
        breadcrumb=[("Home", "index.html"), ("Reference", "reference/index.html"), (title, "")],
        content_html=index_content,
        sidebar_html=reference_sidebar(section),
    )
    add_search_entry(title, index_path, "Reference", f"{len(entries)} entries")
    return entries


def build_reference() -> None:
    if (OUTPUT_DIR / "reference").exists():
        shutil.rmtree(OUTPUT_DIR / "reference")

    resource_entries = build_reference_section("resources", "Resources", DOCS_DIR / "resources") if (DOCS_DIR / "resources").is_dir() else []
    data_source_entries = build_reference_section("data-sources", "Data Sources", DOCS_DIR / "data-sources") if (DOCS_DIR / "data-sources").is_dir() else []

    content = f"""<h1>Reference</h1>
<p>Synced from <a href="{GITHUB_BLOB}/docs/resources">docs/resources</a> and
<a href="{GITHUB_BLOB}/docs/data-sources">docs/data-sources</a> — the same content published to the
<a href="https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs">Terraform Registry</a>,
generated via <code>tfplugindocs</code>. This is a mirror for browsing/search, not a second source of truth.</p>
<ul class="card-grid">
  <li><a href="resources/index.html">Resources</a><p>{len(resource_entries)} resources</p></li>
  <li><a href="data-sources/index.html">Data Sources</a><p>{len(data_source_entries)} data sources</p></li>
</ul>"""
    out_path = OUTPUT_DIR / "reference" / "index.html"
    render_page(
        out_path=out_path,
        title="Reference",
        active="Reference",
        breadcrumb=[("Home", "index.html"), ("Reference", "")],
        content_html=content,
        sidebar_html=reference_sidebar(""),
    )
    add_search_entry("Reference", out_path, "Reference", "Resource and data source schema reference")


# --------------------------------------------------------------------------
# static hand-authored pages
# --------------------------------------------------------------------------

def build_home() -> None:
    content = f"""<h1>Terraform Provider for Spectro Cloud</h1>
<p>The Spectro Cloud Terraform provider lets you manage Palette and Palette VerteX — SaaS or on-prem —
as infrastructure as code: cloud accounts, cluster profiles, clusters, and more.</p>
<p>This site collects working, copy-paste-runnable <a href="examples/index.html">examples</a> pulled directly
from the provider repository's <code>examples/</code> folder, plus the full
<a href="reference/index.html">resource and data source reference</a>.</p>
<h2>Where to start</h2>
<ul>
  <li>New to the provider? Start with <a href="getting-started.html">Getting Started</a>.</li>
  <li>Looking for a working example to copy? Browse <a href="examples/index.html">Examples</a>.</li>
  <li>Need the schema for a specific resource or data source? Jump to <a href="reference/index.html">Reference</a>.</li>
</ul>
<h2>Links</h2>
<ul>
  <li><a href="{GITHUB_BLOB}">GitHub repository</a></li>
  <li><a href="https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs">Terraform Registry page</a></li>
  <li><a href="https://github.com/spectrocloud/terraform-provider-spectrocloud/discussions">Support / discussions</a></li>
</ul>"""
    out_path = OUTPUT_DIR / "index.html"
    render_page(out_path=out_path, title="Home", active="Home", breadcrumb=[("Home", "")], content_html=content)
    add_search_entry("Home", out_path, "Home", "Terraform provider for Spectro Cloud / Palette")


def build_getting_started() -> None:
    content = """<h1>Getting Started</h1>
<h2>Pre-requisites</h2>
<ul>
  <li>A Spectro Cloud account (<a href="https://www.spectrocloud.com/free-trial/">sign up for a free trial</a>)</li>
  <li>Terraform 0.13+</li>
  <li>kubectl 1.16+ (for interacting with provisioned clusters)</li>
</ul>
<h2>Configure the provider</h2>
<p>Create a <code>providers.tf</code> file:</p>
<pre><code class="language-hcl">terraform {
  required_providers {
    spectrocloud = {
      source  = "spectrocloud/spectrocloud"
      version = "&gt;= 0.1"
    }
  }
}

provider "spectrocloud" {
  host    = var.sc_host
  api_key = var.sc_api_key
}</code></pre>
<p>Populate <code>sc_host</code> and <code>sc_api_key</code> (an
<a href="https://docs.spectrocloud.com/user-management/authentication/api-key/create-api-key">API key</a>)
via a <code>terraform.tfvars</code> file, or use environment variables instead:</p>
<pre><code class="language-shell">export SPECTROCLOUD_HOST=api.spectrocloud.com
export SPECTROCLOUD_APIKEY=5b7aad.........</code></pre>
<p>Then:</p>
<pre><code class="language-shell">terraform init &amp;&amp; terraform apply</code></pre>
<h2>Next steps</h2>
<ul>
  <li>Browse <a href="examples/index.html">Examples</a> for full working configurations by use case.</li>
  <li>See the <a href="reference/index.html">Reference</a> for every resource and data source the provider supports.</li>
  <li>Full provider configuration details (environment variables, feature flags, import) are documented on the
      <a href="https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs">Terraform Registry page</a>.</li>
</ul>"""
    out_path = OUTPUT_DIR / "getting-started.html"
    render_page(out_path=out_path, title="Getting Started", active="Getting Started", breadcrumb=[("Home", "index.html"), ("Getting Started", "")], content_html=content)
    add_search_entry("Getting Started", out_path, "Home", "Provider setup, authentication, first apply")


# --------------------------------------------------------------------------
# examples/README.md -- a plain-markdown categorized index, generated from
# the same source of truth, so it stays usable directly on github.com (no
# clone/download needed) without ever drifting from the HTML site above.
# --------------------------------------------------------------------------

def build_examples_readme() -> None:
    lines = [
        "# Examples",
        "",
        "This directory contains examples that are mostly used for documentation, but can also be run/tested",
        "manually via the Terraform CLI.",
        "",
        "The category and example lists below are generated by [`docsite/generate_site.py`](../docsite/generate_site.py) —",
        "edit that script (or an example's own `README.md`), not this file directly.",
        "",
        "A browsable, searchable HTML version of everything below (plus the full resource/data source reference)",
        "lives in [`docsite/`](../docsite/README.md) — clone this repo and open `docsite/output/index.html`.",
        "",
        "## Categories",
        "",
    ]

    for category, title in EXAMPLE_CATEGORIES:
        cat_dir = EXAMPLES_DIR / category
        if not cat_dir.is_dir():
            continue
        lines.append(f"### {title} (`{category}/`)")
        lines.append("")
        subdirs = sorted(p for p in cat_dir.iterdir() if p.is_dir())
        examples = subdirs if subdirs else [cat_dir]
        for example_dir in examples:
            rel = example_dir.relative_to(EXAMPLES_DIR).as_posix()
            readme = example_dir / "README.md"
            desc = first_paragraph(readme.read_text()) if readme.exists() and readme.stat().st_size > 0 else ""
            entry = f"- [`{rel}/`]({rel}/)"
            if desc:
                entry += f" — {desc}"
            lines.append(entry)
        lines.append("")

    lines.append("## Resource & data source examples")
    lines.append("")
    lines.append("- [`resources/<resource name>/`](resources/) — minimal example backing that resource's Registry doc page")
    lines.append("- [`data-sources/<data source name>/`](data-sources/) — minimal example backing that data source's Registry doc page")
    lines.append("")

    (EXAMPLES_DIR / "README.md").write_text("\n".join(lines))


# --------------------------------------------------------------------------
# main
# --------------------------------------------------------------------------

def main() -> None:
    if OUTPUT_DIR.exists():
        shutil.rmtree(OUTPUT_DIR)
    OUTPUT_DIR.mkdir(parents=True)

    shutil.copytree(ASSETS_SRC, OUTPUT_DIR / "assets")

    build_home()
    build_getting_started()
    build_examples()
    build_reference()
    build_examples_readme()

    (OUTPUT_DIR / "search-index.js").write_text(
        "const SEARCH_INDEX = " + json.dumps(search_entries, indent=None) + ";\n"
    )

    print(f"Generated {len(search_entries)} pages into {OUTPUT_DIR}")
    print(f"Updated {EXAMPLES_DIR / 'README.md'}")
    print(f"Open: {OUTPUT_DIR / 'index.html'}")


if __name__ == "__main__":
    main()
