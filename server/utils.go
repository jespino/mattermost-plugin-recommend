package main

import "fmt"

func channelsMessage(text string, channels []string) string {
	if len(channels) == 0 {
		return ""
	}
	channelsList := ""
	for _, channel := range channels {
		channelsList += fmt.Sprintf("~%s ", channel)
	}
	return fmt.Sprintf("%s: %s\n\n", text, channelsList)
}
