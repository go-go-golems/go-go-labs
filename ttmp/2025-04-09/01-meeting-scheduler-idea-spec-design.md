# Meeting Scheduler Application - Specification and Design

## Overview

This document outlines the specification and design for a meeting scheduler application targeted at technology enthusiasts. The primary goal is to facilitate finding common availability for 1-hour group calls among members with diverse schedules.

## Problem Statement

Technology enthusiast groups often face challenges in coordinating meeting times across different schedules, time zones, and availability patterns. The current solutions typically involve:

1. Manual coordination via chat/email (inefficient, time-consuming)
2. Using general-purpose scheduling tools not optimized for group consensus
3. Doodle-like polls that don't effectively visualize time conflicts

## Goals and Requirements

### Primary Goals

- Enable users to find common available time slots for 1-hour meetings
- Minimize the coordination effort required from group members
- Provide clear visualization of availability overlaps
- Ensure a smooth, intuitive user experience

### Functional Requirements

1. **User Management**

   - User registration and authentication
   - Profile management with timezone information
   - Group creation and membership management

2. **Availability Management**

   - Ability to set recurring availability patterns
   - One-time availability exceptions
   - Integration with external calendars (Google, Outlook, etc.)

3. **Meeting Scheduling**

   - Propose potential meeting dates/times
   - Collect availability from all participants
   - Visualize common availability periods
   - Automated suggestion of optimal meeting times
   - Confirmation and calendar event creation

4. **Notifications**
   - Email notifications for new meeting proposals
   - Reminders to provide availability
   - Meeting confirmation alerts

### Non-Functional Requirements

1. **Performance**

   - Fast availability calculation even with many participants
   - Responsive UI with minimal loading times

2. **Security**

   - Secure authentication and data storage
   - Privacy controls for availability information

3. **Usability**
   - Intuitive, clean interface
   - Mobile-responsive design
   - Minimal steps to complete common tasks

## System Architecture

### High-Level Architecture

The application will follow a modern web application architecture with:

1. **Backend**

   - Go-based REST API server
   - PostgreSQL database for persistence
   - Redis for caching and real-time features

2. **Frontend**
   - Web application using HTMX, Bootstrap, and Templ (as per guidelines)
   - Progressive enhancement for JavaScript-optional experience

```
┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│               │      │               │      │               │
│  Web Frontend │◄────►│  Go Backend   │◄────►│  PostgreSQL   │
│  (HTMX/Templ) │      │  (API Server) │      │  Database     │
│               │      │               │      │               │
└───────────────┘      └───────────────┘      └───────────────┘
                              │
                              ▼
                       ┌───────────────┐      ┌───────────────┐
                       │               │      │               │
                       │  Redis Cache  │      │  External     │
                       │               │      │  Calendar APIs│
                       │               │      │               │
                       └───────────────┘      └───────────────┘
```

### Database Schema

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│   Users     │       │   Groups    │       │ Memberships │
├─────────────┤       ├─────────────┤       ├─────────────┤
│ id          │       │ id          │       │ id          │
│ email       │       │ name        │       │ user_id     │
│ name        │       │ description │       │ group_id    │
│ password    │       │ created_at  │       │ role        │
│ timezone    │       │ updated_at  │       │ joined_at   │
│ created_at  │       └─────────────┘       └─────────────┘
│ updated_at  │              │                     │
└─────────────┘              │                     │
       │                     └──────────┬──────────┘
       │                                │
       ▼                                ▼
┌─────────────┐               ┌─────────────┐
│ Availability │               │  Meetings   │
├─────────────┤               ├─────────────┤
│ id          │               │ id          │
│ user_id     │◄──────────────┤ group_id    │
│ start_time  │               │ title       │
│ end_time    │               │ description │
│ recurrence  │               │ duration    │
│ is_available│               │ status      │
│ created_at  │               │ created_by  │
│ updated_at  │               │ created_at  │
└─────────────┘               │ updated_at  │
                              └─────────────┘
                                     │
                                     ▼
                              ┌─────────────┐
                              │ TimeSlots   │
                              ├─────────────┤
                              │ id          │
                              │ meeting_id  │
                              │ start_time  │
                              │ end_time    │
                              │ votes       │
                              │ status      │
                              │ created_at  │
                              │ updated_at  │
                              └─────────────┘
