This is a [Next.js](https://nextjs.org) project that serves as the frontend interface for the Ralts-CMS machine and maintenance management system.

## Prerequisites

Before running the frontend, ensure you have:

- Node.js 18 or later
- npm, yarn, pnpm, or bun package manager
- The Ralts-CMS backend API running (see [root README](../README.md) for backend setup)

## Getting Started

### 1. Install Dependencies

```bash
npm install
# or
yarn install
# or
pnpm install
# or
bun install
```

### 2. Configure Environment Variables

Create a `.env.local` file in the `ui/` directory:

```env
# Backend API Configuration
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_API_TIMEOUT=30000

# Authentication
NEXT_PUBLIC_JWT_TOKEN_KEY=ralts_cms_token
```

### 3. Start the Development Server

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

The page auto-updates as you edit files thanks to Next.js hot reloading.

## Backend Integration

This frontend application integrates with the Go backend API to provide a complete machine management solution.

### API Communication

- **Base URL**: Backend runs on `http://localhost:8080` by default
- **Authentication**: Uses JWT tokens for secure API access
- **Content Type**: All requests use `application/json`
- **Error Handling**: Comprehensive error handling for API failures and network issues

### Key Integration Points

#### Authentication Flow
1. User provides JWT token (stored in localStorage/sessionStorage)
2. Token included in `Authorization: Bearer <token>` header for all API requests
3. Frontend handles token expiration and redirects to login when needed

#### Machine Management
- **GET** `/machines` - List machines with pagination and sorting
- **GET** `/machines/{serial_number}` - Get specific machine details
- **POST** `/machines` - Create new machine
- **PUT** `/machines` - Update existing machine
- **DELETE** `/machines/{serial_number}` - Delete machine

#### Maintenance Records
- **GET** `/machines/{serial_number}/maintenance` - List maintenance records
- **POST** `/machines/{serial_number}/maintenance` - Create maintenance record
- **PUT** `/machines/{serial_number}/maintenance` - Update maintenance record
- **DELETE** `/machines/{serial_number}/maintenance/{work_order_number}` - Delete maintenance record

#### Health Monitoring
- **GET** `/health` - Check backend API health status

### API Client Services

The application includes dedicated service modules in `src/services/` for:
- **Auth Service**: JWT token management and authentication
- **Machine Service**: Machine CRUD operations
- **Maintenance Service**: Maintenance record management
- **API Client**: Centralized HTTP client with error handling

### Development Workflow

1. **Start Backend**: Ensure the Go backend is running on port 8080
2. **Start Frontend**: Run the Next.js development server on port 3000
3. **Test Integration**: Use browser dev tools to monitor API calls
4. **Hot Reloading**: Both frontend and backend support hot reloading for rapid development

## Common npm Commands

| Command                | Description                         |
| ---------------------- | ----------------------------------- |
| `npm run dev`          | Start the development server        |
| `npm run build`        | Build the app for production        |
| `npm start`            | Start the production server         |
| `npm run lint`         | Run ESLint for code quality         |
| `npm run format`       | Format code with Prettier           |
| `npm run format:check` | Check code formatting with Prettier |
| `npm test`             | Run unit tests with Jest            |

## Architecture Overview

### Tech Stack

- **Framework**: Next.js 15 with App Router
- **Language**: TypeScript for type safety
- **Styling**: Tailwind CSS for responsive design
- **Testing**: Jest with Testing Library
- **State Management**: React hooks and context
- **HTTP Client**: Native fetch with custom service layers

### Project Structure

```
ui/src/
├── app/                 # Next.js app router pages and layouts
├── components/          # Reusable UI components
│   ├── ui/             # Base UI components (buttons, forms, etc.)
│   ├── machine/        # Machine-specific components
│   └── maintenance/    # Maintenance-specific components
├── services/           # API client services
│   ├── api.ts         # Base API client
│   ├── auth.ts        # Authentication service
│   ├── machine.ts     # Machine API service
│   └── maintenance.ts # Maintenance API service
├── hooks/              # Custom React hooks
├── types/              # TypeScript type definitions
├── contexts/           # React context providers
├── config/             # Application configuration
└── utils/              # Utility functions
```

## Development Guidelines

### API Integration Best Practices

1. **Error Handling**: All API calls include proper error handling with user-friendly messages
2. **Loading States**: UI provides loading indicators during API requests
3. **Type Safety**: All API responses are typed with TypeScript interfaces
4. **Authentication**: JWT tokens are automatically included in API requests
5. **Retry Logic**: Failed requests include retry mechanisms for network issues

### Component Guidelines

- Use TypeScript for all components
- Implement proper error boundaries
- Follow responsive design principles
- Include comprehensive unit tests
- Use semantic HTML and accessibility features

## Troubleshooting

### Common Issues

#### Backend Connection Failed
- Verify the backend API is running on port 8080
- Check the `NEXT_PUBLIC_API_BASE_URL` environment variable
- Ensure CORS is properly configured in the backend

#### Authentication Errors
- Verify JWT token is valid and not expired
- Check token format in browser dev tools
- Ensure backend JWT_SECRET matches token signature

#### Build Errors
```bash
# Clear Next.js cache
rm -rf .next
npm run build

# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

## Learn More

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API
- [TypeScript Handbook](https://www.typescriptlang.org/docs/) - TypeScript language guide
- [Tailwind CSS Docs](https://tailwindcss.com/docs) - utility-first CSS framework
- [Testing Library](https://testing-library.com/) - testing utilities for React
