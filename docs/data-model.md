# Data Model Design

## Overview

The data model for the Program Manager application is designed to store and manage information about executable programs and shell scripts. The core entity is the `Program` which contains all necessary information to manage a program's lifecycle.

## Program Entity

The `Program` entity represents a single executable program or shell script that can be managed by the application.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| ID | string | Unique identifier for the program (UUID format) |
| Name | string | Human-readable name for the program |
| Command | string | The command to execute the program |
| WorkingDir | string | The working directory where the program should be executed |
| Description | string | Optional description or notes about the program |
| Status | string | Current status of the program ("running" or "stopped") |
| PID | int | Process ID when the program is running (0 when stopped) |

### Example JSON Representation

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Nginx Web Server",
  "command": "/usr/sbin/nginx",
  "working_dir": "/etc/nginx",
  "description": "High performance web server",
  "status": "running",
  "pid": 12345
}
```

## ProgramManager Entity

The `ProgramManager` entity is a container for managing a collection of programs.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| Programs | array of Program | List of managed programs |

### Example JSON Representation

```json
{
  "programs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Nginx Web Server",
      "command": "/usr/sbin/nginx",
      "working_dir": "/etc/nginx",
      "description": "High performance web server",
      "status": "running",
      "pid": 12345
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "Database Service",
      "command": "/usr/bin/mysqld",
      "working_dir": "/var/lib/mysql",
      "description": "MySQL database server",
      "status": "stopped",
      "pid": 0
    }
  ]
}
```

## Data Validation Rules

### Program Validation

1. **ID**: Must be a valid UUID string
2. **Name**: 
   - Required field
   - Maximum length of 100 characters
   - Must not be empty
3. **Command**: 
   - Required field
   - Maximum length of 500 characters
   - Must not be empty
4. **WorkingDir**: 
   - Optional field
   - If provided, must be a valid file system path
   - Maximum length of 500 characters
5. **Description**: 
   - Optional field
   - Maximum length of 1000 characters
6. **Status**: 
   - Must be either "running" or "stopped"
   - Default value is "stopped"
7. **PID**: 
   - Must be a non-negative integer
   - 0 indicates the program is not running
   - Positive values represent actual process IDs

## Relationships

The data model is relatively simple with a one-to-many relationship:
- One `ProgramManager` contains many `Program` entities

There are no complex relationships or foreign keys since all data is stored in a single JSON file.

## Storage Considerations

### JSON Serialization

When serializing to JSON:
- All field names are converted to snake_case
- Boolean values are represented as lowercase strings
- Numeric values are represented as numbers
- Empty values are omitted from the JSON output

### File Storage

The entire `ProgramManager` entity is serialized to a single JSON file:
- Default filename: `programs.json`
- Stored in the application's data directory
- UTF-8 encoding
- Human-readable formatting with indentation

## Concurrency Handling

Since multiple operations may access the data concurrently:
- A mutex is used to protect read/write operations
- File locks may be implemented to prevent corruption
- Atomic file operations are used when updating the data file