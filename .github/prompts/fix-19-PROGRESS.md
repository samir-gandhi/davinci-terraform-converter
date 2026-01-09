# Bug Fix 19 — Escaping in Settings (Progress)

This tracks investigation and fixes for over-escaping in HCL output (e.g., `settings.css`) that causes drift versus imported state.

## Status

- Identified hotspots and implemented heredoc emission for multiline settings.
- Unit tests added for `css` and `intermediate_loading_screen_html`; converter tests pass.
- Pending: broader test run and doc updates.

## Root Cause

- `quoteString()` escapes `\n`, `\"`, etc., and is used by `writeSettingsBlock()` for all strings.
- Multiline fields (CSS/HTML) do not require escaped newlines/quotes for provider payloads, leading to duplicate escaping.

Mitigation added: dynamic, content-hashed heredoc labels (prefix `TFHCL_<KEY>_<HASH>`) to avoid accidental label collisions that can lead to unclosed blocks.

## Plan (TDD)

- Tests: Verify `settings.css` and `intermediate_loading_screen_html` emit heredocs and do not contain `\\n`/`\\\"`. Confirm `settings.csp` remains quoted.
- Implementation: In `writeSettingsBlock()`, use heredoc for `css`, `intermediate_loading_screen_css`, `intermediate_loading_screen_html`, or any string containing newlines.
- Validation: Run converter package tests; then run full repo tests when unrelated build issues are resolved.
- Documentation: Note heredoc rationale and provider behavior.

## Progress Log

- 2026-01-08: Implemented heredoc emission; added unit tests; converter tests passing.

## References

- Converter: [internal/converter/flow_converter.go](internal/converter/flow_converter.go)
- Tests: [internal/converter/flow_converter_test.go](internal/converter/flow_converter_test.go)

## Root-Cause Locations
## Current Behavior

- Imported state `settings.css` example shows raw newlines and quoted URLs.
- Generated HCL currently prints `settings.css` with `\n` and `\"`, leading to drift on `terraform plan` and unnecessary changes on `apply`.

