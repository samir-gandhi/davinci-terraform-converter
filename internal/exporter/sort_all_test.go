package exporter

import (
	"testing"

	"github.com/samir-gandhi/davinci-terraform-converter/internal/utils"
)

func TestSortAllResourceBlocks(t *testing.T) {
	input := "# Header\n\n" +
		"resource \"x\" \"zeta\" {\n}\n\n" +
		"resource \"x\" \"alpha\" {\n}\n\n" +
		"resource \"x\" \"beta\" {\n}\n\n" +
		"# Footer\n"
	out := utils.SortAllResourceBlocks(input)
	// Expect order: alpha, beta, zeta under the header and before the footer
	headerEnd := "# Header\n\n"
	if len(out) < len(headerEnd) || out[:len(headerEnd)] != headerEnd {
		t.Fatalf("header not preserved")
	}
	if indexOf(out, "\"alpha\"") > indexOf(out, "\"beta\"") || indexOf(out, "\"beta\"") > indexOf(out, "\"zeta\"") {
		t.Fatalf("resources not globally sorted: %s", out)
	}
}
