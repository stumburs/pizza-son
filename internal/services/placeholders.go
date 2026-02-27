package services

import (
	"strconv"
	"strings"
	"time"
)

func FillPlaceholders(prompt, channel string) string {
	channelInfo := TwitchServiceInstance.GetStreamInfo(channel)

	// TODO: Replace with proper placeholders
	replacer := strings.NewReplacer(
		"{{game_name}}", channelInfo.GameName,
		"{{channel_name}}", channelInfo.ChannelName,
		"{{stream_title}}", channelInfo.StreamTitle,
		"{{channel_tags}}", strings.Join(channelInfo.ChannelTags, ", "),
		"{{viewer_count}}", strconv.Itoa(channelInfo.ViewerCount),
		"{{thumbnail_url}}", channelInfo.ThumbnailURL,
		"{{time_cest}}", func() string {
			loc, _ := time.LoadLocation("Europe/Paris") // CEST
			return time.Now().In(loc).Format("2006-01-02 15:04")
		}(),
	)

	result := replacer.Replace(prompt)

	return result
}
