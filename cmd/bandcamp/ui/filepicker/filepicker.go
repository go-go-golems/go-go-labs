package filepicker

import (
	tea_fp "github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
)

type Model struct {
	fp    tea_fp.Model
	Title string
	Help  help.Model
}
