There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug. 

## Bug 10

The current implementation gives inconsistent generation results because the generated resources are printed as they are created, leading to non-deterministic ordering. Refactor the generator to print all the resources in alphabetical order after all resources have been created. There should be some tests to verify this behavior. Imports are created correctly and do not need to be changed. 