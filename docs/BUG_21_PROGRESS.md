# BUG 21 Progress

Status

- Phase 1 (Models): Completed
- Phase 2 (Converter): Completed
- Phase 3 (Exporter Integration): Completed
- Phase 4 (Tests): Completed (unit tests added)
- Phase 5 (Docs): Completed

Notes

- Enabled resolution implements conflict detection between `flowStatus` and `enabled`.
- Deploy resource maps `deployed_version` to `publishedVersion` or provider reference based on `--skip-dependencies`.
- Exporter propagates `enabled` and `publishedVersion` to converter.
