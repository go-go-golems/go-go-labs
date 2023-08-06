package plaintext

import (
	"bytes"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/mastoid/pkg"
	"github.com/mattn/go-mastodon"
	"github.com/rs/zerolog/log"
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
	Verbose         bool
	WithHeader      bool
	isFirst         bool
	previousAuthor  string
	indent          string
	firstLinePrefix string
	nextLinePrefix  string
	prefix          string
}

type RenderOption func(*Renderer)

func WithMarkdown() RenderOption {
	return func(r *Renderer) {
		r.indent = "> "
		r.firstLinePrefix = ""
		r.nextLinePrefix = ""
		r.prefix = "> "
	}
}

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
		isFirst:         true,
		indent:          "  ",
		firstLinePrefix: "+ ",
		nextLinePrefix:  "| ",
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

	var buffer bytes.Buffer

	if r.Verbose {
		buffer.WriteString(fmt.Sprintf("ID: %s\n", status.ID))
	}

	if r.previousAuthor != status.Account.Acct {
		buffer.WriteString(fmt.Sprintf("Author: %s (%v)\n", status.Account.Acct, status.CreatedAt))
		r.previousAuthor = status.Account.Acct
	} else {
		buffer.WriteString(fmt.Sprintf("Author: %s (%v)\n", status.Account.Acct, status.CreatedAt))

	}
	if r.Verbose {
		buffer.WriteString(fmt.Sprintf("Replies/Reblogs/Favourites: %d/%d/%d\n",
			status.RepliesCount, status.ReblogsCount, status.FavouritesCount))
	}

	header := buffer.String()

	if isFirst || r.Verbose {
		header += fmt.Sprintf("URL: %s\nAuthor URL: %s\n", status.URL, status.Account.URL)
	}

	_, err := w.Write([]byte(header))
	if err != nil {
		return err
	}

	return nil
}

func isImageUrl(url string) bool {
	s := strings.ToLower(url)
	return strings.HasSuffix(s, ".jpg") || strings.HasSuffix(s, ".jpeg") || strings.HasSuffix(s, ".png") ||
		strings.HasSuffix(s, ".gif") || strings.HasSuffix(s, ".webp")
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
		for i, attachment := range status.MediaAttachments {
			url := attachment.URL

			if isImageUrl(url) {
				url = fmt.Sprintf("![attachment %d](%s)", i, url)
				if _, err := w.Write([]byte(fmt.Sprintf("%s\n\n", url))); err != nil {
					return err
				}
			} else {
				if _, err := w.Write([]byte(fmt.Sprintf("[attachment %d](%s)\n", i, attachment.URL))); err != nil {
					return err
				}
			}

			// split attachment.Description by \n and add "> " to the beginning of each line
			s := strings.Split(attachment.Description, "\n")
			for i := range s {
				s[i] = "> " + s[i]
			}
			attachment.Description = "> (Image Description)\n>\n" + strings.Join(s, "\n")

			if _, err := w.Write([]byte(fmt.Sprintf("%s\n\n", attachment.Description))); err != nil {
				return err
			}
		}
	}

	_, _ = w.Write([]byte("\n"))

	return nil
}

func (r *Renderer) RenderThread(w io.Writer, status *mastodon.Status, context *mastodon.Context) error {

	thread := &pkg.Thread{
		Nodes: map[mastodon.ID]*pkg.Node{},
	}

	thread.AddStatus(status)
	thread.AddContextAndGetMissingIDs(status.ID, context)

	prevDepth := 0
	siblingIdx := 0

	printNode := func(node *pkg.Node, depth int) error {
		buf := bytes.NewBuffer(nil)
		if err := r.RenderStatus(buf, node.Status); err != nil {
			return err
		}

		s := strings.TrimSpace(buf.String())
		// prepend each line with depth * "  "
		lines := strings.Split(s, "\n")
		buf = bytes.NewBuffer(nil)
		var prefix string
		if depth <= prevDepth {
			siblingIdx++
		}
		log.Debug().
			Int("depth", depth).
			Int("prevDepth", prevDepth).
			Int("siblingIdx", siblingIdx).
			Msg("rendering")

		prefix = r.prefix + strings.Repeat(r.indent, siblingIdx-1)
		prevDepth = depth

		for i := range lines {
			if i == 0 {
				lines[i] = prefix + r.firstLinePrefix + lines[i]
			} else {
				lines[i] = prefix + r.nextLinePrefix + lines[i]
			}
			buf.WriteString(lines[i] + "\n")
		}
		buf.WriteString(prefix + "\n")

		_, err := w.Write(buf.Bytes())
		if err != nil {
			return err
		}
		return nil
	}

	err := thread.WalkDepthFirst(printNode)
	if err != nil {
		return err
	}

	return nil
}
