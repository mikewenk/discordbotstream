package botcontroller

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"

	//"github.com/davecgh/go-spew/spew"
	"github.com/mikewenk/discordbotstream/elizabot/eliza"
	"go.uber.org/zap"
)

type discordUser struct {
	userId string
}

var seenUsers []discordUser

var dg discordgo.Session
var sugar *zap.SugaredLogger

func BotInit(s *zap.SugaredLogger, discordToken string) error {
	sugar = s
	// Initialize active users
	seenUsers = make([]discordUser, 1)
	dg, err := discordgo.New(discordToken)
	if err != nil {
		sugar.Errorf("error while creating discordgo: %v", err)
	}
	// Open a websocket connection to Discord
	err = dg.Open()
	if err != nil {
		sugar.Errorf("error opening connection to Discord, %s\n", err)
		return errors.New("error opening connection")
	}
	// Create the mux
	initializeMux(dg)

	return nil
}

func BotClose() {
	dg.Close()
}
func seenUser(targetuser string) bool {
	for _, cur := range seenUsers {
		if cur.userId == targetuser {
			return true
		}
	}
	return false
}

func messageHandler(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	defer sugar.Sync()
	//sugar.Infof("mc=%v", spew.Sdump(mc))
	// Ignore all messages created by the Bot account itself
	if mc.Author.ID == ds.State.User.ID {
		return
	}
	var outMessage string
	if !seenUser(mc.Author.ID) {
		var newUser discordUser
		newUser.userId = mc.Author.ID
		seenUsers = append(seenUsers, newUser)
		outMessage = eliza.Greetings()
	} else {
		outMessage = eliza.ReplyTo(mc.Content)
	}

	sugar.Infof("outMessage=%v", outMessage)
	ds.ChannelMessageSend(mc.ChannelID, outMessage)
	// Fetch the channel for this Message
}

func initializeMux(session *discordgo.Session) *mux.Mux {
	var mux = mux.New()
	session.AddHandler(messageHandler)
	return mux
}
