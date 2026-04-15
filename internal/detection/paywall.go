package detection

import (
	"fmt"
	"strings"

	"github.com/aluknot/scraper-lib/internal/dom"
	"golang.org/x/net/html"
)

// detectPaywall searches for paywall signals in raw HTML.
// MVP: marks with warnings only, does not attempt to bypass.
// Returns whether paywall signals were detected and the list of signals.
func DetectPaywall(rawHTML string) (bool, []string) {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return false, nil
	}

	var signals []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Signal 1: Known paywall CSS classes
			class := dom.Attr(n, "class")
			for _, pc := range paywallClasses {
				if strings.Contains(class, pc) {
					signals = append(signals, fmt.Sprintf("paywall_class:%s", pc))
				}
			}

			// Signal 2: Subscription overlay/modal IDs
			id := dom.Attr(n, "id")
			for _, pid := range paywallIDs {
				if id == pid {
					signals = append(signals, fmt.Sprintf("paywall_id:%s", pid))
				}
			}

			// Signal 3: Subscription text in paragraphs
			if n.Data == "p" {
				text := extractText(n)
				lower := strings.ToLower(text)
				for _, phrase := range paywallPhrases {
					if strings.Contains(lower, phrase) {
						signals = append(signals, "paywall_text_detected")
						return
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return len(signals) > 0, signals
}

var paywallClasses = []string{
	"paywall",
	"subscription-wall",
	"premium-content",
	"registered-only",
	"subscriber-content",
}

var paywallIDs = []string{
	"paywall-cta",
	"subscription-prompt",
	"login-wall",
	"premium-modal",
}

var paywallPhrases = []string{
	"subscribe to continue",
	"premium content",
	"this article is exclusively",
}

func extractText(n *html.Node) string {
	var buf strings.Builder
	var walk func(*html.Node)
	walk = func(c *html.Node) {
		if c.Type == html.TextNode {
			buf.WriteString(c.Data)
		}
		for ch := c.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(n)
	return buf.String()
}
