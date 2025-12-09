package exporter

import (
	"github.com/samir-gandhi/davinci-terraform-converter/internal/utils"
)

// NamedHCL pairs a Terraform resource name with its HCL body.
type NamedHCL = utils.NamedHCL

// joinHCLBlocksSorted sorts by Name alphabetically and joins with blank lines.
func joinHCLBlocksSorted(blocks []NamedHCL) string { return utils.JoinHCLBlocksSorted(blocks) }

// sortAllResourceBlocks takes a full HCL string and returns it with all
// Terraform resource blocks sorted alphabetically by their resource name.
// It preserves any leading non-resource content (headers, provider, variables)
// before the first resource block, and preserves trailing newlines.
func sortAllResourceBlocks(hcl string) string { return utils.SortAllResourceBlocks(hcl) }
