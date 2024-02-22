package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jmorganca/ollama/api"
	"github.com/jmorganca/ollama/core"
	"github.com/spf13/cobra"
)

func ChatCmd() *cobra.Command {
	c := &cobra.Command{
		Use: "chat",
	}
	c.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("model is required")
		}
		model := args[0]
		ok, err := core.HasModel(model)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Pulling model? (y/n)")
			r := bufio.NewReader(os.Stdin)
			text, err := r.ReadString('\n')
			if err != nil {
				return err
			}
			text = strings.TrimSpace(text)
			if text == "y" {
				err := core.PullModel(cmd.Context(), model, func(r api.ProgressResponse) {
					fmt.Println(r.Completed, r.Total)
				})
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("model not found")
			}
		}

		c, err := core.New(model)
		defer c.Close()
		if err != nil {
			return err
		}

		var msg []api.Message

		r := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("You: ")
			text, err := r.ReadString('\n')
			if err != nil {
				return err
			}
			msg = append(msg, api.Message{"user", text, nil})
			cc, err := c.Chat(cmd.Context(), msg, nil)
			if err != nil {
				return err
			}
			fmt.Println()
			fmt.Println("Assistant:")
			fullText := ""
			for m := range cc {
				if m.Err != nil {
					return m.Err
				}
				fmt.Print(m.Result.Content)
				fullText += m.Result.Content
			}
			fmt.Println()
			msg = append(msg, api.Message{"assistant", fullText, nil})
		}
	}
	return c
}
