Use this prompt to create a phased approach and test driven development to work through this bug. For any tests created, use re-usable variables for input data that would change between tests. 

## Feature 01

expand the `davinci-convert davinci-to-hcl` currently can accept one `flow.json` file. With this Some `flow.json` files may have an array of flows and the tool will convert each one into an individual terraform hcl resource. This capability needs to be expanded to also support accepting a string of file names as multiple flows, or a `--flow-json-dir` that contains a set of flows. There will continue to need to be validation on these inputs. Wherever possible keep the code scalable for other capabilities to be added in.