# CLAUDE.md

Guidance for AI assistants working in this repository.

## What this project is

A Discord bot (Go) that polls YouTube playlists every 5 minutes and posts an
embed to a Discord text channel whenever a new video is added to a watched
playlist. State lives in PostgreSQL via GORM. Configuration is per-guild via
the `/playlist-notifier` slash command.

Module path: `github.com/tabo-syu/discord-playlist-notifier`

## Commands

```bash
# Build / static checks (no test suite exists — see "Testing")
go build ./...
go vet ./...
gofmt -l .          # must print nothing; see "Known quirks"

# Run the whole stack (bot + postgres). This is the normal path.
cp .env.example .env   # then fill in real values
docker compose up -d
docker compose logs -f bot

# Run the bot directly. Requires a reachable Postgres and all env vars
# exported in the shell — the app does NOT read a .env file (see "Configuration").
go run cmd/server/main.go
```

There is no Makefile, no CI workflow, and no golangci-lint config in the repo.
The devcontainer sets `go.lintTool: golangci-lint` in VS Code settings only, so
lint is editor-side and not enforced anywhere.

## Architecture

Strict one-way layering; nothing lower ever imports something higher.

```
cmd/server/main.go            composition root — constructs and wires everything
        │
        ├── internal/server/          Discord gateway: session, events, slash commands
        │       └── command/          command definitions + handlers
        ├── internal/scheduler/       5-minute polling loop + Discord embed rendering
        │
        ├── internal/service/         business logic (orchestrates repositories)
        ├── internal/repository/      data access: Postgres (GORM) and YouTube Data API v3
        ├── internal/domain/          GORM models + sentinel errors
        └── internal/env/             process env vars
```

Two independent entry paths drive the app:

1. **Interactive** — Discord gateway events → `internal/server` → `service` → `repository`.
2. **Scheduled** — gocron every 5 min → `scheduler.schedule.Notify` → `service` → `repository` → `scheduler.renderer` posts to Discord.

### Wiring

`cmd/server/main.go` is the only place dependencies are constructed. `init()`
builds the four package-level singletons (`sr` gocron, `db` GORM, `dc` discordgo,
`yt` YouTube service) and `log.Fatalf`s on any failure; `main()` then assembles
repositories → services → server/scheduler and blocks on `SIGINT`. Adding a new
dependency means adding it here — there is no DI container or config framework.

### `internal/server`

- `server.go` — owns the discordgo session and registers three handlers:
  `GuildCreate` (register slash commands + create the guild row), `GuildDelete`
  (delete the guild row), `InteractionCreate` (route to a command handler).
- `registrar.go` — registers commands **per guild** (`ApplicationCommandCreate`
  with a guild ID, not globally) and remembers them in memory so `Stop()` can
  delete them. Consequence: commands only exist while the bot is running, and
  they are re-created on every `GuildCreate` (i.e. on every reconnect).
- `router.go` — maps top-level command name → `command.HandleType`.
- `event.go` — thin adapter from gateway events to `GuildService`.

### `internal/server/command`

`Command` is the interface every command implements (`Handle`, `GetCommand`,
`SetCommand`). Handlers have the signature:

```go
func(request *discordgo.ApplicationCommandInteractionData, guildId, channelId string) string
```

**Handlers return a plain string and never touch the Discord session.**
`server.go` wraps the returned string in an `InteractionResponse`. Keep it that
way — it is what makes handlers trivially testable.

`playlist_notifier/` shows the subcommand layout: `playlist_notifier.go` declares
the top-level command and dispatches on `data.Options[0].Name`; each subcommand
gets its own file holding both its `*discordgo.ApplicationCommandOption` var and
its method on `*PlaylistNotifier` (`add.go`, `list.go`, `delete.go`, `source.go`).

### `internal/scheduler`

- `scheduler.go` — the 5-minute cadence (`Every(5).Minutes()`), passing the
  scheduler's `*time.Location` into the job.
- `schedule.go` — `Notify`: load all playlists → diff against YouTube → bump
  `UpdatedAt` → render. Note the order: the timestamp is persisted **before**
  the message is sent, and only playlists whose update succeeded are notified.
  This trades a possible missed notification for never double-notifying.
- `renderer.go` — builds `discordgo.MessageEmbed`s (Japanese labels) and calls
  `ChannelMessageSendEmbeds`.

### How "new video" is detected

There is no per-video diffing against stored rows. `Playlist.UpdatedAt` is the
watermark: `PlaylistService.GetDiffFromLatest` treats a video as new when
`!video.PublishedAt.Before(last.UpdatedAt)` (`!Before`, deliberately, so videos
published at exactly the watermark are included). `PublishedAt` is the
*playlist item's* timestamp (when it was added to the playlist), while
`OwnerPublishedAt` is the video's own upload time.

`Video` rows are persisted as a side effect: `UpdateUpdatedAt` calls
`playlist.Update` → `db.Save`, and GORM cascades the `Videos` association that
`GetDiffFromLatest` attached. Nothing reads those rows back for deduplication.

## Conventions

### Go style

