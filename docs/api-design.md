# API Design

## Overview

The Program Manager application provides a RESTful API for managing programs, monitoring their status, controlling their execution, and handling data persistence. All API endpoints are prefixed with `/api` and return JSON responses.

## Error Handling

All API endpoints follow a consistent error response format:

```json
{
  "error": "Error message",
  "code": 400
}
```

## Program Management APIs

### Get All Programs

**Endpoint**: `GET /api/programs`

**Description**: Retrieve a list of all managed programs.

**Response**:
- Status: 200 OK
- Body: Array of Program objects

**Example Response**:
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Nginx Web Server",
    "command": "/usr/sbin/nginx",
    "working_dir": "/etc/nginx",
    "description": "High performance web server",
    "status": "running",
    "pid": 12345
  }
]
```

### Add New Program

**Endpoint**: `POST /api/programs`

**Description**: Add a new program to the manager.

**Request Body**:
```json
{
  "name": "Database Service",
  "command": "/usr/bin/mysqld",
  "working_dir": "/var/lib/mysql",
  "description": "MySQL database server"
}
```

**Response**:
- Status: 201 Created
- Body: The created Program object

**Error Responses**:
- 400 Bad Request: Invalid input data
- 500 Internal Server Error: Failed to save program

### Update Program

**Endpoint**: `PUT /api/programs/{id}`

**Description**: Update an existing program.

**Path Parameters**:
- `id`: The unique identifier of the program to update

**Request Body**:
```json
{
  "name": "Database Service",
  "command": "/usr/bin/mysqld",
  "working_dir": "/var/lib/mysql",
  "description": "MySQL database server"
}
```

**Response**:
- Status: 200 OK
- Body: The updated Program object

**Error Responses**:
- 400 Bad Request: Invalid input data
- 404 Not Found: Program with specified ID not found
- 500 Internal Server Error: Failed to update program

### Delete Program

**Endpoint**: `DELETE /api/programs/{id}`

**Description**: Remove a program from the manager.

**Path Parameters**:
- `id`: The unique identifier of the program to delete

**Response**:
- Status: 204 No Content

**Error Responses**:
- 404 Not Found: Program with specified ID not found
- 500 Internal Server Error: Failed to delete program

## Status Monitoring APIs

### Get Program Status

**Endpoint**: `GET /api/programs/{id}/status`

**Description**: Get the current status of a specific program.

**Path Parameters**:
- `id`: The unique identifier of the program

**Response**:
- Status: 200 OK
- Body: Program status object

**Example Response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "pid": 12345
}
```

**Error Responses**:
- 404 Not Found: Program with specified ID not found

### Get All Programs Status

**Endpoint**: `GET /api/programs/status`

**Description**: Get the current status of all programs.

**Response**:
- Status: 200 OK
- Body: Array of program status objects

**Example Response**:
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "running",
    "pid": 12345
  },
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "status": "stopped",
    "pid": 0
  }
]
```

### Refresh All Programs Status

**Endpoint**: `POST /api/programs/refresh`

**Description**: Refresh the status of all programs by checking their process status.

**Response**:
- Status: 200 OK
- Body: Array of updated program objects

**Example Response**:
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Nginx Web Server",
    "command": "/usr/sbin/nginx",
    "working_dir": "/etc/nginx",
    "description": "High performance web server",
    "status": "running",
    "pid": 12345
  }
]
```

## Process Control APIs

### Start Program

**Endpoint**: `POST /api/programs/{id}/start`

**Description**: Start a program execution.

**Path Parameters**:
- `id`: The unique identifier of the program to start

**Response**:
- Status: 200 OK
- Body: Updated Program object

**Error Responses**:
- 404 Not Found: Program with specified ID not found
- 409 Conflict: Program is already running
- 500 Internal Server Error: Failed to start program

### Stop Program

**Endpoint**: `POST /api/programs/{id}/stop`

**Description**: Stop a running program.

**Path Parameters**:
- `id`: The unique identifier of the program to stop

**Response**:
- Status: 200 OK
- Body: Updated Program object

**Error Responses**:
- 404 Not Found: Program with specified ID not found
- 409 Conflict: Program is not running
- 500 Internal Server Error: Failed to stop program

### Batch Start Programs

**Endpoint**: `POST /api/programs/start`

**Description**: Start multiple programs.

**Request Body**:
```json
{
  "ids": [
    "550e8400-e29b-41d4-a716-446655440000",
    "660e8400-e29b-41d4-a716-446655440001"
  ]
}
```

**Response**:
- Status: 200 OK
- Body: Array of updated Program objects

**Error Responses**:
- 400 Bad Request: Invalid input data
- 500 Internal Server Error: Failed to start one or more programs

### Batch Stop Programs

**Endpoint**: `POST /api/programs/stop`

**Description**: Stop multiple running programs.

**Request Body**:
```json
{
  "ids": [
    "550e8400-e29b-41d4-a716-446655440000",
    "660e8400-e29b-41d4-a716-446655440001"
  ]
}
```

**Response**:
- Status: 200 OK
- Body: Array of updated Program objects

**Error Responses**:
- 400 Bad Request: Invalid input data
- 500 Internal Server Error: Failed to stop one or more programs

## Data Persistence APIs

程序管理器目前专注于核心功能，数据持久化通过JSON文件直接管理，暂不提供导入导出功能。