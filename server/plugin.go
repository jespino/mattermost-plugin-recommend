package main

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
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
	if p.configuration.DBDriverName == "" || p.configuration.DBDataSource == "" {
		return errors.New("You need to properly configure your database access")
	}
	store, err := NewDBStore(p.configuration.DBDriverName, p.configuration.DBDataSource)
	if err != nil {
		return err
	}
	p.Store = store
	p.API.RegisterCommand(getCommand())
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
	if !p.configuration.RecommendOnJoinChannel {
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

	suggestions, err := p.GetMostPopularChannelsForTheChannelMembersOfAChannel(channelMember.UserId, channelMember.ChannelId, channel.TeamId)
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
	if !p.configuration.RecommendOnJoinTeam {
		return
	}
	time.Sleep(DelayInSecons * time.Second)

	suggestions, err := p.GetMostActiveChannelsForTeam(teamMember.UserId, teamMember.TeamId)
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

	suggestions, err = p.GetMostPopulatedChannelsForTeam(teamMember.UserId, teamMember.TeamId)
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
