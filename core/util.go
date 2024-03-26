package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmorganca/ollama/api"
	"github.com/jmorganca/ollama/llm"
	"github.com/jmorganca/ollama/server"
)

func modelOptions(model *server.Model, requestOpts map[string]interface{}) (api.Options, error) {
	opts := api.DefaultOptions()
	if err := opts.FromMap(model.Options); err != nil {
		return api.Options{}, err
	}

	if err := opts.FromMap(requestOpts); err != nil {
		return api.Options{}, err
	}

	return opts, nil
}

type llmWrapper struct {
	llm.LLM
	*server.Model
}

func load(modelName string) (*llmWrapper, error) {
	model, err := server.GetModel(modelName)
	if err != nil {
		return nil, err
	}
	var opts, _ = modelOptions(model, nil)

	llmRunner, err := llm.New(model.ModelPath, model.AdapterPaths, model.ProjectorPaths, opts)
	if err != nil {
		// some older models are not compatible with newer versions of llama.cpp
		// show a generalized compatibility error until there is a better way to
		// check for model compatibility
		if errors.Is(llm.ErrUnsupportedFormat, err) || strings.Contains(err.Error(), "failed to load model") {
			err = fmt.Errorf("%v: this model may be incompatible with your version of Ollama. If you previously pulled this model, try updating it by running `ollama pull %s`", err, model.ShortName)
		}

	}
	return &llmWrapper{Model: model, LLM: llmRunner}, err
}
