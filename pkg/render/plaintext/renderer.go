package plaintext

import (
	"bytes"
	"fmt"
	"github.com/mattn/go-mastodon"
	"golang.org/x/net/html"
	"io"
	"strings"
)

func convertHTMLToPlainText(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	var f func(*html.Node)
	var output strings.Builder
	blockTags := map[string]struct{}{
		"p":       {},
		"div":     {},
		"br":      {},
		"article": {},
		"section": {},
		"li":      {},
	}

	f = func(n *html.Node) {
		// If the current node is a text node
		if n.Type == html.TextNode {
			output.WriteString(n.Data)
		}

		// If the current node is one of the block tags types
		if _, present := blockTags[n.Data]; present {
			output.WriteRune('\n')
		}

		// Recurse on the children nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	return output.String()
}

type Renderer struct {
	Verbose    bool
	WithHeader bool
	isFirst    bool
}

type RenderOption func(*Renderer)

func WithHeader(withHeader bool) RenderOption {
	return func(r *Renderer) {
		r.WithHeader = withHeader
	}
}

func WithVerbose(verbose bool) RenderOption {
	return func(r *Renderer) {
		r.Verbose = verbose
	}
}

func NewRenderer(opts ...RenderOption) *Renderer {
	r := &Renderer{
		isFirst: true,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Renderer) RenderHeader(w io.Writer, status *mastodon.Status, isFirst bool) error {
	if !r.WithHeader {
		return nil
	}

	if !r.Verbose && !r.isFirst {
		return nil
	}

	var buffer bytes.Buffer

	if r.Verbose {
		buffer.WriteString(fmt.Sprintf("ID: %s\n", status.ID))
	}

	buffer.WriteString(fmt.Sprintf("Author: %s (%v)\n", status.Account.Acct, status.CreatedAt))
	buffer.WriteString(fmt.Sprintf("Replies/Reblogs/Favourites: %d/%d/%d\n",
		status.RepliesCount, status.ReblogsCount, status.FavouritesCount))

	header := buffer.String()

	if isFirst {
		header += fmt.Sprintf("URL: %s\nAuthor URL: %s\n", status.URL, status.Account.URL)
	}

	_, err := w.Write([]byte(header))
	if err != nil {
		return err
	}

	return nil
}

func (r *Renderer) RenderStatus(w io.Writer, status *mastodon.Status) error {
	var content string
	if r.WithHeader {
		if err := r.RenderHeader(w, status, r.isFirst); err != nil {
			return err
		}
		r.isFirst = false
	}

	if r.Verbose {
		content = fmt.Sprintf("Status ID: %s\nCreated at: %v\nContent: %s\n-----------------\n",
			status.ID, status.CreatedAt, convertHTMLToPlainText(status.Content))
	} else {
		content = fmt.Sprintf("%s\n", convertHTMLToPlainText(status.Content))
	}
	_, _ = w.Write([]byte(content))

	if len(status.MediaAttachments) > 0 {
		_, _ = w.Write([]byte("Attachments:\n"))
		for _, attachment := range status.MediaAttachments {
			if _, err := w.Write([]byte(fmt.Sprintf("%s\n", attachment.URL))); err != nil {
				return err
			}
		}
	}

	_, _ = w.Write([]byte("\n"))

	return nil
}

func (r *Renderer) RenderThread(w io.Writer, status *mastodon.Status, context *mastodon.Context) error {
	for _, ancestor := range context.Ancestors {
		if r.Verbose {
			if _, err := w.Write([]byte("--AN--\n")); err != nil {
				return err
			}
		}
		if err := r.RenderStatus(w, ancestor); err != nil {
			return err
		}
	}

	if r.Verbose {
		if _, err := w.Write([]byte("--OR--\n")); err != nil {
			return err
		}
	}
	if err := r.RenderStatus(w, status); err != nil {
		return err
	}

	for _, descendant := range context.Descendants {
		if r.Verbose {
			if _, err := w.Write([]byte("--DE--\n")); err != nil {
				return err
			}
		}
		if err := r.RenderStatus(w, descendant); err != nil {
			return err
		}
	}

	return nil
}
