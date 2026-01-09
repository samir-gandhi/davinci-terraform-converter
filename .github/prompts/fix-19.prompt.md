There are a number of bugs in the current build. Use this prompt to create a phased approach and test driven development to work through this bug.

## Bug Fix 19

This bug has to do with escaping special characters in json responsed when converting them to HCL. It seems that the way the PingOne Terraform provider applies payloads that these escpaped characters are not needed and lead to duplicate escaping. 

Example of terraform.tfstate after a `pingone_davinci_flow` resource is imported:

```json
"settings": {
  "csp": "worker-src 'self' blob:; script-src 'self' https://cdn.jsdelivr.net https://code.jquery.com https://devsdk.singularkey.com http://cdnjs.cloudflare.com 'unsafe-inline' 'unsafe-eval';",
  "css": ".companyLogo {\n    /* Ping Logo  */\n    content: url(\"https://assets.pingone.com/ux/ui-library/5.0.2/images/logo-pingidentity.png\");\n    width: 65px;\n    height: 65px;\n}",
  "css_links": [],
  ...
}
```

Example `terraform.tfstate` after applying the genereated HCL:

```json
"settings": {
  "csp": "worker-src 'self' blob:; script-src 'self' https://cdn.jsdelivr.net https://code.jquery.com https://devsdk.singularkey.com http://cdnjs.cloudflare.com 'unsafe-inline' 'unsafe-eval';",
  "css": ".companyLogo {\\n    /* Ping Logo  */\\n    content: url(\\\"https://assets.pingone.com/ux/ui-library/5.0.2/images/logo-pingidentity.png\\\");\\n    width: 65px;\\n    height: 65px;\\n}",
  "css_links": [],
  ...
}
```

Find the locations in the code where these special characters are being escaped when they shouldn't be. Create a phased plan on a new markdown file in the `.github/prompts/` directory named `fix-19-PROGRESS.md` that outlines the steps to identify and fix this bug, and provides space to document progress.