package converter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/utils"
)

// ConvertFlowToHCL converts a DaVinci flow JSON structure to Terraform HCL
// This implements Part 2.1 Phase 2.1 - Comprehensive Flow Structure Conversion
// If skipDependencies is true, connection IDs will be left as hardcoded strings instead of Terraform references
func ConvertFlowToHCL(flowData map[string]interface{}, environmentID string, skipDependencies bool) (string, error) {
	var hcl strings.Builder

	// Generate resource name from flow name using pingcli-compatible sanitization
	resourceName := utils.SanitizeResourceName(getString(flowData, "name"))

	hcl.WriteString(fmt.Sprintf("resource \"pingone_davinci_flow\" \"%s\" {\n", resourceName))

	// Handle environment_id - quote if it's a UUID string, otherwise use as-is (for var.environment_id)
	if strings.HasPrefix(environmentID, "var.") {
		hcl.WriteString(fmt.Sprintf("  environment_id = %s\n\n", environmentID))
	} else {
		hcl.WriteString(fmt.Sprintf("  environment_id = %q\n\n", environmentID))
	}

	// Required: name
	if name := getString(flowData, "name"); name != "" {
		hcl.WriteString(fmt.Sprintf("  name        = %s\n", quoteString(name)))
	}

	// Optional: description
	if description := getString(flowData, "description"); description != "" {
		hcl.WriteString(fmt.Sprintf("  description = %s\n", quoteString(description)))
	}

	// Optional: color (flowColor in JSON)
	if color := getString(flowData, "flowColor"); color != "" {
		hcl.WriteString(fmt.Sprintf("  color       = %s\n", quoteString(color)))
	}

	// Settings block
	if settings, ok := flowData["settings"].(map[string]interface{}); ok && len(settings) > 0 {
		hcl.WriteString("\n")
		if err := writeSettingsBlock(&hcl, settings); err != nil {
			return "", fmt.Errorf("failed to write settings: %w", err)
		}
	}

	// Graph data block - complex nested structure
	if graphData, ok := flowData["graphData"].(map[string]interface{}); ok {
		hcl.WriteString("\n")
		if err := writeGraphDataBlock(&hcl, graphData, skipDependencies); err != nil {
			return "", fmt.Errorf("failed to write graph_data: %w", err)
		}
	}

	// Input schema list
	if inputSchema, ok := flowData["inputSchema"].([]interface{}); ok && len(inputSchema) > 0 {
		hcl.WriteString("\n")
		if err := writeInputSchemaBlock(&hcl, inputSchema); err != nil {
			return "", fmt.Errorf("failed to write input_schema: %w", err)
		}
	}

	// Output schema object
	if outputSchema, ok := flowData["outputSchema"].(map[string]interface{}); ok && len(outputSchema) > 0 {
		hcl.WriteString("\n")
		if err := writeOutputSchemaBlock(&hcl, outputSchema); err != nil {
			return "", fmt.Errorf("failed to write output_schema: %w", err)
		}
	}

	// Trigger block
	if trigger, ok := flowData["trigger"].(map[string]interface{}); ok {
		hcl.WriteString("\n")
		if err := writeTriggerBlock(&hcl, trigger); err != nil {
			return "", fmt.Errorf("failed to write trigger: %w", err)
		}
	}

	hcl.WriteString("}\n")
	return hcl.String(), nil
}

