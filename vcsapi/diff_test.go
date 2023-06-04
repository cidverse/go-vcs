package vcsapi

import (
	"testing"
)

func TestDiffSimple(t *testing.T) {
	srcContent := "Hello, world!"
	dstContent := "Hello, Go!"

	expectedDiff := []VCSDiffLine{
		{Operation: -1, Content: "Hello, world!"},
		{Operation: 1, Content: "Hello, Go!"},
	}

	diffLines := Diff(srcContent, dstContent)
	for i, expected := range expectedDiff {
		actual := diffLines[i]

		if actual.Operation != expected.Operation {
			t.Errorf("Unexpected diff operation at index %d. Expected: %d, Got: %d", i, expected.Operation, actual.Operation)
		}
		if actual.Content != expected.Content {
			t.Errorf("Unexpected diff content at index %d. Expected: %s, Got: %s", i, expected.Content, actual.Content)
		}
	}
}

func TestDiffMultiline(t *testing.T) {
	srcContent := "package main\nimport \"fmt\"\nfunc main() {\n   fmt.Println(\"Hello, World!\")\n}"
	dstContent := "package main\nimport (\n        \"fmt\"\n        \"strings\"\n)\nfunc main() {\n   fmt.Println(\"Hello, World!\")\n}"
	expectedDiff := []VCSDiffLine{
		{Operation: 0, Content: `package main`},
		{Operation: 0, Content: ``},
		{Operation: -1, Content: `import "fmt"`},
		{Operation: -1, Content: ``},
		{Operation: 1, Content: "import ("},
		{Operation: 1, Content: "        \"fmt\""},
		{Operation: 1, Content: "        \"strings\""},
		{Operation: 1, Content: ")"},
		{Operation: 1, Content: ``},
		{Operation: 0, Content: `func main() {`},
		{Operation: 0, Content: `   fmt.Println("Hello, World!")`},
		{Operation: 0, Content: `}`},
	}

	diffLines := Diff(srcContent, dstContent)
	for i, expected := range expectedDiff {
		actual := diffLines[i]

		if actual.Operation != expected.Operation {
			t.Errorf("Unexpected diff operation at index %d. Expected: %d, Got: %d", i, expected.Operation, actual.Operation)
		}
		if actual.Content != expected.Content {
			t.Errorf("Unexpected diff content at index %d. Expected: %s, Got: %s", i, expected.Content, actual.Content)
		}
	}
}
