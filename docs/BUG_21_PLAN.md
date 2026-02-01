# BUG 21 Plan: Add `pingone_davinci_flow_enabled` and `pingone_davinci_flow_deploy`

Phases

1. Models

- Add `Enabled` and `PublishedVersion` to API `FlowDetail`.
- Parse fields from raw API response.
- Propagate fields via exporter `convertFlowDetailToMap`.

1. Converter

- Implement `resolveEnabled()` to detect from `flowStatus` (export) or `enabled` (API), with conflict error.
- Extend `ConvertFlowToHCL()` to emit:
  - `pingone_davinci_flow_enabled` with `environment_id`, `flow_id`, and `enabled`.
  - `pingone_davinci_flow_deploy` with `deploy_trigger_values = { "deployed_version" = ... }`.
- Honor `--skip-dependencies` (hardcode IDs/values vs references).

1. CLI/Exporter Integration

- No CLI changes required; exporter gains propagated fields enabling auxiliary resource generation.

1. Tests

- Add converter tests covering:
  - Auxiliary resources generation (reference mode).
  - Auxiliary resources generation (skip-dependencies hardcoded mode).
  - Conflict detection when export `flowStatus` contradicts API `enabled`.
- Extend exporter tests to validate field propagation.

1. Documentation & Tracking

- Maintain progress in `docs/BUG_21_PROGRESS.md`.

Acceptance Criteria

- Conversion from single/multi-flow export emits both new resources.
- Export from API emits both new resources.
- Conflict detection works with clear error.
- All tests pass via `make test`.
