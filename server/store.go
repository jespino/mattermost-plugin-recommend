package main

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	ActivePeriodInMinutes = 7 * 24 * 60 // A week
)

type DBStore struct {
	conn *sql.DB
	sq   sq.StatementBuilderType
}

func NewDBStore(driverName, dataSource string) (*DBStore, error) {
	db, err := sql.Open(driverName, dataSource)
	if err != nil {
		return nil, err
	}

	builder := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	if driverName == model.DATABASE_DRIVER_POSTGRES {
		builder = builder.PlaceholderFormat(sq.Dollar)
	}
	builder = builder.RunWith(db)

	return &DBStore{conn: db, sq: builder}, nil
}

func (db *DBStore) Close() {
	db.conn.Close()
}

func (db *DBStore) MostActiveChannels(userID, teamID string) ([]string, error) {
	myChannels, err := db.getMyChannelsForTeam(userID, teamID)
	if err != nil {
		return nil, err
	}

	lastWeek := model.GetMillis() - (ActivePeriodInMinutes * 60 * 1000)
	query := db.sq.Select("C.Name").
		From("Posts AS P").
		LeftJoin("Channels AS C ON P.ChannelId = C.Id").
		Where(sq.Gt{"P.CreateAt": lastWeek}).
		Where(sq.Eq{"C.Type": model.CHANNEL_OPEN}).
		Where(sq.Eq{"C.TeamId": teamID}).
		Where(sq.NotEq{"C.Id": myChannels}).
		GroupBy("C.Name").
		OrderBy("Count(P.Id) DESC").
		Limit(3)
	rows, err := query.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	channels := []string{}
	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func (db *DBStore) MostPopulatedChannels(userID, teamID string) ([]string, error) {
	myChannels, err := db.getMyChannelsForTeam(userID, teamID)
	if err != nil {
		return nil, err
	}

	query := db.sq.Select("C.Name").
		From("ChannelMembers AS CM").
		LeftJoin("Channels AS C ON CM.ChannelId = C.Id").
		Where(sq.Eq{"C.TeamId": teamID}).
		Where(sq.NotEq{"CM.UserId": userID}).
		Where(sq.NotEq{"C.Id": myChannels}).
		Where(sq.Eq{"C.Type": model.CHANNEL_OPEN}).
		GroupBy("C.Name").
		OrderBy("Count(CM.UserId) DESC").
		Limit(3)

	rows, err := query.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	channels := []string{}
	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func (db *DBStore) getChannelMembers(channelID string) ([]string, error) {
	query := db.sq.Select("UserId").
		From("ChannelMembers").
		Where(sq.Eq{"ChannelId": channelID})

	rows, err := query.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []string{}
	for rows.Next() {
		var user string
		if err := rows.Scan(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (db *DBStore) MostPopularChannelsByChannel(userID, channelID, teamID string) ([]string, error) {
	otherMembersInChannel, err := db.getChannelMembers(channelID)
	if err != nil {
		return nil, err
	}

	myChannels, err := db.getMyChannelsForTeam(userID, teamID)
	if err != nil {
		return nil, err
	}

	query := db.sq.Select("C.Name").
		From("ChannelMembers AS CM").
		LeftJoin("Channels AS C ON CM.ChannelId = C.Id").
		Where(sq.Eq{"CM.UserId": otherMembersInChannel}).
		Where(sq.NotEq{"C.Id": myChannels}).
		Where(sq.Eq{"C.Type": model.CHANNEL_OPEN}).
		Where(sq.Eq{"C.TeamId": teamID}).
		GroupBy("C.Name").
		OrderBy("Count(CM.UserId) DESC").
		Limit(3)
	rows, err := query.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	channels := []string{}
	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func (db *DBStore) getMyChannelsForTeam(userID string, teamID string) ([]string, error) {
	query := db.sq.Select("ChannelId").
		From("ChannelMembers").
		LeftJoin("Channels ON Channels.Id=ChannelMembers.ChannelId").
		Where(sq.Eq{"UserId": userID}).
		Where(sq.Eq{"TeamId": teamID})
	rows, err := query.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	channels := []string{}
	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func (db *DBStore) getMyCoMembersForTeam(userID string, teamID string) ([]string, error) {
	myChannels, err := db.getMyChannelsForTeam(userID, teamID)
	if err != nil {
		return nil, err
	}

	query := db.sq.Select("UserId").
		From("ChannelMembers").
		LeftJoin("Channels AS C ON ChannelMembers.ChannelId=C.Id").
		Where(sq.Eq{"ChannelId": myChannels}).
		Where(sq.NotEq{"Name": "town-square"}).
		Where(sq.NotEq{"UserId": userID})

	rows, err := query.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []string{}
	for rows.Next() {
		var user string
		if err := rows.Scan(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (db *DBStore) MostPopularChannelsByUserCoMembers(userID, teamID string) ([]string, error) {
	myChannels, err := db.getMyChannelsForTeam(userID, teamID)
	if err != nil {
		return nil, err
	}
	myCoMembers, err := db.getMyCoMembersForTeam(userID, teamID)
	if err != nil {
		return nil, err
	}

	query := db.sq.Select("C.Name").
		From("ChannelMembers AS CM").
		LeftJoin("Channels AS C ON CM.ChannelId = C.Id").
		Where(sq.Eq{"C.Type": model.CHANNEL_OPEN}).
		Where(sq.Eq{"C.TeamId": teamID}).
		Where(sq.Eq{"CM.UserId": myCoMembers}).
		Where(sq.NotEq{"CM.ChannelId": myChannels}).
		GroupBy("C.Name").
		OrderBy("Count(CM.UserId) DESC").
		Limit(3)

	rows, err := query.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	channels := []string{}
	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}
