# autoresearch: terraform-provider-spectrocloud docs

Autonomous research loop that improves the schema-derived documentation for
`terraform-provider-spectrocloud`. Same spirit as the nanochat autoresearch:
single measurable metric, fixed-scope editable surface, keep-or-discard branch
dance. No neural net — the "experiment" is a patch to schema descriptions
and/or template files, the "training" is regenerating docs with
`tfplugindocs`, and the "val_bpb" is a documentation-quality score.

## Setup

Work with the user to:

1. **Agree on a run tag**: propose one based on today's date (e.g. `apr17`).
   Branch `autoresearch-docs/<tag>` must not already exist.
2. **Create the branch**: `git checkout -b autoresearch-docs/<tag>` from the
   current default branch (ask the user which one — likely `main` or
   `master`).
3. **Read the in-scope files for context**:
   - `docs/` and `templates/` — the generated docs and their source templates.
   - `spectrocloud/resource_*.go` and `spectrocloud/data_source_*.go` — the
     schema definitions. The `Description` field on every `*schema.Schema`
     is the primary thing you will edit.
   - `main.go` — provider registration, to enumerate which resources and
     data sources exist.
   - `GNUmakefile` — check whether there is already a `docs` target. If not,
     use `go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate`
     directly (the dependency is declared in `tools/tools.go`).
4. **Install tfplugindocs once**: `go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs`.
   Verify `tfplugindocs --version` works.
5. **Build a baseline score** (see "Metric" below). Write
   `autoresearch_docs/results.tsv` with just the header row. The baseline
   will be the first recorded entry.
6. **Confirm and go**. Show the user the baseline score and the top-5
   worst-scoring pages. Then start the loop.

## Scope — what you CAN and CANNOT touch

**You CAN edit:**
- `Description:` string literals in `spectrocloud/*.go` (schema definitions,
  for resources and data sources alike — including nested schemas).
- `templates/**/*.md.tmpl` — the tfplugindocs templates.
- `templates/index.md.tmpl` — provider-level doc.
- `examples/**` ONLY to fix syntactically broken snippets that make a doc
  page render incorrectly. Do not invent new examples there; prefer to add
  example blocks inside templates.

**You CANNOT edit:**
- `docs/**` directly. These are generated. If you edit them by hand the next
  `tfplugindocs generate` will wipe your work. Always edit schemas/templates
  and regenerate.
- Any non-Description field of the schema: no changing `Type`, `Required`,
  `Optional`, `Computed`, `ForceNew`, `Default`, `ValidateFunc`,
  `DiffSuppressFunc`, `Elem`, etc. This is a docs run, not a behavior run.
- Resource CRUD functions, expand/flatten functions, tests, or any
  non-schema logic.
- `prepare.py`-equivalent files: the scoring script at
  `autoresearch_docs/score.py` (once created) is read-only. It defines the
  ground-truth metric.
- Dependencies: do not add Go modules, do not modify `go.mod` / `go.sum`.

If you find a bug in the schema (wrong Required/Optional, missing field) —
do NOT fix it in this run. Note it in `autoresearch_docs/notes.md` for the
human.

## The metric

Lower is better. The score is a weighted sum of documentation defects,
computed over all pages under `docs/resources/` and `docs/data-sources/`
after running `tfplugindocs generate`.

For every schema attribute that appears in a generated doc page:

| defect                                                    | points |
|-----------------------------------------------------------|--------|
| `Description` is empty                                    |  10    |
| `Description` is < 15 chars (effectively a stub)          |   5    |
| `Description` ends without punctuation                    |   1    |
| `Description` is just the field name reformatted          |   4    |
| `Description` references a type/enum that doesn't exist   |   6    |
| `Description` of a list/set field omits element semantics |   2    |

Plus per-page defects:

| defect                                                    | points |
|-----------------------------------------------------------|--------|
| Page has no `## Example Usage` section                    |  15    |
| Page has `## Example Usage` but the fenced block is empty |  10    |
| Example block does not parse as HCL (rough check)         |   8    |
| Page `description:` frontmatter is empty or < 20 chars    |   5    |
| Top-level `{{.Description}}` renders to empty             |   8    |

`tfplugindocs generate` MUST succeed (exit 0). If it fails, the score is
`CRASH` and the run is discarded.

`autoresearch_docs/score.py` is the ground truth. Build it in setup if it
doesn't exist. It must:
- Walk `docs/resources/` and `docs/data-sources/`.
- For attribute-level defects, cross-reference the parsed schema. The
  easiest path: run `terraform providers schema -json` against a trivial
  config that uses the locally built provider, OR parse the `## Schema`
  sections of the generated markdown (simpler, preferred).
