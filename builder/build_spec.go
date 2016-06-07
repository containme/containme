package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

type BuildSpec struct {
	Environment  EnvironmentStage  `json:"environment"`
	Dependencies DependenciesStage `json:"dependencies"`
	Test         TestStage         `json:"test"`
	Package      PackageStage      `json:"package"`
}

type EnvironmentStage struct {
	Profile   string `json:"profile"`
	Workspace string `json:"workspace"`
	BaseStage
}

type BaseStage struct {
	Before   StepList `json:"before"`
	Override StepList `json:"override"`
	After    StepList `json:"after"`
}

type DependenciesStage struct {
	BaseStage
}

type TestStage struct {
	BaseStage
}

type PackageStage struct {
	PackageStageDocker `json:"docker"`
}

type PackageStageDocker struct {
	Tag        string `json:"tag"`
	Dockerfile string `json:"dockerfile"`
}

type Step struct {
	Cmd     string            `json:"cmd"`
	Timeout int               `json:"timeout"`
	Cache   bool              `json:"cache"`
	Env     map[string]string `json:"environment"`
	Pwd     string            `json:"pwd"`
}

type StepList struct {
	Steps []Step
}

func (sl *StepList) UnmarshalJSON(data []byte) error {
	steps := []interface{}{}
	err := json.Unmarshal(data, &steps)
	if err != nil {
		return err
	}

	for idx := range steps {
		step := Step{}
		stepData, _ := json.Marshal(steps[idx])
		json.Unmarshal(stepData, &step)
		// switch stepT := steps[idx].(type) {
		// case string:
		// 	step.Cmd = stepT
		// case map[string]interface{}:
		// 	for cmd, val := range stepT {
		// 		stepData, _ := json.Marshal(val)
		// 		json.Unmarshal(stepData, &step)
		// 		step.Cmd = cmd
		// 		break
		// 	}
		// default:
		// 	fmt.Printf("unexpected type %T\n", stepT)
		//
		// }

		sl.Steps = append(sl.Steps, step)
	}
	return nil
}

func (step *Step) UnmarshalJSON(data []byte) error {
	var v interface{}
	json.Unmarshal(data, &v)
	switch stepT := v.(type) {
	case string:
		step.Cmd = stepT
	case map[string]interface{}:
		for cmd, val := range stepT {
			alias := &struct {
				Timeout int               `json:"timeout"`
				Cache   bool              `json:"cache"`
				Env     map[string]string `json:"environment"`
				Pwd     string            `json:"pwd"`
			}{}
			stepData, _ := json.Marshal(val)
			json.Unmarshal(stepData, &alias)
			step.Cmd = cmd
			step.Timeout = alias.Timeout
			step.Cache = alias.Cache
			step.Env = alias.Env
			step.Pwd = alias.Pwd
			break
		}
	default:
		fmt.Printf("unexpected type %T\n", stepT)

	}
	return nil
}

func ParseBuildSpecFile(file string) (*BuildSpec, error) {
	path, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ParseBuildSpec(f)
}

func ParseBuildSpec(src io.Reader) (*BuildSpec, error) {
	buildSpec := &BuildSpec{}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, src); err != nil {
		return nil, err
	}

	err := yaml.Unmarshal(buf.Bytes(), &buildSpec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse BuildSpec")
	}

	return buildSpec, nil
}
