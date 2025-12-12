# Bug 15: Davinci Flow Export Parity Plan

Purpose: Achieve Terraform plan with "No changes" for exported PingOne DaVinci flows by aligning generated HCL to provider schema and imported state.

## Artifacts Referenced

- Terraform schema: `.github/prompts/prompt-resources/davinci_flow.md`
- API response: `.github/prompts/prompt-resources/agreement-subflow-api-response.json`
- Terraform plan: `.github/prompts/prompt-resources/agreement-subflow-tf-plan.txt`
- Historical UI export: `.github/prompts/prompt-resources/agreement-subflow-export.json`
- Terraform state: `.github/prompts/prompt-resources/agreement-subflow-state.tfstate`
- Current generated HCL: `.github/prompts/prompt-resources/agreement-subflow-hcl.tf`

## Gap Analysis (from provided artifacts)

- Color missing: Provider shows `color` present in state (`#CACED3`) and removed by plan; generated HCL lacks `color`.
- Input schema missing: Plan shows `input_schema` attributes slated for removal → generated HCL does not emit `input_schema` at all.
- Settings key mismatch: API includes `useBetaAlgorithm`; generated HCL emits `usebetaalgorithm` which is not a Terraform setting attribute. Remove or map appropriately; provider schema does not include this field.
- Graph data renderer missing: Provider schema accepts `renderer` as a `String`; state uses a string; generated HCL does not emit `renderer`.
- Graph data pan missing: Provider schema requires `pan { x, y }` optionally; state has `pan`; HCL lacks it.
- Elements structure normalization: Ensure `elements.nodes[*].data` uses Terraform field names (`node_type`, `connection_id`, `connector_id`, etc.) and sensitive `properties` remain as `String` via `jsonencode`. Current HCL aligns, but confirm completeness of keys (`name`, `label`, `status`, `type`).
- Box/zoom/panning flags: State shows booleans for `box_selection_enabled`, `panning_enabled`, etc. HCL omits most; add when present to match state.

## Phased TDD Plan

Phase 1: Minimal Parity (color + input_schema)

- Phase 1a — Tests:
  - Unit: `internal/converter/flow_converter_test.go` asserts `color` and `input_schema` emission.
  - Acceptance: `tests/acceptance/flow_export_contains_color_input_schema_test.go` validates exported HCL contains `color` and `input_schema`.
  - Status: Completed (passing).
- Phase 1b — Implementation:
  - API: `internal/api/flows.go` now parses top-level `inputSchema` and `color`.
  - Exporter: `internal/exporter/flow_exporter.go` forwards `inputSchema` to converter.
  - Converter: `internal/converter/flow_converter.go` emits `color` and `input_schema` (direct and compiled / graph fallbacks).
  - Status: Completed.

  Notes (edits summary to inform future phases):
  - API (`internal/api/flows.go`): Extended `FlowDetail` with `InputSchema` and `Color`; `GetFlow` parses top-level `inputSchema` and preserves `inputSchemaCompiled`.
  - Exporter (`internal/exporter/flow_exporter.go`): `convertFlowDetailToMap` includes `inputSchema`, `inputSchemaCompiled`, `color`, `settings`, and `graphData` for converter input.
  - Converter (`internal/converter/flow_converter.go`): Emits `color`; writes `input_schema` from top-level `inputSchema`; derives from `inputSchemaCompiled` when absent; adds graph-based fallback to reconstruct inputs; ensures `input_schema {}` exists to prevent Terraform diffs.
  - Tests (`internal/converter/flow_converter_test.go`, `tests/acceptance/flow_export_contains_color_input_schema_test.go`): Converter unit tests cover color and input schema emission; acceptance test uses shared helpers and asserts presence of `color` and `input_schema` with spacing-robust color check; input schema absence fails the test.

Phase 2: Settings Normalization

- Phase 2a — Tests:
  - Validate `settings` serialization uses only provider-supported keys and snake_case names: `csp`, `intermediate_loading_screen_css`, `intermediate_loading_screen_html`, `log_level`, etc.
  - Ensure unsupported keys (e.g., `useBetaAlgorithm`) are omitted.
  - Status: Not started.
- Phase 2b — Implementation:
  - Normalize keys from API to Terraform names; create a mapping table.
  - Ensure only schema-approved keys are emitted.
  - Status: Not started.

Phase 3: Connectors Attribute — Not Needed

- Rationale: The Terraform provider defines the `connectors` field as ReadOnly/computed. It must not be set in configuration and should not be emitted in generated HCL.
- Decision: Cancel this phase. Do not add tests or implementation that emit `connectors` in HCL.
- Status: Cancelled.

Phase 4: Graph Data Completeness (pan + renderer + flags)

- Phase 4a — Tests:
  - Ensure `graph_data.pan { x,y }`, `graph_data.renderer` (jsonencoded), and boolean flags (`box_selection_enabled`, `panning_enabled`, `user_panning_enabled`, `user_zooming_enabled`, `zooming_enabled`) are emitted when present.
  - Implemented in `internal/converter/flow_graph_data_test.go`.
  - Status: Completed (passing unit tests).
