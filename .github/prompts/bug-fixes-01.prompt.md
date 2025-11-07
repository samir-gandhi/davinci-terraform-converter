There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through them. 

## Bug 01

When the `--include-imports` flag is provided, the import blocks are generated, but they are generated alongside the resources within the child module. Instead the import blocks should be generated next to the child module, in the root module. The import blocks should map to the id within the child module. Look through the code carefully and first identify why this may have been working before and what is preventing it now. I suspect there may be duplicate or contradicting code. 