package converter

import (
	"fmt"
	"strings"

	"github.com/pingidentity/pingone-go-client/pingone"
)

// ConvertFlowPolicyToTerraform converts a DaVinci flow policy to Terraform HCL format
func ConvertFlowPolicyToTerraform(policy pingone.DaVinciFlowPolicyResponse, resourceName, applicationID, environmentID string, skipDeps bool) (string, error) {
	var hcl strings.Builder

	// Create resource block
	hcl.WriteString(fmt.Sprintf("resource \"pingone_davinci_application_flow_policy\" \"%s\" {\n", resourceName))

	// Environment ID
	if strings.HasPrefix(environmentID, "var.") {
		hcl.WriteString(fmt.Sprintf("  environment_id = %s\n", environmentID))
	} else {
		hcl.WriteString(fmt.Sprintf("  environment_id = %q\n", environmentID))
	}

	// Application ID
	if skipDeps {
		hcl.WriteString(fmt.Sprintf("  application_id = %q\n", applicationID))
	} else {
		appResourceName := sanitizeResourceName(applicationID)
		hcl.WriteString(fmt.Sprintf("  application_id = pingone_davinci_application.%s.id\n", appResourceName))
	}

	// Name
	if name, ok := policy.GetNameOk(); ok {
		hcl.WriteString(fmt.Sprintf("  name           = %q\n", *name))
	}

	// Status
	if status, ok := policy.GetStatusOk(); ok {
		hcl.WriteString(fmt.Sprintf("  status         = %q\n", string(*status)))
	}

	// Flow distributions
	if distributions, ok := policy.GetFlowDistributionsOk(); ok && len(distributions) > 0 {
		hcl.WriteString("\n")
		if !skipDeps {
			hcl.WriteString("  # Note: Flow IDs below should be replaced with pingone_davinci_flow.<resource_name>.id references\n")
		}
		hcl.WriteString("  flow_distributions = [\n")

		for _, dist := range distributions {
			hcl.WriteString("    {\n")

			// Flow ID
			if flowID, ok := dist.GetIdOk(); ok {
				if skipDeps {
					hcl.WriteString(fmt.Sprintf("      id      = %q\n", *flowID))
				} else {
					// Use raw UUID - user needs to manually replace
					hcl.WriteString(fmt.Sprintf("      id      = %q\n", *flowID))
				}
			}

			// Version
			if version, ok := dist.GetVersionOk(); ok {
				hcl.WriteString(fmt.Sprintf("      version = %d\n", int64(*version)))
			}

			// Weight (optional)
			if weight, ok := dist.GetWeightOk(); ok {
				hcl.WriteString(fmt.Sprintf("      weight  = %d\n", int64(*weight)))
			}

			hcl.WriteString("    },\n")
		}

		hcl.WriteString("  ]\n")
	}

	hcl.WriteString("}\n")

	return hcl.String(), nil
}
