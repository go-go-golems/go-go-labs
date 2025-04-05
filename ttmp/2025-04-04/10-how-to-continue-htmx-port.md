# Onboarding: Continuing the Friday Talks HTMX Conversion

**Date:** 2025-04-04

## 1. Goal

The primary goal is to enhance the `friday-talks` web application by integrating HTMX more deeply. This involves converting server-rendered components that currently cause full-page reloads (like form submissions or data filtering) into dynamic updates using HTMX requests. The aim is a more fluid and responsive user experience.

## 2. Application Context

*   **Name:** Friday Talks
*   **Purpose:** Manage and schedule internal technical talks.
*   **Tech Stack:**
    *   Backend: Go
    *   Routing: Chi
    *   Templating: `templ` (Go-based HTML templating)
    *   Frontend Interaction: HTMX
    *   Styling: Bootstrap 5
    *   Database: SQLite
*   **Architecture:** Clean architecture (handlers, services, repositories).
*   **Key Features:** Talk proposal, voting, scheduling, attendance, feedback, resource management, user auth.
*   **Reference:** `ttmp/2025-04-04/07-html-http-flow.md` provides a detailed overview of the application's routes and flow *before* the HTMX conversion started.

## 3. Conversion Pattern

The conversion process follows a consistent pattern:

1.  **Identify Target:** Find an interaction (form submission, button click, link navigation) that causes a full page reload but could be updated dynamically.
2.  **Create Partial Template:** Extract the HTML for the component that needs to be updated into its own `templ` file (e.g., `_myPartial.templ`). This partial should accept the necessary data (like the `*models.User`, `*models.Talk`, error/success messages) as parameters.
3.  **Add HTMX Attributes:** In the *partial* template, add the relevant `hx-*` attributes (`hx-post`, `hx-get`, `hx-delete`, `hx-target`, `hx-swap`, `hx-indicator`, `hx-confirm` etc.) to the interactive element (form, button, link).
    *   `hx-target`: Often points to a wrapper `div` around the partial itself.
    *   `hx-swap`: Usually `outerHTML` if replacing the whole component, or `innerHTML` if replacing content inside a container.
    *   `hx-indicator`: Points to a loading spinner element.
4.  **Modify Main Template:** In the original template file (e.g., `talks.templ`), replace the extracted HTML with a call to the new partial template (e.g., `@_myPartial(...)`). Ensure the partial call is wrapped in an element whose `id` matches the `hx-target` specified in the partial.
5.  **Update Go Handler:**
    *   Import `github.com/go-go-golems/go-go-labs/internal/helpers`.
    *   Use `helpers.IsHtmxRequest(r)` to detect if the request came from HTMX.
    *   **On Error (Validation, Permissions etc.):** If HTMX request, render *only* the partial template (passing the error message) with `Content-Type: text/html`. Otherwise (non-HTMX), perform the original action (e.g., render the full page with the error or redirect with an error query param).
    *   **On Success:**
        *   **Update-in-place:** If the action updates the current view (e.g., adding/deleting a resource, voting), render the updated partial template (passing a success message or fresh data) for HTMX requests. Otherwise, redirect back to the page (often with a success query param).
        *   **Navigation:** If the action *should* navigate away (e.g., after create, edit, schedule), send back an empty `200 OK` with an `HX-Redirect: /new/url` header for HTMX requests. Otherwise, perform a standard `http.Redirect`.
    *   Create helper functions within the handler (like `renderMyPartial`) to avoid code duplication when rendering the partial in different scenarios (initial load, HTMX error, HTMX success). These helpers should fetch the necessary fresh data before rendering the partial.
6.  **Run `templ generate`:** Execute `cd cmd/apps/friday-talks && templ generate` in the terminal to update the Go code (`*.templ.go`) for the template changes. Resolve any generation errors.

## 4. Work Completed So Far (as of 2025-04-04)

The following features have been converted to use HTMX:

*   **Talk List Filtering (`/talks`)**: Status tabs (All, Scheduled, etc.) load dynamically.
    *   Partials: `_talksCards.templ`
    *   Handler: `HandleListTalks` in `handlers/talks.go`
