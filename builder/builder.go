package builder

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/filters"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var DefaultCmdTimeout = 600 * time.Second

type Builder struct {
	docker  *client.Client
	id      string
	workdir string
	//path:volume_name
	volumes map[string]string

	spec    *BuildSpec
	profile *ProfileSpec

	image            string
	environmentImage string
	stepCount        int

	ui UI
}

func NewBuilder(workdir string, build *BuildSpec, profile *ProfileSpec) (*Builder, error) {
	var err error

	builder := &Builder{
		id:      randStringBytesMask(8),
		workdir: workdir,
		spec:    build,
		profile: profile,
		ui:      &ConsolUI{},
	}

	if os.Getenv("DOCKER_HOST") != "" {
		builder.docker, err = client.NewEnvClient()
	} else {
		builder.docker, err = client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, nil)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to create docker client")
	}

	return builder, nil
}

func (b *Builder) ExecInEnvironment(cmd string) error {
	_, err := b.runCmdInImage(b.environmentImage, "environment-exec", cmd, []string{}, DefaultCmdTimeout)
	return err
}

func (b *Builder) ExecuteEnivronmentStage(useCache bool, cacheImage string) (string, error) {
	err := b.createVolumes()
	if err != nil {
		return b.image, errors.Wrap(err, "failed to create volumes for cache directories")
	}

	if useCache && cacheImage != "" {
		_, _, err := b.docker.ImageInspectWithRaw(context.TODO(), cacheImage, false)
		if err != nil {
			b.ui.Printf("recieved err for cached environment, proceeding without cache: %s", err)
			goto EXECUTE_ENVIRONMENT_STAGE_NOCACHE
		}
		b.image = cacheImage
		b.environmentImage = b.image
		return b.image, nil
	}

EXECUTE_ENVIRONMENT_STAGE_NOCACHE:
	b.image = b.profile.Image

	for _, step := range b.spec.Environment.After.Steps {
		err = b.runCmd("environment-after", step.Cmd, []string{}, DefaultCmdTimeout)
		if err != nil {
			return b.image, errors.Wrapf(err, "failed to run cmd(%s)", step.Cmd)
		}
	}
	b.environmentImage = b.image

	return b.image, nil
}

func (b *Builder) ExecuteDependenciesStage(useCache bool, cacheImage string) (string, error) {
	if useCache && cacheImage != "" {
		_, _, err := b.docker.ImageInspectWithRaw(context.TODO(), cacheImage, false)
		if err != nil {
			b.ui.Printf("recieved err for cached dependencies, proceeding without cache: %s", err)
			goto EXECUTE_DEPENDENCIES_STAGE_NOCACHE
		}
		b.image = cacheImage
		return b.image, nil
	}

EXECUTE_DEPENDENCIES_STAGE_NOCACHE:
	var err error
	for _, step := range b.spec.Dependencies.Before.Steps {
		err = b.runCmd("dependancies-before", step.Cmd, []string{}, DefaultCmdTimeout)
		if err != nil {
			return b.image, errors.Wrapf(err, "failed to run cmd(%s)", step.Cmd)
		}
	}

	if len(b.spec.Dependencies.Override.Steps) > 0 {
		for _, step := range b.spec.Dependencies.Override.Steps {
			err = b.runCmd("dependancies-override", step.Cmd, []string{}, DefaultCmdTimeout)
			if err != nil {
				return b.image, errors.Wrapf(err, "failed to run cmd(%s)", step.Cmd)
			}
		}
	} else {
		for _, step := range b.profile.Dependencies {
			err = b.runCmd("dependancies-default", step.Cmd, []string{}, DefaultCmdTimeout)
			if err != nil {
				return b.image, errors.Wrapf(err, "failed to run cmd(%s)", step.Cmd)
			}
		}
	}

	for _, step := range b.spec.Dependencies.After.Steps {
		err = b.runCmd("dependancies-after", step.Cmd, []string{}, DefaultCmdTimeout)
		if err != nil {
			return b.image, errors.Wrapf(err, "failed to run cmd(%s)", step.Cmd)
		}
	}

	return b.image, nil
}

func (b *Builder) ExecuteTestStage() error {
	var err error
	for _, step := range b.spec.Test.Before.Steps {
		err = b.runCmd("test-before", step.Cmd, []string{}, DefaultCmdTimeout)
		if err != nil {
			return errors.Wrapf(err, "failed to run cmd(%s)", step.Cmd)
		}
	}

	if len(b.spec.Test.Override.Steps) > 0 {
		for _, step := range b.spec.Test.Override.Steps {
			err = b.runCmd("test-override", step.Cmd, []string{}, DefaultCmdTimeout)
			if err != nil {
				return errors.Wrapf(err, "failed to run cmd(%s)", step.Cmd)
			}
		}
	} else {
		for _, step := range b.profile.Test {
			err = b.runCmd("test-default", step.Cmd, []string{}, DefaultCmdTimeout)
			if err != nil {
				return errors.Wrapf(err, "failed to run cmd(%s)", step.Cmd)
			}
		}
	}

	for _, step := range b.spec.Test.After.Steps {
		err = b.runCmd("test-after", step.Cmd, []string{}, DefaultCmdTimeout)
		if err != nil {
			return errors.Wrapf(err, "failed to run cmd(%s)", step.Cmd)
		}
	}

	return nil
}

