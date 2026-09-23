# Porting checklist for tutorial-aligned examples

This is the repeatable process used to port each example under `examples/tutorials/` from the source
tutorial code at [github.com/spectrocloud/tutorials](https://github.com/spectrocloud/tutorials)
(`terraform/<name>-tf/`). It was established while porting `custom-pack` (the first example ported here)
and should be followed for every example added to this directory afterward.

1. **Fetch the source.** Download every file under the corresponding `terraform/<name>-tf/` folder in the
   `spectrocloud/tutorials` repo.

2. **Rename to match this repo's convention.**
   - `provider.tf` → `providers.tf`
   - `inputs.tf` → `variables.tf`
   - `terraform.tfvars` → `terraform.template.tfvars` (placeholder values only — never commit a real
     `terraform.tfvars`)
   - Leave other filenames as-is unless they contain an outright typo or naming inconsistency worth
     fixing while porting (fix it if you spot one; note the fix in the PR description).

3. **Adapt the provider configuration.** Replace a hardcoded `project_name` / API-key-via-environment-only
   setup with this repo's convention: `sc_host`, `sc_api_key`, and `sc_project_name` variables feeding the
   `provider "spectrocloud"` block, matching the existing `examples/e2e/*/providers.tf` style.

4. **Verify every attribute against the current provider schema — not just against the source repo.**
   The tutorials repo can lag the provider's release. Check each resource/data source attribute against
   this repo's actual schema source (`spectrocloud/resource_*.go`, `spectrocloud/data_source_*.go`) or the
   generated docs under `docs/`. Do not assume the source example is still accurate.

5. **Clean up descriptions.** Rewrite any terse or unclear variable/resource descriptions so they're
   understandable without needing to read the docs tutorial or the source repo first — state what the
   value is used for and any constraints (format, required/optional).

6. **Fix dangling cross-references.** If a source file's comment points to a README section that doesn't
   actually exist (or no longer applies after porting), fix the comment or add the missing content —
   don't carry over a broken reference.

7. **Write the README.** Open with "This example accompanies: `<docs tutorial URL>`", followed by what the
   example provisions, prerequisites, step-by-step apply instructions, and a clean-up section. Match the
   tone/structure of the other example READMEs in this directory.

8. **Format.** Run `terraform fmt` across the new directory.

9. **Validate.** Run `terraform init -backend=false && terraform validate` against the latest published
   provider. Fix anything that fails before committing. Delete the resulting `.terraform/` directory and
   `.terraform.lock.hcl` before committing — they're already covered by `.gitignore`, but confirm nothing
   slipped through.

10. **Update the index.** Update this directory's top-level `README.md` status table for the example you
    just ported.

11. **Commit, push, and open a PR** from a task branch cut off the epic branch, following this repo's
    existing per-ticket-branch convention.
