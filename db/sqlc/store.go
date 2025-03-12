package db

import(
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute database queries and transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg CreateTransferParams) (TransferTxResult, error)
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

// NewStore creates a new Store
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db: db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rb error: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// TransferTxParams contains the input parameters of the transfer transaction
// type TransferTxParams struct {
// 	FromAccountID int64 `json:"from_account_id"`
// 	ToAccountID int64 `json:"to_account_id"`
// 	Amount int64 `json:"amount"`
// }

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer Transfer `json:"transfer"`
	FromAccount Account `json:"from_account"` // The account that the money was transferred from
	ToAccount Account `json:"to_account"` // The account that the money was transferred to
	FromEntry Entry `json:"from_entry"` // The entry created to record the money being transferred from the account
	ToEntry Entry `json:"to_entry"` // The entry created to record the money being transferred to the account
}

// var txKey = struct{}{}
// type contextKey string
// const txKey contextKey = "txKey"

// TransferTx performs a money transfer from one account to the other
// It creates two entries in the `entries` table - one for the account that the money is transferred from and another for the account that the money is transferred to
// It creates a transfer in the `transfers` table
// If the sender does not have enough balance, it rolls back the transaction and returns an error
func (store *SQLStore) TransferTx(ctx context.Context, arg CreateTransferParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// txName := ctx.Value(txKey).(string)

		// fmt.Println("Create Transfer: ", txName)
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID: arg.ToAccountID,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}

		// fmt.Println("Create Entry 1: ", txName)
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount: -arg.Amount,
		})
		if err != nil {
			return err
		}

		// fmt.Println("Create Entry 2: ", txName)
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}
		
		// fmt.Println("Get Account 1: ", txName)
		result.FromAccount, err = q.GetAccountForUpdate(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		// fmt.Println("Get Account 2: ", txName)
		result.ToAccount, err = q.GetAccountForUpdate(ctx, arg.ToAccountID)
		if err != nil {
			return err
		}

		newFromAccountBalance := result.FromAccount.Balance - arg.Amount
		if newFromAccountBalance < 0 {
			return fmt.Errorf("insufficient balance in account: %d", arg.FromAccountID)
		}

		newToAccountBalance := result.ToAccount.Balance + arg.Amount
		if newToAccountBalance < 0 {
			return fmt.Errorf("overflow in account: %d", arg.ToAccountID)
		}

		if arg.FromAccountID < arg.ToAccountID {
			// fmt.Println("Update Account 1 Balance: ", txName)
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, newFromAccountBalance, arg.ToAccountID, newToAccountBalance)
			if err != nil {
				return err
			}
			
		} else {
			// fmt.Println("Update Account 2 Balance: ", txName)
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, newToAccountBalance, arg.FromAccountID, newFromAccountBalance)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.UpdateAccount(ctx, UpdateAccountParams{
				ID: accountID1,
				Balance: amount1,
			})
	if err != nil {
		return
	}

	account2, err = q.UpdateAccount(ctx, UpdateAccountParams{
				ID: accountID2,
				Balance: amount2,
			})
	if err != nil {
		return
	}

	return
}