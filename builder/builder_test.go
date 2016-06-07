package builder

import (
	"os"
	"testing"
)

func TestStandardBuild(t *testing.T) {
	t.SkipNow()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	profile := &ProfileSpec{
		Image:     "golang:1.6-alpine",
		Workspace: "/go/src",
	}
	spec := &BuildSpec{
		Environment: EnvironmentStage{
			Workspace: "/go/src/github.com/jive/containme",
			BaseStage: BaseStage{
				After: StepList{
					Steps: []Step{
						Step{
							Cmd: "apk add --update git bash curl gcc mercurial py-pip",
						},
					},
				},
			},
		},
		Dependencies: DependenciesStage{
			BaseStage: BaseStage{
				Before: StepList{
					Steps: []Step{
						Step{
							Cmd: "curl -L https://github.com/Masterminds/glide/releases/download/0.10.2/glide-0.10.2-linux-amd64.tar.gz | tar xzfO - linux-amd64/glide > /bin/glide; chmod +x /bin/glide; glide -v",
						},
					},
				},
				Override: StepList{
					Steps: []Step{
						Step{
							Cmd: "glide install",
						},
					},
				},
			},
		},
		Test: TestStage{
			BaseStage: BaseStage{
				Override: StepList{
					Steps: []Step{
						Step{
							Cmd: "go list",
						},
					},
				},
			},
		},
	}

	builder, err := NewBuilder(pwd+"/..", spec, profile)
	if err != nil {
		t.Fatal(err)
	}

	_, err = builder.ExecuteEnivronmentStage(false, "")
	if err != nil {
		t.Fatal(err)
	}

	err = builder.ExecuteDependenciesStage(false)
	if err != nil {
		t.Fatal(err)
	}

	err = builder.ExecuteTestStage(false)
	if err != nil {
		t.Fatal(err)
	}

	err = builder.destroyVolumes()
	if err != nil {
		t.Fatal(err)
	}
}
