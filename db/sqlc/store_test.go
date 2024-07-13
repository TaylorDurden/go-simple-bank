package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	originalAccount1 := createRandAccount(t)
	originalAccount2 := createRandAccount(t)

	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: originalAccount1.ID,
				ToAccountID:   originalAccount2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	// check results
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, originalAccount1.ID, transfer.FromAccountID)
		require.Equal(t, originalAccount2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, originalAccount1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		// check entries
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, originalAccount2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, originalAccount1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, originalAccount2.ID, toAccount.ID)

		// check accounts' balance
		diff1 := originalAccount1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - originalAccount2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		// e.g: amonut = 10, account1.Balance = 1001
		// 1st transfer 10 to toAccount
		// 2nd transfer 10 to toAccount
		// 3rd ...
		// 1st: diff1 = 100 - (100 - 10) = 10
		// 2nd: diff1 = 100 - (90 - 10) = 20
		// 3rd: diff1 = 100 - (80 - 10) = 30
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final updated balances
	updatedAccount1, err := testQueries.GetAccount(context.Background(), originalAccount1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), originalAccount2.ID)
	require.NoError(t, err)

	require.Equal(t, originalAccount1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, originalAccount2.Balance+int64(n)*amount, updatedAccount2.Balance)
}

func TestTransferTxDeadLock(t *testing.T) {
	store := NewStore(testDB)

	originalAccount2 := createRandAccount(t)
	originalAccount1 := createRandAccount(t)

	// run n concurrent transfer transactions
	n := 10
	amount := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := originalAccount1.ID
		toAccountID := originalAccount2.ID

		// switch the from/to accounts
		if i%2 == 1 {
			fromAccountID = originalAccount2.ID
			toAccountID = originalAccount1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// check the final updated balances
	updatedAccount1, err := testQueries.GetAccount(context.Background(), originalAccount1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), originalAccount2.ID)
	require.NoError(t, err)

	fmt.Println(">>> after:", updatedAccount1.Balance, updatedAccount2.Balance)

	// 5 transactions transfer from account1 to account2
	// 5 transactions transfer from account2 to account1
	// so the originalAccount1 and originalAccount2 balances stay unchanged
	require.Equal(t, originalAccount1.Balance, updatedAccount1.Balance)
	require.Equal(t, originalAccount2.Balance, updatedAccount2.Balance)
}
