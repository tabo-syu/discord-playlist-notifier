package notifier

import (
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/subscription"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/view"

	"github.com/bwmarrin/discordgo"
)

// Discord rejects embed descriptions longer than this
const MAX_EMBED_DESCRIPTION_LENGTH = 4096

func (n *notifier) PostWrapped(sub *subscription.Subscription, w *library.Wrapped) error {
	embed := &discordgo.MessageEmbed{
		Color:       color("9b59b6"),
		Title:       view.WrappedTitle(sub.Title, w.Year),
		URL:         "https://www.youtube.com/playlist?list=" + string(sub.YoutubeID),
		Description: view.Truncate(view.Wrapped(w, n.location, view.InEmbed), MAX_EMBED_DESCRIPTION_LENGTH),
	}

	return n.send(sub.SendChannelID, embed)
}