- **Accept interfaces, return structs** (commit `0a14cd7`). Repository
  interfaces are declared in `internal/repository`; constructors return the
  unexported concrete type (`*playlistRepository`, `*registrar`, `*router`,
  `*scheduler`, `*schedule`, `*renderer`, `*event`). Follow this — do not export
  the struct types just to make them nameable.
- Constructors are `NewX(...)` taking already-built dependencies, assigned
  positionally in a bare composite literal (`return &Server{s, rg, e, rt}`).
- Errors are **sentinel values** in `internal/domain/errors.go`, grouped by
  source (Discord / YouTube / DB). Repositories and services return these;
  the command layer maps them to user-facing text.
- Logging is stdlib `log` with space-separated fields
  (`log.Println("Guild record created:", guildId)`). No structured logger.

### Language

- **User-facing text is Japanese** — slash command descriptions, command reply
  strings, embed field labels. Match the existing tone (e.g.
  `"エラー！システムに問題があります！"`).
- **Logs and code comments are English** (a few older Japanese comments survive
  in `playlist_repository.go`).
- `README.md` is Japanese. Commit subjects are `[type] summary` with the summary
  in Japanese or English; types seen in history: `feat`, `fix`, `chore`, `doc`,
  `refactor`, `add`, `update`.

### Adding a new slash command

1. Create a package under `internal/server/command/` implementing
   `command.Command`.
2. Append it to the `commands` slice in `cmd/server/main.go` — the router and
   registrar both derive from that slice, so nothing else needs changing.

### Adding a persisted field

Add it to the struct in `internal/domain/domain.go`. Schema changes are applied
by `db.AutoMigrate(&domain.Guild{}, &domain.Playlist{}, &domain.Video{})` in
`main.go`'s `init()`. There are **no migration files** — AutoMigrate only adds
columns, so renames and drops must be handled manually against the database.

## Configuration

`internal/env/env.go` reads every variable into a package-level `var` at
**package initialization time**. There is no dotenv loading in Go code: the
`.env` file is consumed by `docker compose` (`env_file:`), not by the binary.
Running outside Docker requires exporting the variables yourself. A missing
variable is an empty string, not an error.

| Variable | Purpose |
|---|---|
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | Postgres DSN parts (`sslmode=disable`) |
| `DB_TIMEZONE` | IANA name; used for both the DSN and the gocron scheduler location |
| `DISCORD_ACCESS_TOKEN` | read into `env.DISCORD_TOKEN` |
| `YOUTUBE_APIKEY` | read into `env.YOUTUBE_TOKEN` |

Note the two name mismatches between the env var and the Go identifier.

Never commit `.env` (gitignored) or paste real tokens into code, tests, commit
messages, or PR descriptions.

## YouTube API usage

`youtube_repository.go` batches every call at `MAX_BATCH_SIZE = 50` (playlists,
videos, channels) and paginates playlist items at `MAX_RESULTS_PER_PAGE = 50`.
Deleted videos and channels are skipped rather than erroring. Timestamps parse
with `YOUTUBE_TIMEFORMAT = "2006-01-02T15:04:05Z"`.

Quota is the main constraint: `FindPlaylistsWithVideos` fetches *every* item of
*every* registered playlist on each 5-minute tick. Be deliberate about adding
new per-video API calls — the existing per-uploader channel lookup already falls
back to a `channelMap` cache to avoid repeats.

## Testing

There are currently **no `_test.go` files** in the repository. If you add tests,
`internal/service` is the natural place to start: the repository interfaces in
`internal/repository` make services fakeable without a database or network, and
command handlers return strings so they can be asserted directly.

Verify changes with `go build ./... && go vet ./...` at minimum. Anything
touching Discord or YouTube behavior needs real credentials and a test guild —
say so explicitly rather than claiming it was verified.

## Known quirks

Pre-existing behavior; don't "fix" these incidentally, but be aware of them.

- `gofmt -l .` reports `internal/scheduler/schedule.go` (trailing whitespace).
  It has been that way since it was committed.
- Go version is inconsistent: `go.mod` says `go 1.23.1`, while `Dockerfile` and
  the devcontainer image are `1.24`.
- `PlaylistRepository.FindAll` and `FindByDiscordId` return
  `domain.ErrDBRecordNotFound` for an empty result rather than an empty slice.
  With no playlists registered anywhere, the scheduler therefore logs
  `Could not notify cause: record not found` every 5 minutes.
- Because of the above, `GuildService.Unregister` returns early when a guild has
  no playlists, so the guild row is not deleted on `GuildDelete`.
- `add.go` and `delete.go` compare errors with `switch err { case ... }` (exact
  equality) while `list.go` uses `errors.Is`. Wrapped errors will fall through to
  the `default` branch in the former.
- `PlaylistRepository.DeleteAll` soft-deletes playlists (`gorm.Model`) but
  hard-deletes videos via `Unscoped()`.
- `renderer.RenderUpdatedVideo` sends all new videos as embeds in a single
  message; Discord caps a message at 10 embeds, and the code does not chunk.
- The Dockerfile runs `go run` (no compiled binary, no multi-stage build) and
  `docker-compose.yml` bind-mounts the source into the container.
