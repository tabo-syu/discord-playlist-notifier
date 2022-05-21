package interfaces

import (
	"discord-playlist-notifier/src/usecases"
	"fmt"
)

type PlaylistController struct {
	PlaylistInteractor usecases.PlaylistInteractor
}

func NewPlaylistController(redisHandler RedisHandler, youtubeHandler YouTubeHandler) *PlaylistController {
	return &PlaylistController{
		PlaylistInteractor: usecases.PlaylistInteractor{
			PlaylistRepository: &PlaylistRepository{
				RedisHandler:   redisHandler,
				YouTubeHandler: youtubeHandler,
			},
		},
	}
}

func (c *PlaylistController) Register(playlistId string) {
	fmt.Println("Registering...")
	playlist, err := c.PlaylistInteractor.Register(playlistId)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("==============")
		fmt.Printf("%#v\n", playlist)
		fmt.Println("==============")
		fmt.Println("Registered!")
	}
}

func (c *PlaylistController) AddedVideosSince(playlistId string) {
	fmt.Println("Detecting...")
	videos, err := c.PlaylistInteractor.AddedVideosSince(playlistId)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("==============")
		fmt.Println(videos)
		fmt.Println("==============")
		fmt.Println("Detected!")
	}
}

func (c *PlaylistController) Delete(playlistId string) {
	fmt.Println("deleting...")
	err := c.PlaylistInteractor.Delete(playlistId)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("deleted!")
	}
}
