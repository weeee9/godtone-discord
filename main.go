package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	voice "github.com/weeee9/discord-music-bot/play"
)

var (
	buffer = make([][]byte, 0)
)

func main() {
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	err = loadSound()
	if err != nil {
		fmt.Println("Error loading sound: ", err)
		fmt.Println("Please copy $GOPATH/src/github.com/bwmarrin/examples/airhorn/airhorn.dca to this directory.")
		return
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

// loadSound attempts to load an encoded sound file from disk.
func loadSound() error {

	file, err := os.Open("./airhorn.dca")
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return err
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
}

func playSound(sess *discordgo.Session, guildID, channelID string) error {
	vc, err := sess.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	time.Sleep(250 * time.Millisecond)

	vc.Speaking(true)

	for _, buff := range buffer {
		vc.OpusSend <- buff
	}
	vc.Speaking(false)

	time.Sleep(250 * time.Millisecond)

	return vc.Disconnect()
}
