# Mattermost Recommend

[![Build Status](https://img.shields.io/circleci/project/github/jespino/mattermost-plugin-recommend/master)](https://circleci.com/gh/jespino/mattermost-plugin-recommend)
[![Release](https://img.shields.io/github/v/release/jespino/mattermost-plugin-recommend)](https://github.com/jespino/mattermost-plugin-recommend/releases/latest)
[![HW](https://img.shields.io/github/issues/jespino/mattermost-plugin-recommend/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/jespino/mattermost-plugin-recommend/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)

This plugin recommends you channels based on your memberships and the most popular channels in a team.

![image](https://user-images.githubusercontent.com/290303/90430523-e26b0680-e0c7-11ea-8b25-5f7510223cff.png)

__Requires Mattermost 5.16 or higher.__

## Features

- Use a `/recommend` command to get recommendations based in the current channel and the current team.
- Enable automatic recommendations when a user joins a channel or a team.
- Give a grace period to new members before they start to get automatic recommendations.

## Installation

1. Go to https://github.com/jespino/mattermost-plugin-recommend/releases to download the latest release file in tar.gz format.
2. Upload the file through **System Console > Plugins > Management**, or manually upload it to the Mattermost server under plugin directory. See [documentation](https://docs.mattermost.com/administration/plugins.html#set-up-guide) for more details.

## Configuration

Go to **System Console > Plugins > Recommend** and set the following values:

1. **Enable Plugin**: ``true``
2. **Recommend at team join**: When user joins to a team, recommend bot is going to recommend interesting channels in that team.
3. **Recommend at channel join**: When user joins to a channel, recommend bot is going to recommend other channels in the team based on the people in that channel.
4. **Grace Period**: Give a period of time since the user was created before start sending automatic messages on join.

You're all set! To test it, go to any Mattermost channel and execute the `/recommend` command.

### Manual Builds

To build the project you can use the existing `Makefile` in the repo. Use `make dist` to compile and compress the plugin into a .tar.gz that you can install in your Mattermost instance.

Inside the `/server` directory, you will find the Go files of the plugin.
