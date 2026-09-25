package view

import (
	"strings"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// Target is where a text is shown, which changes how links are written.
type Target int

const (
	// A normal message. Discord adds a preview for each link unless the URL
	// is wrapped in <>, so links are written that way.
	InMessage Target = iota
	// An embed description, where Discord never adds previews.
	InEmbed
)

// Brackets in a title would break the Markdown link, so they are replaced
// with full-width ones.
var linkTitleReplacer = strings.NewReplacer("[", "［", "]", "］")

// VideoLink formats the video title as a Markdown link to the video.
func VideoLink(v *library.Video, target Target) string {
	url := "https://www.youtube.com/watch?v=" + string(v.YoutubeID)
	if target == InMessage {
		url = "<" + url + ">"
	}

	return "[" + linkTitleReplacer.Replace(v.Title) + "](" + url + ")"
}
