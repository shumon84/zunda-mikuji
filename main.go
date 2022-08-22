package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const textListURL = "https://kuroneko6423.com/ExVoice/zundamon/%E3%81%9A%E3%82%93%E3%81%A0%E3%82%82%E3%82%93ExVoice.txt"
const emojiName = "saiyo"
const popularTextFile = "popular_text.txt"

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

func GetRandomPopularText() (msg string, err error) {
	if len(popularTextMap.cache) == 0 {
		return "", errors.New("popularTextMap is empty")
	}
	index := rand.Intn(len(popularTextMap.cache))
	i := 0
	popularTextMap.mtx.Lock()
	for text, _ := range popularTextMap.cache {
		if index == i {
			msg = text
			break
		}
		i++
	}
	popularTextMap.mtx.Unlock()
	return
}

var sentMessageIDs = NewAtomicCache[string, string]()
var popularTextMap = NewAtomicCache[string, struct{}]()

func OmikujiHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	text, err := GetRandomText()
	if err != nil {
		log.Println(err)
		return
	}

	msg, err := s.ChannelMessageSend(m.ChannelID, text)
	if err != nil {
		log.Println(err)
		return
	}
	sentMessageIDs.Store(msg.ID, text)
}

func SuperOmikujiHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	text, err := GetRandomPopularText()
	if err != nil {
		log.Println(err)
		return
	}

	if _, err := s.ChannelMessageSend(m.ChannelID, text); err != nil {
		log.Println(err)
		return
	}
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m == nil {
		return
	}

	fmt.Println("> ", m.Content)
	switch m.Content {
	case "おみくじ":
		OmikujiHandler(s, m)
	case "スーパーおみくじ":
		SuperOmikujiHandler(s, m)
	}
	return
}

func onMessageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.Emoji.Name != emojiName {
		return
	}
	if !sentMessageIDs.Exists(m.MessageID) {
		return
	}
	popularTextMap.Store(sentMessageIDs.Load(m.MessageID), struct{}{})
}

func main() {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")

	rand.Seed(time.Now().UnixNano())

	if err := loadPopularTexts(); err != nil {
		log.Fatal(err)
	}

	client, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalln(err)
	}

	client.AddHandler(onMessageCreate)
	client.AddHandler(onMessageReactionAdd)

	if err := client.Open(); err != nil {
		log.Fatalln(err)
	}
	defer client.Close()
	stopBot := make(chan os.Signal, 1)
	signal.Notify(stopBot, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-stopBot
	if err := savePopularTexts(); err != nil {
		log.Fatal(err)
	}
}

func loadPopularTexts() error {
	f, err := os.OpenFile(popularTextFile, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		text := sc.Text()
		popularTextMap.Store(text, struct{}{})
	}
	return nil
}

func savePopularTexts() error {
	f, err := os.OpenFile(popularTextFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	for text, _ := range popularTextMap.cache {
		_, err := f.Write([]byte(text + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
