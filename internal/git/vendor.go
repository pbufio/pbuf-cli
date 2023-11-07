package git

import (
	"io"
	"log"
	"net/url"
	"os"
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
	"github.com/pbufio/pbuf-cli/internal/model"
)

func VendorGitModule(module *model.Module, netrcAuth *netrc.Netrc) error {
	log.Printf("start vendoring .proto files. repo: %s, path: %s", module.Repository, module.Path)

	var reference plumbing.ReferenceName
	if module.Branch != "" {
		// clone repository with branch
		reference = plumbing.NewBranchReferenceName(module.Branch)
	} else if module.Tag != "" {
		// clone repository with tag
		reference = plumbing.NewTagReferenceName(module.Tag)
	}

	parsed, err := url.Parse(module.Repository)
	if err != nil {
		log.Printf("failed to parse url: %s", module.Repository)
		return err
	}

	var auth transport.AuthMethod
	if netrcAuth != nil {
		machine := netrcAuth.Machine(parsed.Host)
		if machine != nil {
			auth = &http.BasicAuth{Username: machine.Get("login"), Password: machine.Get("password")}
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

	var modulePath string
	if strings.HasSuffix(module.Path, ".proto") {
		modulePath = filepath.Dir(module.Path)
	} else {
		modulePath = module.Path
	}

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

	return nil
}