```

## Technical Stack

### Backend

- **Language**: Go
- **Web Framework**: Standard library or Gin/Echo with proper middleware
- **CLI**: Cobra for command-line operations
- **Database**: PostgreSQL with proper migrations
- **ORM/Query Builder**: GORM or SQLx
- **Authentication**: JWT-based authentication
- **Error Handling**: github.com/pkg/errors
- **Concurrency**: errgroup for goroutine management
- **Calendar Integration**: Google Calendar, Outlook APIs

### Frontend

- **UI Framework**: Bootstrap for responsive design
- **Interactive UI**: HTMX for dynamic content without heavy JavaScript
- **Templating**: Templ for type-safe HTML templates
- **Time/Date Handling**: Day.js or similar for timezone operations
- **Visualization**: Custom time grid component

## Key Features and Implementation Details

### 1. User Management

Users will register with email and password. Account profiles will include:

- Personal details (name, email)
- Timezone preferences
- Default availability patterns
- External calendar connections

### 2. Group Management

Users can create and join multiple groups:

- Technology meetup groups
- Project teams
- Study groups
- Each group has admins and members with different permissions

### 3. Availability Collection

Multiple approaches to set availability:

- Weekly recurring patterns (e.g., "Available weekdays 5-8 PM")
- One-time exceptions (e.g., "Unavailable next Monday")
- Calendar integration to automatically detect conflicts
- Quick response options for specific meeting proposals

### 4. Meeting Proposal and Scheduling

The core scheduling workflow:

1. Group admin creates a meeting proposal with potential date ranges
2. Members indicate availability for proposed ranges
3. System calculates and visualizes optimal times
4. Admin selects final time
5. System creates calendar events and notifies participants

### 5. Availability Visualization

A crucial UI component showing:

- Time grid with user availability overlaps
- Heat map showing optimal slots
- Timezone conversion for distributed teams
- Conflict highlighting

## User Interface Design

### Main Screens

1. **Dashboard**

   - Overview of upcoming meetings
   - Pending availability requests
   - Quick access to groups

2. **Group Management**

   - Member list
   - Past and upcoming meetings
   - Group settings

3. **Availability Settings**

   - Regular schedule configuration
   - Calendar integration options
   - Timezone settings

4. **Meeting Creation**

   - Date range selector
   - Duration settings
   - Description and agenda

5. **Availability Response**

   - Interactive time grid for selection
   - Bulk availability actions
   - Comment/note options

6. **Meeting Results View**
   - Optimal time visualization
   - Final selection interface
   - Notification options

## Implementation Plan

### Phase 1: Core System

1. Set up Go project structure with Cobra CLI
2. Implement user authentication system
3. Create basic group management
4. Develop database schema and migrations
5. Build simple availability collection

### Phase 2: Scheduling Engine

1. Implement availability algorithm
2. Create time slot voting system
3. Develop optimal time calculation
4. Build meeting confirmation flow

### Phase 3: UI and Experience

1. Develop responsive Bootstrap/HTMX interface
2. Create interactive availability selection components
3. Build visualization for availability overlaps
4. Implement email notification system

### Phase 4: Integration and Enhancement

1. Add external calendar integration
2. Implement recurring meeting patterns
3. Add timezone conversion and display
4. Optimize performance for larger groups

## Conclusion

This meeting scheduler application addresses a common pain point for technology enthusiast groups by providing an intuitive way to find common availability. The focus on minimal user effort, clear visualization, and intelligent suggestions will make coordination significantly easier compared to existing solutions.

By leveraging modern Go-based architecture with HTMX and Bootstrap, we can create a responsive, maintainable application that provides an excellent user experience without excessive complexity.
