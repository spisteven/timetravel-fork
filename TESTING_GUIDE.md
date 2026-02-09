# Testing Guide

This guide provides comprehensive instructions for testing the Time Travel API, including both v1 (backward compatible) and v2 (versioned) endpoints.

## Prerequisites

1. **Go 1.17+** installed
2. **curl** or similar HTTP client for testing
3. Database dependencies installed:
   ```bash
   go mod tidy
   ```

## Starting the Server

1. Build the application:
   ```bash
   go build -o timetravel .
   ```

2. Run the server:
   ```bash
   ./timetravel
   ```

   The server will start on `http://127.0.0.1:8000`

3. Verify the server is running:
   ```bash
   curl -X POST http://localhost:8000/api/v1/health
   ```

   Expected response:
   ```json
   {"ok":true}
   ```

## API v1 Testing (Backward Compatible)

The v1 API provides basic record operations without versioning. All data is persisted in SQLite.

### Create a Record

```bash
curl -X POST http://localhost:8000/api/v1/records/1 \
  -H "Content-Type: application/json" \
  -d '{"hello": "world", "status": "active"}'
```

**Expected Response:**
```json
{"id":1,"data":{"hello":"world","status":"active"}}
```

### Get a Record

```bash
curl -X GET http://localhost:8000/api/v1/records/1
```

**Expected Response:**
```json
{"id":1,"data":{"hello":"world","status":"active"}}
```

### Update a Record

```bash
curl -X POST http://localhost:8000/api/v1/records/1 \
  -H "Content-Type: application/json" \
  -d '{"hello": "world 2", "status": "inactive", "newfield": "added"}'
```

**Expected Response:**
```json
{"id":1,"data":{"hello":"world 2","status":"inactive","newfield":"added"}}
```

### Delete a Field (Set to null)

```bash
curl -X POST http://localhost:8000/api/v1/records/1 \
  -H "Content-Type: application/json" \
  -d '{"newfield": null}'
```

**Expected Response:**
```json
{"id":1,"data":{"hello":"world 2","status":"inactive"}}
```

### Error Cases

**Get non-existent record:**
```bash
curl -X GET http://localhost:8000/api/v1/records/999
```

**Expected Response (400 Bad Request):**
```json
{"error":"record of id 999 does not exist"}
```

**Invalid ID:**
```bash
curl -X GET http://localhost:8000/api/v1/records/0
```

**Expected Response (400 Bad Request):**
```json
{"error":"invalid id; id must be a positive number"}
```

## API v2 Testing (Time Travel)

The v2 API provides full versioning capabilities, allowing you to track history and query records at specific points in time.

### Create a Record (Version 1)

```bash
curl -X POST http://localhost:8000/api/v2/records/100 \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com", "role": "admin"}'
```

**Expected Response:**
```json
{"id":100,"data":{"name":"John Doe","email":"john@example.com","role":"admin"}}
```

This creates the record and automatically creates version 1.

### Get Latest Version

```bash
curl -X GET http://localhost:8000/api/v2/records/100
```

**Expected Response:**
```json
{"id":100,"data":{"name":"John Doe","email":"john@example.com","role":"admin"}}
```

### List All Versions

```bash
curl -X GET http://localhost:8000/api/v2/records/100/versions
```

**Expected Response:**
```json
{
  "id": 100,
  "versions": [
    {
      "version": 1,
      "created_at": "2026-02-08T18:04:43.970786-06:00"
    }
  ]
}
```

### Update Record (Creates New Version)

```bash
curl -X POST http://localhost:8000/api/v2/records/100 \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john.doe@example.com", "role": "user", "department": "Engineering"}'
```

**Expected Response:**
```json
{"id":100,"data":{"name":"John Doe","email":"john.doe@example.com","role":"user","department":"Engineering"}}
```

This creates version 2 automatically.

### View Version History

```bash
curl -X GET http://localhost:8000/api/v2/records/100/versions
```

**Expected Response:**
```json
{
  "id": 100,
  "versions": [
    {
      "version": 2,
      "created_at": "2026-02-08T18:05:12.123456-06:00"
    },
    {
      "version": 1,
      "created_at": "2026-02-08T18:04:43.970786-06:00"
    }
  ]
}
```

### Get Specific Version (Time Travel)

**Get Version 1:**
```bash
curl -X GET http://localhost:8000/api/v2/records/100/versions/1
```

**Expected Response:**
```json
{"id":100,"data":{"name":"John Doe","email":"john@example.com","role":"admin"}}
```

**Get Version 2:**
```bash
curl -X GET http://localhost:8000/api/v2/records/100/versions/2
```

**Expected Response:**
```json
{"id":100,"data":{"name":"John Doe","email":"john.doe@example.com","role":"user","department":"Engineering"}}
```

### Update with Field Deletion

```bash
curl -X POST http://localhost:8000/api/v2/records/100 \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john.doe@example.com", "role": "user", "department": null}'
```

