package converter

import (
	"fmt"
	"strings"

	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/samir-gandhi/davinci-terraform-converter/internal/resolver"
)

// ConvertFlowPolicyToTerraform converts a DaVinci flow policy to Terraform HCL format
func ConvertFlowPolicyToTerraform(policy pingone.DaVinciFlowPolicyResponse, resourceName, applicationID, environmentID string, skipDeps bool, graph *resolver.DependencyGraph) (string, error) {
	var hcl strings.Builder

	// Create resource block
	hcl.WriteString(fmt.Sprintf("resource \"pingone_davinci_application_flow_policy\" \"%s\" {\n", resourceName))

	// Environment ID
	if strings.HasPrefix(environmentID, "var.") {
		hcl.WriteString(fmt.Sprintf("  environment_id = %s\n", environmentID))
	} else {
		hcl.WriteString(fmt.Sprintf("  environment_id = %q\n", environmentID))
	}

	// Application ID - use graph for reference if available
	if skipDeps {
		hcl.WriteString(fmt.Sprintf("  da_vinci_application_id = %q\n", applicationID))
	} else {
		if graph != nil {
			appRef, err := resolver.GenerateTerraformReference(graph, "pingone_davinci_application", applicationID, "id")
			if err != nil {
				// Fallback to TODO placeholder if application not found in graph
				hcl.WriteString(fmt.Sprintf("  da_vinci_application_id = \"\" # TODO: %s\n", err.Error()))
			} else {
				hcl.WriteString(fmt.Sprintf("  da_vinci_application_id = %s\n", appRef))
			}
		} else {
			// Fallback to legacy sanitized name
			appResourceName := sanitizeResourceName(applicationID)
			hcl.WriteString(fmt.Sprintf("  da_vinci_application_id = pingone_davinci_application.%s.id\n", appResourceName))
		}
	}

	// Name
	if name, ok := policy.GetNameOk(); ok {
		hcl.WriteString(fmt.Sprintf("  name           = %q\n", *name))
	}

	// Status
	if status, ok := policy.GetStatusOk(); ok {
		hcl.WriteString(fmt.Sprintf("  status         = %q\n", string(*status)))
	}

	// Trigger - Always export to match provider schema defaults
	// If API doesn't return trigger, use schema default values
	hcl.WriteString("\n")
	hcl.WriteString("  trigger = {\n")

	// Type (default: "AUTHENTICATION")
	triggerType := "AUTHENTICATION"
	if trigger, ok := policy.GetTriggerOk(); ok && trigger != nil {
		if t, typeOk := trigger.GetTypeOk(); typeOk && t != nil {
			triggerType = *t
		}
	}
	hcl.WriteString(fmt.Sprintf("    type = %q\n", triggerType))

	// Configuration (always include with defaults if not present)
	hcl.WriteString("\n")
	hcl.WriteString("    configuration = {\n")

	// MFA configuration (defaults: enabled=false, time=0, time_format="min")
	mfaEnabled := false
	mfaTime := float32(0)
	mfaTimeFormat := "min"
	if trigger, ok := policy.GetTriggerOk(); ok && trigger != nil {
		if config, configOk := trigger.GetConfigurationOk(); configOk && config != nil {
			if mfa, mfaOk := config.GetMfaOk(); mfaOk {
				if enabled, enabledOk := mfa.GetEnabledOk(); enabledOk {
					mfaEnabled = *enabled
				}
				if time, timeOk := mfa.GetTimeOk(); timeOk {
					mfaTime = *time
				}
				if timeFormat, formatOk := mfa.GetTimeFormatOk(); formatOk && timeFormat != nil {
					mfaTimeFormat = *timeFormat
				}
			}
		}
	}
	hcl.WriteString("      mfa = {\n")
	hcl.WriteString(fmt.Sprintf("        enabled     = %t\n", mfaEnabled))
	hcl.WriteString(fmt.Sprintf("        time        = %d\n", int64(mfaTime)))
	hcl.WriteString(fmt.Sprintf("        time_format = %q\n", mfaTimeFormat))
	hcl.WriteString("      }\n")

	// Password configuration (defaults: enabled=false, time=0, time_format="min")
	pwdEnabled := false
	pwdTime := float32(0)
	pwdTimeFormat := "min"
	if trigger, ok := policy.GetTriggerOk(); ok && trigger != nil {
		if config, configOk := trigger.GetConfigurationOk(); configOk && config != nil {
			if pwd, pwdOk := config.GetPwdOk(); pwdOk {
				if enabled, enabledOk := pwd.GetEnabledOk(); enabledOk {
					pwdEnabled = *enabled
				}
				if time, timeOk := pwd.GetTimeOk(); timeOk {
					pwdTime = *time
				}
				if timeFormat, formatOk := pwd.GetTimeFormatOk(); formatOk && timeFormat != nil {
					pwdTimeFormat = *timeFormat
				}
			}
		}
	}
	hcl.WriteString("\n")
	hcl.WriteString("      pwd = {\n")
	hcl.WriteString(fmt.Sprintf("        enabled     = %t\n", pwdEnabled))
	hcl.WriteString(fmt.Sprintf("        time        = %d\n", int64(pwdTime)))
	hcl.WriteString(fmt.Sprintf("        time_format = %q\n", pwdTimeFormat))
	hcl.WriteString("      }\n")

	hcl.WriteString("    }\n")
	hcl.WriteString("  }\n")

	// Flow distributions
	if distributions, ok := policy.GetFlowDistributionsOk(); ok && len(distributions) > 0 {
		hcl.WriteString("\n")
		hcl.WriteString("  flow_distributions = [\n")

		for _, dist := range distributions {
			hcl.WriteString("    {\n")

			// Flow ID - use graph for reference if available
			if flowID, ok := dist.GetIdOk(); ok {
				if skipDeps {
					hcl.WriteString(fmt.Sprintf("      id      = %q\n", *flowID))
				} else {
					if graph != nil {
						flowRef, err := resolver.GenerateTerraformReference(graph, "pingone_davinci_flow", *flowID, "id")
						if err != nil {
							// Generate TODO placeholder for missing flow dependency
							placeholder := resolver.GenerateTODOPlaceholder("pingone_davinci_flow", *flowID, err)
							hcl.WriteString(fmt.Sprintf("      id      = %s\n", placeholder))
						} else {
							hcl.WriteString(fmt.Sprintf("      id      = %s\n", flowRef))
						}
					} else {
						// Fallback: use raw UUID with comment
						hcl.WriteString(fmt.Sprintf("      id      = %q # TODO: Replace with pingone_davinci_flow.<resource_name>.id\n", *flowID))
					}
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
