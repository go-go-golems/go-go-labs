package doc

import (
	"embed"
)

//go:embed topics/*.md tutorials/*.md examples/*.md
var Files embed.FS
