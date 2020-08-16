package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kkdai/youtube"
	voice "github.com/weeee9/discord-music-bot/play"
)

const (
	temDir = "./tmp"
)

var (
	servers map[string]queue

	stop = make(chan bool)
)

type queue struct {
	songs []string
}

func init() {
	createDirIfNotExist(temDir)
}

func main() {
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	discord.AddHandler(airhorn)

	if err := discord.Open(); err != nil {
		panic(err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}

func airhorn(sess *discordgo.Session, msg *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if msg.Author.ID == sess.State.User.ID {
		return
	}

	ch, err := sess.State.Channel(msg.ChannelID)
	if err != nil {
		log.Println(err)
		sess.ChannelMessageSend(msg.ChannelID, "sorry, something went wrong")
		return
	}
	guild, err := sess.State.Guild(ch.GuildID)
	if err != nil {
		log.Println(err)
		sess.ChannelMessageSend(msg.ChannelID, "sorry, something went wrong")
		return
	}

	var isInVoiceChannel bool
	var voiceChannelID string
	for _, state := range guild.VoiceStates {
		if state.UserID == msg.Author.ID {
			voiceChannelID = state.ChannelID
			isInVoiceChannel = true
		}
	}

	args := strings.Split(msg.Content, " ")
	switch args[0] {
	case "!play":
		if len(args) == 1 {
			sess.ChannelMessageSend(msg.ChannelID, "you need to provide a link!")
			return
		}
		if !isInVoiceChannel {
			sess.ChannelMessageSend(msg.ChannelID, "you must be in a voice channel to play!")
			return
		}

		if _, ok := servers[guild.ID]; !ok {
			servers[guild.ID] = queue{}
		}

		server := servers[guild.ID]
		_ = server

		vc, err := sess.ChannelVoiceJoin(guild.ID, voiceChannelID, false, true)
		if err != nil {
			log.Println(err)
			sess.ChannelMessageSend(msg.ChannelID, "sorry, something went wrong")
			return
		}
		voice.PlayAudioFile(vc, "", stop)

	case "!skip":
	case "!stop":
		stop <- true
	}
	if strings.HasPrefix(msg.Content, "!airhorn") {
		ch, err := sess.State.Channel(msg.ChannelID)
		if err != nil {
			sess.ChannelMessageSend(msg.ChannelID, "ERROR CHANNEL STATE "+err.Error())
			return
		}

		guild, err := sess.State.Guild(ch.GuildID)
		if err != nil {
			sess.ChannelMessageSend(msg.ChannelID, "ERROR GUILD STATE "+err.Error())
			return
		}

		for _, vs := range guild.VoiceStates {
			if vs.UserID == msg.Author.ID {
				// err = playSound(sess, guild.ID, vs.ChannelID)

				vc, err := sess.ChannelVoiceJoin(guild.ID, vs.ChannelID, false, true)
				if err != nil {
					sess.ChannelMessageSend(msg.ChannelID, "ERROR PLAYING SOUND "+err.Error())
					return
				}

				// dgvoice.PlayAudioFile(vc, "./m4a/carry.mp4", make(chan bool))
				stop := make(chan bool)
				voice.PlayAudioFile(vc, "./m4a/carry.mp4", stop)
			}
		}
	}
}

func downloadYtb(url string) ([]byte, error) {
	client := youtube.Client{}

	video, err := client.GetVideo(url)
	if err != nil {
		return nil, err
	}

	resp, err := client.GetStream(video, &video.Streams[0])
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)

	return b, err
}

func tempSave(guildID string, content []byte) error {
	path := filepath.Join(temDir, guildID)
	err := os.Mkdir(path, 0644)
	if err != nil {
		return err
	}

	return nil
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}
