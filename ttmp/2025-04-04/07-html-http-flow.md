# Friday Talks Application Flow Analysis

## Overview

The Friday Talks application is a web-based platform for managing and scheduling technical talks. It uses a Go backend with Chi router, SQLite database, and follows a clean architecture pattern with handlers, services, and repositories.

## Core Components

### Backend Structure
- Main application (`main.go`)
- Handlers (in `internal/handlers/`)
- Services (in `internal/services/`)
- Models and Repositories (in `internal/models/`)
- Authentication (in `internal/auth/`)

### Key Services
1. Authentication Service (`auth.NewAuth`)
2. Scheduler Service (`services.NewSchedulerService`)
3. Notification Service (`services.NewNotificationService`)

## HTTP Routes and Page Flow

### 1. Authentication Flow
```
GET/POST /login
  ├── Displays login form (GET)
  └── Processes login credentials (POST)
      └── Sets JWT token on success

GET/POST /register
  ├── Shows registration form (GET)
  └── Creates new user account (POST)
      └── Redirects to login

GET /logout
  └── Clears authentication and redirects to home
```

### 2. Main Navigation Flow
```
GET / (Home)
  └── Displays upcoming talks and highlights

GET /calendar
  └── Shows calendar view of scheduled talks

GET /profile (Authenticated)
  └── User profile management
```

### 3. Talk Management Flow
```
GET /talks
  └── Lists all talks

GET /talks/{id}
  └── Shows detailed view of specific talk

GET/POST /talks/propose (Authenticated)
  ├── Shows talk proposal form (GET)
  └── Submits new talk proposal (POST)

GET /talks/my (Authenticated)
  └── Lists user's talks (proposed/scheduled)

POST /talks/{id}/vote (Authenticated)
  └── Records user vote on talk

POST /talks/{id}/attend (Authenticated)
  └── Manages attendance for talk

POST /talks/{id}/feedback (Authenticated)
  └── Submits feedback for completed talks

GET/POST /talks/{id}/edit (Authenticated)
  ├── Shows edit form for talk (GET)
  └── Updates talk details (POST)

POST /talks/{id}/cancel (Authenticated)
  └── Cancels scheduled talk

POST /talks/{id}/complete (Authenticated)
  └── Marks talk as completed

GET/POST /talks/{id}/schedule (Authenticated)
  ├── Shows scheduling form (GET)
  └── Sets talk date/time (POST)
```

## Middleware Chain

The application uses the following middleware stack (in order):
1. Request ID generation
2. Real IP detection
3. Logging
4. Error recovery
5. Request timeout (60s)
6. Authentication state

## Static File Serving

Static files are embedded in the binary and served from the `/static` route prefix. This includes:
- CSS stylesheets
- JavaScript files
- Images and other assets

## Database Interactions

The application uses SQLite with the following key repositories:
- User Repository: Account management
- Talk Repository: Talk proposals and details
- Vote Repository: User votes on talks
- Attendance Repository: Talk attendance tracking
- Resource Repository: Additional talk resources/materials

## Authentication System

Uses JWT (JSON Web Tokens) for authentication with:
- Token-based session management
- Protected route middleware (`RequireAuth`)
- User context injection

## Page-Specific Flows

### 1. Home Page Flow
- Displays upcoming talks
- Shows recent activity
- Quick links to propose/view talks

### 2. Talk Proposal Flow
1. User submits talk proposal
2. Other users can vote
3. Talk gets scheduled based on votes
4. Attendees can register
5. After completion, feedback can be submitted

### 3. Calendar Integration
- Visual calendar of scheduled talks
- Integration with scheduling system
- Attendance tracking

### 4. Profile Management
- User information updates
- Talk history
- Attendance records
- Voting history

## Error Handling

The application implements comprehensive error handling:
- Middleware-level recovery
- Context timeouts
- Database transaction management
- User-friendly error pages

## Future Considerations

1. Email Notifications
   - Currently configured but disabled
   - Ready for SMTP integration

2. Resource Management
   - Support for talk materials
   - File uploads

3. Feedback System
   - Post-talk surveys
   - Rating system

