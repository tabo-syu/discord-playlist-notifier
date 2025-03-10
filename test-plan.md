# Discord Playlist Notifier - Test Plan

This document outlines a comprehensive test plan for the Discord Playlist Notifier application. The tests are designed to verify that the application works correctly and to catch any issues that might have been missed in the code review.

## 1. Unit Tests

Unit tests should be implemented for each component of the application to ensure that individual functions and methods work correctly in isolation.

### 1.1. Domain Models

- Test that the domain models can be properly created, updated, and deleted
- Test that the relationships between models work correctly (e.g., Guild has many Playlists, Playlist has many Videos)
- Test validation logic for model fields

### 1.2. Repositories

- Test that the repository methods correctly interact with the database
- Test error handling for database operations
- Test edge cases like empty result sets, duplicate records, etc.

#### Example Test for GuildRepository:

```go
func TestGuildRepository_Exist(t *testing.T) {
    // Setup test database
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("Failed to create mock DB: %v", err)
    }
    defer db.Close()
    
    gormDB, err := gorm.Open(postgres.New(postgres.Config{
        Conn: db,
    }), &gorm.Config{})
    if err != nil {
        t.Fatalf("Failed to open GORM DB: %v", err)
    }
    
    repo := repository.NewGuildRepository(gormDB)
    
    // Test case: Guild exists
    mock.ExpectQuery(`SELECT count\(\*\) FROM "guilds" WHERE discord_id = \$1`).
        WithArgs("123456789").
        WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
    
    exists, err := repo.Exist("123456789")
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    if !exists {
        t.Errorf("Expected guild to exist, but it doesn't")
    }
    
    // Test case: Guild doesn't exist
    mock.ExpectQuery(`SELECT count\(\*\) FROM "guilds" WHERE discord_id = \$1`).
        WithArgs("987654321").
        WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
    
    exists, err = repo.Exist("987654321")
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    if exists {
        t.Errorf("Expected guild not to exist, but it does")
    }
    
    // Test case: Database error
    mock.ExpectQuery(`SELECT count\(\*\) FROM "guilds" WHERE discord_id = \$1`).
        WithArgs("555555555").
        WillReturnError(fmt.Errorf("database error"))
    
    exists, err = repo.Exist("555555555")
    if err == nil {
        t.Errorf("Expected error, but got nil")
    }
    
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("Unfulfilled expectations: %v", err)
    }
}
```

### 1.3. Services

- Test that the service methods correctly use the repositories
- Test business logic in the services
- Test error handling and edge cases

#### Example Test for PlaylistService:

```go
func TestPlaylistService_GetDiffFromLatest(t *testing.T) {
    // Setup mocks
    mockYouTubeRepo := mocks.NewMockYouTubeRepository(t)
    mockPlaylistRepo := mocks.NewMockPlaylistRepository(t)
    mockGuildRepo := mocks.NewMockGuildRepository(t)
    
    service := service.NewPlaylistService(mockYouTubeRepo, mockPlaylistRepo, mockGuildRepo)
    
    // Create test data
    now := time.Now()
    oldTime := now.Add(-1 * time.Hour)
    
    lastPlaylists := []*domain.Playlist{
        {
            Model:     gorm.Model{ID: 1, UpdatedAt: oldTime},
            YoutubeID: "playlist1",
            Title:     "Test Playlist 1",
        },
    }
    
    latestPlaylists := []*domain.Playlist{
        {
            YoutubeID: "playlist1",
            Title:     "Test Playlist 1 Updated",
            Videos: []domain.Video{
                {
                    YoutubeID:   "video1",
                    Title:       "Test Video 1",
                    PublishedAt: oldTime.Add(30 * time.Minute), // New video
                },
                {
                    YoutubeID:   "video2",
                    Title:       "Test Video 2",
                    PublishedAt: oldTime.Add(-30 * time.Minute), // Old video
                },
            },
        },
    }
    
    // Setup expectations
    mockYouTubeRepo.On("FindPlaylistsWithVideos", []string{"playlist1"}).Return(latestPlaylists, nil)
    
    // Call the method
    diffs, err := service.GetDiffFromLatest(lastPlaylists)
    
    // Verify results
    assert.NoError(t, err)
    assert.Len(t, diffs, 1)
    assert.Equal(t, "Test Playlist 1 Updated", diffs[0].Title)
    assert.Len(t, diffs[0].Videos, 1)
    assert.Equal(t, "video1", diffs[0].Videos[0].YoutubeID)
    
    mockYouTubeRepo.AssertExpectations(t)
}
```

