package main

import "fmt"

func channelsMessage(text string, teamName string, channels []ChannelData, extraText string) string {
	if len(channels) == 0 {
		return ""
	}
	channelsList := ""
	for _, channel := range channels {
		channelsList += fmt.Sprintf("[%s](/%s/channels/%s) ", channel.DisplayName, teamName, channel.Name)
	}
	return fmt.Sprintf("%s%s%s\n\n", text, channelsList, extraText)
}
