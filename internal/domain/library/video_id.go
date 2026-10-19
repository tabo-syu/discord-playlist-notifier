package library

import (
	"net/url"
	"regexp"
	"strings"
)

var videoIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// ParseVideoID は人が入力した動画 ID か、次のような動画の URL を読み取る。
// https://www.youtube.com/watch?v=ID、https://youtu.be/ID、
// https://www.youtube.com/shorts/ID
// どちらとも読めなければ ErrInvalidVideoID を返す。
func ParseVideoID(s string) (VideoID, error) {
	s = strings.TrimSpace(s)
	if videoIDPattern.MatchString(s) {
		return VideoID(s), nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return "", ErrInvalidVideoID
	}

	candidate := u.Query().Get("v")
	if candidate == "" {
		segments := strings.Split(strings.Trim(u.Path, "/"), "/")
		candidate = segments[len(segments)-1]
	}

	if !videoIDPattern.MatchString(candidate) {
		return "", ErrInvalidVideoID
	}

	return VideoID(candidate), nil
}
