package github

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ParseRepoMetadata(doc *goquery.Document, url string) (*RepoMetadata, error) {
	owner, repo := extractPathParts(url)

	description := extractDescription(doc)
	stars := extractStars(doc)
	forks := extractForks(doc)
	openIssues := extractIssues(doc)
	language := extractLanguage(doc)
	license := extractLicense(doc)
	topics := extractTopics(doc)
	isPrivate := extractIsPrivate(doc)
	isFork := extractIsFork(doc)
	homepage := extractHomepage(doc)

	return &RepoMetadata{
		Owner:       owner,
		Name:        repo,
		FullName:    owner + "/" + repo,
		Description: description,
		Stars:       stars,
		Forks:       forks,
		OpenIssues:  openIssues,
		Language:    language,
		License:     license,
		Topics:      topics,
		IsPrivate:   isPrivate,
		IsFork:      isFork,
		Homepage:    homepage,
	}, nil
}

func ParseReadmeContent(doc *goquery.Document) (*ReadmeContent, error) {
	content := doc.Find("article.markdown-body").First().Text()
	if content == "" {
		content = doc.Find("article#readme").First().Text()
	}
	if content == "" {
		content = doc.Find(".markdown-body").First().Text()
	}

	return &ReadmeContent{
		Content: strings.TrimSpace(content),
	}, nil
}

func ParseProfileMetadata(doc *goquery.Document) (*ProfileMetadata, error) {
	username := doc.Find("h1").First().Text()
	name := doc.Find("span.p-name").First().Text()
	bio := doc.Find("p.p-note").First().Text()

	followers := 0
	following := 0

	doc.Find("a.Link--secondary").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "followers") {
			re := regexp.MustCompile(`([0-9,]+)`)
			matches := re.FindStringSubmatch(text)
			if len(matches) > 1 {
				followers, _ = strconv.Atoi(strings.ReplaceAll(matches[1], ",", ""))
			}
		}
		if strings.Contains(text, "following") {
			re := regexp.MustCompile(`([0-9,]+)`)
			matches := re.FindStringSubmatch(text)
			if len(matches) > 1 {
				following, _ = strconv.Atoi(strings.ReplaceAll(matches[1], ",", ""))
			}
		}
	})

	return &ProfileMetadata{
		Username:  username,
		Name:      name,
		Bio:       bio,
		Followers: followers,
		Following: following,
	}, nil
}

func extractPathParts(url string) (string, string) {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com")
	url = strings.TrimPrefix(url, "/")

	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

func extractDescription(doc *goquery.Document) string {
	desc := doc.Find("[data-testid='about-description'] p").First().Text()
	if desc == "" {
		desc = doc.Find(".repository-content .BorderGrid-cell p").First().Text()
	}
	if desc == "" {
		desc, _ = doc.Find("meta[name='description']").Attr("content")
	}
	if desc == "" {
		desc, _ = doc.Find("meta[property='og:description']").Attr("content")
	}
	return strings.TrimSpace(desc)
}

func extractStars(doc *goquery.Document) int {
	starsText := doc.Find("a[href$='/stargazers'] strong").First().Text()
	if starsText == "" {
		starsText = doc.Find("[data-testid='stargazers'] strong").First().Text()
	}
	return parseNumber(starsText)
}

func extractForks(doc *goquery.Document) int {
	forksText := doc.Find("a[href$='/forks'] strong").First().Text()
	if forksText == "" {
		forksText = doc.Find("[data-testid='forks'] strong").First().Text()
	}
	return parseNumber(forksText)
}

func extractIssues(doc *goquery.Document) int {
	issuesText := doc.Find("a[href$='/issues'] span.Counter").First().Text()
	if issuesText == "" {
		issuesText = doc.Find("[data-testid='issues-counter']").First().Text()
	}
	return parseNumber(issuesText)
}

func extractLanguage(doc *goquery.Document) string {
	lang := doc.Find("a[href*='search?l=']").First().AttrOr("href", "")
	if idx := strings.Index(lang, "?l="); idx != -1 {
		lang = lang[idx+3:]
		lang = strings.Title(strings.ToLower(lang))
		return lang
	}

	lang = doc.Find(".Progress-item").First().AttrOr("aria-label", "")
	if idx := strings.Index(lang, " "); idx != -1 {
		return strings.TrimSpace(lang[:idx])
	}

	lang = doc.Find("[data-testid='primary-language'] span").First().Text()
	return strings.TrimSpace(lang)
}

func extractLicense(doc *goquery.Document) string {
	license := ""
	doc.Find(".BorderGrid-cell h3").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "License") {
			license = s.Parent().Find("a").First().Text()
		}
	})
	return strings.TrimSpace(license)
}