- Emit total score + per-page breakdown to stdout, and a machine-readable
  summary to `autoresearch_docs/score.json`.

When in doubt about a rule, keep the scorer strict and mechanical. Do NOT
have the scorer call an LLM. Determinism matters — two runs on the same
tree must produce identical scores.

## Experiment step

Each experiment is a focused edit. A good experiment changes ONE thing:

- "fill empty descriptions on `spectrocloud_cluster_gcp`"
- "rewrite stub descriptions in `cluster_common_fields.go`"
- "add Example Usage block to `templates/resources/password_policy.md.tmpl`"
- "standardize punctuation across all data-source descriptions"

Bad experiments: sweeping stylistic rewrites across 40 files in one commit.
If the score worsens you won't know which edit caused it.

**Per-experiment steps:**

1. `git status` — confirm clean tree on the autoresearch branch.
2. Read `autoresearch_docs/score.json` — identify the highest-value target
   (page or cluster of fields with the most defect points).
3. Edit schemas/templates to address that target. Keep the diff small.
4. Build: `go build ./...` (sanity check; must pass).
5. Regenerate: `tfplugindocs generate` (must exit 0).
6. Score: `python3 autoresearch_docs/score.py > autoresearch_docs/score.json`.
7. Compare total score to the previous score:
   - **Lower**: commit (`git commit -am "<short desc>"`), update
     `results.tsv`, continue.
   - **Equal**: usually discard (`git reset --hard HEAD`). Exception: a
     pure simplification (e.g. template dedup) with equal score — keep.
   - **Higher**: discard (`git reset --hard HEAD`). Log in `results.tsv`
     with status `discard`.
   - **Build or tfplugindocs crash**: discard, log `crash`.

**Do not commit** `autoresearch_docs/results.tsv`, `score.json`, or
`notes.md`. Keep them untracked.

## Output / logging

`autoresearch_docs/results.tsv` columns (tab-separated):

```
commit	score	pages_failing	status	description
```

- `commit`: short hash (7 chars).
- `score`: integer total defect score. `-1` for crash.
- `pages_failing`: count of doc pages with any defect.
- `status`: `keep`, `discard`, or `crash`.
- `description`: short text, no tabs, no commas (TSV-safe).

Example:

```
commit	score	pages_failing	status	description
a1b2c3d	1842	71	keep	baseline
b2c3d4e	1790	69	keep	fill empty descs in cluster_aws schema
c3d4e5f	1791	70	discard	rewrite addon_deployment example
d4e5f6g	-1	0	crash	tfplugindocs parse error in cluster_gcp tmpl
```

## Style rules for Description strings

Enforce these in your own edits (the scorer enforces a subset; the rest is
taste):

1. Complete sentence, ends with a period.
2. First word is a verb or noun phrase matching the field's role — e.g.
   `"The AWS region in which the cluster is deployed."`, not `"region"`.
3. Name the unit when relevant: seconds, minutes, bytes, IP addresses, FQDNs.
4. For enum-ish strings, list the allowed values: `"One of 'aws', 'azure',
   'gcp'."`.
5. For references to other resources, say so: `"The UID of a
   spectrocloud_cluster_profile resource."`.
6. For lists/sets, describe the element: `"List of tag strings of the form
   'key:value'."`.
7. For `ForceNew` fields, end with `"Changing this forces a new resource."`.
   (This field is already known from the schema — you're describing
   observed behavior, not changing it.)
8. No marketing language. No "simply", "easily", "powerful". No emojis.
9. British vs American spelling: match whatever is dominant in the file.
10. Never invent API semantics you cannot verify from the code or existing
    docs. If unsure, write a minimal, strictly-true description rather than
    a detailed guess.

## The loop

On branch `autoresearch-docs/<tag>`:

LOOP FOREVER:

1. Check git state and last score.
2. Pick the next target: highest defect-point page, or a cluster of related
   fields. Prefer breadth-first early (knock out empty descriptions
   everywhere) and depth later (polish prose).
3. Make the edit.
4. Build, regenerate, score.
5. Keep / discard / crash as above. Record in TSV.
6. Every ~20 experiments, re-read this `program.md` and the current top-10
   worst pages to refresh your mental model.

**Never stop** once the loop starts. If you run out of obvious targets:
read a few full doc pages end-to-end for qualitative improvements, compare
against sibling pages (the `cluster_*` family, the `cloudaccount_*` family)
for consistency, or tighten templates to remove duplication.

**Timeout**: any single experiment should finish in under 2 minutes
(schemas are text, `tfplugindocs generate` on this repo is fast). If
something hangs past 5 minutes, kill it, discard, log crash.

**Stopping condition**: the human interrupts. That's it.
