// Package dom provides DOM manipulation utilities for HTML nodes.
package dom

import "golang.org/x/net/html"

// Attr returns the value of the named attribute on an HTML node.
// Returns an empty string if the attribute is not found.
func Attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
