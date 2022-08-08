package scheduler

import (
	"discord-playlist-notifier/handler/schedule"
	"log"

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
	s.scheduler.Every(1).Day().At("18:00").Do(s.schedule.Notify, s.scheduler.Location())
	s.scheduler.StartAsync()

	log.Println("Scheduler started")
}

func (s *scheduler) Stop() {
	s.scheduler.Stop()

	log.Println("Scheduler stopped")
}
