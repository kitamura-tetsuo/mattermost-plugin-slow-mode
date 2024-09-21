package main

import (
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// MyPlugin struct definition, extending MattermostPlugin and adding a map to track the last post time per user
type MyPlugin struct {
	plugin.MattermostPlugin
	lastPostTime map[string]time.Time // Stores the last post time for each user
}

// OnActivate initializes the plugin and sets up the lastPostTime map
func (p *MyPlugin) OnActivate() error {
	p.lastPostTime = make(map[string]time.Time)
	return nil
}

// MessageWillBePosted is a hook that gets triggered before a message is posted
func (p *MyPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	userID := post.UserId
	currentTime := time.Now()

	// Check if the user has posted before and if the time since their last post is less than 30 seconds
	if lastTime, ok := p.lastPostTime[userID]; ok {
		timeSinceLastPost := currentTime.Sub(lastTime)

		if timeSinceLastPost < 30*time.Second {
			return nil, "You cannot post more than once within 30 seconds."
		}
	}

	// Update the last post time for the user
	p.lastPostTime[userID] = currentTime
	return post, ""
}

func main() {
	plugin.ClientMain(&MyPlugin{})
}