// writeSettingsBlock writes the settings nested block
func writeSettingsBlock(hcl *strings.Builder, settings map[string]interface{}) error {
	hcl.WriteString("  settings = {\n")

	// Get keys and sort for consistent output
	keys := make([]string, 0, len(settings))
	for k := range settings {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Map JSON field names to HCL attribute names
	fieldNameMap := map[string]string{
		"csp":                           "csp",
		"css":                           "css",
		"cssLinks":                      "css_links",
		"customErrorScreenBrandLogoUrl": "custom_error_screen_brand_logo_url",
		"customErrorShowFooter":         "custom_error_show_footer",
		"customFaviconLink":             "custom_favicon_link",
		"customLogoURLSelection":        "custom_logo_urlselection",
		"customTitle":                   "custom_title",
		"defaultErrorScreenBrandLogo":   "default_error_screen_brand_logo",
		"flowHttpTimeoutInSeconds":      "flow_http_timeout_in_seconds",
		"flowTimeoutInSeconds":          "flow_timeout_in_seconds",
		"intermediateLoadingScreenCSS":  "intermediate_loading_screen_css",
		"intermediateLoadingScreenHTML": "intermediate_loading_screen_html",
		"jsCustomFlowPlayer":            "js_custom_flow_player",
		"jsLinks":                       "js_links",
		"logLevel":                      "log_level",
		"scrubSensitiveInfo":            "scrub_sensitive_info",
		"sensitiveInfoFields":           "sensitive_info_fields",
		"useCSP":                        "use_csp",
		"useCustomCSS":                  "use_custom_css",
		"useCustomFlowPlayer":           "use_custom_flow_player",
		"useCustomScript":               "use_custom_script",
		"useIntermediateLoadingScreen":  "use_intermediate_loading_screen",
		"validateOnSave":                "validate_on_save",
	}

	for _, key := range keys {
		value := settings[key]
		hclKey := fieldNameMap[key]
		if hclKey == "" {
			hclKey = toSnakeCase(key)
		}

		// Special handling for js_links - array of objects
		if key == "jsLinks" {
			if jsLinks, ok := value.([]interface{}); ok && len(jsLinks) > 0 {
				hcl.WriteString("    js_links = [\n")
				for i, linkInterface := range jsLinks {
					if link, ok := linkInterface.(map[string]interface{}); ok {
						hcl.WriteString("      {\n")
						// Write all required fields for js_links
						if crossorigin := getString(link, "crossorigin"); crossorigin != "" {
							hcl.WriteString(fmt.Sprintf("        crossorigin    = %s\n", quoteString(crossorigin)))
						}
						if deferVal, ok := link["defer"].(bool); ok {
							hcl.WriteString(fmt.Sprintf("        defer          = %t\n", deferVal))
						}
						if integrity := getString(link, "integrity"); integrity != "" {
							hcl.WriteString(fmt.Sprintf("        integrity      = %s\n", quoteString(integrity)))
						}
						if label := getString(link, "label"); label != "" {
							hcl.WriteString(fmt.Sprintf("        label          = %s\n", quoteString(label)))
						}
						if referrerpolicy := getString(link, "referrerpolicy"); referrerpolicy != "" {
							hcl.WriteString(fmt.Sprintf("        referrerpolicy = %s\n", quoteString(referrerpolicy)))
						}
						if linkType := getString(link, "type"); linkType != "" {
							hcl.WriteString(fmt.Sprintf("        type           = %s\n", quoteString(linkType)))
						}
						if value := getString(link, "value"); value != "" {
							hcl.WriteString(fmt.Sprintf("        value          = %s\n", quoteString(value)))
						}
						hcl.WriteString("      }")
						if i < len(jsLinks)-1 {
							hcl.WriteString(",")
						}
						hcl.WriteString("\n")
					}
				}
				hcl.WriteString("    ]\n")
			}
			continue
		}

		switch v := value.(type) {
		case string:
			hcl.WriteString(fmt.Sprintf("    %-36s = %s\n", hclKey, quoteString(v)))
		case float64:
			hcl.WriteString(fmt.Sprintf("    %-36s = %d\n", hclKey, int(v)))
		case bool:
			hcl.WriteString(fmt.Sprintf("    %-36s = %t\n", hclKey, v))
		case []interface{}:
			// Handle array fields like cssLinks, sensitiveInfoFields
			hcl.WriteString(fmt.Sprintf("    %s = [", hclKey))
			for i, item := range v {
				if i > 0 {
					hcl.WriteString(", ")
				}
				hcl.WriteString(quoteString(fmt.Sprintf("%v", item)))
			}
			hcl.WriteString("]\n")
		}
	}

	hcl.WriteString("  }\n")
	return nil
}

// writeGraphDataBlock writes the graph_data nested block
func writeGraphDataBlock(hcl *strings.Builder, graphData map[string]interface{}, skipDependencies bool) error {
	hcl.WriteString("  graph_data = {\n")

	// Elements (nodes and edges) - most complex part
	if elements, ok := graphData["elements"].(map[string]interface{}); ok {
		hcl.WriteString("    elements = {\n")

		// Nodes
		if nodes, ok := elements["nodes"].([]interface{}); ok {
			if err := writeNodesBlock(hcl, nodes, skipDependencies); err != nil {
				return fmt.Errorf("failed to write nodes: %w", err)
			}
		}

		// Edges
		if edges, ok := elements["edges"].([]interface{}); ok {
			if err := writeEdgesBlock(hcl, edges); err != nil {
				return fmt.Errorf("failed to write edges: %w", err)
			}
		}

		hcl.WriteString("    }\n\n")
	}

	// Pan object
	if pan, ok := graphData["pan"].(map[string]interface{}); ok {
		hcl.WriteString("    pan = {\n")
		if x, ok := pan["x"].(float64); ok {
			hcl.WriteString(fmt.Sprintf("      x = %g\n", x))
		}
		if y, ok := pan["y"].(float64); ok {
			hcl.WriteString(fmt.Sprintf("      y = %g\n", y))
		}
		hcl.WriteString("    }\n\n")
	}

	// Simple fields
	if zoom, ok := graphData["zoom"].(float64); ok {
		hcl.WriteString(fmt.Sprintf("    zoom                  = %d\n", int(zoom)))
	}
	if minZoom, ok := graphData["minZoom"].(float64); ok {
		hcl.WriteString(fmt.Sprintf("    min_zoom              = %g\n", minZoom))
	}
	if maxZoom, ok := graphData["maxZoom"].(float64); ok {
		hcl.WriteString(fmt.Sprintf("    max_zoom              = %g\n", maxZoom))
	}
	if zoomingEnabled, ok := graphData["zoomingEnabled"].(bool); ok {
		hcl.WriteString(fmt.Sprintf("    zooming_enabled       = %t\n", zoomingEnabled))
	}
	if panningEnabled, ok := graphData["panningEnabled"].(bool); ok {
		hcl.WriteString(fmt.Sprintf("    panning_enabled       = %t\n", panningEnabled))
	}
	if userZoomingEnabled, ok := graphData["userZoomingEnabled"].(bool); ok {
		hcl.WriteString(fmt.Sprintf("    user_zooming_enabled  = %t\n", userZoomingEnabled))
	}
	if userPanningEnabled, ok := graphData["userPanningEnabled"].(bool); ok {
		hcl.WriteString(fmt.Sprintf("    user_panning_enabled  = %t\n", userPanningEnabled))
	}
	if boxSelectionEnabled, ok := graphData["boxSelectionEnabled"].(bool); ok {
		hcl.WriteString(fmt.Sprintf("    box_selection_enabled = %t\n", boxSelectionEnabled))
	}

	// Renderer - uses jsonencode() because it's jsontypes.NormalizedType
	if renderer, ok := graphData["renderer"].(map[string]interface{}); ok {
		rendererJSON, err := json.Marshal(renderer)
		if err != nil {
			return fmt.Errorf("failed to marshal renderer: %w", err)
		}
		hcl.WriteString(fmt.Sprintf("\n    renderer = jsonencode(%s)\n", string(rendererJSON)))
	}

	hcl.WriteString("  }\n")
	return nil
}

// writeNodesBlock writes the nodes array within elements
func writeNodesBlock(hcl *strings.Builder, nodes []interface{}, skipDependencies bool) error {
	hcl.WriteString("      nodes = [\n")

	for i, nodeInterface := range nodes {
		node, ok := nodeInterface.(map[string]interface{})
		if !ok {
			continue
		}

		hcl.WriteString("        {\n")

		// Node data block - required
		if data, ok := node["data"].(map[string]interface{}); ok {
			hcl.WriteString("          data = {\n")

			// Required: id and node_type
			if id := getString(data, "id"); id != "" {
				hcl.WriteString(fmt.Sprintf("            id              = %s\n", quoteString(id)))
			}
			if nodeType := getString(data, "nodeType"); nodeType != "" {
				hcl.WriteString(fmt.Sprintf("            node_type       = %s\n", quoteString(nodeType)))
			}

			// Optional fields - connection_id needs special handling
			if connectionID := getString(data, "connectionId"); connectionID != "" {
				if skipDependencies {
					// Use hardcoded ID when skipping dependencies
					hcl.WriteString(fmt.Sprintf("            connection_id   = %s\n", quoteString(connectionID)))
				} else {
					// Generate Terraform reference
					connectorID := getString(data, "connectorId")
					ref := generateConnectionReference(connectorID, connectionID)
					hcl.WriteString(fmt.Sprintf("            connection_id   = %s\n", ref))
				}
			}

			if connectorID := getString(data, "connectorId"); connectorID != "" {
				hcl.WriteString(fmt.Sprintf("            connector_id    = %s\n", quoteString(connectorID)))
			}
			if name := getString(data, "name"); name != "" {
				hcl.WriteString(fmt.Sprintf("            name            = %s\n", quoteString(name)))
			}
			if label := getString(data, "label"); label != "" {
				hcl.WriteString(fmt.Sprintf("            label           = %s\n", quoteString(label)))
			}
			if status := getString(data, "status"); status != "" {
				hcl.WriteString(fmt.Sprintf("            status          = %s\n", quoteString(status)))
			}
			if capabilityName := getString(data, "capabilityName"); capabilityName != "" {
				hcl.WriteString(fmt.Sprintf("            capability_name = %s\n", quoteString(capabilityName)))
			}
			if nodeTypeField := getString(data, "type"); nodeTypeField != "" {
				hcl.WriteString(fmt.Sprintf("            type            = %s\n", quoteString(nodeTypeField)))
			}

			// Properties - uses jsonencode() because it's jsontypes.NormalizedType
			if properties, ok := data["properties"].(map[string]interface{}); ok {
				propertiesJSON, err := json.Marshal(properties)
				if err != nil {
					return fmt.Errorf("failed to marshal properties: %w", err)
				}
				// Use base64decode with heredoc to avoid HCL parsing issues with special characters
				hcl.WriteString("            properties = base64decode(<<-EOT\n")
				encoded := base64.StdEncoding.EncodeToString(propertiesJSON)
				hcl.WriteString(encoded)
				hcl.WriteString("\nEOT\n)\n")
			}

			hcl.WriteString("          }\n")
		}

		// Position block - optional
		if position, ok := node["position"].(map[string]interface{}); ok {
			hcl.WriteString("          position = {\n")
			if x, ok := position["x"].(float64); ok {
				hcl.WriteString(fmt.Sprintf("            x = %g\n", x))
			}
			if y, ok := position["y"].(float64); ok {
				hcl.WriteString(fmt.Sprintf("            y = %g\n", y))
			}
			hcl.WriteString("          }\n")
		}

		// Other node attributes
		if group := getString(node, "group"); group != "" {
			hcl.WriteString(fmt.Sprintf("          group      = %s\n", quoteString(group)))
		}
		if removed, ok := node["removed"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          removed    = %t\n", removed))
		}
		if selected, ok := node["selected"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          selected   = %t\n", selected))
		}
		if selectable, ok := node["selectable"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          selectable = %t\n", selectable))
		}
		if locked, ok := node["locked"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          locked     = %t\n", locked))
		}
		if grabbable, ok := node["grabbable"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          grabbable  = %t\n", grabbable))
		}
		if pannable, ok := node["pannable"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          pannable   = %t\n", pannable))
		}
		// Always include classes field (even if empty string)
		classes := getString(node, "classes")
		hcl.WriteString(fmt.Sprintf("          classes    = %s\n", quoteString(classes)))

		hcl.WriteString("        }")
		// Add comma for all but last element
		if i < len(nodes)-1 {
			hcl.WriteString(",")
		}
		hcl.WriteString("\n")
	}

	hcl.WriteString("      ]\n")
	return nil
}

// writeEdgesBlock writes the edges array within elements
func writeEdgesBlock(hcl *strings.Builder, edges []interface{}) error {
	if len(edges) == 0 {
		hcl.WriteString("      edges = []\n")
		return nil
	}

	hcl.WriteString("      edges = [\n")

	for i, edgeInterface := range edges {
		edge, ok := edgeInterface.(map[string]interface{})
		if !ok {
			continue
		}

		hcl.WriteString("        {\n")

		// Edge data block - required
		if data, ok := edge["data"].(map[string]interface{}); ok {
			hcl.WriteString("          data = {\n")

			// Required: id, source, target
			if id := getString(data, "id"); id != "" {
				hcl.WriteString(fmt.Sprintf("            id     = %s\n", quoteString(id)))
			}
			if source := getString(data, "source"); source != "" {
				hcl.WriteString(fmt.Sprintf("            source = %s\n", quoteString(source)))
			}
			if target := getString(data, "target"); target != "" {
				hcl.WriteString(fmt.Sprintf("            target = %s\n", quoteString(target)))
			}

			hcl.WriteString("          }\n")
		}

		// Optional: position object (rarely used for edges but supported)
		if position, ok := edge["position"].(map[string]interface{}); ok {
			hcl.WriteString("          position = {\n")
			if x, ok := position["x"].(float64); ok {
				hcl.WriteString(fmt.Sprintf("            x = %g\n", x))
			}
			if y, ok := position["y"].(float64); ok {
				hcl.WriteString(fmt.Sprintf("            y = %g\n", y))
			}
			hcl.WriteString("          }\n")
		}

		// Optional edge attributes
		if group := getString(edge, "group"); group != "" {
			hcl.WriteString(fmt.Sprintf("          group      = %s\n", quoteString(group)))
		}
		if removed, ok := edge["removed"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          removed    = %t\n", removed))
		}
		if selected, ok := edge["selected"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          selected   = %t\n", selected))
		}
		if selectable, ok := edge["selectable"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          selectable = %t\n", selectable))
		}
		if locked, ok := edge["locked"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          locked     = %t\n", locked))
		}
		if grabbable, ok := edge["grabbable"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          grabbable  = %t\n", grabbable))
		}
		if pannable, ok := edge["pannable"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("          pannable   = %t\n", pannable))
		}
		// Always include classes field (even if empty string)
		classes := getString(edge, "classes")
		hcl.WriteString(fmt.Sprintf("          classes    = %s\n", quoteString(classes)))

		hcl.WriteString("        }")
		// Add comma for all but last element
		if i < len(edges)-1 {
			hcl.WriteString(",")
		}
		hcl.WriteString("\n")
	}

	hcl.WriteString("      ]\n")
	return nil
}

