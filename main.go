package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	tb "gopkg.in/tucnak/telebot.v2"
)

type News struct {
	Entries []Entry `json:"events"`
}

type Entry struct {
	Gid  string `json:"gid"`
	Time int    `json:"rtime32_start_time"`
}

const LATEST_KEY = "latest-dota"

func createBot() *tb.Bot {
	botToken := os.Getenv("BOT_TOKEN")

	bot, err := tb.NewBot(tb.Settings{
		Token:  botToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		panic(err)
	}

	bot.Handle(tb.OnText, handleMe)

	return bot
}

func handleMe(m *tb.Message) {
	log.Println(m.Sender.ID, m.Text)
}

var ctx = context.Background()

func main() {
	chatId, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)

	if err != nil {
		log.Fatalln("CHAT_ID not present")
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPass := os.Getenv("REDIS_PASSWORD")

	bot := createBot()
	log.Println("Bot created")

	go bot.Start()
	log.Println("Bot started")

	log.Println("Starting redis client")
	db := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPass,
		DB:       0,
	})

	log.Println("Starting loop")
	for {
		time.Sleep(1 * time.Minute)

		client := &http.Client{}

		req, err := http.NewRequest("GET", "https://store.steampowered.com/events/ajaxgetpartnereventspageable/?clan_accountid=0&appid=570&offset=0&count=100&l=english&origin=https:%2F%2Fwww.dota2.com", nil)

		if err != nil {
			log.Println("Cannot build new request: ", err)
			continue
		}

		req.Header.Add("Connection", "keep-alive")
		req.Header.Add("Accept", "application/json, text/plain, */*")
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")
		req.Header.Add("Origin", "https://www.dota2.com")
		req.Header.Add("Sec-Fetch-Site", "cross-site")
		req.Header.Add("Sec-Fetch-Mode", "cors")
		req.Header.Add("Sec-Fetch-Dest", "empty")
		req.Header.Add("Referer", "https://www.dota2.com/")
		req.Header.Add("Accept-Language", "en-US,en;q=0.9,pt-BR;q=0.8,pt;q=0.7,sv;q=0.6")

		res, err := client.Do(req)

		if err != nil {
			log.Println("Cannot make request: ", err)
			continue
		}

		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}

		news := News{}
		json.Unmarshal(body, &news)

		value, err := db.Get(ctx, LATEST_KEY).Result()
		if err != nil {
			bot.Send(&tb.Chat{ID: chatId}, "Failed to get redis key on list open")
		}

		latest, err := strconv.Atoi(value)
		if err != nil {
			bot.Send(&tb.Chat{ID: chatId}, "Failed to parse latest key")
		}

		if latest < news.Entries[0].Time {
			err := db.Set(ctx, LATEST_KEY, news.Entries[0].Time, 0).Err()
			if err != nil {
				bot.Send(&tb.Chat{ID: chatId}, "Failed to set redis key on list open")
			}

			log.Println("sending message")
			bot.Send(&tb.Chat{ID: chatId}, "https://www.dota2.com/newsentry/"+news.Entries[0].Gid)
		}

	}

}
