package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var testBin string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "tui-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmp)

	testBin = filepath.Join(tmp, "tui")
	if runtime.GOOS == "windows" {
		testBin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", testBin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s\n", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestCLI_Check(t *testing.T) {
	t2Files, _ := filepath.Glob("testdata/*.t2")
	if len(t2Files) == 0 {
		t.Skip("no testdata/*.t2 files found")
	}

	for _, t2File := range t2Files {
		t.Run(filepath.Base(t2File), func(t *testing.T) {
			cmd := exec.Command(testBin, "check", t2File)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("check %s failed: %v\n%s", t2File, err, out)
			}
		})
	}
}

func TestCLI_Fmt_Stdout(t *testing.T) {
	t2Files, _ := filepath.Glob("testdata/*.t2")
	if len(t2Files) == 0 {
		t.Skip("no testdata/*.t2 files found")
	}

	for _, t2File := range t2Files {
		t.Run(filepath.Base(t2File), func(t *testing.T) {
			cmd := exec.Command(testBin, "fmt", "--stdout", t2File)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("fmt --stdout %s failed: %v\n%s", t2File, err, out)
			}
			if len(out) == 0 {
				t.Errorf("fmt --stdout %s produced empty output", t2File)
			}
		})
	}
}

func TestCLI_Version(t *testing.T) {
	cmd := exec.Command(testBin, "version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("version failed: %v\n%s", err, out)
	}
}

func TestCLI_Help(t *testing.T) {
	cmd := exec.Command(testBin, "help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("help failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Error("help output should not be empty")
	}
}