This creates version 3, removing the `department` field.

### Error Cases

**Get non-existent record:**
```bash
curl -X GET http://localhost:8000/api/v2/records/999
```

**Expected Response (404 Not Found):**
```json
{"error":"record of id 999 does not exist"}
```

**Get non-existent version:**
```bash
curl -X GET http://localhost:8000/api/v2/records/100/versions/999
```

**Expected Response (404 Not Found):**
```json
{"error":"record version 100@999 does not exist"}
```

**Invalid version number:**
```bash
curl -X GET http://localhost:8000/api/v2/records/100/versions/0
```

**Expected Response (400 Bad Request):**
```json
{"error":"invalid version; version must be a positive number"}
```

## Complete Test Scenario

Here's a complete scenario demonstrating the time travel functionality:

### Step 1: Create Initial Record
```bash
curl -X POST http://localhost:8000/api/v2/records/200 \
  -H "Content-Type: application/json" \
  -d '{"policy_number": "POL-001", "coverage": "100000", "status": "active"}'
```

### Step 2: Verify Version 1
```bash
curl -X GET http://localhost:8000/api/v2/records/200/versions/1
```

### Step 3: Update Coverage (Creates Version 2)
```bash
curl -X POST http://localhost:8000/api/v2/records/200 \
  -H "Content-Type: application/json" \
  -d '{"policy_number": "POL-001", "coverage": "200000", "status": "active"}'
```

### Step 4: Change Status (Creates Version 3)
```bash
curl -X POST http://localhost:8000/api/v2/records/200 \
  -H "Content-Type: application/json" \
  -d '{"policy_number": "POL-001", "coverage": "200000", "status": "suspended", "reason": "non-payment"}'
```

### Step 5: View All Versions
```bash
curl -X GET http://localhost:8000/api/v2/records/200/versions
```

### Step 6: Time Travel - Query Historical States

**What was the coverage in version 1?**
```bash
curl -X GET http://localhost:8000/api/v2/records/200/versions/1
```
Response shows: `"coverage": "100000"`

**What was the status in version 2?**
```bash
curl -X GET http://localhost:8000/api/v2/records/200/versions/2
```
Response shows: `"status": "active"`

**What is the current state?**
```bash
curl -X GET http://localhost:8000/api/v2/records/200
```
Response shows: `"status": "suspended"` with `"reason": "non-payment"`

## Testing Backward Compatibility

Verify that v1 and v2 APIs work independently:

1. **Create record in v1:**
   ```bash
   curl -X POST http://localhost:8000/api/v1/records/300 \
     -H "Content-Type: application/json" \
     -d '{"test": "v1"}'
   ```

2. **Verify it's accessible via v1:**
   ```bash
   curl -X GET http://localhost:8000/api/v1/records/300
   ```

3. **Note:** v1 records are NOT versioned. They exist in the database but don't have version history.

4. **Create record in v2 (different ID):**
   ```bash
   curl -X POST http://localhost:8000/api/v2/records/400 \
     -H "Content-Type: application/json" \
     -d '{"test": "v2"}'
   ```

5. **Verify versioning works:**
   ```bash
   curl -X GET http://localhost:8000/api/v2/records/400/versions
   ```

## Database Persistence Testing

Test that data persists across server restarts:

1. **Create a record:**
   ```bash
   curl -X POST http://localhost:8000/api/v2/records/500 \
     -H "Content-Type: application/json" \
     -d '{"persistent": "test"}'
   ```

2. **Stop the server** (Ctrl+C or `pkill -f timetravel`)

3. **Restart the server:**
   ```bash
   ./timetravel
   ```

4. **Verify the record still exists:**
   ```bash
   curl -X GET http://localhost:8000/api/v2/records/500
   ```

   The record should still be there with all its versions intact.

## Performance Testing

For load testing, you can use tools like `ab` (Apache Bench) or `wrk`:

```bash
# Test GET endpoint performance
ab -n 1000 -c 10 http://localhost:8000/api/v2/records/100

# Test POST endpoint performance
ab -n 100 -c 5 -p payload.json -T application/json \
   http://localhost:8000/api/v2/records/100
```

## Troubleshooting

### Server won't start
- Check if port 8000 is already in use
- Verify Go version is 1.17+
- Run `go mod tidy` to ensure dependencies are installed

### Database errors
- Check file permissions for `timetravel.db`
- Ensure SQLite3 is properly installed
- Delete `timetravel.db` to reset (will lose all data)

### 404 errors on v2 endpoints
- Ensure you're using the correct URL format
- Check that the server was rebuilt after code changes
- Verify routes are registered in `server.go`

## Additional Notes

- All record IDs must be **positive integers**
- Version numbers start at 1 and increment automatically
- Versions are immutable - once created, they cannot be modified
- The `created_at` timestamp reflects when the version was created
- Null values in POST requests delete fields from the record
- v1 and v2 APIs use the same database but different service implementations
