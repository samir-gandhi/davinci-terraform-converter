There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug. 

## Bug 07

Right now the generator creates graph_data.elements.nodes.data.properties attributes as base64encoded json strings. This makes it hard to read and debug. Update the generator to create these attributes as terraform acceptable json strings. If this conflicts with tests, identify why. 

