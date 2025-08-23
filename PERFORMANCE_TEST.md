# DB Loader Performance Tester - Implementation Plan

## Project Overview

Create a comprehensive database loading performance tester that generates CSV files with configurable parameters and tests different insertion strategies. This will benchmark database performance and demonstrate optimization techniques.

## Core Requirements

### Frontend Features

- Configuration form with:
  - Row count selection (presets: 10k, 100k, 1M, 10M + custom input)
  - Column count (10-200 range slider)
  - Date column count (0-50% of total columns)
  - Method selection: "Brute Force" vs "Optimized"
- Real-time progress display with WebSocket updates
- Results dashboard showing timing breakdown and performance metrics

### Backend Features

- CSV generation service with realistic fake data
- Date validation service supporting 10 different US date formats
- Two insertion methods:
  - **Brute Force**: Single row inserts
  - **Optimized**: Batch inserts (2000-5000 rows per batch) with transactions
- Progress tracking via WebSocket
- Performance timing for each phase (generation, parsing, insertion)

## Technical Implementation

### Database Schema (GORM Models)

Create these new models in `server/internal/models/`:

**LoadTest Model:**

```go
type LoadTest struct {
    BaseModel
    Rows         int    `json:"rows" gorm:"not null"`
    Columns      int    `json:"columns" gorm:"not null"`
    DateColumns  int    `json:"date_columns" gorm:"not null"`
    Method       string `json:"method" gorm:"not null"` // 'brute_force' or 'optimized'
    Status       string `json:"status" gorm:"not null"` // 'running', 'completed', 'failed'
    CSVGenTime   *int   `json:"csv_gen_time"`   // milliseconds
    ParseTime    *int   `json:"parse_time"`     // milliseconds
    InsertTime   *int   `json:"insert_time"`    // milliseconds
    TotalTime    *int   `json:"total_time"`     // milliseconds
    ErrorMessage *string `json:"error_message,omitempty"`
}
```

**TestData Model (Dynamic Columns):**

```go
type TestData struct {
    BaseModel
    LoadTestID string `json:"load_test_id" gorm:"not null;index"`
    // Dynamic columns col_1 through col_200
    Col1   *string `json:"col_1" gorm:"type:varchar(255)"`
    Col2   *string `json:"col_2" gorm:"type:varchar(255)"`
    // ... continue pattern up to Col200
    Col200 *string `json:"col_200" gorm:"type:varchar(255)"`
}
```

### Backend Components

#### 1. LoadTest Controller (`server/internal/controllers/loadtest/`)

**LoadTestController:**

- `StartTest(c *fiber.Ctx)` - Initialize and start load test
- `GetTestStatus(c *fiber.Ctx)` - Get current test status
- `GetTestHistory(c *fiber.Ctx)` - List previous tests
- `DeleteTest(c *fiber.Ctx)` - Clean up test data

#### 2. Services (`server/internal/services/`)

**CSVGeneratorService:**

- Generate CSV files with configurable parameters
- Use faker library for realistic data:
  - Names, addresses, phone numbers, email addresses
  - Product names, company names, descriptions
  - Random strings for filler columns
- Shuffle column order randomly
- Insert date columns in various formats at random positions

**DateValidationService:**
Support these 10 US date formats:

```
1. MM/DD/YYYY (03/15/2024)
2. M/D/YYYY (3/15/2024)
3. MM/DD/YY (03/15/24)
4. M/D/YY (3/15/24)
5. MM-DD-YYYY (03-15-2024)
6. M-D-YYYY (3-15-2024)
7. MM-DD-YY (03-15-24)
8. M-D-YY (3-15-24)
9. YYYY-MM-DD (2024-03-15)
10. YYYY/MM/DD (2024/03/15)
```

- Parse and validate dates
- Convert to standardized format (YYYY-MM-DD)
- Handle invalid dates gracefully

**LoadTestService:**

- Coordinate the entire test process
- Two insertion methods:
  - **Brute Force**: `INSERT` one row at a time
  - **Optimized**: Batch `INSERT` with transactions (batch size: 3000 rows)
- Progress tracking with WebSocket updates
- Performance timing for each phase

#### 3. Repository Layer

**LoadTestRepository:**

- CRUD operations for LoadTest model
- Batch insert operations for TestData
- Performance-optimized queries

#### 4. WebSocket Integration

Extend existing WebSocket system to broadcast:

- Test start/completion notifications
- Progress updates (percentage complete, current phase)
- Performance metrics in real-time
- Error notifications

### Frontend Components

#### 1. LoadTest Page (`client/src/pages/LoadTest/`)

**LoadTestForm Component:**

- Row count selection with presets and custom input
- Column count slider (10-200)
- Date column percentage slider
- Method radio buttons
- Start test button

**ProgressDisplay Component:**

- Overall progress bar
- Phase-specific progress (CSV Gen → Parsing → Insertion)
- Real-time timing display
- ETA calculation

**ResultsDashboard Component:**

- Performance metrics table
- Time breakdown visualization
- Historical test comparison
- Success/failure status

#### 2. API Integration (`client/src/services/api/loadtest.ts`)

**LoadTest API Service:**

- Start test endpoint
- Status polling
- Historical data fetching
- WebSocket connection for real-time updates

### File Structure

```
server/
├── internal/
│   ├── controllers/
│   │   └── loadtest/
│   │       └── loadTestController.go
│   ├── models/
│   │   ├── loadTest.model.go
│   │   └── testData.model.go
│   ├── services/
│   │   ├── csvGenerator.service.go
│   │   ├── dateValidation.service.go
│   │   └── loadTest.service.go
│   ├── repositories/
│   │   └── loadTest.repository.go
│   └── routes/
│       └── loadTest.routes.go

client/
├── src/
│   ├── pages/
│   │   └── LoadTest/
│   │       ├── LoadTest.tsx
│   │       ├── LoadTestForm.tsx
│   │       ├── ProgressDisplay.tsx
│   │       └── ResultsDashboard.tsx
│   └── services/
│       └── api/
│           └── loadtest.ts
```

### API Endpoints

```
POST   /api/loadtest/start      - Start new load test
GET    /api/loadtest/:id        - Get test status
GET    /api/loadtest            - List test history
DELETE /api/loadtest/:id        - Delete test and data
```

### Performance Targets

- **Goal**: Load 1M rows in under 60 seconds with optimized method
- **Benchmark**: Compare against reported 10-15 minutes for 100k rows
- **Metrics**: Track CSV generation, parsing, and insertion times separately

### Implementation Phases

1. **Phase 1**: Backend models, repositories, and basic controller
2. **Phase 2**: CSV generation and date validation services
3. **Phase 3**: Load test execution with both methods
4. **Phase 4**: WebSocket integration for progress updates
5. **Phase 5**: Frontend UI and API integration
6. **Phase 6**: Performance optimization and testing

### Dependencies

**Backend:**

- Add faker library for Go: `github.com/go-faker/faker/v4`
- CSV parsing: `encoding/csv` (standard library)

**Frontend:**

- Chart visualization (if needed): existing recharts
- Progress indicators: custom components with existing styling

### Success Criteria

- Successfully generate and load 10M rows
- Demonstrate significant performance improvement with optimized method
- Real-time progress updates working smoothly
- Clean, intuitive UI for configuration and monitoring
- Comprehensive error handling and recovery

This implementation will provide a robust testing framework for database performance while showcasing optimization techniques and real-time monitoring capabilities.
