package modules

import (
	"io"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/jdx/go-netrc"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Version string `yaml:"version"`
	Modules []struct {
		Repository   string `yaml:"repository"`
		Path         string `yaml:"path"`
		Branch       string `yaml:"branch"`
		Tag          string `yaml:"tag"`
		OutputFolder string `yaml:"out"`
	} `yaml:"modules"`
}

// NewConfig create a struct for bytes array
func NewConfig(contents []byte) (*Config, error) {
	modulesConfig := &Config{}
	err := yaml.Unmarshal(contents, modulesConfig)
	if err != nil {
		return nil, err
	}
	return modulesConfig, nil
}

// Vendor function that iterate over the modules and vendor proto files from git repositories
func Vendor(config *Config) error {
	for _, module := range config.Modules {
		var reference plumbing.ReferenceName
		if module.Branch != "" {
			// clone repository with branch
			reference = plumbing.NewBranchReferenceName(module.Branch)
		} else if module.Tag != "" {
			// clone repository with tag
			reference = plumbing.NewTagReferenceName(module.Tag)
		}

		usr, err := user.Current()
		n, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
		parsed, err := url.Parse(module.Repository)
		if err != nil {
			log.Printf("failed to parse url: %s", module.Repository)
			return err
		}

		var auth transport.AuthMethod
		if n != nil {
			machine := n.Machine(parsed.Host)
			if machine != nil {
				auth = &http.BasicAuth{Username: usr.Username, Password: machine.Get("password")}
			}
		}

		fs := memfs.New()

		_, err = git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
			URL:           module.Repository,
			Auth:          auth,
			ReferenceName: reference,
			SingleBranch:  true,
			Depth:         1,
			Progress:      os.Stdout,
		})

		if err != nil {
			log.Printf("failed to clone repository: %s", module.Repository)
			return err
		}

		modulePath := filepath.Dir(module.Path)
		baseDir := modulePath
		if module.OutputFolder != "" {
			baseDir = module.OutputFolder
		}

		err = os.MkdirAll(baseDir, os.ModePerm)
		if err != nil {
			log.Printf("failed to create directory: %s", baseDir)
			return err
		}

		err = util.Walk(fs, module.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("failed to walk by path: %s", path)
				return err
			}

			outputPath := strings.ReplaceAll(path, modulePath, baseDir)

			if info.IsDir() {
				err := os.Mkdir(outputPath, os.ModePerm)
				if err != nil {
					if os.IsExist(err) {
						return nil
					}

					log.Printf("failed to create directory: %s", outputPath)
					return err
				}

				return nil
			}

			// skip if not a proto file
			if !strings.HasSuffix(path, ".proto") {
				return nil
			}

			file, err := fs.Open(path)
			if err != nil {
				log.Printf("failed to open file in repository: %s", path)
				return err
			}

			fileContents, err := io.ReadAll(file)
			if err != nil {
				log.Printf("failed to read file contents in repository: %s", path)
				return err
			}

			copiedFile, err := os.Create(outputPath)
			if err != nil {
				log.Printf("failed to create file: %s", outputPath)
				return err
			}

			_, err = copiedFile.Write(fileContents)
			if err != nil {
				log.Printf("failed to write file contents: %s", outputPath)
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}

		log.Printf("successfully vendoring .proto files. repo: %s, path: %s", module.Repository, module.Path)
	}

	return nil
}
