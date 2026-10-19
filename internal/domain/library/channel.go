package library

import "strings"

// ChannelID は YouTube がチャンネルに割り当てる ID。
type ChannelID string

// ParseChannelID は人が入力したチャンネルの ID を読み取る。前後の空白は
// 取り除く。ID に使えない文字があれば ErrInvalidChannelID を返す。
func ParseChannelID(s string) (ChannelID, error) {
	s = strings.TrimSpace(s)
	if !idCharsPattern.MatchString(s) {
		return "", ErrInvalidChannelID
	}

	return ChannelID(s), nil
}

// Channel は通知対象のプレイリストにある動画を投稿した YouTube のチャンネル。
// 動画は ID で参照する。アイコンは保存せず、通知を送る直前に取得する。
type Channel struct {
	YoutubeID ChannelID
	// そのチャンネルの動画のいずれかを最後に更新したときの名前
	Title string
}

// ChannelsOf は取得した動画の詳細から、それを投稿したチャンネルを重複なく、最初に
// 現れた順に返す。同じチャンネルの動画が複数あれば、後の動画の名前を使う。
// チャンネルがわからない動画は除く。
func ChannelsOf(details []*VideoDetails) []*Channel {
	var channels []*Channel

	byID := map[ChannelID]*Channel{}

	for _, d := range details {
		if d.ChannelID == "" {
			continue
		}

		if ch, ok := byID[d.ChannelID]; ok {
			ch.Title = d.ChannelTitle

			continue
		}

		ch := &Channel{YoutubeID: d.ChannelID, Title: d.ChannelTitle}
		byID[d.ChannelID] = ch
		channels = append(channels, ch)
	}

	return channels
}
