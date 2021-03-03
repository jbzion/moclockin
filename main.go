package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/line/line-bot-sdk-go/linebot"
)

var mu sync.Mutex

var subscribeMap map[string]map[string]string // map[groupID][userID]DisplayName

func init() {
	subscribeMap = make(map[string]map[string]string)
}

func main() {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
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
			if len(event.Source.GroupID) == 0 {
				return
			}
			if userProfileResponse, err := bot.GetGroupMemberProfile(event.Source.GroupID, event.Source.UserID).Do(); err != nil {
				log.Print(err)
			} else {
				fmt.Printf("%+v\n", userProfileResponse)
			}
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					switch message.Text {
					case "喵嗚喵": // 取得訂閱者
						var msg string
						mu.Lock()
						userList, exists := subscribeMap[event.Source.GroupID]
						if !exists || len(userList) == 0 {
							msg = "目前沒有人需要提醒喵"
						} else {
							msg = "目前訂閱提醒服務的有: "
							for _, v := range userList {
								msg += v + ", "
							}
							msg = msg[:len(msg)-2] + "，喵"
						}
						mu.Unlock()
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err != nil {
							log.Print(err)
						}
					case "喵嗚": // 訂閱
						userProfileResponse, err := bot.GetGroupMemberProfile(event.Source.GroupID, event.Source.UserID).Do()
						if err != nil {
							log.Print(err)
							return
						}
						mu.Lock()
						defer mu.Unlock()
						_, exists := subscribeMap[event.Source.GroupID]
						if !exists {
							subscribeMap[event.Source.GroupID] = make(map[string]string)
						}
						subscribeMap[event.Source.GroupID][event.Source.UserID] = userProfileResponse.DisplayName
						msg := userProfileResponse.DisplayName + " 已訂閱喵"
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err != nil {
							log.Print(err)
						}
					case "喵喵": // 取消訂閱
						userProfileResponse, err := bot.GetGroupMemberProfile(event.Source.GroupID, event.Source.UserID).Do()
						if err != nil {
							log.Print(err)
							return
						}
						mu.Lock()
						defer mu.Unlock()
						_, exists := subscribeMap[event.Source.GroupID]
						if !exists {
							return
						}
						_, exists = subscribeMap[event.Source.GroupID][event.Source.UserID]
						if !exists {
							msg := userProfileResponse.DisplayName + " 你沒有訂閱喵"
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err != nil {
								log.Print(err)
								return
							}
						}
						delete(subscribeMap[event.Source.GroupID], event.Source.UserID)
						msg := userProfileResponse.DisplayName + " 已取消訂閱喵"
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err != nil {
							log.Print(err)
						}
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
