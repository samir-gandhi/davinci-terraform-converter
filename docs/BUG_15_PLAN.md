# Bug 15 — DaVinci Flow Export → Terraform Plan "No changes" (Phased TDD Plan)

Status: Draft
Owner: GitHub Copilot (GPT-5)
Date: 2025-12-11

Objective

- Achieve `terraform plan` returning "No changes" for exported PingOne DaVinci flows after import, proving full parity between live environment and generated HCL.

Scope

- Resource: `pingone_davinci_flow`
- Areas: API structs, converter output, exporter JSON alignment, acceptance tests in a live environment.

References (provided in prompt attachments)

- `./.github/prompts/prompt-resources/davinci_flow.md` (Terraform schema)
- `./.github/prompts/prompt-resources/agreement-subflow-api-response.json`
- `./.github/prompts/prompt-resources/agreement-subflow-export.json`
- `./.github/prompts/prompt-resources/agreement-subflow-state.tfstate`
- `./.github/prompts/prompt-resources/agreement-subflow-hcl.tf`
- `./.github/prompts/prompt-resources/agreement-subflow-tf-plan.txt`

Constraints

- Provider marks certain attributes as computed/sensitive (e.g., `nodes` in `graph_data.elements`). Diffs can occur even if values are hidden; normalization must ensure exact equality.
- Connectors field is read-only/computed: do not emit connectors directly.
- Provider development overrides can break `terraform init`; acceptance harness must avoid unnecessary init when overrides are present.

Phases

Phase 1 — Gap Analysis and Documentation

- Identify missing/mis-mapped fields comparing:
  - API response vs exported JSON vs generated HCL vs imported state.
- Document discrepancies with concrete examples from agreement subflow artifacts.
- Output: Section listing each discrepancy, impact, and responsible layer (api/converter/exporter).

Phase 2 — Converter Baseline Compliance (already largely present)

- Verify graph_data fields: pan (x,y), zoom, min_zoom, max_zoom, flags (box_selection_enabled, panning_enabled, user_panning_enabled, user_zooming_enabled, zooming_enabled), renderer as `jsonencode`.
- Verify elements data fields: node fields (id, node_type, connection_id reference, connector_id, name, label, status, capability_name, type), properties via `jsonencode`; edges fields (id, source, target).
- TDD: Unit tests asserting presence and formats. Keep passing.
- Output: No code changes if already compliant.

Phase 3 — Cancel Connectors Emission

- Document: connectors attribute is computed; confirm converter does not emit connectors beyond references to connector instance IDs.
- TDD: Test ensures no explicit connectors list emission.


Phase 5 — Settings Parity

- Ensure settings map parity (keys present, defaults applied where state shows them; avoid duplicate keys like `input_schema` attr redefinition).
- TDD: Unit tests targeting known flows where `input_schema` duplication appeared; assert single presence and expected format.

Phase 6 — Acceptance Tests Against Live Environment

- Pre-requisites:
  - macOS + zsh; Terraform CLI installed.
  - PingOne environment with client credentials in env vars: `PINGONE_CLIENT_ID`, `PINGONE_CLIENT_SECRET`, `PINGONE_ENVIRONMENT_ID` (OAuth env for provider auth), and `var.pingone_environment_id` for target environment.
- Provider configuration:
  - Avoid redundant `terraform init` when development overrides are set (as seen in test output warnings). Use lockfile pinned version if not overriding.
- Procedure:
  1. Export environment using tool to HCL.
  2. Import all resources to Terraform state.
  3. For target flows (incl. agreement subflow), run:
     - `terraform validate`
     - `terraform plan`
  4. Expectation: "No changes." If changes reported, capture plan output.
  5. Diagnose sensitive nodes diffs:
     - Compare HCL vs state for non-sensitive fields; enforce normalization from Phase 4.
     - Adjust converter normalization until plan stabilizes.
- Commands (example):
  - Ensure env vars set:
    - `export PINGONE_CLIENT_ID=...`
    - `export PINGONE_CLIENT_SECRET=...`
    - `export PINGONE_ENVIRONMENT_ID=...`
    - `export TF_CLI_CONFIG_FILE=$HOME/.terraformrc` (if using overrides intentionally)
  - In flow module directory:
    - `terraform init`
    - `terraform validate`
    - `terraform plan`

Phase 7 — Regression Suite Hardening

- Expand acceptance harness to skip `init` when provider overrides are active to prevent 503s.
- Add large real-file tests coverage for normalization invariants.

Deliverables

- Updated converter implementation (if needed) for ordering and canonicalization.
- Unit tests per phase under `internal/converter/*_test.go`.
- Acceptance documentation and scripts/notes for live environment runs.
- Progress doc updates in `docs/BUG_15_PROGRESS.md` and this plan.

Risks and Mitigations

- Provider registry 503: Use local provider override or cached lockfile; skip init under overrides.
- Sensitive diffs: Treat equality via normalization; add deterministic ordering and complete flags.
- Schema drift: Track provider versions; align emitted fields strictly to resource schema doc.

Execution Checklist (linked to TODOs)

- Create plan document (this file).
- Add unit test checklist per phase.
- Define acceptance testing steps with commands and prerequisites.
- Execute and iterate: run `make test`; address failing acceptance flows; document deltas.

Notes

- The plan aligns with existing tests that already pass for Phases 4/5; remaining work focuses on normalization and acceptance stability to remove sensitive diffs.
