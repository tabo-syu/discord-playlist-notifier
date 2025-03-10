# Discord Playlist Notifier - Code Improvement Recommendations

This document provides specific code examples and recommendations for addressing the issues identified in the README-analysis.md file.

## 1. Internationalization Issues

### Example Issue:
```go
// internal/scheduler/renderer.go
embed := &discordgo.MessageEmbed{
    Color: red,
    Author: &discordgo.MessageEmbedAuthor{
        Name: playlist.Title + " に追加されました！", // Japanese text
    },
    // ...
    Fields: []*discordgo.MessageEmbedField{
        {
            Name:   "追加日時", // Japanese text
            Value:  video.PublishedAt.In(location).Format("2006/01/02 15:04:05"),
            Inline: true,
        },
        {
            Name:   "再生回数", // Japanese text
            Value:  separator(video.Views),
            Inline: true,
        },
    },
    // ...
}
```

### Recommendation:
Implement a simple internationalization system:

```go
// internal/i18n/i18n.go
package i18n

import (
    "fmt"
    "strings"
)

type Translator struct {
    locale string
    translations map[string]map[string]string
}

func NewTranslator(locale string) *Translator {
    return &Translator{
        locale: locale,
        translations: map[string]map[string]string{
            "en": {
                "playlist_added": "Added to %s!",
                "date_added": "Date Added",
                "view_count": "View Count",
                // Add more translations here
            },
            "ja": {
                "playlist_added": "%s に追加されました！",
                "date_added": "追加日時",
                "view_count": "再生回数",
                // Add more translations here
            },
        },
    }
}

func (t *Translator) T(key string, args ...interface{}) string {
    if translations, ok := t.translations[t.locale]; ok {
        if translation, ok := translations[key]; ok {
            if len(args) > 0 {
                return fmt.Sprintf(translation, args...)
            }
            return translation
        }
    }
    
    // Fallback to English
    if t.locale != "en" {
        if translations, ok := t.translations["en"]; ok {
            if translation, ok := translations[key]; ok {
                if len(args) > 0 {
                    return fmt.Sprintf(translation, args...)
                }
                return translation
            }
        }
    }
    
    return key
}
```

Then use it in the code:

```go
// internal/scheduler/renderer.go
func NewRenderer(s *discordgo.Session, translator *i18n.Translator) *renderer {
    return &renderer{s, translator}
}

func (r *renderer) RenderUpdatedVideo(playlist *domain.Playlist, location *time.Location) error {
    // ...
    embed := &discordgo.MessageEmbed{
        Color: red,
        Author: &discordgo.MessageEmbedAuthor{
            Name: r.translator.T("playlist_added", playlist.Title),
        },
        // ...
        Fields: []*discordgo.MessageEmbedField{
            {
                Name:   r.translator.T("date_added"),
                Value:  video.PublishedAt.In(location).Format("2006/01/02 15:04:05"),
                Inline: true,
            },
            {
                Name:   r.translator.T("view_count"),
                Value:  separator(video.Views),
                Inline: true,
            },
        },
        // ...
    }
    // ...
}
```

## 2. Error Handling Issues

### Example Issue:
```go
// internal/domain/errors.go
var (
    // ...
    ErrYouTubePlaylistCouldNotFound = errors.New("playlist could not found") // Grammatical error
    // ...
    ErrDBRecordCouldNotFound = errors.New("record could not found") // Grammatical error
)

// internal/repository/youtube_repository.go
publishedAt, _ := time.Parse(YOUTUBE_TIMEFORMAT, listVideo.Snippet.PublishedAt) // Ignoring error
ownerPublishedAt, _ := time.Parse(YOUTUBE_TIMEFORMAT, video.Snippet.PublishedAt) // Ignoring error
```

### Recommendation:
Fix grammatical errors and handle all errors properly:

```go
// internal/domain/errors.go
var (
    // ...
    ErrYouTubePlaylistNotFound = errors.New("playlist not found") // Fixed grammar
    // ...
    ErrDBRecordNotFound = errors.New("record not found") // Fixed grammar
)

// internal/repository/youtube_repository.go
publishedAt, err := time.Parse(YOUTUBE_TIMEFORMAT, listVideo.Snippet.PublishedAt)
if err != nil {
    return nil, fmt.Errorf("failed to parse video publish time: %w", err)
}
ownerPublishedAt, err := time.Parse(YOUTUBE_TIMEFORMAT, video.Snippet.PublishedAt)
if err != nil {
    return nil, fmt.Errorf("failed to parse video owner publish time: %w", err)
}
```

For repository functions that return errors for empty result sets:

```go
// internal/repository/playlist_repository.go
func (r *playlistRepository) FindAll() ([]*domain.Playlist, error) {
    var playlists []*domain.Playlist
    if err := r.db.Find(&playlists).Error; err != nil {
        return nil, err
    }
    
    // Return empty slice instead of error for no results
    return playlists, nil
}
```

