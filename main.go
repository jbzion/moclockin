package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	chatgpt "github.com/golang-infrastructure/go-ChatGPT"
	"github.com/line/line-bot-sdk-go/linebot"
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

	jwt := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik1UaEVOVUpHTkVNMVFURTRNMEZCTWpkQ05UZzVNRFUxUlRVd1FVSkRNRU13UmtGRVFrRXpSZyJ9.eyJodHRwczovL2FwaS5vcGVuYWkuY29tL3Byb2ZpbGUiOnsiZW1haWwiOiJlNTAzMTBAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImdlb2lwX2NvdW50cnkiOiJUVyJ9LCJodHRwczovL2FwaS5vcGVuYWkuY29tL2F1dGgiOnsidXNlcl9pZCI6InVzZXItVUhwVWhBUlcyYWk4ems1VWhkallnTndEIn0sImlzcyI6Imh0dHBzOi8vYXV0aDAub3BlbmFpLmNvbS8iLCJzdWIiOiJnb29nbGUtb2F1dGgyfDEwNjgxNDkwODQ2NDgxODY3MzA2MCIsImF1ZCI6WyJodHRwczovL2FwaS5vcGVuYWkuY29tL3YxIiwiaHR0cHM6Ly9vcGVuYWkuYXV0aDAuY29tL3VzZXJpbmZvIl0sImlhdCI6MTY3MTQxMzYzOSwiZXhwIjoxNjcyMDE4NDM5LCJhenAiOiJUZEpJY2JlMTZXb1RIdE45NW55eXdoNUU0eU9vNkl0RyIsInNjb3BlIjoib3BlbmlkIHByb2ZpbGUgZW1haWwgbW9kZWwucmVhZCBtb2RlbC5yZXF1ZXN0IG9yZ2FuaXphdGlvbi5yZWFkIG9mZmxpbmVfYWNjZXNzIn0.X40vjCMeAOsj89npWCy47MZkFLu6mt5ZzPvD9m97q1OG_9SzLYo6kBxSZSpaCKVrDF9AzDRNIASoeDd5FKMvS6VhCtilLo8WT-D1sHsuDdoRKbM4lJEteA-AqZSKSKRj4upwCLSngWqJf0nRrPWMYRmlUr7CJMQQ0575r1UPCu0mcSg-g6aKHvy1UTxR3jaKqfrluWqFU-vD3VwkqtcMhwuhLhT9eKtzWVulzXtqZKMO1M5VMRker4Au-n7KrXwWvgWj3UXECYts0k9Ozzm1Ogg78EXulHpO1xfnTQriqfiLc_68UUMxxrH4Ig4995Al1UsdOWlmvG8O2msHVNK89Q"
	chat := chatgpt.NewChatGPT(jwt)
	talk, err := chat.Talk(input)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return talk.Message.Content.Parts[0], nil
}
