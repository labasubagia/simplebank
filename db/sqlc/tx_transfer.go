package db

import (
	"context"
	"errors"
	"fmt"
)

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		txName := ctx.Value(txKey)

		fmt.Println(txName, "lock both account")
		accounts, err := q.GetAccountsForUpdate(ctx, []int64{arg.FromAccountID, arg.ToAccountID})
		if err != nil {
			return err
		}
		accountMap := map[int64]Account{}
		for _, account := range accounts {
			accountMap[account.ID] = account
		}
		fromAccount, ok := accountMap[arg.FromAccountID]
		if !ok {
			return errors.New("from account not found")
		}
		toAccount, ok := accountMap[arg.ToAccountID]
		if !ok {
			return errors.New("to account not found")
		}

		fmt.Println(txName, "create transfer")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		fmt.Println(txName, "create entry 1")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		fmt.Println(txName, "create entry 2")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		fmt.Println(txName, "subtract account 1")
		result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      fromAccount.ID,
			Balance: fromAccount.Balance - arg.Amount,
		})
		if err != nil {
			return err
		}

		fmt.Println(txName, "add account 2")
		result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      toAccount.ID,
			Balance: toAccount.Balance + arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}
