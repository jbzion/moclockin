package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/line/line-bot-sdk-go/linebot"
	gogpt "github.com/sashabaranov/go-gpt3"
)

var mu sync.Mutex

type User struct {
	DisplayName string
	InHour      int
	InMin       int
	OutHour     int
	OutMin      int
}

var bot *linebot.Client

const chatGptURL = "https://api.openai.com/v1/completions"

func main() {
	client, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	bot = client
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {

		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					// 呼叫 ChatGPT API 並取得回應
					response, err := callChatGptAPI(message.Text)
					if err != nil {
						fmt.Println(err)
						continue
					}
					// 回傳回應給使用者
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(response)).Do(); err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	})

	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func callChatGptAPI(input string) (string, error) {
	c := gogpt.NewClient("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")
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:     "text-davinci-003",
		MaxTokens: 5,
		Prompt:    input,
	}
	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Text, nil
}
