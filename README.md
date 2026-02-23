# 🌐 Pulse - Social Network Platform

> *Where your people pulse back.*

A modern, privacy-focused social network built for intentional sharing with circles that matter. **Pulse** enables users to build communities, host groups, share moments, and stay connected through meaningful conversations.

---

## 🎯 Project Overview

Pulse is a full-stack social networking platform that emphasizes user privacy and intentional content sharing. It features a clean, modern interface with support for posts, groups, real-time chat, followers/following relationships, and granular visibility controls.

**Current Branch:** `development`

---

## 🛠️ Tech Stack

### Frontend (`/frontnew`)
- **Framework:** [Next.js](https://nextjs.org/) (React-based)
- **Language:** TypeScript
- **UI Components:** Custom component library with modern CSS
- **State Management:** [TanStack Query (React Query)](https://tanstack.com/query) - server state management
- **Routing:** Next.js App Router
- **Fonts:** Google Fonts (Fraunces, Space Grotesk)
- **HTTP Client:** Native `fetch` API with custom wrapper
- **Build Tool:** Next.js built-in tooling

### Backend (`/backend`)
- **Language:** [Go](https://go.dev/) (v1.24.6)
- **Database:** [SQLite](https://www.sqlite.org/) with [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) driver
- **Query Generation:** [sqlc](https://sqlc.dev/) - type-safe SQL queries from raw SQL
- **Authentication:** Session-based with bcrypt password hashing
- **UUID Generation:** `github.com/google/uuid` & `github.com/gofrs/uuid`
- **Hot Reload:** [Air](https://github.com/air-verse/air) (`.air.toml` config)
- **File Management:** Custom file upload service with automatic cleanup
- **API Testing:** Postman collections included

---

## 🏗️ Architecture & Techniques

### Backend Architecture
- **Clean Architecture Pattern:** Separation of concerns with distinct layers
  - `cmd/` - Application entry point
  - `app/` - App context & dependency injection
  - `internal/` - Private application code
    - `handlers/` - HTTP request handlers
    - `middleware/` - Authentication, CORS, logging
    - `models/` - Response DTOs
    - `services/` - Business logic (file uploads, image processing)
    - `helpers/` - Utility functions (UUID conversion, validation)
    - `utils/` - HTTP response helpers
  - `pkg/` - Reusable packages
    - `db/queries/` - SQL queries organized by feature
    - `db/sqlite/` - Database connection & management
  - `storage/` - File uploads directory

- **Type-Safe SQL with sqlc:** All database queries are written in pure SQL and compiled to type-safe Go code, providing compile-time safety and excellent performance

- **Middleware Stack:**
  - Session authentication
  - CORS handling
  - Request logging
  - Context injection

- **File Upload Service:** 
  - Automatic cleanup of expired temporary files
  - HMAC-signed URLs for secure file access
  - Support for images with validation

- **Database Design:**
  - Normalized relational schema
  - UUID-based primary keys
  - Support for complex relationships (followers, group members, post visibility)

### Frontend Architecture
- **App Router Structure:**
  - `(auth)/` - Authentication pages (login, register)
  - `(app)/` - Protected application routes with shared layout
  - Component-based architecture with reusable UI components

- **Data Fetching Strategy:**
  - Server state managed by TanStack Query
  - Optimistic updates for better UX
  - Automatic cache invalidation
  - Retry logic and error handling

- **Type Safety:**
  - Full TypeScript coverage
  - Typed API responses
  - Type-safe routing with Next.js

- **Styling Approach:**
  - CSS custom properties (CSS variables) for theming
  - Utility-first component styles
  - Responsive grid system
  - Warm, accessible color palette

---

## ✨ Key Features

### User Management
- **Authentication:** Secure registration and login with session management
- **Profiles:** User profiles with avatars, bio, and personal details
- **Follow System:** Follow/unfollow users with request approval for private accounts
- **Privacy Controls:** Public, semi-private, and private visibility settings

### Content Sharing
- **Posts:** Create posts with text and optional images
- **Visibility Control:** Granular control over who can see your posts
- **Comments:** Nested commenting system
- **Reactions:** Like/react to posts and comments
- **Feed Algorithm:** Personalized feed based on followers and group memberships

### Groups
- **Create & Join:** Create public or private groups
- **Invitations:** Invite users or request to join groups
- **Approval System:** Group creators can approve/reject join requests
- **Group Posts:** Share content within specific groups

### Messaging
- **Direct Messages:** One-on-one conversations
- **Group Chat:** Multi-user conversations
- **Real-time Support:** WebSocket support for live messaging
- **Unread Tracking:** Track unread message counts

### Media
- **Image Uploads:** Secure image upload with validation
- **Temporary Storage:** Auto-cleanup of unused uploads
- **Signed URLs:** Secure file access with HMAC signatures

---

## 🚀 Getting Started

### Prerequisites
- **Node.js** (v18+)
- **Go** (v1.24.6+)
- **SQLite** (usually pre-installed on most systems)
- **sqlc** (for regenerating DB queries) - [Install sqlc](https://docs.sqlc.dev/en/stable/overview/install.html)

### Backend Setup

1. Navigate to the backend directory:
   ```bash
   cd backend
   ```

2. Copy environment variables:
   ```bash
   cp .env.example .env
   ```

3. Install Go dependencies:
   ```bash
   go mod download
   ```

4. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

   Or use Air for hot reload:
   ```bash
   air
   ```

**Default Backend Port:** `8000`

### Frontend Setup

1. Navigate to the frontend directory:
   ```bash
   cd frontnew
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Run the development server:
   ```bash
   PORT=5173 npm run dev
   ```

**Default Frontend Port:** `5173`

---

## 📚 API Documentation

API endpoints are documented in [`endpoints.md`](./endpoints.md).

**Postman Collections:** Available in the `backend/` directory:
- `Social Network.postman_collection.json`
- `Social Network.postman_environment.json`

Import these into Postman for easy API testing.

---

## 🧪 Development Workflow

### Adding New Features (Backend)

We follow a consistent pattern for adding new features. See [`backend/BackendNewFeaturesSteps.md`](backend/BackendNewFeaturesSteps.md) for detailed instructions.

**Quick Overview:**
1. Create SQL queries in `backend/pkg/db/queries/<feature>/`
2. Add queries to `backend/sqlc.yml`
3. Run `sqlc generate` to generate Go code
4. Create route handlers in `backend/internal/handlers/`
5. Register routes in `backend/internal/routes/`
6. Validate inputs using `backend/internal/helpers/inputs.go` pattern
7. Use `backend/internal/utils/responder` for consistent responses
8. Update Postman collection

### Code Standards

- **Go:** Follow standard Go conventions, use `gofmt`
- **TypeScript:** Strict mode enabled, ESLint recommended
- **SQL:** Pure SQL queries, no ORM abstractions
- **Validation:** Always validate user input on the backend
- **Error Handling:** Structured logging with proper error responses

---

## 📦 Project Structure

```
social-network/
├── backend/
│   ├── cmd/server/          # Application entry point
│   ├── app/                 # App context & DI
│   ├── internal/            # Private application code
│   │   ├── handlers/        # HTTP handlers
│   │   ├── middleware/      # Middleware stack
│   │   ├── models/          # Response models
│   │   ├── services/        # Business logic
│   │   ├── helpers/         # Utilities
│   │   └── utils/           # Response helpers
│   ├── pkg/
│   │   └── db/
│   │       ├── queries/     # SQL queries by feature
│   │       └── sqlite/      # DB connection
│   ├── storage/uploads/     # File uploads
│   ├── go.mod               # Go dependencies
│   ├── sqlc.yml            # sqlc configuration
│   └── .air.toml           # Hot reload config
├── frontnew/
│   ├── app/                 # Next.js app directory
│   │   ├── (auth)/         # Auth pages
│   │   ├── (app)/          # Protected routes
│   │   ├── layout.tsx      # Root layout
│   │   ├── providers.tsx   # App providers
│   │   └── globals.css     # Global styles
│   ├── components/          # Reusable UI components
│   ├── hooks/              # Custom React hooks
│   ├── lib/                # Utilities & API client
│   ├── next.config.mjs     # Next.js config
│   └── tsconfig.json       # TypeScript config
└── endpoints.md            # API documentation
```

---

## 👥 Team

This is a collaborative project. For questions or assistance:
- **Giannis** (Discord)
- **Eddie** (Discord)

---

## 📝 Notes

- **Development Branch:** Active development happens on the `development` branch
- **Database:** SQLite is used for simplicity; migrations are handled manually
- **File Cleanup:** Unused uploads are automatically cleaned every 5 minutes
- **Session Management:** Sessions are stored server-side; adjust timeout in config
- **WebSocket:** Real-time chat available at `/ws/chat` endpoint

---

## 🔮 Future Enhancements

- [ ] Notifications system
- [ ] Advanced search functionality
- [ ] Stories/ephemeral content
- [ ] Video upload support
- [ ] Progressive Web App (PWA) support
- [ ] Email notifications
- [ ] Advanced moderation tools
- [ ] Analytics dashboard

---

## 📄 License

*To be determined by the team.*

---

**Happy coding! 🚀**

*Built with ❤️ by the Pulse team.*
