package main

import (
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

const (
	delayInSecons           = 6
	moreRecommendationsText = "\n\nGet more recommendations using the `/recommend` command."
)

type Plugin struct {
	plugin.MattermostPlugin
	configurationLock sync.RWMutex
	configuration     *configuration
	Store             *DBStore
	botID             string
}

func (p *Plugin) OnActivate() error {
	config := p.API.GetUnsanitizedConfig()
	var err error
	var store *DBStore
	if len(config.SqlSettings.DataSourceReplicas) > 0 {
		store, err = NewDBStore(*config.SqlSettings.DriverName, config.SqlSettings.DataSourceReplicas[0])
	} else {
		store, err = NewDBStore(*config.SqlSettings.DriverName, *config.SqlSettings.DataSource)
	}
	if err != nil {
		return err
	}
	p.Store = store
	if err = p.API.RegisterCommand(getCommand()); err != nil {
		return err
	}

	recommendBot := &model.Bot{
		Username:    "channel-recommender",
		DisplayName: "Channel Recommender Bot",
		Description: "A bot account created by the Channel Recommender plugin",
	}

	options := []plugin.EnsureBotOption{
		plugin.ProfileImagePath("assets/icon.png"),
	}

	botID, ensureBotError := p.Helpers.EnsureBot(recommendBot, options...)
	if ensureBotError != nil {
		return errors.Wrap(ensureBotError, "failed to ensure recommend bot user.")
	}

	p.botID = botID

	return nil
}

func (p *Plugin) OnDeactivate() error {
	p.Store.Close()
	return nil
}

func (p *Plugin) isInGracePeriod(user *model.User) bool {
	return user.CreateAt+int64(p.getConfiguration().GracePeriod*1000) > model.GetMillis()
}

func (p *Plugin) UserHasJoinedChannel(c *plugin.Context, channelMember *model.ChannelMember, actor *model.User) {
	if !p.getConfiguration().RecommendOnJoinChannel {
		return
	}

	user, appErr := p.API.GetUser(channelMember.UserId)
	if appErr != nil {
		p.API.LogError(appErr.Error())
		return
	}

	if p.isInGracePeriod(user) {
		return
	}
	time.Sleep(delayInSecons * time.Second)

	channel, appErr := p.API.GetChannel(channelMember.ChannelId)
	if appErr != nil {
		p.API.LogError(appErr.Error())
		return
	}
	if !p.canReceiveRecommendations(user.Id, channel.TeamId) {
		return
	}

	team, appErr := p.API.GetTeam(channel.TeamId)
	if appErr != nil {
		p.API.LogError(appErr.Error())
		return
	}

	if channel.Name == "town-square" || channel.Name == "off-topic" {
		return
	}

	suggestions, err := p.Store.MostPopularChannelsByChannel(channelMember.UserId, channelMember.ChannelId, channel.TeamId)
	if err != nil {
		p.API.LogError(err.Error())
		return
	}
	if len(suggestions) == 0 {
		return
	}
	message := channelsMessage("Others who joined this channel also joined ", team.Name, suggestions, ". You may be interested joining them too!")
	channelsMentions := channelsMentionsMetadata(team.Name, suggestions)
	message += moreRecommendationsText
	post := model.Post{
		UserId:    p.botID,
		ChannelId: channelMember.ChannelId,
		Message:   message,
		Props: map[string]interface{}{
			"channel_mentions":        channelsMentions,
			"disable_group_highlight": true,
		},
	}
	p.API.SendEphemeralPost(channelMember.UserId, &post)
}

func (p *Plugin) UserHasJoinedTeam(c *plugin.Context, teamMember *model.TeamMember, actor *model.User) {
	if !p.getConfiguration().RecommendOnJoinTeam {
		return
	}

	user, appErr := p.API.GetUser(teamMember.UserId)
	if appErr != nil {
		p.API.LogError(appErr.Error())
		return
	}

	if p.isInGracePeriod(user) {
		return
	}
	if !p.canReceiveRecommendations(user.Id, teamMember.TeamId) {
		return
	}

	time.Sleep(delayInSecons * time.Second)

	message := ""

	team, appErr := p.API.GetTeam(teamMember.TeamId)
	if appErr != nil {
		p.API.LogError(appErr.Error())
		return
	}

	activityThreshold := p.getConfiguration().ActivityThreshold
	if activityThreshold == 0 {
		activityThreshold = 7 * 24 * 60 // A week
	}

	suggestions, err := p.Store.MostActiveChannels(teamMember.UserId, teamMember.TeamId, activityThreshold)
	if err != nil {
		p.API.LogError(err.Error())
	}
	message += channelsMessage("Currently, the most active channels in this team are: ", team.Name, suggestions, "")
	channelsMentions := channelsMentionsMetadata(team.Name, suggestions)
	allChannelsMentions := channelsMentions

	suggestions, err = p.Store.MostPopulatedChannels(teamMember.UserId, teamMember.TeamId)
	if err != nil {
		p.API.LogError(err.Error())
	}
	message += channelsMessage("The most popular channels in this team are: ", team.Name, suggestions, "")
	channelsMentions = channelsMentionsMetadata(team.Name, suggestions)
	for k, v := range channelsMentions {
		allChannelsMentions[k] = v
	}

	if message == "" {
		return
	}

	defaultChannel, appErr := p.API.GetChannelByName(teamMember.TeamId, "town-square", false)
	if appErr != nil {
		p.API.LogError(appErr.Error())
		return
	}
	message += moreRecommendationsText

	post := model.Post{
		UserId:    p.botID,
		ChannelId: defaultChannel.Id,
		Message:   message,
		Props: map[string]interface{}{
			"channel_mentions":        allChannelsMentions,
			"disable_group_highlight": true,
		},
	}
	p.API.SendEphemeralPost(teamMember.UserId, &post)
}

func (p *Plugin) canReceiveRecommendations(userID string, teamID string) bool {
	canListChannels := p.API.HasPermissionToTeam(userID, teamID, model.PERMISSION_LIST_TEAM_CHANNELS)
	canJoinPublicChannels := p.API.HasPermissionToTeam(userID, teamID, model.PERMISSION_JOIN_PUBLIC_CHANNELS)
	return canListChannels && canJoinPublicChannels
}
