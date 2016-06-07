package builder

import (
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

type ProfileResolver interface {
	ResolveProfile() (*ProfileSpec, error)
}

type LocalProfileResolver struct {
	Path string
}

func (res *LocalProfileResolver) ResolveProfile() (*ProfileSpec, error) {
	return ParseProfileSpecFile(res.Path)
}

func NewProfileResolver(profile string) (ProfileResolver, error) {
	var profilePath string
	switch {
	case strings.HasPrefix(profile, "./"):
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		profilePath = path.Join(pwd, profile)
		fallthrough
	case strings.HasPrefix(profile, "/"):
		if profilePath == "" {
			profilePath = profile
		}

		if _, err := os.Stat(profilePath); err == nil {
			return &LocalProfileResolver{
				Path: profilePath,
			}, nil
		}

		// check with .yaml extension
		if _, err := os.Stat(profilePath + ".yaml"); err == nil {
			return &LocalProfileResolver{
				Path: profilePath + ".yaml",
			}, nil
		}

		// check with .yml extension
		if _, err := os.Stat(profilePath + ".yml"); err == nil {
			return &LocalProfileResolver{
				Path: profilePath + ".yml",
			}, nil
		}

		return nil, errors.Errorf("%s: profile could not be found", profilePath)

	default:
		return nil, errors.Errorf("could not determine profile resolution strategy")
	}
}
