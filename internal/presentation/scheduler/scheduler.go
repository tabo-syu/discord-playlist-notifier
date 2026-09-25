// Package scheduler runs the scheduled use cases.
package scheduler

import (
	"log"

	"github.com/tabo-syu/discord-playlist-notifier/internal/application"

	"github.com/go-co-op/gocron"
)

type scheduler struct {
	scheduler    *gocron.Scheduler
	notification *application.NotificationService
}

func NewScheduler(sdr *gocron.Scheduler, n *application.NotificationService) *scheduler {
	return &scheduler{sdr, n}
}

func (s *scheduler) Start() {
	s.scheduler.Every(5).Minutes().Do(guard("Notify", s.notification.NotifyUpdates))
	s.scheduler.Every(6).Hours().WaitForSchedule().Do(guard("RefreshVideos", s.notification.RefreshVideos))
	s.scheduler.Cron(WRAPPED_CRON).Do(guard("Wrapped", s.notification.PostWrapped))
	s.scheduler.Every(1).Day().At(PICK_TIME).Do(guard("Pick", s.notification.PostPicks))
	s.scheduler.StartAsync()

	log.Println("Scheduler started")
}

func (s *scheduler) Stop() {
	s.scheduler.Stop()

	log.Println("Scheduler stopped")
}
