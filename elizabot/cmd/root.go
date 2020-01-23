/*
Copyright © 2020 Michael Wenk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"github.com/caser/eliza"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/disgord/x/mux"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var router *mux.Mux
var sugar *zap.SugaredLogger
var cfgFile string
var discordToken string
// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "elizabot",
	Short: "Eliza simulates interactions with psychotherapist",
	Long:  `Eliza simulates interactions with psychotherapist.`,
	Run: func(cmd *cobra.Command, args []string) {
		defer sugar.Sync()
		// Init
		discordInit()

		// Copied from disgord example
		if discordToken == "" {
			sugar.Infof("You must provide a Discord authentication token.")
			return
		}
		dg, err := discordgo.New(discordToken)
		if err != nil { 
			sugar.Errorf("error while creating discordgo: %v", err)
		}
		// Open a websocket connection to Discord
		err = dg.Open()
		if err != nil {
			sugar.Errorf("error opening connection to Discord, %s\n", err)
			os.Exit(1)
		}

		// Create the mux 
		initializeMux(dg)

		// Wait for a CTRL-C
		sugar.Infof(`Now running. Press CTRL-C to exit.`)
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc

		// Clean up
		dg.Close()

	},
}
func messageHandler(ds *discordgo.Session,  mc *discordgo.MessageCreate ) { 
	defer sugar.Sync()
	sugar.Infof("mc=%v", spew.Sdump(nil))
	// Ignore all messages created by the Bot account itself
	if mc.Author.ID == ds.State.User.ID {
		return
	}
	sugar.Infof("content=%v", mc.Content)
	parsed := eliza.ParseInput(mc.Content) 
	preproced := eliza.PreProcess(parsed)
	result := eliza.PostProcess(parsed)

	sugar.Infof("result=%v", result)
	//ds.ChannelMessageSend(mc.ChannelID,fmt.Sprintf("Received: %v", mc.Content))
	// Fetch the channel for this Message

}
func initializeMux(session *discordgo.Session) *mux.Mux { 
	var mux = mux.New()
	session.AddHandler(messageHandler)
	return mux
}
// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Version is a constant that stores the Disgord version information.
const Version = "v0.0.1-alpha"

// Grab token from config
func discordInit() {
	defer sugar.Sync()
	discordToken = viper.GetString("discord.token")
}

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("error found in creating logger: %v", err)
		os.Exit(1)
	}
	sugar = logger.Sugar()
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.elizabot.yaml)")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	defer sugar.Sync()
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".elizabot" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".elizabot")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		sugar.Infof("Viper is using config file: %v", viper.ConfigFileUsed())
	}
}
