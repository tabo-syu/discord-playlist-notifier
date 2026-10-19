package application

// EventPublisher は、集約を保存した後にドメインイベントをハンドラへ配る。
// 発行する側は、誰が受け取るかを知らない。ハンドラは cmd/bot/main.go が
// Subscribe で登録する。
//
// 配るのは発行した goroutine の中で、登録した順に同期的に呼ぶ。Go のメソッドは
// 型引数を持てないので、登録と発行はパッケージの関数にしている。
type EventPublisher struct {
	// 型ごとのハンドラ。要素は func([]E)
	handlers []any
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{}
}

// Subscribe は E のイベントを受け取るハンドラを登録する。ハンドラは 1 回の発行で
// 起きたイベントをまとめて受け取る。チャンネルごとの重複の排除や、YouTube への
// 問い合わせを 1 回にまとめるため。
func Subscribe[E any](p *EventPublisher, handler func(events []E)) {
	p.handlers = append(p.handlers, handler)
}

// Publish は E のイベントを、E を購読した全てのハンドラに配る。イベントがなければ
// 何もしない。イベントを起こした集約を保存した後に呼ぶ。
func Publish[E any](p *EventPublisher, events []E) {
	if len(events) == 0 {
		return
	}

	for _, h := range p.handlers {
		if handle, ok := h.(func([]E)); ok {
			handle(events)
		}
	}
}