func (b *Builder) createVolumes() error {
	b.volumes = map[string]string{}
	for idx := range b.profile.CacheDirectories {
		ops := types.VolumeCreateRequest{
			Driver: "local",
			Labels: map[string]string{
				"containme_id":  b.id,
				"containme_dir": b.profile.CacheDirectories[idx],
			},
		}
		vol, err := b.docker.VolumeCreate(context.TODO(), ops)
		if err != nil {
			return errors.Wrap(err, "failed to create volume")
		}
		b.volumes[b.profile.CacheDirectories[idx]] = vol.Name
	}
	return nil
}

func (b *Builder) destroyVolumes() error {
	for idx := range b.profile.CacheDirectories {
		err := b.docker.VolumeRemove(context.TODO(), b.volumes[b.profile.CacheDirectories[idx]])
		if err != nil {
			return errors.Wrap(err, "failed to create volume")
		}
		delete(b.volumes, b.profile.CacheDirectories[idx])
	}
	return nil
}

func (b *Builder) runCmd(stage, cmd string, env []string, timeout time.Duration) error {
	nextImage, err := b.runCmdInImage(b.image, stage, cmd, env, timeout)
	if err != nil {
		return err
	}

	b.image = nextImage
	b.stepCount++
	return nil
}

func (b *Builder) runCmdInImage(image, stage, cmd string, env []string, timeout time.Duration) (string, error) {
	config, hostConfig := b.containerConfig()
	config.Cmd = b.profile.MakeCmd(cmd)
	//config.Env = append(config.Env, env...)

	b.ui.Printf("Stage(%s): Running command (%s) for build-step-%d in image(%s):\n\n", stage, cmd, b.stepCount, b.image)

	resp, err := b.docker.ContainerCreate(context.TODO(), config, hostConfig, nil, fmt.Sprintf("cme_builder_%s-%d", b.id, b.stepCount))
	if err != nil {
		return "", errors.Wrapf(err, "failed to create build-step-%d container", b.stepCount)
	}
	defer b.docker.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{
		Force: true,
	})

	err = b.docker.ContainerStart(context.TODO(), resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "failed to start build-step-%d container", b.stepCount)
	}

	wg := sync.WaitGroup{}

	go func() {
		wg.Add(1)
		defer wg.Done()
		var hyjack types.HijackedResponse
		hyjack, err = b.docker.ContainerAttach(context.TODO(), resp.ID, types.ContainerAttachOptions{
			Stream: true,
			Stderr: true,
			Stdout: true,
		})
		if err != nil {
			b.ui.Printf("failed to attached to build-step-%d output", b.stepCount)
			return
		}
		defer hyjack.Close()
		io.Copy(b.ui, hyjack.Reader)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = b.docker.ContainerWait(ctx, resp.ID)
	if err != nil {
		return "", errors.Wrapf(err, "build-step-%d exceeded timeout", b.stepCount)
	}

	wg.Wait()

	commit, err := b.docker.ContainerCommit(context.TODO(), resp.ID, types.ContainerCommitOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "failed to commit build-step-%d", b.stepCount)
	}

	return commit.ID, nil
}

func (b *Builder) containerConfig() (*container.Config, *container.HostConfig) {
	config := &container.Config{
		Image:      b.image,
		WorkingDir: b.profile.Workspace,
		Labels: map[string]string{
			"containme_id": b.id,
		},
		Tty:       true,
		StdinOnce: true,
	}

	if b.spec.Environment.Workspace != "" {
		if filepath.IsAbs(b.spec.Environment.Workspace) {
			config.WorkingDir = b.spec.Environment.Workspace
		} else {
			config.WorkingDir = path.Join(config.WorkingDir, b.spec.Environment.Workspace)
		}
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:%s", b.workdir, config.WorkingDir)},
	}
	for path, vol := range b.volumes {
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s", vol, path))
	}

	return config, hostConfig
}

func (b *Builder) cmdHash(stage, cmd string) string {
	return fmt.Sprintf("%x", (sha256.Sum256([]byte(fmt.Sprintf("%s.%s.%s", b.spec.Environment.Profile, stage, cmd)))))
}

func (b *Builder) getCachedImage(hash string) (string, error) {
	args := filters.NewArgs()
	args.Add("containme_hash", hash)
	ops := types.ImageListOptions{

		Filters: args,
	}
	resp, err := b.docker.ImageList(context.TODO(), ops)
	if err != nil {
		return "", errors.Wrap(err, "failed to list images")
	}

	switch len(resp) {
	case 0:
		return "", nil
	case 1:
		return resp[0].ID, nil
	default:
		return "", errors.Errorf("multiple images for hash(%s)", hash)
	}
}