## Security Considerations

1. Authentication
   - JWT token management
   - Secure password handling
   - Session management

2. Authorization
   - Route protection
   - Resource access control
   - User role management

3. Data Protection
   - Input validation
   - SQL injection prevention
   - XSS protection

## Template Structure and HTML Components

### Template Engine

The application uses the Go `templ` templating engine, which allows for strongly-typed HTML templates with Go logic embedded. All templates are defined in the `internal/templates` directory and are organized by functionality.

### Core Templates

1. **Layout Template (`layout.templ`)**
   - Acts as the base template for all pages
   - Defines common HTML structure:
     - DOCTYPE, head with meta tags, Bootstrap and HTMX includes
     - Navbar with dynamic menu based on user authentication status
     - Container for content
     - Footer
   - Provides common components:
     - Alert messages
     - Loading spinner

2. **Home Template (`home.templ`)**
   - Landing page with hero section
   - Sections for upcoming, proposed, and recent talks
   - Feature cards explaining the talk flow process
   - Reusable `TalkCard` component for displaying talk previews
   - Dynamic content based on user authentication state

3. **Auth Templates (`auth.templ`)**
   - Login form with email and password fields
   - Registration form with validation
   - Profile management form with optional password change functionality
   - Success/error message handling

4. **Talks Templates (`talks.templ`)**
   - List view with filtering tabs (All, Scheduled, Proposed, Past)
   - Detail view with:
     - Talk information (title, description, status)
     - Voting interface for proposed talks
     - Attendance management for scheduled talks
     - Feedback submission for completed talks
     - Resource listing
   - Talk management forms:
     - Proposal form
     - Edit form
     - Scheduling form

5. **Calendar Template (`calendar.templ`)**
   - Month-based calendar view
   - Navigation between months
   - Visual indicators for days with scheduled talks
   - Links to talk details from calendar entries

### Component Reuse

The application makes extensive use of reusable components:

1. **TalkCard Component**
   - Used across multiple pages (home, talks list)
   - Displays key talk information in a card format
   - Includes status badge, title, presenter, and date

2. **Alert Component**
   - Provides consistent styling for success/error messages
   - Used throughout the application for user feedback

3. **LoadingSpinner Component**
   - Used for HTMX loading states
   - Provides visual feedback during async operations

### UI Framework and Libraries

1. **Bootstrap 5**
   - Provides responsive grid system
   - Styling for forms, cards, tables, and other components
   - Responsive navigation

2. **HTMX**
   - Used for dynamic interactions without full page reloads
   - Powers the voting, attendance, and other interactive features
   - Enhances user experience with minimal JavaScript

### CSS Customizations

The application includes custom CSS for specific components:
- Talk cards with hover effects
- Calendar day styling with visual indicators
- Status badges with semantic colors

### HTML Data Flow

1. **Server-Side Rendering**
   - Templates are rendered on the server
   - Data is injected into templates through handler functions
   - Dynamic content based on database state

2. **Form Submissions**
   - Traditional form submissions for most operations
   - HTMX-enhanced submissions for interactive elements
   - Validation happens both client and server side

3. **Conditional Rendering**
   - Templates use Go conditionals to show/hide elements based on:
     - User authentication status
     - Talk status (proposed, scheduled, completed)
     - User's relationship to a talk (speaker, attendee, voter)

### Mobile Responsiveness

The application is designed to be mobile-friendly with:
- Responsive navbar that collapses on small screens
- Responsive grid system for content layout
- Appropriate sizing for touch interactions on mobile devices

### Accessibility Considerations

1. **Semantic HTML**
   - Proper heading hierarchy (h1, h2, etc.)
   - Appropriate form labels
   - ARIA attributes where needed

2. **Keyboard Navigation**
   - Focusable elements properly tab-indexed
   - Logical tab order for form elements
   - Visible focus states

### JavaScript Integration

The application uses minimal JavaScript, primarily leveraging:
- Bootstrap's JS for interactive components (dropdowns, tabs)
- HTMX for dynamic content updates
- No custom JavaScript required for core functionality 