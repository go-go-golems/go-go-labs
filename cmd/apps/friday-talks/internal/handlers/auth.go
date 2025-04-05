package handlers

import (
	"net/http"
	"strings"

	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/auth"
	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/models"
	"github.com/go-go-golems/go-go-labs/cmd/apps/friday-talks/internal/templates"
)

// AuthHandler handles authentication related routes
type AuthHandler struct {
	userRepo models.UserRepository
	auth     *auth.Auth
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo models.UserRepository, auth *auth.Auth) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		auth:     auth,
	}
}

// HandleLogin handles the login page and form submission
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to home
	if user := auth.UserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var errorMsg string

	// Process login form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		// Validate input
		if email == "" || password == "" {
			errorMsg = "Email and password are required"
		} else {
			// Authenticate user
			user, err := h.auth.Authenticate(r.Context(), email, password)
			if err != nil {
				errorMsg = "Invalid email or password"
			} else {
				// Generate token and set cookie
				token, err := h.auth.GenerateToken(user.ID)
				if err != nil {
					http.Error(w, "Failed to generate token", http.StatusInternalServerError)
					return
				}

				h.auth.SetTokenCookie(w, token)

				// Redirect to home page
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}
	}

	// Render login page
	templates.Login(errorMsg).Render(r.Context(), w)
}

// HandleRegister handles the registration page and form submission
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to home
	if user := auth.UserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var errorMsg string

	// Process registration form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Validate input
		if name == "" || email == "" || password == "" {
			errorMsg = "All fields are required"
		} else if password != confirmPassword {
			errorMsg = "Passwords do not match"
		} else if len(password) < 8 {
			errorMsg = "Password must be at least 8 characters"
		} else {
			// Check if email is already in use
			_, err := h.userRepo.FindByEmail(r.Context(), email)
			if err == nil {
				errorMsg = "Email is already in use"
			} else {
				// Create new user
				hashedPassword, err := models.HashPassword(password)
				if err != nil {
					http.Error(w, "Failed to hash password", http.StatusInternalServerError)
					return
				}

				user := &models.User{
					Name:         name,
					Email:        email,
					PasswordHash: hashedPassword,
				}

				if err := h.userRepo.Create(r.Context(), user); err != nil {
					http.Error(w, "Failed to create user", http.StatusInternalServerError)
					return
				}

				// Generate token and set cookie
				token, err := h.auth.GenerateToken(user.ID)
				if err != nil {
					http.Error(w, "Failed to generate token", http.StatusInternalServerError)
					return
				}

				h.auth.SetTokenCookie(w, token)

				// Redirect to home page
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}
	}

	// Render registration page
	templates.Register(errorMsg).Render(r.Context(), w)
}

// HandleLogout handles user logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear authentication cookie
	h.auth.ClearTokenCookie(w)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleProfile handles the user profile page and updates
func (h *AuthHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := auth.UserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var errorMsg string
	var successMsg string

	// Process profile update form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		currentPassword := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_password")

		// Check if email is already in use by another user
		if email != user.Email {
			existingUser, err := h.userRepo.FindByEmail(r.Context(), email)
			if err == nil && existingUser.ID != user.ID {
				errorMsg = "Email is already in use by another account"
				templates.Profile(user, successMsg, errorMsg).Render(r.Context(), w)
				return
			}
		}

		// Update user information
		user.Name = name
		user.Email = email

		// Update password if provided
		if currentPassword != "" && newPassword != "" {
			if !models.CheckPassword(currentPassword, user.PasswordHash) {
				errorMsg = "Current password is incorrect"
				templates.Profile(user, successMsg, errorMsg).Render(r.Context(), w)
				return
			}

			if newPassword != confirmPassword {
				errorMsg = "New passwords do not match"
				templates.Profile(user, successMsg, errorMsg).Render(r.Context(), w)
				return
			}

			if len(newPassword) < 8 {
				errorMsg = "New password must be at least 8 characters"
				templates.Profile(user, successMsg, errorMsg).Render(r.Context(), w)
				return
			}

			hashedPassword, err := models.HashPassword(newPassword)
			if err != nil {
				http.Error(w, "Failed to hash password", http.StatusInternalServerError)
				return
			}

			user.PasswordHash = hashedPassword
		}

		// Save user updates
		if err := h.userRepo.Update(r.Context(), user); err != nil {
			http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			return
		}

		successMsg = "Profile updated successfully"
	}

	// Render profile page
	templates.Profile(user, successMsg, errorMsg).Render(r.Context(), w)
}

// ValidateEmail checks if the email is in a valid format
func ValidateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
