# Standalone CLI Enhancement - COMPLETE

## ✅ Dual-Mode Operation Added

### What Was Implemented

Enhanced the binary to operate in **two modes**:

1. **Plugin Mode**: Launched by `pingcli` as gRPC plugin (original functionality)
2. **Standalone CLI Mode**: Run directly from command line with full flag support

### Detection Mechanism

The binary automatically detects which mode to run based on environment:

```go
// Check for plugin environment variable set by go-plugin
if os.Getenv("PLUGIN_PROTOCOL_VERSIONS") != "" {
    runAsPlugin()   // gRPC server for pingcli
    return
}

runAsStandalone()  // Direct CLI execution
```

### Standalone CLI Features

**Command-Line Flags:**
- `-f, --flow-json`: Input DaVinci flow JSON file (required)
- `-o, --out`: Output HCL file (optional, defaults to stdout)
- `-d, --out-dir`: Output directory for multi-flow exports (optional)
- `-h, --help`: Show help message
- `-v, --version`: Show version information

**Input Format Detection:**
Automatically detects single-flow vs multi-flow exports:
```go
var check map[string]interface{}
json.Unmarshal(fileData, &check)

if _, hasFlows := check["flows"]; hasFlows {
    // Multi-flow: Use ConvertMultiFlow()
} else {
    // Single flow: Use Convert()
}
```

**Output Modes:**

1. **Single Flow → Stdout**
   ```bash
   ./davinci-convert --flow-json flow.json
   ```

2. **Single Flow → File**
   ```bash
   ./davinci-convert --flow-json flow.json --out output.tf
   ```

3. **Multi-Flow → Directory** (separate files)
   ```bash
   ./davinci-convert --flow-json multiflow.json --out-dir ./flows
   # Creates: flow_name_1.tf, flow_name_2.tf, etc.
   ```

4. **Multi-Flow → Single File** (concatenated)
   ```bash
   ./davinci-convert --flow-json multiflow.json --out combined.tf
   ```

5. **Multi-Flow → Stdout** (concatenated)
   ```bash
   ./davinci-convert --flow-json multiflow.json
   ```

### Implementation Details

**New Functions in main.go:**

1. `runAsPlugin()` - Original gRPC plugin server
2. `runAsStandalone()` - New CLI interface with pflag parsing
3. `handleSingleFlow()` - Process single flow export
4. `handleMultiFlow()` - Process multi-flow export with output options
5. `extractFlowName()` - Extract resource name from HCL for filenames

**File Naming Logic:**
For multi-flow directory output, filenames are extracted from HCL resource names:
```go
resource "pingone_davinci_flow" "pingone_sign_on_with_sessions" {
  // Filename: pingone_sign_on_with_sessions.tf
}
```

### Usage Examples

**Simple Single Flow:**
```bash
$ ./davinci-convert --flow-json simple-demo-flow.json
resource "pingone_davinci_flow" "simple_demo_flow" {
  environment_id = var.environment_id
  name        = "Simple Demo Flow"
  ...
}
```

**Real Multi-Flow Export:**
```bash
$ ./davinci-convert --flow-json multiflow-export.json --out-dir /tmp/flows
Flow 1 written to: /tmp/flows/pingone_sign_on_with_sessions.tf
Flow 2 written to: /tmp/flows/pingone_sign_on_with_registration_password_reset_and_recovery.tf

Successfully converted 2 flows to: /tmp/flows
```

**Check Version:**
```bash
$ ./davinci-convert --version
davinci-convert v0.1.0
```

