package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"gopkg.in/yaml.v2"
)

// ChannelConfig represents the configuration for each channel, loaded from YAML.
type ChannelConfig struct {
	PostLimit string `yaml:"post_limit"`
}

// MyPlugin is the main structure of the plugin.
type MyPlugin struct {
	plugin.MattermostPlugin
	lastPostTime   map[string]map[string]time.Time // Stores the last post time per user per channel.
	channelConfigs map[string]*ChannelConfig       // Caches channel configurations.
}

// OnActivate is called when the plugin is activated.
func (p *MyPlugin) OnActivate() error {
	p.lastPostTime = make(map[string]map[string]time.Time)
	p.channelConfigs = make(map[string]*ChannelConfig)
	return nil
}

// MessageWillBePosted is a hook that runs before a message is posted.
func (p *MyPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	userID := post.UserId
	channelID := post.ChannelId
	currentTime := time.Now()

	// Get the channel configuration.
	config, appErr := p.getChannelConfig(channelID)
	if appErr != nil {
		p.API.LogError("Failed to get channel config", "channel_id", channelID, "error", appErr.Error())
		return nil, "An internal error has occurred."
	}

	// Parse the post limit duration.
	limitDuration, err := time.ParseDuration(config.PostLimit)
	if err != nil {
		p.API.LogError("Invalid post_limit format", "post_limit", config.PostLimit)
		limitDuration = 30 * time.Second // Default value.
	}

	// Initialize the map for user last post times in this channel.
	if _, exists := p.lastPostTime[channelID]; !exists {
		p.lastPostTime[channelID] = make(map[string]time.Time)
	}

	// Check if the user is allowed to post.
	if lastTime, ok := p.lastPostTime[channelID][userID]; ok {
		timeSinceLastPost := currentTime.Sub(lastTime)
		if timeSinceLastPost < limitDuration {
			remainingTime := limitDuration - timeSinceLastPost
			return nil, fmt.Sprintf("Please wait %v before posting again in this channel.", remainingTime.Round(time.Second))
		}
	}

	// Update the user's last post time.
	p.lastPostTime[channelID][userID] = currentTime
	return post, ""
}

// getChannelConfig retrieves the channel configuration, using the cache if possible.
func (p *MyPlugin) getChannelConfig(channelID string) (*ChannelConfig, *model.AppError) {
	// Check if the configuration is cached.
	if config, exists := p.channelConfigs[channelID]; exists {
		return config, nil
	}

	// If not cached, retrieve from the channel header.
	channel, appErr := p.API.GetChannel(channelID)
	if appErr != nil {
		return nil, appErr
	}

	yamlContent, err := extractYAMLFromHeader(channel.Header)
	if err != nil {
		// If YAML is not found, use default configuration.
		defaultConfig := &ChannelConfig{PostLimit: "30s"}
		p.channelConfigs[channelID] = defaultConfig
		return defaultConfig, nil
	}

	config, err := parseChannelConfig(yamlContent)
	if err != nil {
		// On parsing error, use default configuration.
		defaultConfig := &ChannelConfig{PostLimit: "30s"}
		p.channelConfigs[channelID] = defaultConfig
		return defaultConfig, nil
	}

	// Cache the configuration.
	p.channelConfigs[channelID] = config
	return config, nil
}

// extractYAMLFromHeader extracts the YAML content from the channel header.
func extractYAMLFromHeader(header string) (string, error) {
	parts := strings.Split(header, "---")
	if len(parts) < 3 {
		return "", fmt.Errorf("no YAML configuration found")
	}
	yamlContent := parts[1]
	return yamlContent, nil
}

// parseChannelConfig parses the YAML content into a ChannelConfig.
func parseChannelConfig(yamlContent string) (*ChannelConfig, error) {
	var config ChannelConfig
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// ChannelHasBeenUpdated is a hook that runs when a channel is updated; it refreshes the cache.
func (p *MyPlugin) ChannelHasBeenUpdated(c *plugin.Context, newChannel *model.Channel, oldChannel *model.Channel) {
	if newChannel.Header != oldChannel.Header {
		// Update the cache.
		delete(p.channelConfigs, newChannel.Id)
	}
}

func main() {
	plugin.ClientMain(&MyPlugin{})
}
