package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	FindByID(ctx context.Context, id int) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	List(ctx context.Context) ([]*User, error)
}

// SQLiteUserRepository implements UserRepository for SQLite
type SQLiteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository creates a new SQLiteUserRepository
func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

// FindByID finds a user by ID
func (r *SQLiteUserRepository) FindByID(ctx context.Context, id int) (*User, error) {
	query := `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	user := &User{}
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "error querying user")
	}

	return user, nil
}

// FindByEmail finds a user by email
func (r *SQLiteUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = ?`

	row := r.db.QueryRowContext(ctx, query, email)

	user := &User{}
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "error querying user")
	}

	return user, nil
}

// Create creates a new user
func (r *SQLiteUserRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (name, email, password_hash, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.PasswordHash,
		now,
		now,
	)

	if err != nil {
		return errors.Wrap(err, "error creating user")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "error getting last insert id")
	}

	user.ID = int(id)
	user.CreatedAt = now
	user.UpdatedAt = now

	return nil
}

// Update updates an existing user
func (r *SQLiteUserRepository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users 
		SET name = ?, email = ?, password_hash = ?, updated_at = ? 
		WHERE id = ?
	`

	now := time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.PasswordHash,
		now,
		user.ID,
	)

	if err != nil {
		return errors.Wrap(err, "error updating user")
	}

	user.UpdatedAt = now

	return nil
}

// List returns all users
func (r *SQLiteUserRepository) List(ctx context.Context) ([]*User, error) {
	query := `SELECT id, name, email, password_hash, created_at, updated_at FROM users ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "error querying users")
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.PasswordHash,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, errors.Wrap(err, "error scanning user row")
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating user rows")
	}

	return users, nil
}

// HashPassword generates a bcrypt hash from a password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPassword checks if the provided password matches the hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