// writeInputSchemaBlock writes the input_schema list
func writeInputSchemaBlock(hcl *strings.Builder, inputSchema []interface{}) error {
	hcl.WriteString("  input_schema = [\n")

	for i, schemaInterface := range inputSchema {
		schema, ok := schemaInterface.(map[string]interface{})
		if !ok {
			continue
		}

		hcl.WriteString("    {\n")

		if propertyName := getString(schema, "propertyName"); propertyName != "" {
			hcl.WriteString(fmt.Sprintf("      property_name           = %s\n", quoteString(propertyName)))
		}
		if preferredDataType := getString(schema, "preferredDataType"); preferredDataType != "" {
			hcl.WriteString(fmt.Sprintf("      preferred_data_type     = %s\n", quoteString(preferredDataType)))
		}
		if preferredControlType := getString(schema, "preferredControlType"); preferredControlType != "" {
			hcl.WriteString(fmt.Sprintf("      preferred_control_type  = %s\n", quoteString(preferredControlType)))
		}
		if required, ok := schema["required"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("      required                = %t\n", required))
		}
		if isExpanded, ok := schema["isExpanded"].(bool); ok {
			hcl.WriteString(fmt.Sprintf("      is_expanded             = %t\n", isExpanded))
		}
		// Always include description field (even if empty string)
		description := getString(schema, "description")
		hcl.WriteString(fmt.Sprintf("      description             = %s\n", quoteString(description)))

		hcl.WriteString("    }")
		// Add comma for all but last element
		if i < len(inputSchema)-1 {
			hcl.WriteString(",")
		}
		hcl.WriteString("\n")
	}

	hcl.WriteString("  ]\n")
	return nil
}

