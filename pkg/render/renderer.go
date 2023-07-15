package render

import (
	"github.com/mattn/go-mastodon"
	"io"
)

type Renderer interface {
	RenderStatus(w io.Writer, status *mastodon.Status) error
	RenderThread(w io.Writer, status *mastodon.Status, context *mastodon.Context) error
}
