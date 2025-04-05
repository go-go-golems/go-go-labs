# Friday Talks: Continuing the HTMX Conversion

**Date:** 2025-04-04

## 1. Purpose and Scope

This document outlines the ongoing effort to enhance the `friday-talks` web application by integrating HTMX more deeply. The goal is to create a more dynamic and responsive user experience by reducing full-page reloads for common actions like form submissions, filtering, and UI updates.

The conversion process follows a pattern of identifying interactive elements, extracting their rendering logic into partial `templ` templates, and modifying the corresponding Go handlers to return these partials for HTMX requests.

## 2. Current Status / Accomplishments

Significant progress has been made in converting key parts of the application to use HTMX:

*   **Talk List Filtering (`/talks`)**: The status filter tabs (All, Scheduled, Proposed, Past) now dynamically load the talk list without a full page refresh. (`_talksCards.templ`, `handlers/talks.go::HandleListTalks`)
*   **Talk Detail Interactions (`/talks/{id}`)**:
    *   Voting on proposed talks (`_talkVoteForm.templ`, `handlers/talk_interaction.go::HandleVoteOnTalk`)
    *   Managing attendance for scheduled talks (`_talkAttendance.templ`, `handlers/talk_interaction.go::HandleManageAttendance`)
    *   Submitting feedback for completed talks (`_talkFeedbackForm.templ`, `handlers/talk_interaction.go::HandleProvideFeedback`)
*   **Talk Proposal Form (`/talks/propose`)**: Submits via HTMX, re-renders form with errors, uses `HX-Redirect` on success. (`_proposeTalkForm.templ`, `handlers/talk_management.go::HandleProposeTalk`)
*   **Talk Edit Form (`/talks/{id}/edit`)**: Submits via HTMX, re-renders form with errors, uses `HX-Redirect` on success. (`_editTalkForm.templ`, `handlers/talk_management.go::HandleEditTalk`)
*   **Authentication Forms**:
    *   Login (`/login`) (`_loginForm.templ`, `handlers/auth.go::HandleLogin`)
    *   Registration (`/register`) (`_registerForm.templ`, `handlers/auth.go::HandleRegister`)
    *   Profile Update (`/profile`) (`_profileForm.templ`, `handlers/auth.go::HandleProfile`)
*   **Linting Errors Fixed**: Addressed `templ` syntax errors and ran `templ generate` to update Go code.

## 3. Approach / Pattern Used

The conversion generally follows these steps:

1.  **Identify Target:** Locate a form submission or interactive element currently causing a full page reload.
2.  **Create Partial Template:** Extract the relevant HTML rendering logic from the main `templ` file into a new partial template file (e.g., `_myPartialForm.templ`).
3.  **Add HTMX Attributes:** Add `hx-*` attributes (like `hx-post`, `hx-target`, `hx-swap="outerHTML"`) to the form or interactive element in the *partial* template. The `hx-target` usually points to a container div that will wrap the partial.
4.  **Modify Main Template:** Replace the original HTML in the main template with a call to the new partial, wrapped in a container `div` with an `id` matching the `hx-target`.
5.  **Update Go Handler:**
    *   Check for the `HX-Request: true` header to detect HTMX requests.
    *   On **validation error**: If it's an HTMX request, render *only* the partial template (including the error message) and send it back with `Content-Type: text/html`. Otherwise, render the full page template.
    *   On **success**: 
        *   If the action should result in a navigation (e.g., after login, register, create, edit), send back an empty `200 OK` response with an `HX-Redirect: /target/url` header for HTMX requests. For standard requests, perform a normal `http.Redirect`.
        *   If the action should just update a section of the current page (e.g., submitting a vote, toggling attendance, adding/deleting a resource), render the updated partial template and send it back for HTMX requests. For standard requests, redirect back to the page, often with a query parameter.
6.  **Run `templ generate`:** Regenerate the Go template code to make the new partials available to the handlers.

## 4. Remaining Tasks

The following areas still need to be converted to use HTMX:

1.  **Resource Management (Talk Detail Page):**
    *   "Add Resource" form.
    *   "Delete" button for existing resources.
    *   Handlers: `HandleAddResource`, `HandleDeleteResource` in `handlers/talk_interaction.go`.
    *   Templates: `TalkDetail` in `talks.templ`.
    *   *Goal:* Adding/deleting resources should update the resource list dynamically without a full page reload.

2.  **Talk Management Actions (Talk Detail Page):**
    *   "Schedule This Talk" button/link (leads to `/talks/{id}/schedule` GET/POST).
    *   "Mark as Completed" button.
    *   "Cancel Talk" button.
    *   Handlers: `HandleScheduleTalk`, `HandleCompleteTalk`, `HandleCancelTalk` in `handlers/talk_management.go`.
    *   Templates: `TalkDetail` in `talks.templ` (specifically the "Manage Talk" card).
    *   *Goal:* These actions should update the talk's status badge and potentially the available actions in the "Manage Talk" card dynamically.
The "Schedule" action involves both a GET (to show the form) and a POST (to submit). Both could potentially be made more dynamic, perhaps loading the scheduling form into a modal or inline section.

3.  **Calendar View (`/calendar`):**
    *   Month navigation (Previous/Next month links).
    *   Handler: `HandleCalendarView` in `handlers/calendar.go`.
    *   Template: `CalendarView` in `calendar.templ`.
    *   *Goal:* Clicking month navigation should load the calendar grid for the target month via HTMX, replacing the existing calendar grid.

## 5. Key Files

*   **Handlers:**
    *   `cmd/apps/friday-talks/internal/handlers/talks.go`
    *   `cmd/apps/friday-talks/internal/handlers/talk_interaction.go`
    *   `cmd/apps/friday-talks/internal/handlers/talk_management.go`
    *   `cmd/apps/friday-talks/internal/handlers/auth.go`
    *   `cmd/apps/friday-talks/internal/handlers/calendar.go`
*   **Templates:**
    *   `cmd/apps/friday-talks/internal/templates/` (all `.templ` files, especially the newly created partials prefixed with `_`)
*   **Reference:**
    *   `ttmp/2025-04-04/07-html-http-flow.md` (Provides overview of original routes/flow)

## 6. Next Steps

1.  Review the "Approach / Pattern Used" section above.
2.  Choose one of the "Remaining Tasks" to start with. Good candidates are:
    *   **Resource Management:** Relatively self-contained section on the talk detail page.
    *   **Talk Management Actions:** Modifies existing elements on the talk detail page.
3.  Apply the HTMX conversion pattern described.
4.  Remember to run `templ generate` after adding/modifying templates.
5.  Test thoroughly. 