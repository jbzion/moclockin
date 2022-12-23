package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

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

const chatGptURL = "https://api.openai.com/v1/chat/gpt"

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
	// 建立 HTTP 請求
	req, err := http.NewRequest("POST", chatGptURL, strings.NewReader(`{
        "model": "chatgpt","prompt": input,
		"max_tokens": 64,
		"temperature": 0.5,
		"top_p": 1,
		"frequency_penalty": 0,
		"presence_penalty": 0
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer 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")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 讀取回應內容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
