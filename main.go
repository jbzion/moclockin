package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

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

var subscribeMap map[string]map[string]*User // map[groupID][userID]*User

func init() {
	subscribeMap = make(map[string]map[string]*User)
	ok = make(map[string]map[string]time.Time)
}

var bot *linebot.Client

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
			if len(event.Source.GroupID) == 0 {
				return
			}
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					arr := strings.Split(message.Text, " ")
					switch arr[0] {
					case "喵嗚喵": // 取得訂閱者
						var msg string
						mu.Lock()
						defer mu.Unlock()
						userList, exists := subscribeMap[event.Source.GroupID]
						if !exists || len(userList) == 0 {
							msg = "目前沒有人需要提醒喵"
						} else {
							msg = "目前訂閱提醒服務的有:\n"
							for _, user := range userList {
								inHour := user.InHour + 8
								if inHour >= 24 {
									inHour -= 24
								}
								outHour := user.OutHour + 8
								if outHour >= 24 {
									outHour -= 24
								}
								msg += fmt.Sprintf("[%s] %02d:%02d - %02d:%02d\n", user.DisplayName, inHour, user.InMin, outHour, user.OutMin)
							}
							msg += "喵"
						}
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err != nil {
							log.Print(err)
						}
					case "喵嗚": // 訂閱
						if len(arr) != 5 {
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("格式錯了喵")).Do(); err != nil {
								log.Print(err)
							}
							return
						}
						userProfileResponse, err := bot.GetGroupMemberProfile(event.Source.GroupID, event.Source.UserID).Do()
						if err != nil {
							log.Print(err)
							return
						}
						mu.Lock()
						defer mu.Unlock()
						_, exists := subscribeMap[event.Source.GroupID]
						if !exists {
							subscribeMap[event.Source.GroupID] = make(map[string]*User)
						}
						inHour, err := strconv.Atoi(arr[1])
						if err != nil {
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("格式錯了喵")).Do(); err != nil {
								log.Print(err)
							}
							return
						}
						inHour -= 8
						if inHour < 0 {
							inHour += 24
						}
						inMin, err := strconv.Atoi(arr[2])
						if err != nil {
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("格式錯了喵")).Do(); err != nil {
								log.Print(err)
							}
							return
						}
						outHour, err := strconv.Atoi(arr[3])
						if err != nil {
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("格式錯了喵")).Do(); err != nil {
								log.Print(err)
							}
							return
						}
						outHour -= 8
						if outHour < 0 {
							outHour += 24
						}
						outMin, err := strconv.Atoi(arr[4])
						if err != nil {
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("格式錯了喵")).Do(); err != nil {
								log.Print(err)
							}
							return
						}
						subscribeMap[event.Source.GroupID][event.Source.UserID] = &User{
							DisplayName: userProfileResponse.DisplayName,
							InHour:      inHour,
							InMin:       inMin,
							OutHour:     outHour,
							OutMin:      outMin,
						}
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
					case "喵": // 打卡
						userProfileResponse, err := bot.GetGroupMemberProfile(event.Source.GroupID, event.Source.UserID).Do()
						if err != nil {
							log.Print(err)
							return
						}
						mu.Lock()
						if user, exists := subscribeMap[event.Source.GroupID][event.Source.UserID]; exists {
							now := time.Now()
							if (now.Hour() == user.InHour && now.Minute()-15 <= user.InMin && now.Minute() >= user.InMin) ||
								(now.Hour() == user.OutHour && now.Minute()-15 <= user.OutMin && now.Minute() >= user.OutMin) {
								muok.Lock()
								var msg string
								if _, exists := ok[event.Source.GroupID]; !exists {
									ok[event.Source.GroupID] = make(map[string]time.Time)
								}
								if _, exists := ok[event.Source.GroupID][event.Source.UserID]; exists {
									msg = userProfileResponse.DisplayName + " 你打過卡了喵"
								} else {
									ok[event.Source.GroupID][event.Source.UserID] = time.Now().Add(time.Minute * 15)
									msg = userProfileResponse.DisplayName + " 已打卡喵"
								}
								muok.Unlock()
								if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err != nil {
									log.Print(err)
								}
							}
						}
						mu.Unlock()
					}
				}
			}
		}
	})
	go run()
	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

var ok map[string]map[string]time.Time
var muok sync.Mutex

func run() {
	for {
		now := time.Now()
		next := now.Add(time.Minute)
		next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), 0, 0, next.Location())
		t := time.NewTimer(next.Sub(now))
		tt := <-t.C
		if tt.Weekday() == 0 || tt.Weekday() == 6 {
			continue
		}
		go func() {
			muok.Lock()
			for groupID := range ok {
				for userID, t := range ok[groupID] {
					if t.Before(tt) {
						delete(ok[groupID], userID)
					}
				}
			}
			muok.Unlock()
		}()
		mu.Lock()
		for groupID, userList := range subscribeMap {
			var inMsg, outMsg string
			for userID, user := range userList {
				if _, exists := ok[groupID]; exists {
					if _, exists := ok[groupID][userID]; exists {
						continue
					}
				}
				fmt.Println(tt)
				fmt.Println(user)
				if tt.Hour() == user.InHour && tt.Minute()-15 <= user.InMin && tt.Minute() >= user.InMin {
					inMsg += user.DisplayName + ", "
				}
				if tt.Hour() == user.OutHour && tt.Minute()-15 <= user.OutMin && tt.Minute() >= user.OutMin {
					outMsg += user.DisplayName + ", "
				}
			}
			if len(inMsg) != 0 {
				inMsg += "快點打卡上班喵！"
				if _, err := bot.PushMessage(groupID, linebot.NewTextMessage(inMsg)).Do(); err != nil {
					log.Print(err)
				}
			}
			if len(outMsg) != 0 {
				outMsg += "快點打卡下班喵！"
				if _, err := bot.PushMessage(groupID, linebot.NewTextMessage(outMsg)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
		mu.Unlock()
	}
}
