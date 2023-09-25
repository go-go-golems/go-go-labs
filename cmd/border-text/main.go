package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"golang.org/x/term"
	"os"
)

func drawBorderedMessage(msg string) string {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}

	width = width / 2

	style := lipgloss.NewStyle().
		Padding(1, 1).
		Border(lipgloss.RoundedBorder())
	//Width(width - 4)

	w := wordwrap.NewWriter(width - 4)
	_, err = w.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
	return style.Render(w.String())
}

const veryLongLoremIpsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec a diam lectus. " +
	"Sed sit amet ipsum mauris. Maecenas congue ligula ac quam viverra nec consectetur ante hendrerit. " +
	"Donec et mollis dolor. Praesent et diam eget libero egestas mattis sit amet vitae augue. " +
	"Nam tincidunt congue enim, ut porta lorem lacinia consectetur. " +
	"Donec ut libero sed arcu vehicula ultricies a non tortor. Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
	"Aenean ut gravida lorem. Ut turpis felis, pulvinar a semper sed, adipiscing id dolor. " +
	"Pellentesque auctor nisi id magna consequat sagittis. " +
	"Curabitur dapibus enim sit amet elit pharetra tincidunt feugiat nisl imperdiet. " +
	"Ut convallis libero in urna ultrices accumsan. Donec sed odio eros."

func main() {
	v := drawBorderedMessage(veryLongLoremIpsum)
	fmt.Println(v)
}
