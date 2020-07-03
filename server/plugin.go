package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	DelayInSecons = 2
)

type Plugin struct {
	plugin.MattermostPlugin
	configurationLock sync.RWMutex
	configuration     *configuration
	Store             *DBStore
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
	return nil
}

func (p *Plugin) OnDeactivate() error {
	p.Store.Close()
	return nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()
	router.ServeHTTP(w, r)
}

func (p *Plugin) UserHasJoinedChannel(c *plugin.Context, channelMember *model.ChannelMember, user *model.User) {
	if !p.getConfiguration().RecommendOnJoinChannel {
		return
	}
	time.Sleep(DelayInSecons * time.Second)

	channel, appErr := p.API.GetChannel(channelMember.ChannelId)
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

	if len(suggestions) > 0 {
		formattedSuggestions := ""
		for _, suggestion := range suggestions {
			formattedSuggestions += "~" + suggestion + " "
		}

		post := model.Post{
			ChannelId: channelMember.ChannelId,
			Message:   fmt.Sprintf("Other people who joined this channel also joined this other channels: %s\nmaybe you are interested in joining them too", formattedSuggestions),
		}
		p.API.SendEphemeralPost(channelMember.UserId, &post)
	}
}

func (p *Plugin) UserHasJoinedTeam(c *plugin.Context, teamMember *model.TeamMember, user *model.User) {
	if !p.getConfiguration().RecommendOnJoinTeam {
		return
	}
	time.Sleep(DelayInSecons * time.Second)

	suggestions, err := p.Store.MostActiveChannels(teamMember.UserId, teamMember.TeamId)
	if err != nil {
		p.API.LogError(err.Error())
	}

	message := ""
	if len(suggestions) > 0 {
		formattedSuggestions := ""
		for _, suggestion := range suggestions {
			formattedSuggestions += "~" + suggestion + " "
		}
		message += fmt.Sprintf("The most active channels in this team lately are: %s\n", formattedSuggestions)
	}

	suggestions, err = p.Store.MostPopulatedChannels(teamMember.UserId, teamMember.TeamId)
	if err != nil {
		p.API.LogError(err.Error())
	}

	if len(suggestions) > 0 {
		formattedSuggestions := ""
		for _, suggestion := range suggestions {
			formattedSuggestions += "~" + suggestion + " "
		}
		message += fmt.Sprintf("The most popular channels in this team are: %s\n", formattedSuggestions)
	}

	defaultChannel, appErr := p.API.GetChannelByName(teamMember.TeamId, "town-square", false)
	if appErr != nil {
		p.API.LogError(appErr.Error())
		return
	}
	post := model.Post{
		ChannelId: defaultChannel.Id,
		Message:   message,
	}
	p.API.SendEphemeralPost(teamMember.UserId, &post)
}
