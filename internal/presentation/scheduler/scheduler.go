// Package scheduler は定期実行するユースケースを動かす。
package scheduler

import (
	"errors"
	"log"
	"time"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"

	"github.com/go-co-op/gocron/v2"
)

type scheduler struct {
	scheduler    gocron.Scheduler
	notification *application.NotificationService
	library      *application.LibraryService
}

func NewScheduler(sdr gocron.Scheduler, n *application.NotificationService, l *application.LibraryService) *scheduler {
	return &scheduler{sdr, n, l}
}

func (s *scheduler) Start() error {
	// ジョブの登録に失敗したら（cron の書式の誤りなど）、起動させずにまとめて返す
	var errs []error

	add := func(def gocron.JobDefinition, name string, job func(), opts ...gocron.JobOption) {
		_, err := s.scheduler.NewJob(def, gocron.NewTask(guard(name, job)), opts...)
		errs = append(errs, err)
	}
	immediately := gocron.WithStartAt(gocron.WithStartImmediately())

	add(gocron.DurationJob(5*time.Minute), "Notify", s.notification.NotifyUpdates, immediately)
	add(gocron.DurationJob(6*time.Hour), "RefreshVideos", s.notification.RefreshVideos)
	add(gocron.CronJob(WRAPPED_CRON, false), "Wrapped", s.notification.PostWrapped)
	add(gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(PICK_HOUR, 0, 0))), "Pick", s.notification.PostPicks)
	// 新しい DB でも名前がそろうように、起動直後にも実行する
	add(gocron.DurationJob(24*time.Hour), "RefreshCategories", s.library.RefreshCategories, immediately)

	if err := errors.Join(errs...); err != nil {
		return err
	}

	s.scheduler.Start()

	log.Println("Scheduler started")

	return nil
}

func (s *scheduler) Stop() {
	if err := s.scheduler.Shutdown(); err != nil {
		log.Println("Could not stop the scheduler cause:", err)

		return
	}

	log.Println("Scheduler stopped")
}
