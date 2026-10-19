#!/usr/bin/env bash
# Stop の hook。main から Go のコードなどを変えていたら、Claude が終わる前に
# go tool task check を実行し、通らなければ終わらせずに直させる。
#
# - 一度通った状態（HEAD と、コミットしていない変更）は記録し、変わるまで実行し直さない
# - 失敗した状態から何も変えずにもう一度終わろうとしたら、直すのを諦めたとみなして止めない
# - Docker が使えないときは、DB のテストを飛ばして確かめる（CLAUDE.md の「コマンド」と同じ）
cd "${CLAUDE_PROJECT_DIR:-.}" || exit 0

base=$(git merge-base HEAD origin/main 2>/dev/null || git rev-parse HEAD)
changed=$(
	{
		git diff --name-only "$base"
		git ls-files --others --exclude-standard
	} | grep -E '(\.go|go\.mod|go\.sum|\.golangci\.yml|\.mockery\.yml|Taskfile\.yml|/testdata/.+)$'
)
if [ -z "$changed" ]; then
	exit 0
fi

state=$(git rev-parse --git-path claude-check)
mkdir -p "$state"
fingerprint=$(
	{
		git rev-parse HEAD
		git diff HEAD --binary
		git ls-files -z --others --exclude-standard | xargs -0 -r sha1sum
	} | sha1sum | cut -d' ' -f1
)

if [ "$(cat "$state/passed" 2>/dev/null)" = "$fingerprint" ]; then
	exit 0
fi

if [ "$(cat "$state/failed" 2>/dev/null)" = "$fingerprint" ]; then
	echo '{"systemMessage": "go tool task check が通らないまま終了しました。"}'
	exit 0
fi

if docker info >/dev/null 2>&1; then
	check="go tool task check"
	note=""
else
	check="go build ./... && go tool task tidy:check && go tool task lint && go tool task test:short"
	note="Docker が使えないので、DB のテストを飛ばして確かめました。"
fi

if output=$(bash -c "$check" 2>&1); then
	echo "$fingerprint" >"$state/passed"
	if [ -n "$note" ]; then
		echo "{\"systemMessage\": \"$note\"}"
	fi
	exit 0
fi

echo "$fingerprint" >"$state/failed"
{
	echo "$check が通りません。原因を直してから終えてください。${note}"
	echo
	echo "$output" | tail -n 80
} >&2
exit 2
