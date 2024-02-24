package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmorganca/ollama/api"
	"github.com/jmorganca/ollama/server"
)

func HasModel(model string) (bool, error) {
	_, err := server.GetModel(model)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ListModel() ([]*server.Model, error) {
	manifestsPath, err := server.GetManifestPath()
	if err != nil {
		return nil, err
	}
	models := make([]*server.Model, 0)

	walkFunc := func(path string, info os.FileInfo, _ error) error {
		if !info.IsDir() {
			path, tag := filepath.Split(path)
			model := strings.Trim(strings.TrimPrefix(path, manifestsPath), string(os.PathSeparator))
			modelPath := strings.Join([]string{model, tag}, ":")
			canonicalModelPath := strings.ReplaceAll(modelPath, string(os.PathSeparator), "/")

			m, err := server.GetModel(canonicalModelPath)
			if err == nil {
				models = append(models, m)
			}
		}

		return nil
	}

	if err := filepath.Walk(manifestsPath, walkFunc); err != nil {
		return nil, err
	}
	return models, nil
}

func PullModel(ctx context.Context, model string, fn func(r api.ProgressResponse)) error {
	regOpts := &server.RegistryOptions{
		Insecure: false,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := server.PullModel(ctx, model, regOpts, fn); err != nil {
		return err
	}
	return nil
}
