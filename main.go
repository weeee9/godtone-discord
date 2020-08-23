package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/weeee9/godtone-discord/voice"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

var servers = make(map[string]*voiceChannel)

type voiceChannel struct {
	vc        *discordgo.VoiceConnection
	isPlaying bool
}

func main() {
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	discord.AddHandler(godtone)

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

func godtone(sess *discordgo.Session, msg *discordgo.MessageCreate) {
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
	switch msg.Content {
	case "!操你媽過來一下":
		if !isInVoiceChannel {
			sess.ChannelMessageSend(msg.ChannelID, "you must be in a voice channel")
			return
		}
		if _, ok := servers[voiceChannelID]; !ok {
			vc, err := sess.ChannelVoiceJoin(guild.ID, voiceChannelID, false, true)
			if err != nil {
				log.Printf("failed to joined voice channel: %s\n", err.Error())
				sess.ChannelMessageSend(msg.ChannelID, "something went worng...")
				return
			}
			server := &voiceChannel{
				vc: vc,
			}
			servers[voiceChannelID] = server
		} else {
			return
		}
	case "!carry":
		play(servers[voiceChannelID], "./m4a/carry.m4a")
	case "!7414":
		play(servers[voiceChannelID], "./m4a/goDie.m4a")
	case "!那我也要睡拉":
		play(servers[voiceChannelID], "./m4a/imGoingToSleep.m4a")
	case "!你要先講":
		play(servers[voiceChannelID], "./m4a/youHaveToSaidItFirst.m4a")
	case "!滾蛋":
		if _, ok := servers[voiceChannelID]; !ok {
			return
		}
		servers[voiceChannelID].vc.Disconnect()
		delete(servers, voiceChannelID)
	}
}

func play(server *voiceChannel, file string) {
	if server == nil || server.isPlaying {
		return
	}
	server.isPlaying = true
	done := make(chan bool)
	stop := make(chan bool)
	// voice pkg if modified from dbvoice
	// please see: https://github.com/bwmarrin/dgvoice
	go voice.PlayAudioFile(server.vc, file, done, stop)
	go func() {
		select {
		case <-done:
			server.isPlaying = false
		}
	}()
}
