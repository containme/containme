package builder

import (
	"bytes"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestParseBuildSpec(t *testing.T) {
	assert := assert.New(t)
	var data = `
environment:
  profile: containme/profiles/example
  after:
  - env after

dependencies:
  override:
    - dep override
    - bundle install: # note the colon here
        timeout: 240
        environment:
          - foo=bar
          - foo2=bar2
        pwd:
          test_dir

test:
  before:
    - test before
  override:
    - mvn test:
        timeout: 600
`

	buf := bytes.NewBuffer([]byte(data))

	spec, err := ParseBuildSpec(buf)
	assert.NoError(err)

	assert.Equal("containme/profiles/example", spec.Environment.Profile)
	assert.Equal(1, len(spec.Environment.After.Steps))
	assert.Equal("env after", spec.Environment.After.Steps[0].Cmd)

	assert.Equal(0, len(spec.Dependencies.Before.Steps))
	assert.Equal(2, len(spec.Dependencies.Override.Steps))
	assert.Equal("dep override", spec.Dependencies.Override.Steps[0].Cmd)
	assert.Equal("bundle install", spec.Dependencies.Override.Steps[1].Cmd)
	assert.Equal(2, len(spec.Dependencies.Override.Steps[1].Env))
	assert.Equal(240, spec.Dependencies.Override.Steps[1].Timeout)
	assert.Equal("foo=bar", spec.Dependencies.Override.Steps[1].Env[0])
	assert.Equal("foo2=bar2", spec.Dependencies.Override.Steps[1].Env[1])
	assert.Equal("test_dir", spec.Dependencies.Override.Steps[1].Pwd)
	assert.Equal(0, len(spec.Dependencies.After.Steps))

	assert.Equal(1, len(spec.Test.Before.Steps))
	assert.Equal("test before", spec.Test.Before.Steps[0].Cmd)
	assert.Equal(1, len(spec.Test.Override.Steps))
	assert.Equal("mvn test", spec.Test.Override.Steps[0].Cmd)
	assert.Equal(600, spec.Test.Override.Steps[0].Timeout)
	assert.Equal(0, len(spec.Test.After.Steps))

}
