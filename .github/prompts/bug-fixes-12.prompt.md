There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug. 

## Bug 12

The current implementation misses a description attribute for pingone_davinci_variable resources that are generated. This attribute is returned by the API and should be included in the generated Terraform code. Update the generator to include this attribute in the pingone_davinci_variable resources.