package transpiler

import (
	"bytes"
	goast "go/ast"
	"go/importer"
	goparser "go/parser"
	gotoken "go/token"
	"go/types"
	"os"
	"os/exec"
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

				// The expected output must be valid, type-correct Go.
				typeCheck(t, expectedFile, expectedContent)
			})
		}
	}
}

func typeCheck(t *testing.T, filename string, src []byte) {
	t.Helper()
	fset := gotoken.NewFileSet()
	file, err := goparser.ParseFile(fset, filename, src, 0)
	if err != nil {
		t.Fatalf("expected output does not parse: %v", err)
	}
	conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil)}
	if _, err := conf.Check("main", fset, []*goast.File{file}, nil); err != nil {
		t.Errorf("expected output does not type-check: %v", err)
	}
}

// TestTranspiledPrograms transpiles each testdata/run/<name>.mygo, runs the
// result and compares its output with <name>.out, so evaluation order,
// scoping and short-circuit behavior are checked, not just the text.
func TestTranspiledPrograms(t *testing.T) {
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go tool not found")
	}
	files, err := filepath.Glob(filepath.Join("testdata", "run", "*.mygo"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mygoFile := range files {
		name := strings.TrimSuffix(filepath.Base(mygoFile), ".mygo")
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(mygoFile)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(strings.TrimSuffix(mygoFile, ".mygo") + ".out")
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			if err := Transpile(bytes.NewReader(src), &output); err != nil {
				t.Fatalf("Transpile failed: %v", err)
			}
			goFile := filepath.Join(t.TempDir(), "main.go")
			if err := os.WriteFile(goFile, output.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := exec.Command(goTool, "run", goFile).CombinedOutput()
			if err != nil {
				t.Fatalf("go run failed: %v\n%s\ngenerated code:\n%s", err, got, output.String())
			}
			if string(got) != string(want) {
				t.Errorf("output mismatch\n got:\n%s\nwant:\n%s\ngenerated code:\n%s", got, want, output.String())
			}
		})
	}
}

func TestTranspileErrors(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "package-level var",
			src:  "package main\n\nvar x = f()?\n",
			want: "3:12: ? is not supported in this position",
		},
		{
			name: "nested in call statement",
			src:  "package main\n\nfunc g() error {\n\th(f()?)\n\treturn nil\n}\n",
			want: "4:7: ? is not supported in this position",
		},
		{
			name: "for init",
			src:  "package main\n\nfunc g() error {\n\tfor x := f()?; x < 3; x++ {\n\t}\n\treturn nil\n}\n",
			want: "4:6: ? is not supported in this statement position",
		},
		{
			name: "function without error result",
			src:  "package main\n\nfunc g() int {\n\tx := f()?\n\treturn x\n}\n",
			want: "4:2: try expression used in function that does not return an error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			err := Transpile(strings.NewReader(tt.src), &output)
			if err == nil {
				t.Fatalf("Transpile succeeded, want error %q; output:\n%s", tt.want, output.String())
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want it to contain %q", err, tt.want)
			}
		})
	}
}
