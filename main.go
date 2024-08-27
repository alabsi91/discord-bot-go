package main

import (
	"discord-bot/discord"
	"discord-bot/firebase"
	"discord-bot/httpServer"
	"discord-bot/torrentClient"
	"discord-bot/utils"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

var Log = &utils.Log

func main() {
	configPath := flag.String("config", ".config.json", "The path to the config file")
	envPath := flag.String("env", ".env", "the Path to the .env file")
	generateConfigTemplateCmd := flag.Bool("config-template", false, "Generate config json template")
	generateEnvFileTemplate := flag.Bool("env-template", false, "Generate .env file template")
	flag.Parse()

	// generate config template
	if *generateConfigTemplateCmd {
		err := utils.GenerateConfigJsonTemplate()
		if err != nil {
			Log.Fatal("\nError generating config template", err.Error())
		}
		Log.Success("\nTemplate generated successfully at \"./.config.json\"")
		return
	}

	// generate .env file template
	if *generateEnvFileTemplate {
		err := utils.GenerateEnvFileTemplate()
		if err != nil {
			Log.Fatal("\nError generating .env file template", err.Error())
		}
		Log.Success("\nTemplate generated successfully at \"./.env\"")
		return
	}

	// load .config file
	Log.Info("\nLoading \".config\" file...")
	config, err := utils.PrepareAppConfig(*configPath)
	if err != nil {
		Log.Fatal("\nError loading \".config\" file", err.Error())
	}

	// set log settings
	Log.SetLogToFile(config.Log.Enabled)
	Log.SetLogFilePath(config.Log.Path)

	// load .env file
	Log.Info("\nLoading \".env\" file...")
	Log.Debug(Log.Level.Info, `Loading ".env" file...`)
	err = godotenv.Load(*envPath)
	if err != nil {
		Log.Fatal("\nError loading \".env\" file", err.Error())
		Log.Debug(Log.Level.Fatal, `Error loading ".env" file:`, err.Error())
	}

	// firebase
	Log.Info("\nInitializing Firebase...")
	Log.Debug(Log.Level.Info, `Initializing Firebase...`)
	firebaseClient := firebase.Initialize()

	// http server
	Log.Info("\nInitializing HTTP Server...")
	Log.Debug(Log.Level.Info, `Initializing HTTP Server...`)
	httpServer.Initialize()

	// torrent client
	Log.Info("\nInitializing Torrent Client...")
	Log.Debug(Log.Level.Info, `Initializing Torrent Client...`)
	torrentClient.Initialize()

	// discord
	Log.Info("\nStarting Discord session...")
	Log.Debug(Log.Level.Info, `Starting Discord session...`)
	dg := discord.StartDiscordBotSession()

	// Wait here until CTRL-C or other term signal is received.
	Log.Tip("\nPress CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	Log.Info("\nClosing Discord session...")
	Log.Debug(Log.Level.Info, `Closing Discord session...`)
	dg.Close()

	// close firebase
	Log.Info("\nClosing Firebase...")
	Log.Debug(Log.Level.Info, `Closing Firebase...`)
	firebaseClient.Close()

	// close http server
	Log.Info("\nClosing HTTP Server...")
	Log.Debug(Log.Level.Info, `Closing HTTP Server...`)
	httpServer.Close()

	// close torrent client
	Log.Info("\nClosing Torrent Client...")
	Log.Debug(Log.Level.Info, `Closing Torrent Client...`)
	torrentClient.CloseSession()
}
