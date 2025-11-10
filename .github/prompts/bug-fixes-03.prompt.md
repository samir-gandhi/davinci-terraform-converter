There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug. 

## Bug 03

This will be a set of bug fixes to be more clear and consistent on naming. 

### Part 1

When the `--module` flag is used, there is a mismatch in the pingone names. 

```
The child module (davinci-module) defined a variable named pingone_environment_id, but all the resources within that module were referencing var.environment_id. This naming mismatch caused Terraform to think that environment_id was undefined in the child module, which made it look for it in the root module, creating a cascade of "missing variable" errors.
```

### Part 2

The default module folder is `davinci-module` and the module name is `davinci`. The generated import blocks look like:
```
import {
  to = module.davinci-module.pingone_davinci_variable.pingcli__companyLogo
  id = "62f10a04-6c54-40c2-a97d-80a98522ff9a/0da83099-d149-4309-bf2b-6cc82884b577"
}
```
So there is a mismatch in name. However, this project is not meant to be davinci only anyway. So we need to:
1. make the default module folder name `ping-export-module`
2. make the default child module name `ping-export`
3. make the import blocks use the module name. ensure it is not using the folder name or some hard-coded value. 

### Part 3

The generated child modules files should use the full resource name. Right now we have files like:

```
flows.tf
applications.tf
connections.tf
..etc..
```
instead it should be:
```
pingone_davinci_application.tf
pingone_davinci_connector_instance.tf
pingone_davinci_flow.tf
```