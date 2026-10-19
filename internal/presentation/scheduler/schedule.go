package scheduler

import (
	"log"
	"runtime/debug"
)

// ピックを投稿する時（スケジューラのロケーションでの時刻）。0 分 0 秒に投稿する。
// 毎週のピックは月曜日、毎月のピックは 1 日に投稿する。
const PICK_HOUR = 12

// 年間のまとめを投稿する cron 式（スケジューラのロケーション）: 12/31 21:00
const WRAPPED_CRON = "0 21 31 12 *"

// guard はジョブの 1 回の実行での panic でボット全体が落ちないようにする。
func guard(name string, job func()) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in %s: %v\n%s", name, r, debug.Stack())
			}
		}()

		job()
	}
}
