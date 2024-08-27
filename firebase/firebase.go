package firebase

import (
	"context"
	"discord-bot/utils"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/bwmarrin/discordgo"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var Log = utils.Log

var client *firestore.Client
var ctx = context.Background()

var cache = make(map[string]*FirebaseData)

func Initialize() *firestore.Client {
	var err error

	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(os.Getenv("SERVICE_ACCOUNT_KEY"))))
	if err != nil {
		Log.Fatal("\nFirebase:", err.Error())
		Log.Debug(Log.Level.Fatal, err.Error())
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		Log.Fatal("\nFirebase:", err.Error())
		Log.Debug(Log.Level.Fatal, err.Error())
	}

	return client
}

func GetGuildData(guildId string) (*FirebaseData, error) {
	// try cache first
	if data, ok := cache[guildId]; ok {
		return data, nil
	}

	data := FirebaseData{}
	data.SetDefaults() // set default values

	// get from firebase
	dsnap, err := client.Collection("Guilds").Doc(guildId).Get(ctx)

	// guild data doesn't exist, create it with defaults
	if status.Code(err) == codes.NotFound {
		_, err := client.Collection("Guilds").Doc(guildId).Set(ctx, map[string]interface{}{
			"voiceMessages":  []interface{}{},
			"customCommands": []interface{}{},
			"savedList":      []interface{}{},
			"prefix":         "!",
		})

		if err != nil {
			return &data, err
		}

		// add to cash
		cache[guildId] = &data

		return &data, nil
	}

	if err != nil {
		return &data, err
	}

	m := dsnap.Data()

	// convert from map to struct and save it in `data`
	data.CreateFromMap(m)

	// add to cash
	cache[guildId] = &data

	return &data, nil
}

func GetBotActivity() (BotActivity, error) {
	data := BotActivity{
		Activity:     "/",
		ActivityType: discordgo.ActivityTypeListening,
	}

	dsnap, err := client.Collection("Shared").Doc("bot").Get(ctx)
	if err != nil {
		return data, err
	}

	m := dsnap.Data()

	if botActivity, ok := m["botActivity"].(map[string]interface{}); ok {
		if activity, ok := botActivity["activity"].(string); ok {
			data.Activity = activity
		}

		if activityType, ok := botActivity["type"].(int64); ok {
			data.ActivityType = discordgo.ActivityType(activityType)
		}
	}

	return data, nil
}

// * Setters

func SetCustomCommand(guildId string, newCommands *[]map[string]interface{}) error {
	_, err := client.Collection("Guilds").Doc(guildId).
		Set(ctx,
			map[string]interface{}{"customCommands": newCommands},
			firestore.MergeAll,
		)

	return err
}

func SetSavedList(guildId string, newSavedList []string) error {
	_, err := client.Collection("Guilds").Doc(guildId).Update(ctx, []firestore.Update{
		{
			Path:  "savedList",
			Value: newSavedList,
		},
	})
	return err
}

func SetVoiceMessages(guildId string, newVoiceMessages *[]map[string]interface{}) error {
	_, err := client.Collection("Guilds").Doc(guildId).
		Set(ctx,
			map[string]interface{}{"voiceMessages": newVoiceMessages},
			firestore.MergeAll,
		)

	return err
}

func SetCommandPrefix(guildId string, newPrefix string) error {
	// get current data
	currentData, err := GetGuildData(guildId)
	if err != nil {
		return err
	}

	// update firebase data
	_, err = client.Collection("Guilds").Doc(guildId).Update(ctx, []firestore.Update{
		{
			Path:  "prefix",
			Value: newPrefix,
		},
	})

	if err != nil {
		return err
	}

	// update cache
	currentData.Prefix = newPrefix

	return nil
}

func SetBotActivity(newActivity BotActivity) error {
	_, err := client.Collection("Shared").Doc("bot").Set(ctx, map[string]interface{}{
		"botActivity": map[string]interface{}{
			"activity": newActivity.Activity,
			"type":     newActivity.ActivityType,
		},
	})

	return err
}
