package youtube

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func ParseVideoMetadata(doc *goquery.Document, videoURL string) (*VideoMetadata, error) {
	videoID := extractVideoID(videoURL)

	channelName := extractChannelName(doc)
	channelURL := extractChannelURL(doc)
	publishedDate := extractPublishedDate(doc)
	duration := extractDuration(doc)
	thumbnailURL := getThumbnailURL(videoID)

	return &VideoMetadata{
		ChannelName:   channelName,
		ChannelURL:    channelURL,
		ViewCount:     0,
		LikeCount:     0,
		Duration:      duration,
		PublishedDate: publishedDate,
		VideoID:       videoID,
		ThumbnailURL:  thumbnailURL,
	}, nil
}

func ParseChannelMetadata(doc *goquery.Document) (*ChannelMetadata, error) {
	title := doc.Find("title").First().Text()
	if idx := strings.Index(title, " - YouTube"); idx != -1 {
		title = title[:idx]
	}

	description, _ := doc.Find("meta[name='description']").Attr("content")
	thumbnail, _ := doc.Find("meta[property='og:image']").Attr("content")

	return &ChannelMetadata{
		Name:            title,
		Description:     description,
		ThumbnailURL:    thumbnail,
		SubscriberCount: 0,
		VideoCount:      0,
	}, nil
}

func ParseVideoContent(doc *goquery.Document) (*VideoContent, error) {
	title := extractTitle(doc)
	description := extractDescription(doc)

	return &VideoContent{
		Title:       title,
		Description: description,
		Tags:        extractTags(doc),
	}, nil
}

func extractVideoID(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	if u.Host == "youtu.be" {
		return strings.TrimPrefix(u.Path, "/")
	}

	if u.Host == "www.youtube.com" || u.Host == "youtube.com" {
		if strings.Contains(u.Path, "/shorts/") {
			parts := strings.Split(u.Path, "/")
			if len(parts) >= 3 {
				return parts[2]
			}
		}
		return u.Query().Get("v")
	}

	return ""
}

func extractTitle(doc *goquery.Document) string {
	title := doc.Find("title").First().Text()
	title = strings.TrimSpace(strings.TrimSuffix(title, "- YouTube"))
	title = strings.TrimSpace(strings.TrimSuffix(title, "- YouTube Music"))
	title = strings.TrimSpace(strings.TrimSuffix(title, "- YouTube Kids"))
	return title
}

func extractChannelName(doc *goquery.Document) string {
	name := doc.Find("link[itemprop='name']").AttrOr("content", "")
	if name != "" {
		return name
	}

	name = doc.Find("yt-formatted-string#owner-name").First().Text()
	if name != "" {
		return strings.TrimSpace(name)
	}

	name = doc.Find("#channel-name a").First().Text()
	if name != "" {
		return strings.TrimSpace(name)
	}

	name, _ = doc.Find("meta[property='og:site_name']").Attr("content")
	return name
}

func extractChannelURL(doc *goquery.Document) string {
	url, _ := doc.Find("span[itemprop='author'] link[itemprop='url']").Attr("href")
	if url != "" {
		return url
	}

	url = doc.Find("#owner-link").AttrOr("href", "")
	if url != "" {
		return "https://www.youtube.com" + url
	}

	return ""
}

func extractPublishedDate(doc *goquery.Document) time.Time {
	dateStr := doc.Find("meta[itemprop='uploadDate']").AttrOr("content", "")
	if dateStr == "" {
		dateStr = doc.Find("meta[itemprop='datePublished']").AttrOr("content", "")
	}

	if dateStr == "" {
		return time.Time{}
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

func extractDuration(doc *goquery.Document) time.Duration {
	durationStr := doc.Find("meta[itemprop='duration']").AttrOr("content", "PT0S")
	return parseISODuration(durationStr)
}

func extractDescription(doc *goquery.Document) string {
	desc := doc.Find("meta[name='description']").AttrOr("content", "")
	if desc == "" {
		desc, _ = doc.Find("meta[property='og:description']").Attr("content")
	}
	return desc
}

func extractTags(doc *goquery.Document) string {
	tags := doc.Find("meta[name='keywords']").AttrOr("content", "")
	if tags == "" {
		tags, _ = doc.Find("meta[property='og:video:tag']").Attr("content")
	}
	return tags
}

func getThumbnailURL(videoID string) string {
	if videoID == "" {
		return ""
	}
	return "https://img.youtube.com/vi/" + videoID + "/hqdefault.jpg"
}

func parseISODuration(duration string) time.Duration {
	re := regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)
	matches := re.FindStringSubmatch(duration)

	if len(matches) == 0 {
		return 0
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])

	return time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes) + time.Second*time.Duration(seconds)
}

func DetectContentType(url string) ContentType {
	if strings.Contains(url, "/shorts/") {
		return ContentTypeShort
	}
	if strings.Contains(url, "/channel/") || strings.Contains(url, "/@") || strings.Contains(url, "/c/") {
		return ContentTypeChannel
	}
	if strings.Contains(url, "/playlist?") {
		return ContentTypePlaylist
	}
	return ContentTypeVideo
}