*   **Talk Detail Interactions (`/talks/{id}`)**:
    *   Voting: `_talkVoteForm.templ`, `HandleVoteOnTalk` in `handlers/talk_interaction.go`
    *   Attendance: `_talkAttendance.templ`, `HandleManageAttendance` in `handlers/talk_interaction.go`
    *   Feedback: `_talkFeedbackForm.templ`, `HandleProvideFeedback` in `handlers/talk_interaction.go`
    *   Resource Management (Add/Delete): `_talkResourcesSection.templ`, `HandleAddResource`, `HandleDeleteResource` in `handlers/talk_interaction.go`
*   **Forms**: Submit via HTMX, show errors inline, `HX-Redirect` on success.
    *   Propose Talk (`/talks/propose`): `_proposeTalkForm.templ`, `HandleProposeTalk` in `handlers/talk_management.go`
    *   Edit Talk (`/talks/{id}/edit`): `_editTalkForm.templ`, `HandleEditTalk` in `handlers/talk_management.go`
    *   Login (`/login`): `_loginForm.templ`, `HandleLogin` in `handlers/auth.go`
    *   Register (`/register`): `_registerForm.templ`, `HandleRegister` in `handlers/auth.go`
    *   Profile Update (`/profile`): `_profileForm.templ`, `HandleProfile` in `handlers/auth.go`

*   **Linting Errors Fixed**: Addressed previous `templ` syntax errors.

## 5. Remaining Tasks

Consult `ttmp/2025-04-04/08-how-to-continue-htmx-port.md` (section 4) for the original list. The next tasks are:

1.  **Talk Management Actions (Talk Detail Page):**
    *   Actions: "Schedule This Talk" (GET/POST), "Mark as Completed", "Cancel Talk".
    *   Target: Dynamically update the "Manage Talk" card and potentially the talk status badge without a full reload. The "Schedule" action might involve loading a form dynamically (e.g., in a modal or inline) for the GET request and handling the POST with HTMX.
    *   Handlers: `HandleScheduleTalk` (GET/POST), `HandleCompleteTalk`, `HandleCancelTalk` in `handlers/talk_management.go`.
    *   Templates: `TalkDetail` in `talks.templ`. Create a new `_talkManagementCard.templ` partial.

2.  **Calendar View (`/calendar`):**
    *   Actions: Previous/Next month navigation links.
    *   Target: Load the calendar grid for the target month via HTMX, replacing the existing grid.
    *   Handler: `HandleCalendarView` in `handlers/calendar.go`.
    *   Templates: `CalendarView` in `calendar.templ`. Create a new `_calendarGrid.templ` partial.

## 6. Key Files

*   **Handlers:**
    *   `cmd/apps/friday-talks/internal/handlers/talks.go`
    *   `cmd/apps/friday-talks/internal/handlers/talk_interaction.go`
    *   `cmd/apps/friday-talks/internal/handlers/talk_management.go`
    *   `cmd/apps/friday-talks/internal/handlers/auth.go`
    *   `cmd/apps/friday-talks/internal/handlers/calendar.go`
*   **Templates:**
    *   `cmd/apps/friday-talks/internal/templates/` (all `.templ` files, especially partials prefixed with `_`)
*   **Models:**
    *   `cmd/apps/friday-talks/internal/models/`
*   **Helper:**
    *   `internal/helpers/htmx.go` (for `IsHtmxRequest`)
*   **Reference Docs:**
    *   `ttmp/2025-04-04/07-html-http-flow.md`
    *   `ttmp/2025-04-04/08-how-to-continue-htmx-port.md`
    *   This file (`ttmp/2025-04-04/10-how-to-continue-htmx-port.md`)

## 7. How to Continue

1.  **Address Linter Errors:** Before starting new work, review the linter errors reported in the previous step (related to `talks.templ` and `talk_interaction.go` after recent changes) and fix them. Run `templ generate` and `go build ./...` to confirm fixes. *Self-note: These likely involve missing template imports or undefined functions/methods.*
2.  **Choose a Task:** Pick one of the "Remaining Tasks" (Talk Management Actions or Calendar View).
3.  **Follow the Pattern:** Apply the HTMX conversion pattern described in Section 3.
4.  **Generate & Build:** Run `cd cmd/apps/friday-talks && templ generate` and `go build ./...` frequently to catch errors early.
5.  **Test:** Thoroughly test the changes in the browser, checking both HTMX and non-HTMX (JavaScript disabled or full page refresh) scenarios. Check browser developer console for errors.
6.  **Commit:** Commit the changes for the completed task.

Good luck! 