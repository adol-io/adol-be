package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/utils"
)

// PostgreSQLUserRepository implements the UserRepository interface for PostgreSQL
type PostgreSQLUserRepository struct {
	db *sql.DB
}

// NewPostgreSQLUserRepository creates a new PostgreSQL user repository
func NewPostgreSQLUserRepository(db *sql.DB) repositories.UserRepository {
	return &PostgreSQLUserRepository{
		db: db,
	}
}

// Create creates a new user
func (r *PostgreSQLUserRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (id, username, email, first_name, last_name, role, status, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Role,
		user.Status,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "username") {
					return errors.NewConflictError("username already exists")
				}
				if strings.Contains(pqErr.Detail, "email") {
					return errors.NewConflictError("email already exists")
				}
				return errors.NewConflictError("user already exists")
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *PostgreSQLUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	query := `
		SELECT id, username, email, first_name, last_name, role, status, password_hash, created_at, updated_at, last_login_at
		FROM users 
		WHERE id = $1`

	user := &entities.User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("user")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *PostgreSQLUserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	query := `
		SELECT id, username, email, first_name, last_name, role, status, password_hash, created_at, updated_at, last_login_at
		FROM users 
		WHERE username = $1`

	user := &entities.User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("user")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgreSQLUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := `
		SELECT id, username, email, first_name, last_name, role, status, password_hash, created_at, updated_at, last_login_at
		FROM users 
		WHERE email = $1`

	user := &entities.User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("user")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// Update updates an existing user
func (r *PostgreSQLUserRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users 
		SET username = $2, email = $3, first_name = $4, last_name = $5, 
		    password_hash = $6, role = $7, status = $8, last_login_at = $9, updated_at = $10
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.FirstName,
		user.LastName,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.LastLoginAt,
		user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "username") {
					return errors.NewConflictError("username already exists")
				}
				if strings.Contains(pqErr.Detail, "email") {
					return errors.NewConflictError("email already exists")
				}
				return errors.NewConflictError("user already exists")
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("user")
	}

	return nil
}

// Delete deletes a user (soft delete)
func (r *PostgreSQLUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users 
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("user")
	}

	return nil
}

// List retrieves users with pagination and filtering
func (r *PostgreSQLUserRepository) List(ctx context.Context, filter repositories.UserFilter, pagination utils.PaginationInfo) ([]*entities.User, utils.PaginationInfo, error) {
	// Build WHERE clause
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	whereConditions = append(whereConditions, "deleted_at IS NULL")

	if filter.Role != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, *filter.Role)
		argIndex++
	}

	if filter.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Search != "" {
		searchCondition := fmt.Sprintf("(username ILIKE $%d OR email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", argIndex, argIndex, argIndex, argIndex)
		whereConditions = append(whereConditions, searchCondition)
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Build ORDER BY clause
	orderBy := "created_at DESC"
	if filter.OrderBy != "" {
		direction := "ASC"
		if filter.OrderDir == "DESC" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", filter.OrderBy, direction)
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to count users: %w", err)
	}

	// Calculate pagination
	offset := (pagination.Page - 1) * pagination.Limit
	totalPages := int((total + int64(pagination.Limit) - 1) / int64(pagination.Limit))

	// Build main query
	query := fmt.Sprintf(`
		SELECT id, username, email, first_name, last_name, role, status, password_hash, created_at, updated_at, last_login_at
		FROM users 
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, pagination.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &entities.User{}
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.Status,
			&user.PasswordHash,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLoginAt,
		)
		if err != nil {
			return nil, pagination, fmt.Errorf("failed to scan user: %w", err)
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, pagination, fmt.Errorf("failed to iterate users: %w", err)
	}

	// Update pagination info
	resultPagination := utils.PaginationInfo{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		TotalCount: int(total),
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}

	return users, resultPagination, nil
}

// ExistsByUsername checks if a user exists by username
func (r *PostgreSQLUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)`
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists by username: %w", err)
	}

	return exists, nil
}

// ExistsByEmail checks if a user exists by email
func (r *PostgreSQLUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists by email: %w", err)
	}

	return exists, nil
}
