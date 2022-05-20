package infrastructure

import (
	"discord-playlist-notifier/src/interfaces"
	"fmt"
)

func Dispatch(redisHandler interfaces.RedisHandler, youtubeHandler interfaces.YouTubeHandler) {
	controller := interfaces.NewPlaylistController(redisHandler, youtubeHandler)

	fmt.Println("start...")

	id := "PLyjGgL6vbeJwv887F0aVWKVr47GknPzuU"
	controller.Register(id)
	controller.PlaylistInteractor.PlaylistRepository.FindById(id)
	controller.Delete(id)

	fmt.Println("end!")
}