- Phase 4b — Implementation:
  - Converter `writeGraphDataBlock` already emits `pan`, `zoom`, `min_zoom`, `max_zoom`, flags, and `renderer = jsonencode(...)` when present.
  - No API/exporter changes required; `graphData` already forwarded end-to-end.
  - Status: Completed.

Phase 5: Elements Consistency and Sensitive Fields

- Phase 5a — Tests:
  - Confirm `elements.nodes[*].data.properties` and `elements.edges[*].data` align with schema; properties must be emitted as JSON string via `jsonencode` and marked sensitive in Terraform, not expanded.
  - Status: Not started.
- Phase 5b — Implementation:
  - Validate converter maintains stringification for properties; align `name`, `label`, `status`, `type` when present.
  - Status: Not started.

Phase 6: End-to-End Acceptance (Plan No Changes)

- Phase 6a — Tests:
  - Add an acceptance test that exports a flow, imports to Terraform, and asserts `terraform plan` reports no changes.
  - Status: Not started.
- Phase 6b — Implementation:
  - Wire exporter if needed to include updated fields; ensure consistent HCL.
  - Status: Not started.

## Test Strategy

- Unit tests: `internal/converter/flow_converter_test.go` — construct representative API structs and assert HCL fragments for:
  - `color` value round-trip
  - `input_schema` entries with `description`, `preferred_control_type`, `preferred_data_type`, `property_name`, `required`, `is_expanded`
  - `settings` allowed keys only and correct snake_case
  - `graph_data.pan` and `renderer` presence and types
  - Flags booleans in `graph_data`
- Integration tests: Parser/serializer for `graph_data` elements; ensure `jsonencode` used for `properties`.
- Acceptance test: Run exporter → import → `terraform plan` equals "No changes" for target flow.

## Acceptance Tests per Phase

- Phase 1: Export HCL contains `color` and `input_schema`.
- Phase 2: Export HCL `settings` contains only provider-supported keys in snake_case; verify absence of unsupported keys (e.g., `useBetaAlgorithm`).
- Phase 3: Export HCL `graph_data` contains `pan { x, y }`, `renderer` string, and present boolean flags.
- Phase 4: Export HCL elements use correct fields and `properties` are emitted as JSON strings.
- Phase 5: End-to-end: After import, `terraform plan` reports "No changes" for flows.

## Work Log / Status

- Phase 1a (Tests): Completed — unit + acceptance passing. Date: 2025-12-10.
- Phase 1b (Implementation): Completed — API parses `inputSchema`, exporter forwards, converter emits.
- Phase 2: Not started
- Phase 3: Cancelled — provider `connectors` is ReadOnly/computed; omit from HCL.
- Phase 4a (Tests): Completed — `flow_graph_data_test.go` validates pan/renderer/flags.
- Phase 4b (Implementation): Completed — converter emits these fields when present.
- Phase 5: Not started

## Notes

- Sensitive attributes will show as "(sensitive value)" in plan; compare against state and schema to detect missing structural fields.
- Do not emit unsupported settings; verify against `davinci_flow.md`.
Known Issue: In some environments, `required = false` for `input_schema` entries may be omitted by Terraform/HCL rendering, leading Terraform to display `required = (known after apply)` in plan even when the API/state show `required: false`. Tracked for normalization; non-blocking for later phases.

Observation: Current Terraform plan for the target flow shows no changes for `settings`. This suggests our emitted keys and values already align with the provider schema for this environment. We will proceed with Phase 2 while retaining tests to guard against regressions.

Provider Constraint: `connectors` is ReadOnly/computed in the provider schema and should not be represented in user configuration. Ensure exporter/converter do not rely on emitting this field in HCL.

Phase 4 — Normalization for Sensitive Equality

- Deterministic ordering:
  - Ensure `pingone_davinci_flow.graphdata.elements.nodes[]` is printed in HCL in the same order that the nodes are in the API response.
- Canonical JSON for properties:
  - Ensure consistent key ordering and consistent string escaping for `jsonencode` fields.
- Flags defaults mirroring state:
  - Emit all UI flags and defaults that appear in state to avoid hidden diffs.
- TDD:
  - Unit tests for ordering determinism.
  - Unit tests checking emitted flags and defaults match expected provider/state values.

Phase 8: End-to-End Acceptance (Plan No Changes)

- Tests:
  - Acceptance test plan: export environment to HCL, import into Terraform state, and assert `terraform plan` returns "No changes" for target flows (e.g., agreement subflow).
  - Pre-requisites: macOS + zsh; Terraform CLI installed; PingOne environment credentials via env vars (`PINGONE_CLIENT_ID`, `PINGONE_CLIENT_SECRET`, `PINGONE_ENVIRONMENT_ID`).
  - Provider configuration: avoid redundant `terraform init` when development overrides are set; use lockfile pinned version if not overriding.
- Procedure:
  1. Export environment using this tool to HCL.
  2. Import all resources to Terraform state.
  3. In the flow module directory, run `terraform validate` then `terraform plan`.
  4. Expectation: "No changes." If changes appear, capture plan output and iterate normalization (ordering, flags, canonical JSON) until stable.
- Status: Planned. Initial runs show provider registry 503 errors in some acceptance flows; mitigation is to skip `init` under provider dev overrides or use cached provider.