func extractTopics(doc *goquery.Document) []string {
	var topics []string
	doc.Find("[data-testid='topic-tag'] a, .topic-tag").Each(func(i int, s *goquery.Selection) {
		topic := strings.TrimSpace(s.Text())
		if topic != "" {
			topics = append(topics, topic)
		}
	})
	return topics
}

func extractIsPrivate(doc *goquery.Document) bool {
	label := doc.Find("[data-testid='private-badge'], .Label:contains('Private')").First().Text()
	return strings.Contains(strings.ToLower(label), "private")
}

func extractIsFork(doc *goquery.Document) bool {
	text := doc.Find("[data-testid='fork-badge'], .fork-flag").First().Text()
	return strings.Contains(strings.ToLower(text), "fork")
}

func extractHomepage(doc *goquery.Document) string {
	homepage, _ := doc.Find("[data-testid='about-homepage'] a, .repository-content a[rel='nofollow']").First().Attr("href")
	if homepage != "" && !strings.Contains(homepage, "github.com") {
		return homepage
	}
	return ""
}

func parseNumber(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	text = strings.ReplaceAll(text, ",", "")

	lower := strings.ToLower(text)
	multiplier := 1
	if strings.HasSuffix(lower, "k") {
		multiplier = 1000
		text = strings.TrimSuffix(lower, "k")
	} else if strings.HasSuffix(lower, "m") {
		multiplier = 1000000
		text = strings.TrimSuffix(lower, "m")
	} else if strings.HasSuffix(lower, "b") {
		multiplier = 1000000000
		text = strings.TrimSuffix(lower, "b")
	}

	re := regexp.MustCompile(`[\d.]+`)
	match := re.FindString(text)
	if match == "" {
		return 0
	}

	num, _ := strconv.ParseFloat(strings.TrimSpace(match), 64)
	return int(num * float64(multiplier))
}

func isRepoURL(url string) bool {
	if !strings.Contains(strings.ToLower(url), "github.com") {
		return false
	}

	owner, repo := extractPathParts(url)
	if owner == "" || repo == "" {
		return false
	}

	if strings.Contains(repo, ".") || strings.Contains(repo, "?") {
		return false
	}

	knownNonRepos := []string{
		"settings", "explore", "marketplace", "notifications",
		"issues", "pulls", "trending", "new", "login", "logout",
		"signup", "features", "security", "enterprise", "team",
		"pricing", "topics", "collections", "events", "sponsors",
	}

	for _, nonRepo := range knownNonRepos {
		if strings.EqualFold(owner, nonRepo) {
			return false
		}
	}

	return true
}

func isProfileURL(url string) bool {
	if !strings.Contains(strings.ToLower(url), "github.com") {
		return false
	}

	parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
	if len(parts) >= 1 {
		username := parts[0]
		if username != "" && !strings.Contains(username, "/") {
			return true
		}
	}

	return false
}

func DetectContentType(url string) ContentType {
	if isProfileURL(url) {
		return ContentTypeProfile
	}
	if isRepoURL(url) {
		return ContentTypeRepo
	}
	if strings.Contains(url, "/releases") {
		return ContentTypeRelease
	}
	return ContentTypeRepo
}
