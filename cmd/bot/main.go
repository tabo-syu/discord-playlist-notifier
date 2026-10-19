package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"
	"github.com/tabo-syu/discord-playlist-notifier/internal/domain/library"
	"github.com/tabo-syu/discord-playlist-notifier/internal/env"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/persistence"
	"github.com/tabo-syu/discord-playlist-notifier/internal/infrastructure/youtube"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/discord/command/playlist_notifier"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/notifier"
	"github.com/tabo-syu/discord-playlist-notifier/internal/presentation/scheduler"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"google.golang.org/api/option"
	ytapi "google.golang.org/api/youtube/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	sr *gocron.Scheduler
	db *gorm.DB
	dc *discordgo.Session
	yt *ytapi.Service
)

func init() {
	var err error

	location, err := time.LoadLocation(env.DB_TIMEZONE)
	if err != nil {
		log.Fatalf("Could not load time location: %v", err)
	}

	sr = gocron.NewScheduler(location)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable Timezone=%s",
		env.DB_HOST,
		env.DB_USER,
		env.DB_PASSWORD,
		env.DB_NAME,
		env.DB_PORT,
		env.DB_TIMEZONE,
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect the db: %v", err)
	}

	dc, err = discordgo.New("Bot " + env.DISCORD_TOKEN)
	if err != nil {
		log.Fatalf("Invalid discord token: %v", err)
	}

	yt, err = ytapi.NewService(context.Background(), option.WithAPIKey(env.YOUTUBE_TOKEN))
	if err != nil {
		log.Fatalf("Invalid youtube token: %v", err)
	}
}

func main() {
	location := sr.Location()

	yc := youtube.NewClient(yt)
	gr := persistence.NewGuildRepository(db)
	sbr := persistence.NewSubscriptionRepository(db)
	pr := persistence.NewPlaylistRepository(db)
	vr := persistence.NewVideoRepository(db)
	cr := persistence.NewChannelRepository(db)
	car := persistence.NewCategoryRepository(db)
	vsr := persistence.NewVideoSnapshotRepository(db)

	// アプリケーションサービスは互いを呼ばず、読み取りの共通部分はこのドメインサービスを共有する
	catalog := library.NewCatalog(pr, vr)
	events := application.NewEventPublisher()

	ss := application.NewSubscriptionService(yc, sbr, gr, time.Now)
	gs := application.NewGuildService(gr, sbr)
	ls := application.NewLibraryService(yc, sbr, pr, vr, cr, car, vsr, catalog, events, location, time.Now)
	nt := notifier.New(dc, location)
	ns := application.NewNotificationService(sbr, catalog, yc, nt, events, location, time.Now)

	// ドメインイベントのハンドラ。通知の種類を足すときは、ここに登録を足す
	nh := application.NewNotificationHandlers(sbr, catalog, yc, nt)
	application.Subscribe(events, nh.NotifyHiddenVideos)
	application.Subscribe(events, nh.NotifyMilestones)
	application.Subscribe(events, nh.NotifyNewVideos)
	// 同期の結果から新着を選び、通知設定を進める（VideosAdded を発行する）
	application.Subscribe(events, ns.TakeNewVideos)

	commands := []command.Command{playlist_notifier.NewPlaylistNotifier(ss, ls, location, time.Now)}

	server := discord.NewServer(
		dc,
		discord.NewRegisterer(dc, commands),
		discord.NewEvent(gs),
		discord.NewRouter(commands),
	)
	if err := server.Serve(); err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	defer func() {
		if err := server.Stop(); err != nil {
			log.Println("Could not stop the session cause:", err)
		}
	}()

	scheduler := scheduler.NewScheduler(sr, ns, ls)
	if err := scheduler.Start(); err != nil {
		// log.Fatalf では defer の server.Stop が動かず、コマンドの登録が残るので return する
		log.Println("Cannot start the scheduler:", err)

		return
	}
	defer scheduler.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
}
