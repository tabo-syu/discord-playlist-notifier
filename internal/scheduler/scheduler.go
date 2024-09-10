package scheduler

import (
	"log"

	"github.com/go-co-op/gocron"
)

type scheduler struct {
	scheduler *gocron.Scheduler
	schedule  *schedule
}

func NewScheduler(sdr *gocron.Scheduler, sdl *schedule) *scheduler {
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
