package library

import (
	"net/url"
	"regexp"
	"strings"
)

var videoIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// ParseVideoID reads a video ID, or a URL of the video such as
// https://www.youtube.com/watch?v=ID, https://youtu.be/ID or
// https://www.youtube.com/shorts/ID.
func ParseVideoID(s string) (VideoID, bool) {
	s = strings.TrimSpace(s)
	if videoIDPattern.MatchString(s) {
		return VideoID(s), true
	}

	u, err := url.Parse(s)
	if err != nil {
		return "", false
	}
	candidate := u.Query().Get("v")
	if candidate == "" {
		segments := strings.Split(strings.Trim(u.Path, "/"), "/")
		candidate = segments[len(segments)-1]
	}
	if !videoIDPattern.MatchString(candidate) {
		return "", false
	}

	return VideoID(candidate), true
}
