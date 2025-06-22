# Overview

Ralts-CMS is a web service built in Golang. It's primarily used by a web application.
It provides functions to query and mutate records represented as `machines`.
Each `machine` has historical data and these are representated as `maintenance`.

# API

Each API endpoint is authenticated using Basic Auth.
The username and password must match the `CREDENTIALS` environment variable.
The variable contains a string token with the target username and password concatenated by ":".
For example, "admin:12345". The token is encoded in a base64 format.

Each API endpoint is expected to return the appropriate HTTP status code for successful and non-successful responses.

- GET /machines/:serial_number
- POST /machines
- PUT /machines
- DELETE /machines/:serial_number
- GET /machines/:serial_number/maintenance
- GET /machines/:serial_number/maintenance/:work_order_number
- POST /machines/:serial_number/maintenance
- PUT /machines/:serial_number/maintenance
- DELETE /machines/:serial_number/maintenance/:work_order_number

# Data Storage

The service uses AWS DynamoDB for persistent data storage.

### Schema
```json
{
  "TableName": "ralts",
  "KeySchema": [
    {
      "KeyType": "HASH",
      "AttributeName": "PK"
    },
    {
      "KeyType": "RANGE",
      "AttributeName": "SK"
    }
  ],
  "AttributeDefinitions": [
    {
      "AttributeName": "PK",
      "AttributeType": "S"
    },
    {
      "AttributeName": "SK",
      "AttributeType": "S"
    },
    {
      "AttributeName": "GSI1_PK",
      "AttributeType": "S"
    },
    {
      "AttributeName": "GSI1_SK",
      "AttributeType": "S"
    }
  ],
  "BillingMode": "PAY_PER_REQUEST",
  "GlobalSecondaryIndexes": [
    {
      "IndexName": "gsi_1",
      "Projection": {
        "ProjectionType": "ALL"
      },
      "KeySchema": [
        {
          "AttributeName": "GSI1_PK",
          "KeyType": "HASH"
        },
        {
          "AttributeName": "GSI1_SK",
          "KeyType": "RANGE"
        }
      ],
      "BillingMode": "PAY_PER_REQUEST"
    }
  ]
}
```

### Example
| PK | SK | Item |
| --- | --- |  --- |
| Machine123 | # | {...} |
| Machine#123 | Maintenance#001 | {...} |
| Machine456 | # | {...} |

# Domain Models

## Machines

- Partition Key: Machine#<SerialNumber>
- Sort key: #
- `CreatedAt` is defined upon the creation of the table record.
- `UpdatedAt` is updated when the table record is updated.

```go
type Machine struct {
	SerialNumber    string
	Customer        string     
	State           string     
	AccountType     string     
	Model           string     
	Status          string     
	Brand           string     
	District        string     
	PersonInCharge  string     
	ReportedBy      string     
	AdditionalNotes string     
	Attachment      string     
	PpmStatus       string     
	TncDate         string // ISO 8601 format. E.g. 2025-05-12
	PpmDate         string // ISO 8601 format. E.g. 2025-05-12
	CreatedAt       string // ISO 8601 format. E.g. 2015-12-21T17:42:34Z
	UpdatedAt       string // ISO 8601 format. E.g. 2015-12-21T17:42:34Z
}
```

## Maintenance

- Partition Key: Machine#<MachineSerialNumber>
- Sort key: Maintenance#<WorkOrderNumber>
- `CreatedAt` is defined upon the creation of the table record.
- `UpdatedAt` is updated when the table record is updated.

```go
type Maintenance struct {
	MachineSerialNumber string
	WorkOrderNumber     string
	WorkOrderDate       string // ISO 8601 format. E.g. 2025-05-12
	ActionTaken         string
	ReportedBy          string
	WorkerOrderType     string
	Attachment          string
	CreatedAt           string // ISO 8601 format. E.g. 2015-12-21T17:42:34Z
	UpdatedAt           string // ISO 8601 format. E.g. 2015-12-21T17:42:34Z
}
```

# Key Features

- Able to fetch a target `machine` record.
- Able to create a new `machine` record.
- Able to update an existing `machine` record.
- Able to delete an existing `machine` record.
- Able to fetch a list of `maintenance` records for a given `machine` record.
- Able to fetch a target `maintenance` record.
- Able to create a new `maintenance` record.
- Able to update an existing `maintenance` record.
- Able to delete an existing `maintenance` record.

# Implementation Checklist

- [x] Setup and configure local DynamoDB instance
- [x] Create DynamoDB table
- [x] Create a web server
- [x] Construct a `/health` endpoint
