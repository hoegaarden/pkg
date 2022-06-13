package testing

import (
	"io"
	"io/ioutil"

	"github.com/vmware-tanzu/carvel-ytt/pkg/cmd/template"
	"github.com/vmware-tanzu/carvel-ytt/pkg/cmd/ui"
	"github.com/vmware-tanzu/carvel-ytt/pkg/files"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
)

type JsonProvider interface {
	Provide() ([]byte, error)
}

type Ytt struct {
	SrcDir     string
	DataValues []string
}

var _ JsonProvider = Ytt{}

func (y Ytt) Provide() ([]byte, error) {
	stdout := ioutil.Discard
	stderr := ioutil.Discard
	debug := false

	ui := ui.NewCustomWriterTTY(debug, stdout, stderr)
	opts := template.NewOptions()

	opts.DataValuesFlags.KVsFromStrings = y.DataValues

	pkgPath := or(y.SrcDir, "src")
	filesToProcess, err := files.NewSortedFilesFromPaths([]string{pkgPath}, files.SymlinkAllowOpts{})
	if err != nil {
		return []byte{}, err
	}

	out := opts.RunWithFiles(template.Input{Files: filesToProcess}, ui)
	if out.Err != nil {
		return []byte{}, out.Err
	}

	jsonPrinter := func(w io.Writer) yamlmeta.DocumentPrinter { return yamlmeta.NewJSONPrinter(w) }
	bytes, err := out.DocSet.AsBytesWithPrinter(jsonPrinter)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func or(options ...string) string {
	for _, s := range options {
		if s != "" {
			return s
		}
	}
	return ""
}