- Usage impacting `settings`:
  - Direct calls within `writeSettingsBlock()` for `js_links` fields (e.g., `value`, `label`) at [internal/converter/flow_converter.go#L315](internal/converter/flow_converter.go#L315), [internal/converter/flow_converter.go#L325](internal/converter/flow_converter.go#L325), [internal/converter/flow_converter.go#L329](internal/converter/flow_converter.go#L329), [internal/converter/flow_converter.go#L333](internal/converter/flow_converter.go#L333), [internal/converter/flow_converter.go#L336](internal/converter/flow_converter.go#L336), [internal/converter/flow_converter.go#L340](internal/converter/flow_converter.go#L340).
- Other areas that use `quoteString()` (nodes, edges, schemas) are not implicated by the bug example and should remain unchanged unless tests prove otherwise.
- Note: `writeJSONAsHCLMap()` in [internal/converter/flow_converter.go](internal/converter/flow_converter.go#L880-L941) uses `strconv.Quote` and `"$" -> "$$"` within `jsonencode(...)`. This is correct for JSON blocks and does not directly affect `settings.css`.

## Root-Cause Locations

- Escaping implementation:
   - `quoteString()` in [internal/converter/flow_converter.go](internal/converter/flow_converter.go#L835-L846): replaces `\`, `"`, `\n`, `\r`, `\t` then wraps with `%q`.

- Usage impacting `settings`:
   - `writeSettingsBlock()` in [internal/converter/flow_converter.go](internal/converter/flow_converter.go#L261): string values are emitted via `quoteString(v)`.
   - Direct calls within `writeSettingsBlock()` for `js_links` fields (e.g., `value`, `label`) at [internal/converter/flow_converter.go](internal/converter/flow_converter.go#L315), [internal/converter/flow_converter.go](internal/converter/flow_converter.go#L325), [internal/converter/flow_converter.go](internal/converter/flow_converter.go#L329), [internal/converter/flow_converter.go](internal/converter/flow_converter.go#L333), [internal/converter/flow_converter.go](internal/converter/flow_converter.go#L336), [internal/converter/flow_converter.go](internal/converter/flow_converter.go#L340).
   - Generic string emission in settings: [internal/converter/flow_converter.go](internal/converter/flow_converter.go#L355).

- Other areas that use `quoteString()` (nodes, edges, schemas) are not implicated by the bug example and should remain unchanged unless tests prove otherwise.

- Note: `writeJSONAsHCLMap()` in [internal/converter/flow_converter.go](internal/converter/flow_converter.go#L880-L941) uses `strconv.Quote` and `"$" -> "$$"` within `jsonencode(...)`. This is correct for JSON blocks and does not directly affect `settings.css`.

   - Unit test: converting a flow with `settings.css` containing newlines and quoted URLs should produce HCL that preserves content without `\\n` or `\\\"`.
   - Unit test: `settings.intermediate_loading_screen_css` and `settings.intermediate_loading_screen_html` behave similarly.
   - Assert HCL uses Terraform heredoc (`<<-EOT ... EOT`) for multiline strings, or equivalent raw-safe formatting, and round-trips to expected state formatting.
2. Design and Implement Escaping Strategy
   - Keep `quoteString()` for simple scalar fields.
## Phased Plan (TDD)

1. Define Failing Tests

    - Unit test: converting a flow with `settings.css` containing newlines and quoted URLs should produce HCL that preserves content without `\\n` or `\\\"`.
    - Unit test: `settings.intermediate_loading_screen_css` and `settings.intermediate_loading_screen_html` behave similarly.
    - Assert HCL uses Terraform heredoc (`<<-EOT ... EOT`) for multiline strings, or equivalent raw-safe formatting, and round-trips to expected state formatting.

   - For known multiline settings keys (`css`, `intermediate_loading_screen_css`, `intermediate_loading_screen_html`), emit heredoc blocks to avoid over-escaping and maintain readability.
   - For single-line but complex strings (e.g., `csp`), retain normal quoting; do not alter unless tests show drift.
   - Guard against Terraform interpolation by ensuring heredoc content does not break `${...}` semantics; consider `jsonencode()` only where semantically correct.
3. Integrate Converter Changes
   - Update `writeSettingsBlock()` to conditionally use heredoc for targeted keys when value contains newline characters or is known multiline.
2. Design and Implement Escaping Strategy

    - Keep `quoteString()` for simple scalar fields.
    - For known multiline settings keys (`css`, `intermediate_loading_screen_css`, `intermediate_loading_screen_html`), emit heredoc blocks to avoid over-escaping and maintain readability.
    - For single-line but complex strings (e.g., `csp`), retain normal quoting; do not alter unless tests show drift.
    - Guard against Terraform interpolation by ensuring heredoc content does not break `${...}` semantics; consider `jsonencode()` only where semantically correct.

   - Ensure sorting and key mapping logic remain deterministic.
4. Validate and Refine
   - Run `make test`; ensure new tests pass and no regressions in existing converter tests.
3. Integrate Converter Changes

    - Update `writeSettingsBlock()` to conditionally use heredoc for targeted keys when value contains newline characters or is known multiline.
    - Ensure sorting and key mapping logic remain deterministic.

   - If necessary, add integration test harness to parse generated HCL and compare with simulated state normalization.
5. Documentation
   - Add a short note to docs explaining heredoc usage in settings and rationale based on provider behavior.
4. Validate and Refine

    - Run `make test`; ensure new tests pass and no regressions in existing converter tests.
    - If necessary, add integration test harness to parse generated HCL and compare with simulated state normalization.


## Test Coverage Targets
5. Documentation

    - Add a short note to docs explaining heredoc usage in settings and rationale based on provider behavior.

- `internal/converter/flow_converter_test.go`
  - `TestSettingsIntermediateCSSHTML_HeredocEmission`: verifies heredoc for `intermediate_loading_screen_css` and `_html`.
  - `TestSettingsCSP_SimpleString`: verifies `csp` remains quoted, no heredoc, no extra escaping beyond what HCL requires.

## Test Coverage Targets

- `internal/converter/flow_converter_test.go`
   - `TestSettingsCSSMultiline_HeredocEmission`: verifies heredoc and absence of `\\n` / `\\\"` in `settings.css`.
   - `TestSettingsIntermediateCSSHTML_HeredocEmission`: verifies heredoc for `intermediate_loading_screen_css` and `_html`.
   - `TestSettingsCSP_SimpleString`: verifies `csp` remains quoted, no heredoc, no extra escaping beyond what HCL requires.

## Acceptance Criteria
## Acceptance Criteria

- No duplicate escaping in generated HCL for multiline settings.
- `terraform plan` against imported state shows no changes due to string escaping in `settings`.
- Existing conversions for graph data, jsonencode blocks, and references remain stable.


## Progress Log

- 2025-12-19: Identified escaping hotspots in `quoteString()` and `writeSettingsBlock()`; drafted TDD plan.
- [Pending]: Write failing unit tests for multiline settings.
- [Pending]: Implement heredoc emission.
- [Pending]: Validate via `make test`.
- [Pending]: Update documentation.

- [Pending]: Update documentation.
## Notes / Decisions

- Prefer heredoc for readability and minimal escaping; avoids manual character replacements.
- Do not alter `jsonencode` handling or `$` escaping in JSON map printer.
- Limit changes to `settings` keys shown to cause drift; expand only with evidence from tests.

- Do not alter `jsonencode` handling or `$` escaping in JSON map printer.
- Limit changes to `settings` keys shown to cause drift; expand only with evidence from tests.
