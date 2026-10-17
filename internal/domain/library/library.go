// Package library は通知対象の YouTube プレイリストの中身（アイテムと動画）を
// 持つ。ギルドごとの設定である通知設定とは違い、中身は YouTube の
// プレイリストごとに 1 つだけ保存し、全てのギルドで共有する。
package library

// PlaylistID は YouTube のプレイリストの ID。
type PlaylistID string

// VideoID は YouTube の動画の ID。
type VideoID string

// ItemID はプレイリストのアイテムの ID で、その動画の ID とは異なる。
type ItemID string
