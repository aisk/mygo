package transpiler

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aisk/mygo/internal/diff"
)

func TestTranspiler(t *testing.T) {
	testdata := "testdata"
	
	entries, err := os.ReadDir(testdata)
	if err != nil {
		t.Fatalf("Failed to read testdata directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".mygo") {
			mygoFile := filepath.Join(testdata, entry.Name())
			expectedFile := filepath.Join(testdata, strings.TrimSuffix(entry.Name(), ".mygo")+"_expected.go")
			
			t.Run(entry.Name(), func(t *testing.T) {
				// Read .mygo file
				mygoContent, err := os.ReadFile(mygoFile)
				if err != nil {
					t.Fatalf("Failed to read .mygo file: %v", err)
				}
				
				// Read expected output
				expectedContent, err := os.ReadFile(expectedFile)
				if err != nil {
					t.Fatalf("Failed to read expected.go file: %v", err)
				}
				
				// Transpile the .mygo content
				input := bytes.NewReader(mygoContent)
				var output bytes.Buffer
				
				err = Transpile(input, &output)
				if err != nil {
					t.Fatalf("Transpile failed: %v", err)
				}
				
				// Compare with expected
				if output.String() != string(expectedContent) {
					diffOutput := diff.Diff(expectedFile, expectedContent, "transpiled", []byte(output.String()))
					t.Errorf("Transpiled result does not match expected:\n%s", diffOutput)
				}
			})
		}
	}
}