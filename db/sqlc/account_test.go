package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/dibrito/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func createTestAccount(t *testing.T) Account {
	u := createRandomUser(t)
	arg := CreateAccountParams{
		Owner:    u.Username,
		Balance:  util.RandonMoney(),
		Currency: util.RandonCurrency(),
	}
	account1, err := testQueries.CreateAccount(context.Background(), arg)
	// require stops the test if fails
	require.NoError(t, err)
	require.NotEmpty(t, account1)
	require.Equal(t, arg.Owner, account1.Owner)
	require.Equal(t, arg.Balance, account1.Balance)
	require.Equal(t, arg.Currency, account1.Currency)

	require.NotZero(t, account1.ID)
	require.NotZero(t, account1.CreatedAt)
	return account1
}

func TestCreateAccount(t *testing.T) {
	createTestAccount(t)
}

func TestGetAccount(t *testing.T) {
	acc1 := createTestAccount(t)
	acc2, err := testQueries.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, acc2)

	require.Equal(t, acc1.Owner, acc2.Owner)
	require.Equal(t, acc1.Balance, acc2.Balance)
	require.Equal(t, acc1.Currency, acc2.Currency)

	require.Equal(t, acc1.ID, acc2.ID)
	require.WithinDuration(t, acc1.CreatedAt, acc2.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	acc1 := createTestAccount(t)

	args := UpdateAccountParams{
		ID:      acc1.ID,
		Balance: util.RandonMoney(),
	}

	acc2, err := testQueries.UpdateAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, acc2)

	require.Equal(t, acc1.Owner, acc2.Owner)
	require.Equal(t, args.Balance, acc2.Balance)
	require.Equal(t, acc1.Currency, acc2.Currency)

	require.Equal(t, acc1.ID, acc2.ID)
	require.WithinDuration(t, acc1.CreatedAt, acc2.CreatedAt, time.Second)
}

func TestDeleteAccount(t *testing.T) {
	acc1 := createTestAccount(t)
	err := testQueries.DeleteAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	acc2, err := testQueries.GetAccount(context.Background(), acc1.ID)
	require.Error(t, err, sql.ErrNoRows)
	require.Empty(t, acc2)
}

func TestListAccount(t *testing.T) {
	var lastAccount Account
	for i := 0; i < 10; i++ {
		lastAccount = createTestAccount(t)
	}

	accounts, err := testQueries.ListAccounts(context.Background(), ListAccountsParams{
		Owner:  lastAccount.Owner,
		Limit:  5,
		Offset: 0,
	})
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	for _, v := range accounts {
		require.NotEmpty(t, v)
		require.Equal(t, v.Owner, lastAccount.Owner)
	}
}
