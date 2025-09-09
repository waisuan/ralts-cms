# Development Plan: User Approval System

## Overview

Implementation of a user approval system that allows users with "ADMIN" role to manage user statuses through a dedicated admin interface accessible via the user menu dropdown.

## Current State Analysis

✅ **What's already in place:**
- User model with status constants (`pending_approval`, `approved`, `suspended`, `inactive`)
- Role-based system with `ADMIN` and `NON_ADMIN` roles
- Authentication middleware with JWT tokens
- User repository with basic CRUD operations
- Frontend auth context with role information
- UserMenu component that can be extended

## Development Plan

### Phase 1: Backend API Development

#### 1.1 Extend User Repository
- **Add new methods to `Repository` interface:**
  - `ListUsers(ctx context.Context, limit, offset int) ([]*User, error)`
  - `GetByID(ctx context.Context, id int64) (*User, error)`
  - `UpdateStatus(ctx context.Context, userID int64, status string) error`
  - `UpdateMultipleStatuses(ctx context.Context, userIDs []int64, status string) error`

#### 1.2 Create Admin Authorization Middleware
- **New file:** `internal/middlewares/admin.go`
- **Function:** `AdminOnlyMiddleware()` - validates that authenticated user has `ADMIN` role
- **Integration:** Chain with existing `AuthenticationMiddleware`

#### 1.3 Create Admin User Management Handler
- **New file:** `internal/handlers/admin_users.go`
- **Endpoints:**
  - `GET /api/v1/admin/users` - List all users with pagination
  - `PUT /api/v1/admin/users/{id}/status` - Update single user status
  - `PUT /api/v1/admin/users/bulk-status` - Update multiple user statuses

#### 1.4 Update Router Configuration
- **Add admin routes** in `internal/router/router.go`
- **Apply admin middleware** to new endpoints

### Phase 2: Frontend Development

#### 2.1 Create Admin User Management Page
- **New file:** `ui/src/app/admin/users/page.tsx`
- **Features:**
  - User list with pagination
  - Multi-select functionality
  - Status update dropdowns
  - Bulk actions for status changes

#### 2.2 Extend UserMenu Component
- **Modify:** `ui/src/components/UserMenu.tsx`
- **Add:** "Manage Users" menu item (visible only to ADMIN role)
- **Navigation:** Link to `/admin/users`

#### 2.3 Create Admin User Service
- **New file:** `ui/src/services/adminUserService.ts`
- **Methods:**
  - `getUsers(limit, offset)` - Fetch user list
  - `updateUserStatus(userId, status)` - Update single user
  - `updateMultipleUserStatuses(userIds, status)` - Bulk update

#### 2.4 Create User Management Components
- **New files:**
  - `ui/src/components/admin/UserList.tsx` - Main user listing
  - `ui/src/components/admin/UserRow.tsx` - Individual user row
  - `ui/src/components/admin/BulkActions.tsx` - Bulk action controls
  - `ui/src/components/admin/StatusBadge.tsx` - Status display component

### Phase 3: Security & Validation

#### 3.1 Backend Security
- **Input validation** for status values (only allow valid constants)
- **Authorization checks** in all admin endpoints
- **Audit logging** for user status changes
- **Rate limiting** on admin endpoints

#### 3.2 Frontend Security
- **Role-based UI rendering** (hide admin features from non-admins)
- **Client-side validation** before API calls
- **Error handling** for unauthorized access

### Phase 4: Testing & Documentation

#### 4.1 Backend Tests
- **Repository tests** for new methods
- **Handler tests** for admin endpoints
- **Middleware tests** for admin authorization
- **Integration tests** for complete flow

#### 4.2 Frontend Tests
- **Component tests** for new admin components
- **Service tests** for admin user service
- **E2E tests** for user approval workflow

## Technical Implementation Details

### Database Schema
The existing `users` table already supports the required fields:
- `role` (ADMIN/NON_ADMIN)
- `status` (pending_approval/approved/suspended/inactive)
- `approved` (boolean field)

