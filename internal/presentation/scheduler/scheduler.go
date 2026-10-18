// Package scheduler は定期実行するユースケースを動かす。
package scheduler

import (
	"errors"
	"log"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"

	"github.com/go-co-op/gocron"
)

type scheduler struct {
	scheduler    *gocron.Scheduler
	notification *application.NotificationService
	library      *application.LibraryService
}

func NewScheduler(sdr *gocron.Scheduler, n *application.NotificationService, l *application.LibraryService) *scheduler {
	return &scheduler{sdr, n, l}
}

func (s *scheduler) Start() error {
	// ジョブの登録に失敗したら（cron の書式の誤りなど）、起動させずにまとめて返す
	var errs []error

	add := func(_ *gocron.Job, err error) { errs = append(errs, err) }
	add(s.scheduler.Every(5).Minutes().Do(guard("Notify", s.notification.NotifyUpdates)))
	add(s.scheduler.Every(6).Hours().WaitForSchedule().Do(guard("RefreshVideos", s.notification.RefreshVideos)))
	add(s.scheduler.Cron(WRAPPED_CRON).Do(guard("Wrapped", s.notification.PostWrapped)))
	add(s.scheduler.Every(1).Day().At(PICK_TIME).Do(guard("Pick", s.notification.PostPicks)))
	// 新しい DB でも名前がそろうように、起動直後にも実行する
	add(s.scheduler.Every(1).Day().Do(guard("RefreshCategories", s.library.RefreshCategories)))

	if err := errors.Join(errs...); err != nil {
		return err
	}

	s.scheduler.StartAsync()

	log.Println("Scheduler started")

	return nil
}

func (s *scheduler) Stop() {
	s.scheduler.Stop()

	log.Println("Scheduler stopped")
}
