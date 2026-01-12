# Provider-Managed Fields and Import Drift

Last Updated: 2026-01-12

This note documents known Terraform plan diffs that appear after importing existing DaVinci flows and the converter’s behavior to suppress churn.

## Import Drift on PingOne DaVinci Flow

- Attributes frequently reported for in-place updates post-import:
  - `connectors`
  - `current_version`
  - `enabled`
  - `published_version`

### Why These Diffs Appear

- Computed/server-managed: Values are decided by the backend (e.g., publish counters, enabled flags, connector set derived from `graphData`). They may change outside Terraform.
- State vs config mismatch: Import state may include provider-managed values that are not explicitly configured, producing no-op diffs.

### Converter Behavior

- Every generated `pingone_davinci_flow` includes a lifecycle block:

```hcl
  lifecycle {
    ignore_changes = [
      connectors,
      current_version,
      enabled,
      published_version
    ]
  }
```

- Terraform may warn that some ignored attributes are redundant because they are purely computed. This is acceptable; the block suppresses plan churn after import.

## Empty `graph_data.data` Inclusion

- When the API returns `"graphData": { "data": {} }`, the provider persists the empty object.
- The converter includes this as:

```hcl
  graph_data = {
    data = jsonencode({})
    # ...
  }
```

This aligns HCL with provider state and avoids false-positive diffs.
