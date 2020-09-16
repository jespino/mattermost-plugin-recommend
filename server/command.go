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

func (p *Plugin) sendResponse(userID string, channelID string, text string, channelsMentions map[string]interface{}) {
	post := model.Post{
		UserId:    p.botID,
		ChannelId: channelID,
		Message:   text,
	}
	if channelsMentions != nil {
		post.Props = map[string]interface{}{
			"channel_mentions":        channelsMentions,
			"disable_group_highlight": true,
		}
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

	if !p.canReceiveRecommendations(args.UserId, args.TeamId) {
		p.sendResponse(args.UserId, args.ChannelId, "You don't have permission to get channels recommendations.", nil)
		return &model.CommandResponse{}, nil
	}

	if command != "/recommend" {
		return &model.CommandResponse{}, nil
	}

	if action != "" {
		p.sendResponse(args.UserId, args.ChannelId, commandHelp, nil)
		return &model.CommandResponse{}, nil
	}

	message := ""

	team, appErr := p.API.GetTeam(args.TeamId)
	if appErr != nil {
		p.sendResponse(args.UserId, args.ChannelId, "Error: Unable to get the current team", nil)
		return &model.CommandResponse{}, nil
	}

	activityThreshold := p.getConfiguration().ActivityThreshold
	if activityThreshold == 0 {
		activityThreshold = 7 * 24 * 60 // A week
	}

	channels, err := p.Store.MostActiveChannels(args.UserId, args.TeamId, activityThreshold)
	if err != nil {
		p.sendResponse(args.UserId, args.ChannelId, err.Error(), nil)
		return &model.CommandResponse{}, nil
	}
	message += channelsMessage("Most active channels for the current team: ", team.Name, channels, "")
	channelsMentions := channelsMentionsMetadata(team.Name, channels)
	allChannelsMentions := channelsMentions

	channels, err = p.Store.MostPopulatedChannels(args.UserId, args.TeamId)
	if err != nil {
		p.sendResponse(args.UserId, args.ChannelId, err.Error(), nil)
		return &model.CommandResponse{}, nil
	}
	message += channelsMessage("Most populated channels for the current team: ", team.Name, channels, "")
	channelsMentions = channelsMentionsMetadata(team.Name, channels)
	for k, v := range channelsMentions {
		allChannelsMentions[k] = v
	}

	channels, err = p.Store.MostPopularChannelsByChannel(args.UserId, args.ChannelId, args.TeamId)
	if err != nil {
		p.sendResponse(args.UserId, args.ChannelId, err.Error(), nil)
		return &model.CommandResponse{}, nil
	}
	message += channelsMessage("Suggested channels for the current team (based on the users of the current channel): ", team.Name, channels, "")
	channelsMentions = channelsMentionsMetadata(team.Name, channels)
	for k, v := range channelsMentions {
		allChannelsMentions[k] = v
	}

	channels, err = p.Store.MostPopularChannelsByUserCoMembers(args.UserId, args.TeamId)
	if err != nil {
		p.sendResponse(args.UserId, args.ChannelId, err.Error(), nil)
		return &model.CommandResponse{}, nil
	}
	message += channelsMessage("Suggested channels for the current team (based on the users of the channels you are member of): ", team.Name, channels, "")
	channelsMentions = channelsMentionsMetadata(team.Name, channels)
	for k, v := range channelsMentions {
		allChannelsMentions[k] = v
	}

	if message == "" {
		message = "No recommendations found for you."
	}
	p.sendResponse(args.UserId, args.ChannelId, message, allChannelsMentions)
	return &model.CommandResponse{}, nil
}
