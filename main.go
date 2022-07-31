package main

import (
	"bufio"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

const textListURL = "https://kuroneko6423.com/ExVoice/zundamon/%E3%81%9A%E3%82%93%E3%81%A0%E3%82%82%E3%82%93ExVoice.txt"

func GetTextList() ([]string, error) {
	resp, err := http.Get(textListURL)
	if err != nil {
		return nil, err
	}

	body := resp.Body
	defer body.Close()

	decoder := transform.NewReader(body, japanese.ShiftJIS.NewDecoder())

	scanner := bufio.NewScanner(decoder)
	textList := make([]string, 0)
	for scanner.Scan() {
		textList = append(textList, scanner.Text())
	}

	return textList, nil
}

func GetRandomText() (string, error) {
	textList, err := GetTextList()
	if err != nil {
		return "", err
	}

	index := rand.Intn(len(textList))
	return textList[index], nil
}

func main() {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")

	rand.Seed(time.Now().UnixNano())

	client, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalln(err)
	}
	check, err := regexp.Compile(".*おみくじ.*")
	if err != nil {
		log.Fatalln(err)
	}

	client.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m == nil {
			return
		}

		fmt.Println("> ", m.Content)

		if !check.MatchString(m.Content) {
			fmt.Println("not おみくじ")
			return
		}

		text, err := GetRandomText()
		if err != nil {
			log.Println(err)
			return
		}

		if _, err := client.ChannelMessageSend(m.ChannelID, text); err != nil {
			log.Println(err)
			return
		}
	})

	if err := client.Open(); err != nil {
		log.Fatalln(err)
	}

	defer client.Close()
	stopBot := make(chan os.Signal, 1)
	signal.Notify(stopBot, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-stopBot
}