### API Endpoints Structure
```
GET /api/v1/admin/users?limit=50&offset=0
PUT /api/v1/admin/users/123/status
PUT /api/v1/admin/users/bulk-status
```

### Frontend Routing
```
/admin/users - Admin user management page (protected by role)
```

### State Management
- Use React state for user list and selections
- Implement optimistic updates for better UX
- Handle loading states and error scenarios

## Security Considerations

1. **Role Verification:** Both frontend and backend must verify ADMIN role
2. **Input Sanitization:** Validate all status updates against allowed constants
3. **Audit Trail:** Log all administrative actions for compliance
4. **Rate Limiting:** Prevent abuse of admin endpoints

## User Experience Flow

1. **Admin Login:** Admin user logs in with ADMIN role
2. **Menu Access:** "Manage Users" appears in UserMenu dropdown
3. **User List:** Admin sees paginated list of all registered users
4. **Selection:** Admin can select individual users or multiple users
5. **Status Update:** Admin changes status via dropdown or bulk action
6. **Confirmation:** System confirms changes and updates UI

## Implementation Notes

This plan leverages the existing architecture patterns and follows the clean architecture principles already established in the codebase. The implementation will be modular, testable, and secure.

### Key Design Decisions

- **Middleware Chain:** Admin middleware will build upon existing auth middleware
- **Repository Pattern:** Extend existing user repository interface
- **Component Architecture:** Create reusable admin components following existing patterns
- **API Design:** RESTful endpoints following established conventions
- **Security First:** Role-based access control at every layer

### Status Values

The system will support these user statuses:
- `pending_approval` - New users awaiting admin approval
- `approved` - Active users who can log in
- `suspended` - Temporarily disabled users
- `inactive` - Permanently disabled users

### Bulk Operations

Admins can perform bulk status updates by:
1. Selecting multiple users via checkboxes
2. Choosing desired status from bulk action dropdown
3. Confirming the operation
4. System updates all selected users atomically

## Implementation Checklist

### Phase 1: Backend Foundation

#### Repository Layer
- [x] Add `ListUsers(ctx context.Context, limit, offset int) ([]*User, int, error)` to Repository interface
- [x] Add `GetByID(ctx context.Context, id int64) (*User, error)` to Repository interface
- [x] Add `UpdateStatus(ctx context.Context, userID int64, status string) error` to Repository interface
- [x] Add `UpdateMultipleStatuses(ctx context.Context, userIDs []int64, status string) error` to Repository interface
- [x] Implement `ListUsers` method in `internal/users/repository.go`
- [x] Implement `GetByID` method in `internal/users/repository.go`
- [x] Implement `UpdateStatus` method in `internal/users/repository.go`
- [x] Implement `UpdateMultipleStatuses` method in `internal/users/repository.go`
- [x] Add validation in `UpdateStatus` to only allow valid status constants
- [x] Add validation in `UpdateMultipleStatuses` to only allow valid status constants

#### Middleware Layer
- [x] Create `internal/middlewares/admin.go` file
- [x] Implement `AdminOnlyMiddleware()` function that checks user role is "ADMIN"
- [x] Add helper function to extract user ID from JWT token in auth middleware
- [x] Update `UserContext` struct to include user ID and role information
- [x] Modify `AuthenticationMiddleware` to fetch and store user role in context

#### Handler Layer
- [x] ~~Create `internal/handlers/admin_users.go` file~~ (Extended existing UsersHandler instead)
- [x] ~~Create `AdminUsersHandler` struct with dependencies~~ (Extended existing UsersHandler instead)
- [x] Implement `ListUsers` handler for `GET /api/v1/admin/users`
- [x] Implement `UpdateUserStatus` handler for `PUT /api/v1/admin/users/{id}/status`
- [x] Implement `BulkUpdateStatus` handler for `PUT /api/v1/admin/users/bulk-status`
- [x] Add input validation for status values in all handlers
- [x] Add proper error handling and HTTP status codes
- [x] Add request/response logging for audit trail