// writeOutputSchemaBlock writes the output_schema object
func writeOutputSchemaBlock(hcl *strings.Builder, outputSchema map[string]interface{}) error {
	hcl.WriteString("  output_schema = {\n")

	// The output field typically contains a JSON object that should be encoded
	if output, ok := outputSchema["output"]; ok {
		// Convert output to JSON string
		outputBytes, err := json.Marshal(output)
		if err != nil {
			return fmt.Errorf("failed to marshal output schema: %w", err)
		}
		hcl.WriteString(fmt.Sprintf("    output = jsonencode(%s)\n", string(outputBytes)))
	}

	hcl.WriteString("  }\n")
	return nil
}

// writeTriggerBlock writes the trigger nested block
func writeTriggerBlock(hcl *strings.Builder, trigger map[string]interface{}) error {
	hcl.WriteString("  trigger = {\n")

	if triggerType := getString(trigger, "type"); triggerType != "" {
		hcl.WriteString(fmt.Sprintf("    type = %s\n", quoteString(triggerType)))
	}

	if config, ok := trigger["configuration"].(map[string]interface{}); ok {
		hcl.WriteString("    configuration = {\n")

		// MFA configuration
		if mfa, ok := config["mfa"].(map[string]interface{}); ok {
			hcl.WriteString("      mfa = {\n")
			if enabled, ok := mfa["enabled"].(bool); ok {
				hcl.WriteString(fmt.Sprintf("        enabled     = %t\n", enabled))
			}
			if time, ok := mfa["time"].(float64); ok {
				hcl.WriteString(fmt.Sprintf("        time        = %d\n", int(time)))
			}
			if timeFormat := getString(mfa, "timeFormat"); timeFormat != "" {
				hcl.WriteString(fmt.Sprintf("        time_format = %s\n", quoteString(timeFormat)))
			}
			hcl.WriteString("      }\n")
		}

		// Password configuration
		if pwd, ok := config["pwd"].(map[string]interface{}); ok {
			hcl.WriteString("      pwd = {\n")
			if enabled, ok := pwd["enabled"].(bool); ok {
				hcl.WriteString(fmt.Sprintf("        enabled     = %t\n", enabled))
			}
			if time, ok := pwd["time"].(float64); ok {
				hcl.WriteString(fmt.Sprintf("        time        = %d\n", int(time)))
			}
			if timeFormat := getString(pwd, "timeFormat"); timeFormat != "" {
				hcl.WriteString(fmt.Sprintf("        time_format = %s\n", quoteString(timeFormat)))
			}
			hcl.WriteString("      }\n")
		}

		hcl.WriteString("    }\n")
	}

	hcl.WriteString("  }\n")
	return nil
}

// Helper functions

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func quoteString(s string) string {
	// Escape special characters in HCL strings
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return fmt.Sprintf("%q", s)
}

// generateResourceName is deprecated - use utils.SanitizeResourceName instead
// Kept for backwards compatibility in case external code references it
func generateResourceName(name string) string {
	return utils.SanitizeResourceName(name)
}

func generateConnectionReference(connectorID, connectionID string) string {
	// Generate Terraform reference for connection_id
	// Format: pingone_davinci_connector_instance.<connector_id>_<connection_id>.id
	connectorName := toSnakeCase(connectorID)
	return fmt.Sprintf("pingone_davinci_connector_instance.%s_%s.id", connectorName, connectionID)
}

func toSnakeCase(s string) string {
	// Convert to lowercase and remove non-alphanumeric characters
	// This creates a simple identifier without underscores between camelCase words
	// Example: "httpConnector" -> "httpconnector"
	re := regexp.MustCompile(`[^\w]+`)
	result := re.ReplaceAllString(s, "")
	return strings.ToLower(result)
}
