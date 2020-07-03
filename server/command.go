package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-server/v5/model"
)

const commandHelp = `* |/recommend| - Recommend me channels
* |/recommend help| - Show command's help`

func getCommand() *model.Command {
	return &model.Command{
		Trigger:          "recommend",
		DisplayName:      "Recommend",
		Description:      "Mattermost Recommend command",
		AutoComplete:     true,
		AutoCompleteDesc: "It recommends channels for you",
	}
}

func (p *Plugin) getCommandResponse(responseType, text string) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: responseType,
		Text:         text,
		Username:     "Mattermost Recommend",
		Type:         model.POST_DEFAULT,
	}
}

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

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	split := strings.Fields(args.Command)
	command := split[0]
	action := ""
	if len(split) > 1 {
		action = split[1]
	}

	if command != "/recommend" {
		return &model.CommandResponse{}, nil
	}

	if action != "" {
		return p.getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, commandHelp), nil
	}

	message := ""

	channels, err := p.GetMostActiveChannelsForTeam(args.UserId, args.TeamId)
	if err != nil {
		return p.getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, err.Error()), nil
	}
	message += channelsMessage("Most active channels for the current team", channels)

	channels, err = p.GetMostPopulatedChannelsForTeam(args.UserId, args.TeamId)
	if err != nil {
		return p.getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, err.Error()), nil
	}
	message += channelsMessage("Most populated channels for the current team", channels)

	channels, err = p.GetMostPopularChannelsForTheChannelMembersOfAChannel(args.UserId, args.ChannelId, args.TeamId)
	if err != nil {
		return p.getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, err.Error()), nil
	}
	message += channelsMessage("Suggested channels for the current team (based on the users of the current channel)", channels)

	channels, err = p.GetMostPopularChannelsForTheChannelMembersOfMyChannels(args.UserId, args.TeamId)
	if err != nil {
		return p.getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, err.Error()), nil
	}
	message += channelsMessage("Suggested channels for the current team (based on the users the channels that you are member)", channels)

	if message == "" {
		return p.getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, "No recomendations found for you"), nil
	}
	return p.getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, message), nil
}
