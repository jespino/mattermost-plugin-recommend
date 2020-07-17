package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

const (
	delayInSecons = 6
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
	store, err := NewDBStore(*config.SqlSettings.DriverName, *config.SqlSettings.DataSource)
	if err != nil {
		return err
	}
	p.Store = store
	if err = p.API.RegisterCommand(getCommand()); err != nil {
		return err
	}

	recommendBot := &model.Bot{
		Username:    "recommend-bot",
		DisplayName: "Recommend Bot",
		Description: "A bot account created by the com.github.jespino.recommend plugin",
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
	if user.CreateAt+int64(p.getConfiguration().GracePeriod*1000) > model.GetMillis() {
		return true
	}
	return false
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()
	router.ServeHTTP(w, r)
}

func (p *Plugin) UserHasJoinedChannel(c *plugin.Context, channelMember *model.ChannelMember, user *model.User) {
	if !p.getConfiguration().RecommendOnJoinChannel {
		return
	}
	if p.isInGracePeriod(user) {
		return
	}
	time.Sleep(delayInSecons * time.Second)

	channel, appErr := p.API.GetChannel(channelMember.ChannelId)
	if appErr != nil {
		p.API.LogError(appErr.Error())
	}

	team, appErr := p.API.GetTeam(channel.TeamId)
	if appErr != nil {
		p.API.LogError(appErr.Error())
	}

	if channel.Name == "town-square" || channel.Name == "off-topic" {
		return
	}

	suggestions, err := p.Store.MostPopularChannelsByChannel(channelMember.UserId, channelMember.ChannelId, channel.TeamId)
	if err != nil {
		p.API.LogError(err.Error())
	}
	if len(suggestions) == 0 {
		return
	}
	message := channelsMessage("Others who joined this channel also joined ", team.Name, suggestions, ". You may be interested joining them too!")
	post := model.Post{
		UserId:    p.botID,
		ChannelId: channelMember.ChannelId,
		Message:   message,
	}
	p.API.SendEphemeralPost(channelMember.UserId, &post)
}

func (p *Plugin) UserHasJoinedTeam(c *plugin.Context, teamMember *model.TeamMember, user *model.User) {
	if !p.getConfiguration().RecommendOnJoinTeam {
		return
	}
	if p.isInGracePeriod(user) {
		return
	}
	time.Sleep(delayInSecons * time.Second)

	message := ""

	team, appErr := p.API.GetTeam(teamMember.TeamId)
	if appErr != nil {
		p.API.LogError(appErr.Error())
	}

	suggestions, err := p.Store.MostActiveChannels(teamMember.UserId, teamMember.TeamId)
	if err != nil {
		p.API.LogError(err.Error())
	}
	message += channelsMessage("Currently the most active channels in this team are: ", team.Name, suggestions, "")

	suggestions, err = p.Store.MostPopulatedChannels(teamMember.UserId, teamMember.TeamId)
	if err != nil {
		p.API.LogError(err.Error())
	}
	message += channelsMessage("The most popular channels in this team are: ", team.Name, suggestions, "")

	if message == "" {
		return
	}

	defaultChannel, appErr := p.API.GetChannelByName(teamMember.TeamId, "town-square", false)
	if appErr != nil {
		p.API.LogError(appErr.Error())
		return
	}
	post := model.Post{
		UserId:    p.botID,
		ChannelId: defaultChannel.Id,
		Message:   message,
	}
	p.API.SendEphemeralPost(teamMember.UserId, &post)
}
