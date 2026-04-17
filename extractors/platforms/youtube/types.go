package youtube

import "time"

type VideoMetadata struct {
	ChannelName   string
	ChannelURL    string
	ViewCount     int64
	LikeCount     int64
	Duration      time.Duration
	PublishedDate time.Time
	VideoID       string
	ThumbnailURL  string
}

type ChannelMetadata struct {
	Handle          string
	Name            string
	SubscriberCount int64
	VideoCount      int
	Description     string
	ThumbnailURL    string
}

type VideoContent struct {
	Title       string
	Description string
	Tags        string
}

type ShortContent struct {
	Title       string
	Description string
}

type ContentType string

const (
	ContentTypeVideo    ContentType = "video"
	ContentTypeShort    ContentType = "short"
	ContentTypeChannel  ContentType = "channel"
	ContentTypePlaylist ContentType = "playlist"
)

func (c ContentType) String() string {
	return string(c)
}
