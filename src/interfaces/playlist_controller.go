package interfaces

import (
	"discord-playlist-notifier/src/usecases"
	"fmt"
	"time"
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
	}

	fmt.Println("==============")
	fmt.Printf("%#v\n", playlist)
	fmt.Println("==============")
	fmt.Println("Registered!")
}

func (c *PlaylistController) AddedVideosSince(playlistId string, since time.Time) {
	videos := c.PlaylistInteractor.AddedVideosSince(playlistId, since)

	fmt.Println(videos)
}

func (c *PlaylistController) Update(playlistId string) {
	playlist, err := c.PlaylistInteractor.Update(playlistId)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(playlist)
}

func (c *PlaylistController) Delete(playlistId string) {
	fmt.Println("deleting...")
	err := c.PlaylistInteractor.Delete(playlistId)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("deleted!")
	}
}
