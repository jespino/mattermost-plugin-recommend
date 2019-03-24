package main

func (p *Plugin) GetMostActiveChannelsForTeam(userID, teamID string) ([]string, error) {
	return p.Store.MostActiveChannels(userID, teamID)
}

func (p *Plugin) GetMostPopulatedChannelsForTeam(userID, teamID string) ([]string, error) {
	return p.Store.MostPopulatedChannels(userID, teamID)
}

func (p *Plugin) GetMostPopularChannelsForTheChannelMembersOfAChannel(userID, channelID, teamID string) ([]string, error) {
	return p.Store.MostPopularChannelsByChannel(userID, channelID, teamID)
}

func (p *Plugin) GetMostPopularChannelsForTheChannelMembersOfMyChannels(userID, teamID string) ([]string, error) {
	return p.Store.MostPopularChannelsByUserCoMembers(userID, teamID)
}
