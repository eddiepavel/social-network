# 🌐 Pulse - Social Network Platform

> *Where your people pulse back.*

A modern, privacy-focused social network built for intentional sharing with circles that matter. **Pulse** enables users to build communities, host groups, share moments, and stay connected through meaningful conversations.

---

## 🎯 Project Overview

Pulse is a full-stack social networking platform that emphasizes user privacy and intentional content sharing. It features a clean, modern interface with support for posts, groups, real-time chat, followers/following relationships, and granular visibility controls.

---

## 🛠️ Tech Stack

### Frontend (`/frontend`)
- **Framework:** [Next.js](https://nextjs.org/) (React-based)
- **Language:** TypeScript
- **UI Components:** Custom component library with modern CSS
- **State Management:** [TanStack Query (React Query)](https://tanstack.com/query) - server state management
- **Real-time Communication:** WebSocket integration for chat and notifications
- **Routing:** Next.js App Router
- **HTTP Client:** Native `fetch` API with custom wrapper
- **Build Tool:** Next.js built-in tooling

### Backend (`/backend`)
- **Language:** [Go](https://go.dev/) (v1.24.6)
- **Database:** [SQLite](https://www.sqlite.org/) with [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) driver
- **Query Generation:** [sqlc](https://sqlc.dev/) - type-safe SQL queries from raw SQL
- **Authentication:** Session-based with bcrypt password hashing and cookies
- **UUID Generation:** `github.com/google/uuid` & `github.com/gofrs/uuid`
- **WebSocket:** Real-time bidirectional communication for chat and notifications
- **Hot Reload:** [Air](https://github.com/air-verse/air) (`.air.toml` config)
- **File Management:** Custom file upload service with automatic cleanup
- **Migrations:** Automated database migrations with versioning
- **API Testing:** Postman collections included

### Deployment & Infrastructure
- **Containerization:** Docker for both frontend and backend services
- **Development:** docker-compose for local development environment

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
  - Foreign key constraints with cascading deletes
  - Automated migrations with up/down scripts

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
- **Authentication:** 
  - Secure registration with email, password, name, date of birth
  - Optional profile fields: avatar, nickname, about me
  - Session-based authentication with secure cookies
  - Persistent login sessions until explicit logout
- **Profiles:** 
  - User profiles with customizable avatars and bio
  - Public vs. private profile settings
  - View user activity, posts, followers, and following lists
  - Profile visibility respects privacy settings
- **Follow System:** 
  - Send follow requests to other users
  - Automatic approval for public profiles
  - Request approval workflow for private profiles
  - Unfollow functionality
  - View followers and following lists with privacy controls
- **Privacy Controls:** Toggle between public and private profile modes

### Content Sharing
- **Posts:** 
  - Create posts with text content and optional images/GIFs
  - Support for JPEG, PNG, and GIF formats
  - Edit and delete your own posts
  - Personalized feed algorithm
- **Visibility Control:** 
  - **Public:** Visible to all users on the network
  - **Semi-Private (Almost Private):** Only visible to your followers
  - **Private:** Visible only to selected followers you choose
  - Profile privacy affects default post visibility
- **Comments:** 
  - Comment on posts you can view
  - Edit and delete your own comments
  - Image/GIF support in comments
- **Reactions:** 
  - React to posts and comments
  - View reaction counts and who reacted
- **Feed Algorithm:** 
  - Personalized feed based on followers
  - Group posts included for your groups
  - Respects privacy settings and visibility rules

### Groups
- **Create & Join:** 
  - Create groups with title, description, and optional image
  - Browse all available groups
  - Request to join private groups
  - Receive invitations from group members
- **Approval System:** 
  - Group creators manage join requests
  - Accept or decline membership requests
  - Invite system for existing members
- **Group Posts:** 
  - Create posts visible only to group members
  - Comment and react within the group
  - Group-specific feed and content
- **Group Events:**
  - Create events with title, description, and date/time
  - RSVP system with "Going" and "Not Going" options
  - Event notifications for all group members
  - View event attendance and responses
- **Group Chat:**
  - Dedicated chat room for each group
  - Real-time messaging for all group members
  - Persistent message history

### Messaging
- **Direct Messages:** 
  - One-on-one private conversations
  - Send messages to users you follow or who follow you
  - Instant delivery for public profiles
- **Group Chat:** 
  - Dedicated chat rooms for each group
  - All group members can participate
  - Real-time message synchronization
- **Real-time Support:** 
  - WebSocket-powered instant messaging
  - Live message delivery without page refresh
  - Connection status indicators
- **Emoji Support:** 
  - Emoji picker for expressive messaging
  - Send and receive emojis in all chats
- **Chat Features:**
  - Unread message tracking
  - Message timestamps
  - Chat room management
  - Persistent message history

### Media & File Management
- **Image Uploads:** 
  - Secure image upload with validation
  - Support for JPEG, PNG, and GIF formats
  - Image handling for posts, comments, profiles, and groups
- **File Storage:**
  - Organized storage structure
  - Temporary storage for upload previews
  - Auto-cleanup of unused uploads (every 5 minutes)
- **Security:**
  - File type validation
  - Size limits and validation
  - Secure file access patterns

### Notifications
- **Real-time Notifications:**
  - WebSocket-powered instant notifications
  - Distinct from chat messages
  - Visible across all pages
- **Notification Types:**
  - Follow requests (for private profiles)
  - Group invitations
  - Group join requests (for group creators)
  - New group events
  - Group membership approvals
- **Notification Management:**
  - Mark as seen/unseen
  - Notification count badges
  - View notification details with user information
  - Delete individual notifications
  - Mark all as read functionality

---

## 🚀 Getting Started

### Prerequisites
- **Node.js** (v18+)
- **Go** (v1.24.6+)
- **SQLite** (usually pre-installed on most systems)
- **Docker** (for containerized deployment)
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
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Run the development server:
   ```bash
   npm run dev
   ```

**Default Frontend Port:** `3000`

### Docker Setup (Recommended)

The project is fully containerized with Docker for easy deployment:

1. Build and run both services:
   ```bash
   docker-compose up --build
   ```

2. Access the application:
   - **Frontend:** `http://localhost:3000`
   - **Backend API:** `http://localhost:8000`

3. Stop the services:
   ```bash
   docker-compose down
   ```

**Docker Images:**
- `pulse-frontend` - Next.js frontend application
- `pulse-backend` - Go backend server

Both containers are configured for optimal development and production environments.

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
│   │   │   ├── auth.go      # Authentication
│   │   │   ├── chat.go      # Real-time messaging
│   │   │   ├── followers.go # Follow system
│   │   │   ├── groups.go    # Groups & events
│   │   │   ├── notifications.go
│   │   │   ├── posts.go     # Posts, comments, reactions
│   │   │   ├── users.go     # User profiles
│   │   │   ├── upload.go    # File uploads
│   │   │   └── websocket.go # WebSocket connection
│   │   ├── middleware/      # Middleware stack
│   │   │   ├── auth.go      # Session authentication
│   │   │   ├── cors.go      # CORS handling
│   │   │   ├── guest.go     # Guest session
│   │   │   └── log.go       # Request logging
│   │   ├── models/          # Response DTOs
│   │   ├── services/        # Business logic
│   │   ├── helpers/         # Input validation & utilities
│   │   ├── utils/           # Response helpers
│   │   ├── websocket/       # WebSocket manager
│   │   │   ├── client.go
│   │   │   ├── event.go
│   │   │   └── manager.go
│   │   └── constants/       # App constants
│   ├── pkg/
│   │   ├── db/
│   │   │   ├── migrations/  # Database migrations
│   │   │   │   └── sqlite/
│   │   │   │       ├── up/  # Migration up scripts
│   │   │   │       └── down/# Migration down scripts
│   │   │   ├── queries/     # SQL queries by feature
│   │   │   │   ├── users/
│   │   │   │   ├── posts/
│   │   │   │   ├── groups/
│   │   │   │   ├── chat/
│   │   │   │   └── ...
│   │   │   └── sqlite/      # DB connection & migrations
│   │   └── environment/     # Environment config
│   ├── storage/uploads/     # File uploads
│   ├── go.mod               # Go dependencies
│   ├── sqlc.yml            # sqlc configuration
│   ├── .air.toml           # Hot reload config
│   ├── Dockerfile          # Backend container
│   └── *.postman_*         # API testing collections
├── frontend/
│   ├── app/                 # Next.js app directory
│   │   ├── (auth)/         # Auth pages (login, register)
│   │   ├── (app)/          # Protected routes
│   │   │   ├── chat/       # Messaging UI
│   │   │   ├── feed/       # Main feed
│   │   │   ├── groups/     # Group pages
│   │   │   └── profile/    # User profiles
│   │   ├── layout.tsx      # Root layout
│   │   ├── providers.tsx   # App providers (React Query, WS)
│   │   └── globals.css     # Global styles
│   ├── components/          # Reusable UI components
│   │   ├── ChatRoom.tsx
│   │   ├── PostCard.tsx
│   │   ├── NotificationDropdown.tsx
│   │   └── ... (30+ components)
│   ├── hooks/              # Custom React hooks
│   │   ├── useSession.ts   # Auth state
│   │   ├── useWebSocket.tsx # WebSocket connection
│   │   └── useInfiniteScroll.ts
│   ├── lib/                # Utilities & API client
│   │   ├── api.ts          # HTTP client wrapper
│   │   ├── types.ts        # TypeScript types
│   │   └── utils.ts        # Helper functions
│   ├── next.config.mjs     # Next.js config
│   ├── tsconfig.json       # TypeScript config
│   └── Dockerfile          # Frontend container
├── docker-compose.yml      # Multi-container orchestration
└── endpoints.md            # API documentation
```

---

## 👥 Team

This is a collaborative project. For questions or assistance:
- [**Giannis Geo**](https://platform.zone01.gr/git/ggeorgako)
- [**Eddie**](https://platform.zone01.gr/git/epavel) 
- [**Giannis Vos**](https://platform.zone01.gr/git/ivossos)
- [**Uipko**](https://platform.zone01.gr/git/ustikker)
- [**Pavlos**](https://platform.zone01.gr/git/pkerasid)
- [**Giorgos**](https://platform.zone01.gr/git/gpavrian)

---

## 🔑 Key Technical Highlights

### Authentication & Sessions
- **Session-based authentication** with secure HTTP-only cookies
- **bcrypt password hashing** for secure credential storage
- Sessions persist until explicit logout
- Middleware-based route protection

### Database Migrations
- **Automated migration system** using numbered SQL files
- Migrations run automatically on application start
- Up and down scripts for version control
- Migration tracking in `schema_migrations` table
- Example migration structure:
  ```
  000001_create_users_table.sql
  000002_create_sessions_table.sql
  000003_create_followers_table.sql
  ...
  ```

### Real-time Features
- **WebSocket connection** at `/ws/connect`
- Bi-directional communication for chat and notifications
- Automatic reconnection with exponential backoff
- Event-based message routing
- Connection state management

### Privacy & Security
- Profile-level privacy (public/private)
- Post-level visibility controls
- Follow request approval system
- Secure file upload validation
- Foreign key constraints with cascading deletes

## 📝 Notes

- **Development Branch:** Active development happens on the `development` branch
- **Database:** SQLite with automated migrations applied on startup
- **File Cleanup:** Unused uploads are automatically cleaned every 5 minutes
- **Session Management:** Sessions are stored server-side in the database
- **WebSocket:** Real-time connection at `/ws/connect` endpoint
- **Type Safety:** Full TypeScript on frontend, type-safe SQL with sqlc on backend

