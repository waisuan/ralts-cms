# CLI Data Generator

This CLI tool generates random test data for the Ralts-CMS application. It can create machines with maintenance records and users with random data.

## Usage

```bash
# Create 10 machines with random maintenance records
go run cmd/cli/main.go -type machine -count 10

# Create 5 users with random data
go run cmd/cli/main.go -type user -count 5
```

## Features

### Machine Generation

- **Random Data**: All machine fields are populated with random realistic data
- **Optional Fields**: Not all optional fields are necessarily populated
- **Random Dates**: Dates are distributed across past, present, and future:
  - 30% past dates (up to 2 years ago)
  - 40% present dates (within 30 days)
  - 30% future dates (up to 1 year ahead)
- **Maintenance Records**: Each machine gets 0-5 random maintenance records
- **Unique Serial Numbers**: Serial numbers follow the pattern `SN-XXXXXX`

### User Generation

- **Random Names**: Combines random first and last names
- **Unique Emails**: Emails follow the pattern `FirstName.LastName.N@domain.com`
- **Random Roles**: user, admin, manager, technician, supervisor
- **Random Status**: active, inactive, suspended, pending
- **Optional Avatars**: 20% chance of having an avatar

## Sample Data

### Machine Fields

- **Customers**: Acme Corp, Beta Industries, Gamma Solutions, etc.
- **States**: Malaysian states (Selangor, Kuala Lumpur, Penang, etc.)
- **Account Types**: Premium, Standard, Basic, Enterprise, Professional
- **Models**: X100, Y200, Z300, A400, B500, etc.
- **Statuses**: Active, Inactive, Maintenance, Repair, Operational, etc.
- **Brands**: BrandA through BrandJ
- **Districts**: Petaling, Klang, Shah Alam, Subang Jaya, etc.
- **Persons**: John Doe, Jane Smith, Bob Johnson, etc.
- **Reporters**: Tech Team A, Maintenance Crew, Service Department, etc.

### Maintenance Fields

- **Actions**: Routine maintenance, Filter replacement, Oil change, etc.
- **Types**: Preventive, Corrective, Emergency, Scheduled, etc.
- **Technicians**: Mike Johnson, Sarah Williams, David Brown, etc.

### User Fields

- **Names**: 20 first names × 20 last names combinations
- **Domains**: example.com, test.org, demo.net, etc.
- **Passwords**: Common test passwords (password123, test123, etc.)

## Examples

### Create 5 machines with maintenance records

```bash
go run cmd/cli/main.go -type machine -count 5
```

Output:
```
Created machine: SN-000001
Created maintenance: WO-SN-000001-001 for machine: SN-000001
Created maintenance: WO-SN-000001-002 for machine: SN-000001
Created machine: SN-000002
Created machine: SN-000003
Created maintenance: WO-SN-000003-001 for machine: SN-000003
Created machine: SN-000004
Created machine: SN-000005
Successfully created 5 machines with maintenance records
```

### Create 3 users

```bash
go run cmd/cli/main.go -type user -count 3
```

Output:
```
Created user: John.Smith.1@example.com (John Smith)
Created user: Jane.Johnson.2@test.org (Jane Johnson)
Created user: Bob.Williams.3@demo.net (Bob Williams)
Successfully created 3 users
```

## Error Handling

- The tool continues processing even if individual records fail to create
- Failed creations are logged with error details
- Database constraint violations (duplicate keys) are handled gracefully

## Database Requirements

- PostgreSQL database must be running
- Database schema must be migrated
- Environment variables must be configured (see main README)

## Notes

- Serial numbers are sequential (SN-000001, SN-000002, etc.)
- Work order numbers follow the pattern `WO-{SerialNumber}-{Sequence}`
- Email addresses include a counter to ensure uniqueness
- All dates are in UTC timezone
- Random generation uses crypto/rand for better randomness 