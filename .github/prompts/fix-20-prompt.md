# Fix 20

reference:

- (./data/dv-new/davinci_flow.md)
- (./data/dv-new/resource_davinci_flow_gen.go)
- (./data/dv-old/davinci_flow.md)
- (./data/dv-old/resource_davinci_flow_gen.go)
- (./ARCHITECTURE.md)

Look at the reference documents listed above to understand this repository.
This repo works to convert API responses to Terraform HCL for `pingone_davinci_*` resources that are being developed. 
There was a schema update to the Flow resource that turned some blocks (`graph_data.elements.edges`, `graph_data.elements.nodes`)
There were also some updates to typing on number or positional style attributes. 

Take a phased approach that tracks progress on `fix-20-PROGRESS.md` and complete the following tasks:

1. Identify ALL differences between the two schemas
2. Identify how these differences may impact the converter/exporter tool.
3. Make phases for how to implement these fixes. The phases should include plans for the fixes as well as plans for test updates.
4. Implement