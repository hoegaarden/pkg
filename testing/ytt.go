package testing

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type DV interface {
	Prepare(tmpDir string, idx int) ([]string, error)
}

type fileDV struct {
	content string
}

var _ DV = fileDV{}

func FileDV(lines ...string) DV {
	return fileDV{strings.Join(lines, "\n")}
}

func (d fileDV) Prepare(tmpDir string, idx int) ([]string, error) {
	file := filepath.Join(tmpDir, fmt.Sprintf("%d.yml", idx))
	if err := os.WriteFile(file, []byte(d.content), 0644); err != nil {
		return []string{}, fmt.Errorf("cannot write data-values file %v: %v", file, err)
	}
	return []string{"-f", file}, nil
}

type plainDV struct {
	name  string
	value string
}

var _ DV = plainDV{}

func PlainDV(name, value string) DV {
	return plainDV{name, value}
}

func (d plainDV) Prepare(_ string, _ int) ([]string, error) {
	return []string{"-v", fmt.Sprintf("%s=%s", d.name, d.value)}, nil
}

type JsonProvider interface {
	Provide(*testing.T) []byte
}

type Ytt struct {
	Cmd    string
	SrcDir string
	Env    []string
	DVs    []DV
}

var _ JsonProvider = Ytt{}

// TODO: use ytt as a library instead of forking the CLI
func (y Ytt) Provide(t *testing.T) []byte {
	t.Helper()

	args := []string{"-o", "json", "-f", or(y.SrcDir, "src")}

	tmpDir := t.TempDir()
	for i, dv := range y.DVs {
		addArgs, err := dv.Prepare(tmpDir, i)
		if err != nil {
			t.Fatalf("cannot prepare data-values: %v", err)
		}
		args = append(args, addArgs...)
	}

	cmd := exec.Command(
		or(y.Cmd, "ytt"),
		args...,
	)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cmd.Stdout = io.Writer(&stdout)
	cmd.Stderr = io.Writer(&stderr)
	cmd.Env = y.Env

	if err := cmd.Run(); err != nil {
		t.Fatalf("error running ytt: %v, stderr: %v", err, stderr.String())
		return []byte{}
	}

	return stdout.Bytes()
}

func or(options ...string) string {
	for _, s := range options {
		if s != "" {
			return s
		}
	}
	return ""
}
