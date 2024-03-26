package core

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"log/slog"

	"github.com/jmorganca/ollama/api"
	"github.com/jmorganca/ollama/gpu"
	"github.com/jmorganca/ollama/llm"
	"github.com/jmorganca/ollama/server"
)

type Core struct {
	mu    sync.Mutex
	model *llmWrapper
}

func New(model string) (*Core, error) {
	if err := llm.Init(); err != nil {
		return nil, fmt.Errorf("unable to initialize llm library %w", err)
	}
	if runtime.GOOS == "linux" { // TODO - windows too
		// check compatibility to log warnings
		if _, err := gpu.CheckVRAM(); err != nil {
			slog.Info(err.Error())
		}
	}

	c := &Core{
		//workDir: workDir,
	}

	cm, err := load(model)
	if err != nil {
		return nil, err
	}
	c.model = cm

	return c, nil
}

func (c *Core) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.model != nil {
		c.model.Close()
	}
	//os.RemoveAll(c.workDir)
}

func (c *Core) Reload(ctx context.Context, model string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.model != nil && c.model.Name == model {
		return nil
	}

	_, err := server.GetModel(model)
	if err != nil {
		return err
	}

	if c.model != nil {
		c.model.Close()
	}
	c.model = nil
	lw, err := load(model)
	if err != nil {
		return err
	}
	c.model = lw
	return nil
}

type chatResponse struct {
	Result llm.PredictResult
	Err    error
}

func (c *Core) Chat(ctx context.Context, messages []api.Message, options map[string]interface{}) (<-chan *chatResponse, error) {
	return c.chat(ctx, messages, options, false)
}
func (c *Core) ChatJSON(ctx context.Context, messages []api.Message, options map[string]interface{}) (<-chan *chatResponse, error) {
	return c.chat(ctx, messages, options, true)
}
func (c *Core) chat(ctx context.Context, messages []api.Message, options map[string]interface{}, json bool) (<-chan *chatResponse, error) {
	c.mu.Lock()

	if nil == c.model {
		c.mu.Unlock()
		return nil, fmt.Errorf("model is nil")
	}

	if c.model.IsEmbedding() {
		c.mu.Unlock()
		return nil, fmt.Errorf("model is not an embedding model")
	}

	opts, err := modelOptions(c.model.Model, options)
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}

	encode := func(s string) ([]int, error) {
		return c.model.Encode(ctx, s)
	}
	prompt, err := server.ChatPrompt(c.model.Template, messages, opts.NumCtx, encode)
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}

	ch := make(chan *chatResponse)
	//checkpointStart := time.Now()
	format := ""
	if json {
		format = "json"
	}

	go func() {
		defer c.mu.Unlock()
		defer close(ch)

		fn := func(r llm.PredictResult) {

			select {
			case <-ctx.Done():
				return
			case ch <- &chatResponse{r, nil}:
			}
		}

		// Start prediction
		predictReq := llm.PredictOpts{
			Prompt: prompt,
			Format: format,
			//Images:  images,
			Options: opts,
		}
		if err := c.model.Predict(ctx, predictReq, fn); err != nil {
			select {
			case <-ctx.Done():
			case ch <- &chatResponse{Err: err}:
			}
			return
		}
	}()
	return ch, nil
}

func (c *Core) Embedding(ctx context.Context, prompt string) ([]float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if nil == c.model {
		return nil, fmt.Errorf("model is nil")
	}
	if !c.model.IsEmbedding() {
		return nil, fmt.Errorf("model is not an embedding model")
	}

	embedding, err := c.model.Embedding(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return embedding, nil
}
