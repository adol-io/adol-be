package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/pkg/utils"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entities.User) error
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	
	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	
	// Update updates an existing user
	Update(ctx context.Context, user *entities.User) error
	
	// Delete deletes a user (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error
	
	// List retrieves users with pagination and filtering
	List(ctx context.Context, filter UserFilter, pagination utils.PaginationInfo) ([]*entities.User, utils.PaginationInfo, error)
	
	// ExistsByUsername checks if a user exists by username
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	
	// ExistsByEmail checks if a user exists by email
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// UserFilter represents filters for user queries
type UserFilter struct {
	Role     *entities.UserRole   `json:"role,omitempty"`
	Status   *entities.UserStatus `json:"status,omitempty"`
	Search   string               `json:"search,omitempty"` // Search in username, email, first_name, last_name
	OrderBy  string               `json:"order_by,omitempty"`
	OrderDir string               `json:"order_dir,omitempty"` // ASC or DESC
}