#### Router Configuration
- [x] Add admin routes to `internal/router/router.go`
- [x] Chain `AdminOnlyMiddleware` with `AuthenticationMiddleware` for admin routes
- [x] Test that non-admin users get 403 Forbidden on admin endpoints

### Phase 2: Frontend Implementation

#### Service Layer
- [ ] Create `ui/src/services/adminUserService.ts` file
- [ ] Implement `getUsers(limit?: number, offset?: number)` method
- [ ] Implement `updateUserStatus(userId: number, status: string)` method
- [ ] Implement `updateMultipleUserStatuses(userIds: number[], status: string)` method
- [ ] Add proper error handling for API responses
- [ ] Add TypeScript interfaces for API responses

#### User Menu Enhancement
- [ ] Modify `ui/src/components/UserMenu.tsx` to show "Manage Users" for ADMIN role
- [ ] Add navigation link to `/admin/users` page
- [ ] Add appropriate icon for the menu item
- [ ] Ensure menu item is only visible to users with ADMIN role

#### Admin Page Structure
- [ ] Create `ui/src/app/admin/` directory
- [ ] Create `ui/src/app/admin/users/` directory
- [ ] Create `ui/src/app/admin/users/page.tsx` main admin page
- [ ] Add role-based route protection to admin page
- [ ] Create basic page layout with title and breadcrumbs

#### Core Components
- [ ] Create `ui/src/components/admin/` directory
- [ ] Create `StatusBadge.tsx` component for displaying user status
- [ ] Create `UserRow.tsx` component for individual user display
- [ ] Create `UserList.tsx` component for user table with selection
- [ ] Create `BulkActions.tsx` component for bulk status updates
- [ ] Add pagination controls to `UserList.tsx`
- [ ] Add loading states to all components
- [ ] Add error handling and display to all components

#### State Management
- [ ] Implement user list state management in admin page
- [ ] Implement multi-select functionality with checkboxes
- [ ] Implement bulk action state (selected users, target status)
- [ ] Add optimistic updates for better UX
- [ ] Add confirmation dialogs for bulk operations

### Phase 3: Integration & Polish

#### API Integration
- [ ] Connect frontend components to backend services
- [ ] Test all CRUD operations work end-to-end
- [ ] Test pagination works correctly
- [ ] Test bulk operations work correctly
- [ ] Test error scenarios (network errors, validation errors)

#### Security Verification
- [ ] Verify non-admin users cannot access admin endpoints (backend)
- [ ] Verify non-admin users cannot see admin UI elements (frontend)
- [ ] Test that invalid status values are rejected
- [ ] Test that unauthorized requests return proper error codes

#### User Experience
- [ ] Add loading spinners during API calls
- [ ] Add success/error toast notifications
- [ ] Add confirmation dialogs for destructive actions
- [ ] Ensure responsive design works on mobile devices
- [ ] Add keyboard navigation support

### Phase 4: Testing & Documentation

#### Backend Tests
- [x] Write unit tests for new repository methods
- [x] Write unit tests for admin middleware
- [x] Write unit tests for admin handlers
- [ ] Write integration tests for admin API endpoints
- [ ] Test error scenarios and edge cases

#### Frontend Tests
- [ ] Write unit tests for admin service methods
- [ ] Write component tests for all new admin components
- [ ] Write integration tests for admin page functionality
- [ ] Test role-based access control in UI

#### Documentation
- [ ] Update API documentation with new admin endpoints
- [ ] Document the admin user management workflow
- [ ] Update README.md with admin functionality
- [ ] Document any new environment variables or configuration

### Deployment Checklist
- [ ] Verify database schema supports all required fields
- [ ] Test with production-like data volumes
- [ ] Verify admin user exists and has proper role
- [ ] Test in staging environment
- [ ] Create deployment rollback plan
