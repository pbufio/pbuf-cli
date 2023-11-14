package registry

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "github.com/pbufio/pbuf-cli/gen/pbuf-registry/v1"
	"github.com/pbufio/pbuf-cli/internal/model"
)

const timeout = 60 * time.Second

func VendorRegistryModule(module *model.Module, client v1.RegistryClient) error {
	log.Printf("start vendoring .proto files. module name: %s, path: %s", module.Name, module.Path)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	response, err := client.PullModule(ctx, &v1.PullModuleRequest{
		Name: module.Name,
		Tag:  module.Tag,
	})

	if err != nil {
		log.Printf("failed to pull module: %v", err)
		return err
	}

	for _, protoFile := range response.Protofiles {
		originalFilename := protoFile.Filename
		outputPath := module.OutputFolder

		if module.Path != "" {
			modulePath := filepath.Dir(module.Path)

			if strings.HasSuffix(module.Path, ".proto") {
				// skip if the file is not in the module path
				if originalFilename != module.Path {
					continue
				}
			} else {
				// skip if the file is not in the module path
				if !strings.HasPrefix(originalFilename, modulePath) {
					continue
				}
			}

			if outputPath != "" {
				originalFilename = strings.Replace(originalFilename, filepath.Dir(module.Path), outputPath, 1)
			}
		} else {
			if outputPath != "" {
				originalFilename = filepath.Join(outputPath, originalFilename)
			}
		}

		err := os.MkdirAll(filepath.Dir(originalFilename), os.ModePerm)
		if err != nil {
			log.Printf("failed to create directory: %s", outputPath)
			return err
		}

		copiedFile, err := os.Create(originalFilename)
		if err != nil {
			log.Printf("failed to create file: %s", outputPath)
			return err
		}

		_, err = copiedFile.Write([]byte(protoFile.Content))
		if err != nil {
			log.Printf("failed to write file contents: %s", outputPath)
			return err
		}
	}

	log.Printf("successfully vendoring .proto files. module name: %s, path: %s", module.Name, module.Path)

	return nil
}
