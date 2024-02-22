package core

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/jmorganca/ollama/api"
	"github.com/stretchr/testify/require"
)

func TestCore_Predict(t *testing.T) {
	c, err := New("qwen:0.5b")
	defer c.Close()
	require.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	msg := []api.Message{{"user", "hello", nil}}
	for i := 0; i < 10; i++ {

		msg = append(msg, api.Message{"user", "why the sky is blue" + strconv.Itoa(i), nil})
		cc, err := c.Chat(ctx, msg, nil)
		require.Nil(t, err)
		require.NotNil(t, cc)
		result := ""
		for m := range cc {
			if m.Err != nil {
				fmt.Println(m.Err)
			}
			result += m.Result.Content
		}
		fmt.Println(result)

		msg = append(msg, api.Message{"assistant", result, nil})
	}
}

func TestCore_List(t *testing.T) {
	a, err := ListModel()
	require.Nil(t, err)
	require.NotNil(t, a)
	fmt.Println(a)
}

func TestCore_Pull(t *testing.T) {
	model := "qwen:0.5b"
	model = "nomic-embed-text"
	err := PullModel(context.Background(), model, func(r api.ProgressResponse) {
		fmt.Println(r.Completed, r.Total)
	})
	require.Nil(t, err)
}

func TestCore_Embed(t *testing.T) {
	model := "qwen:0.5b"
	model = "nomic-embed-text"
	c, err := New(model)
	defer c.Close()
	require.Nil(t, err)

	b, err := c.Embedding(context.Background(), "hello")
	require.Nil(t, err)
	require.NotNil(t, b)
	fmt.Println(b)
}
