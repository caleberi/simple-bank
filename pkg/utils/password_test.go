package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func Test_Password(t *testing.T) {
	hasher := NewHasher(bcrypt.DefaultCost)

	password := RandomString(10)
	hashedPassword, err := hasher.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	err = hasher.CheckPassword(password, hashedPassword)
	require.NoError(t, err)

	wrongPassword := RandomString(10)
	err = hasher.CheckPassword(wrongPassword, hashedPassword)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	rehashedPassword, err := hasher.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)
	require.NotEqual(t, hashedPassword, rehashedPassword)

}