## 3. YouTube API Limitations

### Example Issue:
```go
// internal/repository/youtube_repository.go
const (
    YOUTUBE_TIMEFORMAT = "2006-01-02T15:04:05Z"
    MAX_RESULTS        = 20 // Hard limit
)

// ...
lists, err := r.youtube.Playlists.List([]string{"id", "snippet"}).MaxResults(MAX_RESULTS).
    Id(ids...).Do() // No pagination
```

### Recommendation:
Implement pagination and handle API quota limits:

```go
// internal/repository/youtube_repository.go
const (
    YOUTUBE_TIMEFORMAT = "2006-01-02T15:04:05Z"
    MAX_RESULTS_PER_PAGE = 50 // Maximum allowed by YouTube API
)

func (r *youTubeRepository) FindPlaylistsWithVideos(ids ...string) ([]*domain.Playlist, error) {
    var response = []*domain.Playlist{}
    
    // Process playlists in batches to respect YouTube API limits
    for i := 0; i < len(ids); i += MAX_RESULTS_PER_PAGE {
        end := i + MAX_RESULTS_PER_PAGE
        if end > len(ids) {
            end = len(ids)
        }
        
        batchIds := ids[i:end]
        lists, err := r.youtube.Playlists.List([]string{"id", "snippet"}).MaxResults(int64(len(batchIds))).
            Id(batchIds...).Do()
        if err != nil {
            return nil, err
        }
        
        // Process each playlist...
        for _, playlist := range lists.Items {
            // Implement pagination for playlist items
            var allItems []*youtube.PlaylistItem
            nextPageToken := ""
            
            for {
                call := r.youtube.PlaylistItems.List([]string{"snippet"}).
                    MaxResults(MAX_RESULTS_PER_PAGE).
                    PlaylistId(playlist.Id)
                
                if nextPageToken != "" {
                    call = call.PageToken(nextPageToken)
                }
                
                playlistItems, err := call.Do()
                if err != nil {
                    return nil, err
                }
                
                allItems = append(allItems, playlistItems.Items...)
                
                nextPageToken = playlistItems.NextPageToken
                if nextPageToken == "" {
                    break
                }
            }
            
            // Continue processing with allItems instead of playlistItems.Items
            // ...
        }
    }
    
    return response, nil
}
```

## 4. Potential Race Conditions and Logic Errors

### Example Issue:
```go
// internal/scheduler/schedule.go
for _, playlist := range diffs {
    if err := s.playlist.UpdateUpdatedAt(playlist, now); err != nil {
        log.Println("Could not update cause:", err)
    }
}

for _, playlist := range diffs {
    // 登録済みの各チャンネルへの送信処理
    if err := s.renderer.RenderUpdatedVideo(playlist, location); err != nil {
        log.Println("Message could not send to", playlist.SendChannelID)
    }
}

// internal/service/playlist_service.go
if video.PublishedAt.After(last.UpdatedAt) {
    updated = append(updated, video)
}
```

### Recommendation:
Fix race conditions and logic errors:

```go
// internal/scheduler/schedule.go
for _, playlist := range diffs {
    // Only send notification if update succeeds
    if err := s.playlist.UpdateUpdatedAt(playlist, now); err != nil {
        log.Println("Could not update cause:", err)
        continue // Skip sending notification if update fails
    }
    
    // Send notification only if update succeeded
    if err := s.renderer.RenderUpdatedVideo(playlist, location); err != nil {
        log.Println("Message could not send to", playlist.SendChannelID)
    }
}

// internal/service/playlist_service.go
// Include videos published at exactly the same time as the last update
if !video.PublishedAt.Before(last.UpdatedAt) {
    updated = append(updated, video)
}
```

For the timestamp format issue:

```go
// internal/scheduler/renderer.go
// Change from:
Timestamp: video.OwnerPublishedAt.Format("2006-01-02 15:04:05"),

// To:
Timestamp: video.OwnerPublishedAt.Format(time.RFC3339),
```

## 5. Database and Repository Issues

### Example Issue:
```go
// internal/repository/guild_repository.go
func (r *guildRepository) Exist(guildId string) (bool, error) {
    var guild domain.Guild
    var count int64
    err := r.db.Where(domain.Guild{DiscordID: guildId}).Find(&guild).Count(&count).Error
    if err != nil {
        return false, err
    }

    return count > 0, err // err is always nil here
}

// internal/repository/guild_repository.go
func (r *guildRepository) Add(guildId string) error {
    result := r.db.Save(&domain.Guild{DiscordID: guildId}) // Save instead of Create
    if err := result.Error; err != nil {
        return err
    }

    return nil
}
```

### Recommendation:
Fix database queries and error handling:

```go
// internal/repository/guild_repository.go
func (r *guildRepository) Exist(guildId string) (bool, error) {
    var count int64
    err := r.db.Model(&domain.Guild{}).Where("discord_id = ?", guildId).Count(&count).Error
    if err != nil {
        return false, err
    }

    return count > 0, nil
}

// internal/repository/guild_repository.go
func (r *guildRepository) Add(guildId string) error {
    // Use Create instead of Save to ensure a new record is created
    return r.db.Create(&domain.Guild{DiscordID: guildId}).Error
}
```

## 6. Docker and Deployment Issues

### Example Issue:
```dockerfile
# Dockerfile
FROM golang:1.24-bookworm

WORKDIR /go/src/bot

CMD [ "go", "run", "cmd/server/main.go" ]
```

```yaml
# docker-compose.yml
services:
  bot:
    build: .
    volumes:
      - .:/go/src/bot
    env_file:
      - .env
    depends_on:
      - db
    # No restart policy
```

### Recommendation:
Improve Docker configuration:

```dockerfile
# Dockerfile
FROM golang:1.24-bookworm AS builder

WORKDIR /go/src/bot

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/bot cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /go/bin/bot .

CMD ["./bot"]
```

```yaml
# docker-compose.yml
services:
  bot:
    build: .
    restart: unless-stopped
    env_file:
      - .env
    depends_on:
      db:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

  db:
    image: "postgres:14.4-alpine"
    restart: always
    env_file:
      - .env
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      TZ: ${DB_TIMEZONE}
    volumes:
      - "db-data:/var/lib/postgresql/data"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s

volumes:
  db-data:
```

## 7. Command Handling Issues

### Example Issue:
```go
// internal/server/command/playlist_notifier/playlist_notifier.go
func (c *PlaylistNotifier) Handle(data *discordgo.ApplicationCommandInteractionData, guildId string, channelId string) string {
    subcommand := data.Options[0] // Potential panic if Options is empty
    
    var message string
    switch subcommand.Name {
    case listSubCommand.Name:
        message = c.list(guildId)
    case addSubCommand.Name:
        options := command.ParseArguments(subcommand.Options)
        message = c.add(
            guildId,
            channelId,
            options[playlistIdOption.Name].StringValue(), // Potential panic if option doesn't exist
        )
    // ...
    // No default case
    }
    
    return message
}
```

### Recommendation:
Improve command handling:

```go
// internal/server/command/playlist_notifier/playlist_notifier.go
func (c *PlaylistNotifier) Handle(data *discordgo.ApplicationCommandInteractionData, guildId string, channelId string) string {
    if len(data.Options) == 0 {
        return "Error: No subcommand provided. Please use one of the available subcommands."
    }
    
    subcommand := data.Options[0]
    if subcommand.Type != discordgo.ApplicationCommandOptionSubCommand {
        return "Error: Invalid command format. Please use one of the available subcommands."
    }
    
    var message string
    switch subcommand.Name {
    case listSubCommand.Name:
        message = c.list(guildId)
    case addSubCommand.Name:
        options := command.ParseArguments(subcommand.Options)
        playlistOption, exists := options[playlistIdOption.Name]
        if !exists {
            return "Error: Playlist ID is required."
        }
        
        playlistId := playlistOption.StringValue()
        if playlistId == "" {
            return "Error: Playlist ID cannot be empty."
        }
        
        message = c.add(guildId, channelId, playlistId)
    // ...
    default:
        message = "Error: Unknown subcommand. Please use one of the available subcommands."
    }
    
    return message
}
```

## 8. Performance Issues

### Example Issue:
```go
// internal/server/command/playlist_notifier/list.go
message := "通知登録されているプレイリスト一覧\n"
for _, playlist := range playlists {
    message += fmt.Sprintf("https://www.youtube.com/playlist?list=%s\n", playlist.YoutubeID) // Inefficient string concatenation
}

// internal/scheduler/renderer.go
func separator(integer uint64) string {
    arr := strings.Split(fmt.Sprintf("%d", integer), "")
    var (
        str string
        i2  int
    )
    for i := len(arr) - 1; i >= 0; i-- {
        if i2 > 2 && i2%3 == 0 {
            str = fmt.Sprintf(",%s", str)
        }
        str = fmt.Sprintf("%s%s", arr[i], str)
        i2++
    }
    
    return str
}
```

### Recommendation:
Improve performance:

```go
// internal/server/command/playlist_notifier/list.go
var sb strings.Builder
sb.WriteString("Registered playlist list:\n")
for _, playlist := range playlists {
    sb.WriteString(fmt.Sprintf("https://www.youtube.com/playlist?list=%s\n", playlist.YoutubeID))
}
message := sb.String()

// internal/scheduler/renderer.go
func separator(integer uint64) string {
    // Use the built-in formatting in Go
    return fmt.Sprintf("%d", integer) // In newer Go versions, you can use fmt.Sprintf("%,d", integer)
}
```

## Conclusion

By implementing these recommendations, you can significantly improve the quality, reliability, and maintainability of the Discord Playlist Notifier codebase. These changes will make the application more robust, easier to maintain, and provide a better user experience.