**Help:**
```bash
$ ./davinci-convert --help
DaVinci Flow to HCL Converter

Usage:
  ./davinci-convert [flags]

Flags:
  -f, --flow-json string   Path to input DaVinci flow JSON file (required)
  -h, --help               Show help message
  -o, --out string         Path to output HCL file (optional, defaults to stdout)
  -d, --out-dir string     Directory for multi-flow output (optional, for multi-flow exports)
  -v, --version            Show version information

Examples:
  # Convert single flow to stdout
  ./davinci-convert --flow-json flow.json

  # Convert single flow to file
  ./davinci-convert --flow-json flow.json --out output.tf

  # Convert multi-flow export to directory
  ./davinci-convert --flow-json multiflow.json --out-dir ./flows
```

### Real-World Testing

**Test with Production Multi-Flow Export:**
- Input: `PingOne_Sign On with Sessions_multiflow.json` (8,464 lines, 2 flows)
- Output: 2 separate .tf files
  - `pingone_sign_on_with_sessions.tf` (31KB)
  - `pingone_sign_on_with_registration_password_reset_and_recovery.tf` (351KB)
- Result: ✅ Complete conversion with all data preserved

**Demo Flow Testing:**
- Created: `simple-demo-flow.json` for quick demonstrations
- Features: Nodes, edges, settings
- Result: ✅ Clean HCL output

### Error Handling

**Comprehensive error messages:**
- Missing required flags
- File read errors
- JSON parsing errors
- Conversion errors (with flow names/indices)
- File write errors

**Exit Codes:**
- `0`: Success
- `1`: Error (with descriptive message to stderr)

### Benefits for Demonstrations

1. **No Dependencies**: No need to install/configure `pingcli`
2. **Instant Feedback**: Direct stdout output for quick testing
3. **File Output**: Easy to share/inspect generated HCL
4. **Directory Mode**: Organized output for multi-flow exports
5. **Help Documentation**: Built-in usage examples

### Backwards Compatibility

✅ **Plugin mode unchanged**: Original `pingcli` plugin functionality preserved
✅ **No breaking changes**: Existing plugin integration continues to work
✅ **Environment detection**: Automatic mode selection based on launch context

### Files Modified

```
main.go  # Enhanced with dual-mode support
  - Added runAsPlugin() (original logic)
  - Added runAsStandalone() (new CLI interface)
  - Added handleSingleFlow() (file I/O for single flows)
  - Added handleMultiFlow() (file I/O for multi-flows)
  - Added extractFlowName() (filename extraction)
```

### Test Results

```bash
$ make all
✓ All tests pass (25 tests)
✓ Coverage: 91.6%
✓ Binary builds successfully

$ ./davinci-convert --version
davinci-convert v0.1.0

$ ./davinci-convert --help
[Help output displayed]

$ ./davinci-convert --flow-json simple-demo-flow.json
[HCL output to stdout]

$ ./davinci-convert --flow-json multiflow.json --out-dir /tmp/test
Flow 1 written to: /tmp/test/flow1.tf
Flow 2 written to: /tmp/test/flow2.tf
Successfully converted 2 flows to: /tmp/test
```

### Integration Notes

**For Part 4 (if needed):**
The command implementation in `cmd/convert.go` can now leverage the standalone functions:
- Read file using `os.ReadFile()`
- Call `converter.Convert()` or `converter.ConvertMultiFlow()`
- Write output using `os.WriteFile()` or logger.Message()

**For Demonstrations:**
- Simple: `./davinci-convert -f flow.json`
- With output: `./davinci-convert -f flow.json -o output.tf`
- Multi-flow: `./davinci-convert -f export.json -d ./flows`

## Summary

Dual-mode operation complete. The binary now functions as both a standalone CLI tool for demos/testing AND as a pingcli plugin for production integration. Zero breaking changes to plugin functionality. Full file I/O support with multiple output modes.

**Benefits:**
- ✅ Instant demonstrations without pingcli setup
- ✅ Easy testing with real flow files
- ✅ Flexible output options (stdout/file/directory)
- ✅ Production-ready plugin mode preserved
- ✅ Comprehensive help and error messages
- ✅ Automatic format detection
- ✅ Version information

Ready for human demonstrations and Part 3 development.
