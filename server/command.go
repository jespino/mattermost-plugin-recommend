package main

import (
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

func (p *Plugin) sendResponse(userID string, channelID string, text string) {
	post := model.Post{
		UserId:    p.botID,
		ChannelId: channelID,
		Message:   text,
	}
	p.API.SendEphemeralPost(userID, &post)
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
		p.sendResponse(args.UserId, args.ChannelId, commandHelp)
		return &model.CommandResponse{}, nil
	}

	message := ""

	team, appErr := p.API.GetTeam(args.TeamId)
	if appErr != nil {
		p.sendResponse(args.UserId, args.ChannelId, "Error: Unable to get the current team")
		return &model.CommandResponse{}, nil
	}

	channels, err := p.Store.MostActiveChannels(args.UserId, args.TeamId)
	if err != nil {
		p.sendResponse(args.UserId, args.ChannelId, err.Error())
		return &model.CommandResponse{}, nil
	}
	message += channelsMessage("Most active channels for the current team", team.Name, channels)

	channels, err = p.Store.MostPopulatedChannels(args.UserId, args.TeamId)
	if err != nil {
		p.sendResponse(args.UserId, args.ChannelId, err.Error())
		return &model.CommandResponse{}, nil
	}
	message += channelsMessage("Most populated channels for the current team", team.Name, channels)

	channels, err = p.Store.MostPopularChannelsByChannel(args.UserId, args.ChannelId, args.TeamId)
	if err != nil {
		p.sendResponse(args.UserId, args.ChannelId, err.Error())
		return &model.CommandResponse{}, nil
	}
	message += channelsMessage("Suggested channels for the current team (based on the users of the current channel)", team.Name, channels)

	channels, err = p.Store.MostPopularChannelsByUserCoMembers(args.UserId, args.TeamId)
	if err != nil {
		p.sendResponse(args.UserId, args.ChannelId, err.Error())
		return &model.CommandResponse{}, nil
	}
	message += channelsMessage("Suggested channels for the current team (based on the users the channels that you are member)", team.Name, channels)

	if message == "" {
		message = "No recommendations found for you"
	}
	p.sendResponse(args.UserId, args.ChannelId, message)
	return &model.CommandResponse{}, nil
}
