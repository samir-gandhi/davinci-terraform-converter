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

Phase 2: Settings Normalization

- Tests: Validate `settings` serialization uses only provider-supported keys and snake_case names: `csp`, `intermediate_loading_screen_css`, `intermediate_loading_screen_html`, `log_level`, etc. Ensure unsupported keys (e.g., `useBetaAlgorithm`) are omitted.
- Implementation:
  - Normalize keys from API to Terraform names; create a mapping table.
  - Ensure only schema-approved keys are emitted.

Phase 3: Graph Data Completeness (pan + renderer + flags)

- Tests: Ensure `graph_data.pan { x,y }`, `graph_data.renderer` (string), and boolean flags (`box_selection_enabled`, `panning_enabled`, `user_panning_enabled`, `zooming_enabled`, etc.) are emitted when present.
- Implementation:
  - Extend API structs for `GraphData` to include these fields.
  - Update converter to output these fields exactly per schema types.

Phase 4: Elements Consistency and Sensitive Fields

- Tests: Confirm `elements.nodes[*].data.properties` and `elements.edges[*].data` align with schema; properties must be emitted as JSON string via `jsonencode` and marked sensitive in Terraform, not expanded.
- Implementation:
  - Validate converter maintains stringification for properties; align `name`, `label`, `status`, `type` when present.

Phase 5: End-to-End Acceptance (Plan No Changes)

- Tests: Add an acceptance test that exports a flow, imports to Terraform, and asserts `terraform plan` reports no changes.
- Implementation:
  - Wire exporter if needed to include updated fields; ensure consistent HCL.

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
- Phase 3: Not started
- Phase 4: Not started
- Phase 5: Not started

## Notes

- Sensitive attributes will show as "(sensitive value)" in plan; compare against state and schema to detect missing structural fields.
- Do not emit unsupported settings; verify against `davinci_flow.md`.
