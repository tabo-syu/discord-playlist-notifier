package schedule

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Schedule interface {
	Notify()
}

type schedule struct {
	session *discordgo.Session
}

func NewSchedule(s *discordgo.Session) Schedule {
	return &schedule{s}
}

func (s *schedule) Notify() {
	log.Println("Scheduled!")
}
