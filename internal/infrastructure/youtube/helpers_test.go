package youtube_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	ytapi "google.golang.org/api/youtube/v3"
)

// TestMain はテストの出力を読みやすくするため、ログを捨てる。
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

// response は偽の API が返すレスポンス。
type response struct {
	status int
	body   string
}

// ok は JSON を返すレスポンス。
func ok(body string) response {
	return response{http.StatusOK, body}
}

// failure は API のエラーを返すレスポンス。
func failure() response {
	return response{http.StatusInternalServerError, `{"error": {"code": 500, "message": "backend error"}}`}
}

// api は YouTube Data API の偽のサーバー。エンドポイント（/playlists など）ごとに、
// 呼ばれた順に responses を返し、受けたリクエストを記録する。
type api struct {
	mu        sync.Mutex
	responses map[string][]response
	requests  []string
}

func (a *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	endpoint := strings.TrimPrefix(r.URL.Path, "/youtube/v3")
	a.requests = append(a.requests, request(endpoint, r))

	queue := a.responses[endpoint]
	if len(queue) == 0 {
		http.Error(w, "unexpected request: "+endpoint, http.StatusNotImplemented)

		return
	}

	a.responses[endpoint] = queue[1:]

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(queue[0].status)
	fmt.Fprint(w, queue[0].body)
}

// request はリクエストを「/エンドポイント キー=値 ...」の形にする。クエリはキーの順に並べ、
// クライアントライブラリが付ける alt と prettyPrint は除く。
func request(endpoint string, r *http.Request) string {
	query := r.URL.Query()

	var keys []string

	for k := range query {
		if k != "alt" && k != "prettyPrint" {
			keys = append(keys, k)
		}
	}

	slices.Sort(keys)

	parts := make([]string, 0, 1+len(keys))

	parts = append(parts, endpoint)
	for _, k := range keys {
		parts = append(parts, k+"="+strings.Join(query[k], ","))
	}

	return strings.Join(parts, " ")
}

// serve は偽の API を立て、そこに接続する YouTube のサービスを返す。
func serve(t *testing.T, responses map[string][]response) (*ytapi.Service, *api) {
	t.Helper()

	a := &api{responses: responses}
	ts := httptest.NewServer(a)
	t.Cleanup(ts.Close)

	yt, err := ytapi.NewService(context.Background(), option.WithEndpoint(ts.URL+"/"), option.WithoutAuthentication())
	require.NoError(t, err)

	return yt, a
}

// numbered は prefix に 01 から n までの番号を付けた ID を返す。
func numbered[T ~string](prefix string, n int) []T {
	var ids []T
	for i := 1; i <= n; i++ {
		ids = append(ids, T(fmt.Sprintf("%s%02d", prefix, i)))
	}

	return ids
}

// joined は ID をカンマでつなぐ。
func joined[T ~string](ids []T) string {
	s := make([]string, 0, len(ids))
	for _, id := range ids {
		s = append(s, string(id))
	}

	return strings.Join(s, ",")
}

// message はエラーの文言を返す。エラーがなければ空文字列。
func message(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
