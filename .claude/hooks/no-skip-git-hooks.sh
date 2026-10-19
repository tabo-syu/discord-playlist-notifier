#!/usr/bin/env bash
# PreToolUse の hook。lefthook の Git フック（コミット前の整形・tidy・lint、push 前のテスト）を
# 飛ばすコマンドを Claude が実行しようとしたら止める。フックが検証の最後の砦なので、
# 通らないときは飛ばさずに原因を直させる。
command=$(jq -r '.tool_input.command // empty')

if echo "$command" | grep -qE -- '--no-verify|LEFTHOOK=(0|false)|LEFTHOOK_EXCLUDE=|core\.hooksPath|\bcommit\b[^|;&]*[[:space:]]-[A-Za-z]*n[A-Za-z]*([[:space:]]|$)'; then
	echo "lefthook の Git フックを飛ばすコマンドは実行できません。フックが通らない原因を直してください。" >&2
	exit 2
fi

exit 0
