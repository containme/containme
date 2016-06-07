package builder

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/engine-api/types/strslice"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

type ProfileSpec struct {
	Shell            string   `json:"shell"`
	Image            string   `json:"image"`
	Workspace        string   `json:"workspace"`
	CacheDirectories []string `json:"cache_directories"`
	Dependencies     []Step   `json:"dependencies"`
	Test             []Step   `json:"test"`
}

func (profile *ProfileSpec) MakeCmd(cmd string) strslice.StrSlice {
	shell := "/bin/sh"
	if profile.Shell != "" {
		shell = profile.Shell
	}

	return strslice.StrSlice([]string{shell, "-c", cmd})
}

func ParseProfileSpecFile(file string) (*ProfileSpec, error) {
	path, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ParseProfileSpec(f)
}

func ParseProfileSpec(src io.Reader) (*ProfileSpec, error) {
	spec := &ProfileSpec{}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, src); err != nil {
		return nil, err
	}

	err := yaml.Unmarshal(buf.Bytes(), &spec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse ProfileSpec")
	}

	return spec, nil
}
