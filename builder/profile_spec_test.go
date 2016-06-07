package builder

import (
	"bytes"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestParseProfileSpec(t *testing.T) {
	assert := assert.New(t)
	var data = `
image: foo/bar
workspace: /use/src/app
cache_directories:
  - /home/admin/.m2
dependencies:
  - mvn install
test:
  - mvn test
`

	buf := bytes.NewBuffer([]byte(data))

	spec, err := ParseProfileSpec(buf)
	assert.NoError(err)

	assert.Equal("foo/bar", spec.Image)
	assert.Equal("/use/src/app", spec.Workspace)
	assert.Equal(1, len(spec.CacheDirectories))
	assert.Equal("/home/admin/.m2", spec.CacheDirectories[0])
	assert.Equal(1, len(spec.Dependencies))
	assert.Equal("mvn install", spec.Dependencies[0].Cmd)
	assert.Equal(1, len(spec.Test))
	assert.Equal("mvn test", spec.Test[0].Cmd)
}