### 1.4. Commands

- Test that the command handlers correctly process command inputs
- Test error handling for invalid inputs
- Test that the commands return the expected responses

### 1.5. Scheduler

- Test that the scheduler correctly identifies new videos
- Test that the scheduler correctly sends notifications
- Test error handling for scheduler operations

## 2. Integration Tests

Integration tests should verify that the different components of the application work correctly together.

### 2.1. Repository Integration

- Test that the repositories correctly interact with a real database
- Test that the relationships between models are correctly maintained in the database
- Test database migrations

### 2.2. Service Integration

- Test that the services correctly use the repositories with a real database
- Test that the business logic works correctly with real data

### 2.3. Discord API Integration

- Test that the application correctly registers commands with the Discord API
- Test that the application correctly handles command interactions
- Test that the application correctly sends notifications to Discord channels

### 2.4. YouTube API Integration

- Test that the application correctly fetches playlist and video data from the YouTube API
- Test error handling for YouTube API rate limits and errors
- Test handling of different playlist and video scenarios (e.g., private playlists, deleted videos)

## 3. End-to-End Tests

End-to-end tests should verify that the entire application works correctly from a user's perspective.

### 3.1. Command Functionality

- Test that users can add, list, and delete playlists using the Discord commands
- Test error handling for invalid inputs
- Test that the commands provide helpful error messages

### 3.2. Notification Functionality

- Test that the application correctly sends notifications when new videos are added to a playlist
- Test that the notifications include the correct information (video title, thumbnail, etc.)
- Test that notifications are only sent once for each new video

### 3.3. Edge Cases

- Test handling of deleted playlists
- Test handling of deleted videos
- Test handling of private playlists
- Test handling of playlists with a large number of videos
- Test handling of videos with special characters in the title

## 4. Performance Tests

Performance tests should verify that the application performs well under load.

### 4.1. Database Performance

- Test database query performance with a large number of guilds, playlists, and videos
- Test database connection pooling and resource usage

### 4.2. API Performance

- Test YouTube API usage with a large number of playlists
- Test Discord API usage with a large number of notifications

### 4.3. Scheduler Performance

- Test scheduler performance with a large number of playlists
- Test memory usage during playlist checking

## 5. Security Tests

Security tests should verify that the application is secure and doesn't expose sensitive information.

### 5.1. Authentication

- Test that the application correctly authenticates with the Discord and YouTube APIs
- Test handling of invalid or expired tokens

### 5.2. Authorization

- Test that the application only allows authorized users to use commands
- Test that the application only sends notifications to the correct channels

### 5.3. Data Protection

- Test that sensitive information (e.g., API tokens) is not exposed in logs or error messages
- Test that the application doesn't store sensitive information unnecessarily

## 6. Deployment Tests

Deployment tests should verify that the application can be deployed correctly.

### 6.1. Docker Deployment

- Test that the application can be built and run using Docker
- Test that the Docker container correctly handles environment variables
- Test that the Docker container correctly connects to the database

### 6.2. Environment Configuration

- Test that the application correctly reads environment variables
- Test handling of missing or invalid environment variables

### 6.3. Database Migrations

- Test that database migrations are applied correctly during deployment
- Test that the application works correctly with the migrated database

## 7. Monitoring and Logging Tests

Monitoring and logging tests should verify that the application provides adequate information for troubleshooting and monitoring.

### 7.1. Logging

- Test that the application logs important events and errors
- Test that log messages include relevant information for troubleshooting
- Test log rotation and storage

### 7.2. Monitoring

- Test that the application exposes metrics for monitoring
- Test that the application's health can be checked
- Test that the application's status can be monitored

## Conclusion

This test plan provides a comprehensive approach to testing the Discord Playlist Notifier application. By implementing these tests, you can ensure that the application works correctly, is reliable, and provides a good user experience. The tests should be automated where possible and run regularly to catch regressions early.