package main

import (
	"context"
	"discord-playlist-notifier/src/infrastructure"
	"fmt"
)

func main() {
	ctx := context.Background()
	redisHandler, err := infrastructure.NewRedisHandler(ctx)
	if err != nil {
		fmt.Println("redis error")
	}
	youtubeHandler, err := infrastructure.NewYouTubeHandler(ctx)
	if err != nil {
		fmt.Println("youtube error")
	}

	infrastructure.Dispatch(redisHandler, youtubeHandler)
}
