package db

import (
	"context"
	"testing"
	"time"

	"github.com/caleberi/simple-bank/pkg/utils"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func createRandomUser(t *testing.T) User {
	hasher := utils.NewHasher(bcrypt.DefaultCost)
	hashPassword, err := hasher.HashPassword(utils.RandomString(10))
	require.NoError(t, err)
	require.NotEmpty(t, hashPassword)

	arg := CreateUserParams{
		Username:       utils.RandomOwner(),
		FullName:       utils.RandomOwner(),
		HashedPassword: hashPassword,
		Email:          utils.RandomEmail(),
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func Test_CreateUser(t *testing.T) {
	createRandomUser(t)
}

func Test_GetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, user1.Username, user1.Username)
	require.Equal(t, user1.HashedPassword, user.HashedPassword)
	require.Equal(t, user1.FullName, user.FullName)
	require.Equal(t, user1.Email, user.Email)

	require.WithinDuration(t, user.PasswordChangedAt, user.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user1.CreatedAt, user.CreatedAt, time.Second)

}
