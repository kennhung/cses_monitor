package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Problem struct {
	Name   string
	Status int
	URL    string
}

type UserInfo struct {
	Name     string
	Problems map[string]*Problem
}

type UserInfos = map[string]*UserInfo

func checkDiff(prev map[string]Problem, now map[string]Problem, callback func(prev Problem, now Problem)) {
	for k, v := range prev {
		if now[k].Status != v.Status {
			fmt.Println(now[k].Status, v.Status)
			callback(v, now[k])
		}
	}
}

func sendNotification(userName string, probName string, probURL string, status int) {
	webhook := os.Getenv("DISCORD_WEBHOOK")

	var emoji string
	var statusStr string
	if status == 2 {
		emoji = ":white_check_mark:"
		statusStr = "pass"
	} else if status == 1 {
		emoji = ":x:"
		statusStr = "failed on"
	} else {
		emoji = ":question:"
		statusStr = "???"
	}

	if len(webhook) > 0 {
		sendStr := fmt.Sprintf("{"+
			"\"content\": \""+
			"%s `%s` %s ||**%s**||\\n%s"+
			"\""+
			"}",
			emoji,
			userName,
			statusStr,
			probName,
			probURL,
		)

		_, err := http.Post(webhook, "application/json", strings.NewReader(sendStr))
		if err != nil {
			log.Printf("Error on sending notification\n%s", err)
		}
	}
}

func getURL(id string) string {
	return fmt.Sprintf("https://cses.fi/problemset/user/%s?userId=%s", id, id)
}

func run(userInfos *UserInfos, list []string, delayTime int) {
	c := colly.NewCollector()

	quit := make(chan struct{})

	c.OnHTML("h2", func(e *colly.HTMLElement) {
		userId := e.Request.Ctx.Get("userId")

		(*userInfos)[userId].Name = strings.Split(e.Text, " ")[2]
	})

	c.OnHTML(".task-score", func(e *colly.HTMLElement) {
		userId := e.Request.Ctx.Get("userId")
		probId := e.Attr("href")

		if (*userInfos)[userId].Problems[probId] == nil {
			(*userInfos)[userId].Problems[probId] = new(Problem)
		}

		problem := (*userInfos)[userId].Problems[probId]
		if strings.Index(e.Attr("class"), "full") != -1 {
			problem.Status = 2
		} else if strings.Index(e.Attr("class"), "zero") != -1 {
			problem.Status = 1
		} else {
			problem.Status = 0
		}
		problem.Name = e.Attr("title")
		problem.URL = e.Attr("href")
	})

	c.OnRequest(func(r *colly.Request) {
		userId := r.URL.Query().Get("userId")

		if (*userInfos)[userId] == nil {
			userInfo := new(UserInfo)
			userInfo.Problems = map[string]*Problem{}
			(*userInfos)[userId] = userInfo
		}

		r.Ctx.Put("userId", userId)
	})

	i := 0

	c.OnResponse(func(r *colly.Response) {
		userId := r.Ctx.Get("userId")

		prev := map[string]Problem{}
		for k, v := range (*userInfos)[userId].Problems {
			prev[k] = *v
		}

		go func() {
			c.Wait()

			now := map[string]Problem{}
			for k, v := range (*userInfos)[userId].Problems {
				now[k] = *v
			}

			checkDiff(prev, now, func(p Problem, n Problem) {
				sendNotification(
					(*userInfos)[userId].Name,
					n.Name,
					fmt.Sprintf("https://cses.fi%s", n.URL),
					n.Status)

				fmt.Println((*userInfos)[userId].Name, n.Name, n.Status)
			})

			time.Sleep(time.Duration(delayTime) * time.Microsecond)
			i++
			if i < len(list) {
				c.Visit(getURL(list[i]))
			} else {
				close(quit)
			}
		}()
	})

	if i < len(list) {
		c.Visit(getURL(list[i]))

		for {
			select {
			case <-quit:
				return
			}
		}
	}
}

func main() {

	var userInfos = map[string]*UserInfo{}

	list := strings.Split(os.Getenv("USER_IDS"), ",")

	if len(list) == 0 {
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})

	delayTime, err := strconv.ParseInt(os.Getenv("FETCH_DELAY"), 10, 32)
	if err != nil {
		log.Println(err)
		delayTime = 2000
	}

	for {
		run(&userInfos, list, int(delayTime))

		select {
		case <-ticker.C:

		case <-quit:
			ticker.Stop()
			return
		}
	}
}
