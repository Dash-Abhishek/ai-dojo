package structuredoutput

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

var client openai.Client
var clientOnce sync.Once

type ModelResp struct {
	Content string `json:"content"`
}

type ChatContext struct {
	Id     int
	Memory Memory
}

type Memory struct {
	Messages []openai.ChatCompletionMessageParamUnion
}

func getClient() openai.Client {

	clientOnce.Do(func() {
		client = openai.NewClient()
	})
	return client
}

func NewChatContext(id int) ChatContext {
	return ChatContext{
		Id: id,
		Memory: Memory{
			Messages: []openai.ChatCompletionMessageParamUnion{},
		},
	}
}

func (c *ChatContext) AddMessage(message openai.ChatCompletionMessageParamUnion) {
	c.Memory.Messages = append(c.Memory.Messages, message)
}

func (c *ChatContext) ViewConversation() {
	for _, msg := range c.Memory.Messages {
		if msg.OfAssistant != nil {
			fmt.Println("Assistant:", msg.OfAssistant.Content.OfString)
		}
		if msg.OfUser != nil {
			fmt.Println("User:", msg.OfUser.Content.OfString)
		}
		if msg.OfSystem != nil {
			fmt.Println("System:", msg.OfSystem.Content.OfString)
		}
		if msg.OfFunction != nil {
			fmt.Println("Function:", msg.OfFunction.Name)
		}
	}
}

func (c *ChatContext) GenerateResponseFromModel(respSchema shared.ResponseFormatJSONSchemaJSONSchemaParam) (string, error) {
	client := getClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModelGPT4o,
		Messages: c.Memory.Messages,
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{JSONSchema: respSchema},
		},
		Temperature: openai.Float(0.0),
	})

	if err != nil {
		return "", err
	}

	c.AddMessage(openai.ChatCompletionMessageParamUnion{
		OfAssistant: &openai.ChatCompletionAssistantMessageParam{
			Content: openai.ChatCompletionAssistantMessageParamContentUnion{
				OfString: openai.String(resp.Choices[0].Message.Content),
			},
		}})
	return resp.Choices[0].Message.RawJSON(), nil

}
