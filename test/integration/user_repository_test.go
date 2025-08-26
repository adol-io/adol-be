package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/repositories"
	infraRepos "github.com/nicklaros/adol/internal/infrastructure/repositories"
	"github.com/nicklaros/adol/pkg/utils"
)

func TestUserRepository_Integration(t *testing.T) {
	// Setup test database
	testDB := SetupTestDB(t)
	defer TeardownTestDB(t, testDB)

	// Setup test context
	ctx, _ := SetupTestContext(t)

	// Create repository
	userRepo := infraRepos.NewPostgreSQLUserRepository(testDB.DB)

	t.Run("Create and Get User", func(t *testing.T) {
		// Create test user
		user := &entities.User{
			ID:        uuid.New(),
			Username:  "testuser123",
			Email:     "testuser123@example.com",
			FirstName: "Test",
			LastName:  "User",
			Role:      entities.RoleManager,
			Status:    entities.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Set password
		err := user.UpdatePassword("testpassword123")
		require.NoError(t, err)

		// Create user
		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Get user by ID
		retrievedUser, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrievedUser.ID)
		assert.Equal(t, user.Username, retrievedUser.Username)
		assert.Equal(t, user.Email, retrievedUser.Email)
		assert.Equal(t, user.FirstName, retrievedUser.FirstName)
		assert.Equal(t, user.LastName, retrievedUser.LastName)
		assert.Equal(t, user.Role, retrievedUser.Role)
		assert.Equal(t, user.Status, retrievedUser.Status)

		// Get user by username
		userByUsername, err := userRepo.GetByUsername(ctx, user.Username)
		require.NoError(t, err)
		assert.Equal(t, user.ID, userByUsername.ID)

		// Get user by email
		userByEmail, err := userRepo.GetByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, userByEmail.ID)

		// Cleanup
		err = userRepo.Delete(ctx, user.ID)
		require.NoError(t, err)
	})

	t.Run("Update User", func(t *testing.T) {
		// Create test user
		user := &entities.User{
			ID:        uuid.New(),
			Username:  "updateuser",
			Email:     "updateuser@example.com",
			FirstName: "Update",
			LastName:  "User",
			Role:      entities.RoleEmployee,
			Status:    entities.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := user.UpdatePassword("password123")
		require.NoError(t, err)

		// Create user
		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Update user
		user.FirstName = "Updated"
		user.LastName = "Name"
		user.Role = entities.RoleManager
		user.UpdatedAt = time.Now()

		err = userRepo.Update(ctx, user)
		require.NoError(t, err)

		// Get updated user
		updatedUser, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", updatedUser.FirstName)
		assert.Equal(t, "Name", updatedUser.LastName)
		assert.Equal(t, entities.RoleManager, updatedUser.Role)

		// Cleanup
		err = userRepo.Delete(ctx, user.ID)
		require.NoError(t, err)
	})

	t.Run("List Users with Pagination", func(t *testing.T) {
		// Create multiple test users
		users := make([]*entities.User, 3)
		for i := 0; i < 3; i++ {
			users[i] = &entities.User{
				ID:        uuid.New(),
				Username:  fmt.Sprintf("listuser%d", i),
				Email:     fmt.Sprintf("listuser%d@example.com", i),
				FirstName: fmt.Sprintf("First%d", i),
				LastName:  fmt.Sprintf("Last%d", i),
				Role:      entities.RoleEmployee,
				Status:    entities.UserStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := users[i].UpdatePassword("password123")
			require.NoError(t, err)

			err = userRepo.Create(ctx, users[i])
			require.NoError(t, err)
		}

		// Test pagination
		filter := repositories.UserFilter{}
		pagination := utils.PaginationInfo{
			Page:  1,
			Limit: 2,
		}

		userList, resultPagination, err := userRepo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(userList), 2) // At least our test users
		assert.Equal(t, 2, resultPagination.Limit)
		assert.Equal(t, 1, resultPagination.Page)

		// Test filtering by role
		filter.Role = func() *entities.UserRole { r := entities.RoleEmployee; return &r }()
		filteredUsers, _, err := userRepo.List(ctx, filter, pagination)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(filteredUsers), 3) // Our test users

		// Cleanup
		for _, user := range users {
			err = userRepo.Delete(ctx, user.ID)
			require.NoError(t, err)
		}
	})

	t.Run("User Status Operations", func(t *testing.T) {
		// Create test user
		user := &entities.User{
			ID:        uuid.New(),
			Username:  "statususer",
			Email:     "statususer@example.com",
			FirstName: "Status",
			LastName:  "User",
			Role:      entities.RoleEmployee,
			Status:    entities.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := user.UpdatePassword("password123")
		require.NoError(t, err)

		// Create user
		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Activate user (should work) - change status to active
		user.ChangeStatus(entities.UserStatusActive)
		err = userRepo.Update(ctx, user)
		require.NoError(t, err)

		// Deactivate user
		user.ChangeStatus(entities.UserStatusInactive)
		err = userRepo.Update(ctx, user)
		require.NoError(t, err)

		// Verify status changed
		updatedUser, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, entities.UserStatusInactive, updatedUser.Status)

		// Suspend user
		user.ChangeStatus(entities.UserStatusSuspended)
		err = userRepo.Update(ctx, user)
		require.NoError(t, err)

		// Verify status changed
		suspendedUser, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, entities.UserStatusSuspended, suspendedUser.Status)

		// Cleanup
		err = userRepo.Delete(ctx, user.ID)
		require.NoError(t, err)
	})

	t.Run("Update Last Login", func(t *testing.T) {
		// Create test user
		user := &entities.User{
			ID:        uuid.New(),
			Username:  "loginuser",
			Email:     "loginuser@example.com",
			FirstName: "Login",
			LastName:  "User",
			Role:      entities.RoleEmployee,
			Status:    entities.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := user.UpdatePassword("password123")
		require.NoError(t, err)

		// Create user
		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Update last login
		user.UpdateLastLogin()
		err = userRepo.Update(ctx, user)
		require.NoError(t, err)

		// Verify last login was updated
		updatedUser, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, updatedUser.LastLoginAt)

		// Cleanup
		err = userRepo.Delete(ctx, user.ID)
		require.NoError(t, err)
	})

	t.Run("Change Password", func(t *testing.T) {
		// Create test user
		user := &entities.User{
			ID:        uuid.New(),
			Username:  "pwduser",
			Email:     "pwduser@example.com",
			FirstName: "Password",
			LastName:  "User",
			Role:      entities.RoleEmployee,
			Status:    entities.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := user.UpdatePassword("oldpassword")
		require.NoError(t, err)
		oldHash := user.PasswordHash

		// Create user
		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Change password
		err = user.UpdatePassword("newpassword123")
		require.NoError(t, err)
		err = userRepo.Update(ctx, user)
		require.NoError(t, err)

		// Verify password was changed
		updatedUser, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotEqual(t, oldHash, updatedUser.PasswordHash)

		// Verify new password works
		isValid := updatedUser.ValidatePassword("newpassword123")
		assert.True(t, isValid)

		// Cleanup
		err = userRepo.Delete(ctx, user.ID)
		require.NoError(t, err)
	})

	t.Run("Unique Constraints", func(t *testing.T) {
		// Create first user
		user1 := &entities.User{
			ID:        uuid.New(),
			Username:  "uniqueuser",
			Email:     "unique@example.com",
			FirstName: "Unique",
			LastName:  "User",
			Role:      entities.RoleEmployee,
			Status:    entities.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := user1.UpdatePassword("password123")
		require.NoError(t, err)

		err = userRepo.Create(ctx, user1)
		require.NoError(t, err)

		// Try to create user with same username
		user2 := &entities.User{
			ID:        uuid.New(),
			Username:  "uniqueuser", // Same username
			Email:     "different@example.com",
			FirstName: "Different",
			LastName:  "User",
			Role:      entities.RoleEmployee,
			Status:    entities.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = user2.UpdatePassword("password123")
		require.NoError(t, err)

		err = userRepo.Create(ctx, user2)
		assert.Error(t, err) // Should fail due to unique constraint

		// Try to create user with same email
		user3 := &entities.User{
			ID:        uuid.New(),
			Username:  "differentuser",
			Email:     "unique@example.com", // Same email
			FirstName: "Different",
			LastName:  "User",
			Role:      entities.RoleEmployee,
			Status:    entities.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = user3.UpdatePassword("password123")
		require.NoError(t, err)

		err = userRepo.Create(ctx, user3)
		assert.Error(t, err) // Should fail due to unique constraint

		// Cleanup
		err = userRepo.Delete(ctx, user1.ID)
		require.NoError(t, err)
	
})
}