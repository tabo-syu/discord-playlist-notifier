# Discord Playlist Notifier - Prioritized Action Plan

This document provides a prioritized action plan for addressing the issues identified in the codebase. The issues are categorized by severity and impact, with recommendations for addressing them in order of priority.

## Critical Issues (High Priority)

These issues could lead to bugs, crashes, or security vulnerabilities and should be addressed immediately.

### 1. Fix Potential Panic Points

**Issue**: Several places in the code could cause panics under certain conditions.

**Affected Files**:
- `internal/server/command/playlist_notifier/playlist_notifier.go` - No check for empty Options array
- `internal/server/command/playlist_notifier/add.go` - No validation for playlistId
- `internal/server/server.go` - No error handling for InteractionRespond

**Action Items**:
- Add null checks and length validation before accessing arrays and maps
- Add input validation for all user-provided data
- Add error handling for all external API calls

### 2. Fix Race Conditions in Notification Logic

**Issue**: If updating a playlist's timestamp fails, notifications are still sent, which could lead to repeated notifications.

**Affected Files**:
- `internal/scheduler/schedule.go`

**Action Items**:
- Only send notifications if the timestamp update succeeds
- Consider using transactions to ensure atomicity

### 3. Fix YouTube API Pagination and Limits

**Issue**: The code only fetches the first 20 items from YouTube API, which could lead to missing videos.

**Affected Files**:
- `internal/repository/youtube_repository.go`

**Action Items**:
- Implement pagination to fetch all items from YouTube API
- Handle YouTube API quota limits properly
- Add error handling for API rate limiting

## Major Issues (Medium Priority)

These issues affect functionality, performance, or maintainability but are not likely to cause crashes.

### 1. Improve Error Handling

**Issue**: Error handling is inconsistent and sometimes inadequate.

**Affected Files**:
- `internal/domain/errors.go` - Grammatical errors in error messages
- `internal/repository/youtube_repository.go` - Ignored errors from time.Parse
- Multiple repository files - Inconsistent error handling patterns

**Action Items**:
- Fix grammatical errors in error messages
- Handle all errors properly, including those from time.Parse
- Standardize error handling patterns across the codebase

### 2. Fix Timestamp Format for Discord

**Issue**: The timestamp format used for Discord embeds is incorrect.

**Affected Files**:
- `internal/scheduler/renderer.go`

**Action Items**:
- Change the timestamp format to RFC3339 (e.g., "2006-01-02T15:04:05Z")

### 3. Improve Database Queries

**Issue**: Some database queries are inefficient or could lead to unexpected behavior.

**Affected Files**:
- `internal/repository/guild_repository.go`
- `internal/repository/playlist_repository.go`

**Action Items**:
- Use First/Take consistently for single record queries
- Use Create for new records and Update for existing records
- Simplify error handling in repository functions

### 4. Fix Docker Configuration

**Issue**: The Docker configuration is not optimized for production use.

**Affected Files**:
- `Dockerfile`
- `docker-compose.yml`

**Action Items**:
- Update Dockerfile to build the application before running it
- Add restart policy for the bot service
- Add health checks for the services

## Minor Issues (Low Priority)

These issues don't affect functionality but would improve code quality and maintainability.

### 1. Internationalization

**Issue**: The codebase mixes Japanese and English, which makes it difficult for non-Japanese speakers to understand and maintain.

**Affected Files**:
- Multiple files with Japanese comments and user-facing messages

**Action Items**:
- Translate all comments to English
- Implement proper internationalization for user-facing messages

### 2. Improve Performance

**Issue**: Some parts of the code could be optimized for better performance.

**Affected Files**:
- `internal/server/command/playlist_notifier/list.go` - Inefficient string concatenation
- `internal/scheduler/renderer.go` - Complex separator function

**Action Items**:
- Use strings.Builder for string concatenation
- Simplify the separator function

### 3. Improve Command Responses

**Issue**: Command responses are limited to strings, which limits the types of responses that can be sent.

**Affected Files**:
- `internal/server/command/command.go`
- `internal/server/server.go`

**Action Items**:
- Support more complex response types (embeds, components, etc.)
- Add better error messages for command failures

## Implementation Strategy

To address these issues efficiently, we recommend the following approach:

1. **Start with Critical Issues**: Fix potential panic points, race conditions, and YouTube API limitations first to ensure the application is stable and reliable.

2. **Address Major Issues**: Once the critical issues are fixed, address the major issues to improve functionality and maintainability.

3. **Tackle Minor Issues**: Finally, address the minor issues to improve code quality and user experience.

4. **Add Tests**: Throughout the process, add unit and integration tests to ensure the changes don't introduce new issues.

5. **Refactor Gradually**: Instead of making all changes at once, refactor the code gradually to minimize the risk of introducing new issues.

## Conclusion

By following this prioritized action plan, you can systematically address the issues in the Discord Playlist Notifier codebase and improve its quality, reliability, and maintainability. The most critical issues should be addressed first to ensure the application is stable and reliable, followed by the major and minor issues to improve functionality, performance, and maintainability.