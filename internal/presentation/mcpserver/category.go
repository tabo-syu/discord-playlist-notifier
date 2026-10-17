package mcpserver

import (
	"slices"
	"strings"

	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
)

// Names of the YouTube video categories that can be assigned to videos. The
// IDs are fixed by YouTube, so they are listed here instead of being fetched
// with videoCategories.list.
var categoryNames = map[library.CategoryID]string{
	"1":  "映画とアニメ",
	"2":  "自動車と乗り物",
	"10": "音楽",
	"15": "ペットと動物",
	"17": "スポーツ",
	"19": "旅行とイベント",
	"20": "ゲーム",
	"22": "ブログ",
	"23": "コメディー",
	"24": "エンターテイメント",
	"25": "ニュースと政治",
	"26": "ハウツーとスタイル",
	"27": "教育",
	"28": "科学と技術",
	"29": "非営利団体と社会活動",
}

// categoryByName returns the ID of the category with the given name.
func categoryByName(name string) (library.CategoryID, bool) {
	for id, n := range categoryNames {
		if n == name {
			return id, true
		}
	}

	return "", false
}

// categoryList lists the category names for error messages.
func categoryList() string {
	var names []string
	for _, n := range categoryNames {
		names = append(names, n)
	}
	slices.Sort(names)

	return strings.Join(names, "、")
}
