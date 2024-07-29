package scheduler

import (
	"log"

	"github.com/tabo-syu/discord-playlist-notifier/internal/handler/schedule"

	"github.com/go-co-op/gocron"
)

type Scheduler interface {
	Start()
	Stop()
}

type scheduler struct {
	scheduler *gocron.Scheduler
	schedule  schedule.Schedule
}

func NewScheduler(sdr *gocron.Scheduler, sdl schedule.Schedule) Scheduler {
	return &scheduler{sdr, sdl}
}

func (s *scheduler) Start() {
	s.scheduler.Every(5).Minutes().Do(s.schedule.Notify, s.scheduler.Location())
	s.scheduler.StartAsync()

	log.Println("Scheduler started")
}

func (s *scheduler) Stop() {
	s.scheduler.Stop()

	log.Println("Scheduler stopped")
}
