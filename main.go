//default driver provided by mattermost

package main

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
	"github.com/mattermost/mattermost-server/v6/model"

	// "github.com/mattermost/mattermost/server/public/model"
	"fmt"
)

var (
	// App manifest
	appManifest = apps.Manifest{
		AppID:       apps.AppID("hello-world"),
		Version:     apps.AppVersion("0.1.0"),
		HomepageURL: "https://github.com/MY_USERNAME/MY_APP_REPO",
		DisplayName: "Hello world!",
		Description: "Example Golang App for Mattermost",
		Icon:        "icon.png",
		RequestedPermissions: apps.Permissions{
			apps.PermissionActAsBot,
		},
		RequestedLocations: apps.Locations{
			apps.LocationChannelHeader,
			apps.LocationCommand,
		},
		Deploy: apps.Deploy{
			HTTP: &apps.HTTP{
				RootURL: "http://my-app-hostname:4000",
			},
		},
	}

	// A form with a single text input field and a Submit button
	appForm = apps.Form{
		Title: "I'm a form!",
		Icon:  "icon.png",
		Fields: []apps.Field{
			{
				Type:                 apps.FieldTypeText,
				Name:                 "message",
				Label:                "message",
				AutocompletePosition: 1,
			},
		},
		Submit: apps.NewCall("/submit"),
	}

	// Bind a button that shows a Form to the Channel Header
	channelHeaderBinding = apps.Binding{
		Location: apps.LocationChannelHeader,
		Bindings: []apps.Binding{
			{
				Location: "send-button",
				Icon:     "icon.png",
				Label:    "send hello message",
				Form:     &appForm,
			},
		},
	}

	// Bind a slash command using a Form for input
	commandBinding = apps.Binding{
		Location: apps.LocationCommand,
		Bindings: []apps.Binding{
			{
				Location:    "send-command",
				Label:       "send hello message",
				Description: appManifest.Description,
				Hint:        "[send]",
				Bindings: []apps.Binding{
					{
						Location: "send",
						Label:    "send",
						Form:     &appForm,
					},
				},
			},
		},
	}

	// Collect all App bindings into a slice
	bindings = []apps.Binding{
		channelHeaderBinding,
		commandBinding,
	}
)

// Handlers

func submitHandler(w http.ResponseWriter, r *http.Request) {
	// Unmarshal the request body into an apps.CallRequest struct
	callRequest := new(apps.CallRequest)
	err := json.NewDecoder(r.Body).Decode(callRequest)
	if err != nil {
		// handle the error
		return
	}

	// Construct the response message using the input form's `message` value
	message := "Hello, world!"
	submittedMessage, ok := callRequest.Values["message"].(string)
	if ok {
		message += " ...and " + submittedMessage + "!"
	}

	// Create an instance of the API Client
	botClient := appclient.AsBot(callRequest.Context)

	// Post a DM
	channel, _, err := botClient.CreateDirectChannel(
		callRequest.Context.BotUserID,
		callRequest.Context.ActingUser.Id,
	)
	if err != nil {
		// handle the error
		return
	}
	post := &model.Post{
		ChannelId: channel.Id,
		Message:   message,
	}
	_, err = botClient.CreatePost(post)
	if err != nil {
		// handle the error
		return
	}

	// Construct the call response and send it
	callResponse := apps.NewTextResponse("Created a post in your DM channel.")
	err = httputils.WriteJSON(w, callResponse)
	if err != nil {
		// handle the error
		return
	}
}

func main() {
	fmt.Println("Mattermost GoProxy started...")
	// Create an instance of the endpoint handler
	handler := httputils.NewHandler()

	// Manifest
	handler.HandleFunc("/manifest.json", httputils.DoHandleJSON(appManifest))

	// Bindings and forms
	handler.HandleFunc("/bindings", httputils.DoHandleJSON(apps.NewDataResponse(bindings)))

	// Handlers
	handler.HandleFunc("/submit", submitHandler)

	// Static assets
	handler.
		PathPrefix("/static/").
		Handler(http.FileServer(http.Dir("static")))

	// Start the HTTP server using the endpoint handler
	server := http.Server{
		Addr:    "http://localhost:8065",
		Handler: handler,
	}
	_ = server.ListenAndServe()
	fmt.Println("Finishing GoProxy")
}
