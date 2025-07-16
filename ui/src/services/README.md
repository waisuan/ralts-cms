# API Integration Documentation

This directory contains the client-side API integration for the Ralts CMS frontend.

## Overview

The API integration is built with TypeScript and provides a clean, type-safe interface for communicating with the backend API. It includes:

- **API Client**: Base HTTP client with error handling
- **Service Classes**: Business logic for specific API endpoints
- **React Hooks**: Custom hooks for data fetching with loading states
- **Type Definitions**: TypeScript interfaces for API requests/responses

## File Structure

```
src/
├── utils/
│   └── api.ts              # Base API client and error handling
├── services/
│   ├── machineService.ts   # Machine API operations
│   └── README.md          # This documentation
├── hooks/
│   ├── useMachines.ts     # Hook for machine list operations
│   └── useMachine.ts      # Hook for individual machine operations
├── config/
│   └── api.ts             # API configuration and constants
└── types/
    └── machine.ts         # Machine type definitions
```

## Configuration

### Environment Variables

Create a `.env.local` file in the root directory:

```bash
# API Configuration
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

### API Configuration

The API configuration is centralized in `src/config/api.ts`:

```typescript
export const API_CONFIG = {
  BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
  ENDPOINTS: {
    MACHINES: '/api/machines',
    MAINTENANCE: '/api/maintenance',
    USERS: '/api/users',
  },
  TIMEOUT: 10000,
  DEFAULT_PAGE_SIZE: 10,
  MAX_PAGE_SIZE: 100,
};
```

## Usage Examples

### Using the Machine Service Directly

```typescript
import { MachineService } from '../services/machineService';

// Fetch machines with pagination and filters
const response = await MachineService.getMachines(1, 10, {
  limit: 10,
  sort: 'created_at_desc',
  due_ppm: true
});

// Create a new machine
const newMachine = await MachineService.createMachine({
  serial_number: 'SN123456',
  customer: 'Acme Corp',
  // ... other required fields
});

// Update a machine
const updatedMachine = await MachineService.updateMachine('SN123456', {
  status: 'maintenance'
});

// Delete a machine
await MachineService.deleteMachine('SN123456');
```

### Using React Hooks

#### Machine List Hook

```typescript
import { useMachines } from '../hooks/useMachines';

function MachineListComponent() {
  const {
  machines,
  total,
  page,
  loading,
  error,
  refetch,
  setPage,
  setFilters
} = useMachines({
  page: 1,
  limit: 10,
  filters: { due_ppm: true, sort: 'created_at_desc' },
  autoFetch: true
});

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      {machines.map(machine => (
        <div key={machine.serial_number}>
          {machine.serial_number} - {machine.customer}
        </div>
      ))}
    </div>
  );
}
```

#### Individual Machine Hook

```typescript
import { useMachine } from '../hooks/useMachine';

function MachineDetailComponent({ serialNumber }: { serialNumber: string }) {
  const {
    machine,
    loading,
    error,
    fetchMachine,
    updateMachine,
    deleteMachine
  } = useMachine();

  useEffect(() => {
    fetchMachine(serialNumber);
  }, [serialNumber]);

  const handleUpdate = async () => {
    const updated = await updateMachine(serialNumber, {
      status: 'maintenance'
    });
    if (updated) {
      console.log('Machine updated successfully');
    }
  };

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;
  if (!machine) return <div>Machine not found</div>;

  return (
    <div>
      <h1>{machine.serial_number}</h1>
      <p>Customer: {machine.customer}</p>
      <p>Status: {machine.status}</p>
      <button onClick={handleUpdate}>Update Status</button>
    </div>
  );
}
```

## Error Handling

The API client provides comprehensive error handling:

```typescript
import { ApiError, handleApiError } from '../utils/api';

try {
  const response = await MachineService.getMachines();
  // Handle success
} catch (error) {
  const apiError = handleApiError(error);
  console.error('API Error:', apiError.message);
  console.error('Status:', apiError.status);
  console.error('Details:', apiError.details);
}
```

## Type Safety

All API operations are fully typed:

```typescript
import { Machine, MachineFilters, MachineListResponse } from '../types/machine';

// Type-safe machine data
const machines: Machine[] = response.data.machines;

// Type-safe filters
const filters: MachineFilters = {
  customer: 'Acme Corp',
  status: 'active'
};

// Type-safe response
const response: ApiResponse<MachineListResponse> = await MachineService.getMachines();
```

## Testing

The API integration is designed to be easily testable:

```typescript
// Mock the API client for testing
jest.mock('../utils/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
  handleApiError: jest.fn(),
}));

// Test the service
describe('MachineService', () => {
  it('should fetch machines', async () => {
    const mockResponse = { data: { machines: [], total: 0 } };
    (apiClient.get as jest.Mock).mockResolvedValue(mockResponse);
    
    const result = await MachineService.getMachines();
    expect(result).toEqual(mockResponse);
  });
});
```

## Best Practices

1. **Always use the hooks for React components** - They provide loading states and error handling
2. **Use TypeScript interfaces** - Ensures type safety across the application
3. **Handle errors gracefully** - Use the error handling utilities provided
4. **Cache responses when appropriate** - Consider implementing caching for frequently accessed data
5. **Validate input data** - Always validate data before sending to the API
6. **Use environment variables** - Configure API URLs for different environments

## Backend API Endpoints

The frontend expects the following backend API endpoints:

- `GET /api/v1/machines` - List machines with pagination and filtering
- `GET /api/v1/machines/{serial_number}` - Get a specific machine
- `POST /api/v1/machines` - Create a new machine
- `PUT /api/v1/machines/{serial_number}` - Update a machine
- `DELETE /api/v1/machines/{serial_number}` - Delete a machine

### Query Parameters for GET /api/v1/machines:

- `limit` (number) - Number of machines to return (default: 10)
- `offset` (number) - Number of machines to skip for pagination
- `sort` (string) - Sort order: "created_at_desc" or "created_at_asc"
- `due_ppm` (boolean) - Filter for machines due for PPM maintenance

All endpoints should return JSON responses with the following structure:

```json
{
  "data": { /* response data */ },
  "message": "Success message",
  "error": null
}
```

For error responses:

```json
{
  "data": null,
  "message": "Error message",
  "error": "Error details"
}
``` 