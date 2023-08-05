package db

import (
	"context"
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

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		txName := ctx.Value(txKey)

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

		// Lock flow --> from_id < to_id
		if arg.FromAccountID < arg.ToAccountID {

			fmt.Println(txName, "get account 1")
			account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
			if err != nil {
				return err
			}

			fmt.Println(txName, "update account 1")
			result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
				ID:      account1.ID,
				Balance: account1.Balance - arg.Amount,
			})
			if err != nil {
				return err
			}

			fmt.Println(txName, "get account 2")
			account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
			if err != nil {
				return err
			}

			fmt.Println(txName, "update account 2")
			result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
				ID:      account2.ID,
				Balance: account2.Balance + arg.Amount,
			})
			if err != nil {
				return err
			}

		} else {

			fmt.Println(txName, "get account 2")
			account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
			if err != nil {
				return err
			}

			fmt.Println(txName, "update account 2")
			result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
				ID:      account2.ID,
				Balance: account2.Balance + arg.Amount,
			})
			if err != nil {
				return err
			}

			fmt.Println(txName, "get account 1")
			account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
			if err != nil {
				return err
			}

			fmt.Println(txName, "update account 1")
			result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
				ID:      account1.ID,
				Balance: account1.Balance - arg.Amount,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	return result, err
}
