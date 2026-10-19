#!/usr/bin/env bash
# SessionStart の hook。clone したばかりの環境（クラウドのセッションなど）でも、
# lefthook の Git フックと Go の依存がそろった状態で作業を始められるようにする。
# 標準出力は Claude のコンテキストに入るので、失敗したときだけ書く。
cd "${CLAUDE_PROJECT_DIR:-.}" || exit 0

hooks=$(git rev-parse --git-path hooks)
if ! grep -qs lefthook "$hooks/pre-commit"; then
	go tool task hooks >/dev/null 2>&1 || echo "lefthook の Git フックを入れられませんでした。go tool task hooks を実行してください。"
fi

go mod download >/dev/null 2>&1 || echo "go mod download に失敗しました。"

exit 0
