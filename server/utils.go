package main

import (
	"fmt"
)

func channelsMessage(text string, teamName string, channels []ChannelData, extraText string) string {
	if len(channels) == 0 {
		return ""
	}
	channelsList := ""
	for _, channel := range channels {
		channelsList += fmt.Sprintf("~%s ", channel.Name)
	}
	return fmt.Sprintf("%s%s%s\n\n", text, channelsList, extraText)
}

func channelsMentionsMetadata(teamName string, suggestions []ChannelData) map[string]interface{} {
	result := map[string]interface{}{}
	for _, suggestion := range suggestions {
		result[suggestion.Name] = map[string]interface{}{
			"display_name": suggestion.DisplayName,
			"team_name":    teamName,
		}
	}
	return result
}
