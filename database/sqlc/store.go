package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

type SQLStore struct {
	*Queries         // Kind of inheritance taken from Queries present inside the db.go
	db       *sql.DB // instance of sql.DB
}

func NewStore(db *sql.DB) Store {
	// Takes argument of db instance that belongs to sql.DB
	return &SQLStore{
		// returns store struct which database instance
		// And also Queries struct.
		db:      db,
		Queries: New(db),
	}
}

// It is the reciever for the SQLStore struct whcih takes context as the argument and function.
// execTx returns the error
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	// This is the function to create a clean database transaction
	// it begins the transaction
	tx, err := store.db.BeginTx(ctx, nil)
	// errors checks
	if err != nil {
		return err
	}

	// Running db.go New function so that we can Queries struct in return with
	q := New(tx)
	err = fn(q) // The function which we recieves in the argument as it only returns for the error
	if err != nil {
		// if the function which we got in the argument returns the error then we go for rollback
		// providing rollback with error check
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("txt err: %v, rbErr: %v", err, rbErr)
		}
		return err
	}

	// If everything is fine then the we need to commit that database transaction.
	return tx.Commit()
}

// Creating a custom params
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

// IN THE CONTEXT OF THIS PROJECT THERE ARE 5 ELEMENTS OF DATABASE TRANSACTIONS:
// 1. CREATE A TRANSFER RECORD
// 2. CREATE AN ENTRY FOR SENDER
// 3. CREATE AN ENTRY FOR RECIEVER
// 4. DEDUCT THE BALANCE FROM SENDER FOR THE AMOUNT SENT
// 5. ADD THE BALANCE IN THE RECIEVER FOR THE AMOUNT RECIEVED

// CREATING TRANSFER RECORD
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	// result will have Transfer, FromAccount, ToAcccount, FromEntry, ToEntry

	// Calling the function created for transaction
	var err error
	err = store.execTx(ctx, func(q *Queries) error {
		//Creating a transfer and storing in the instance of the result.

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})

		// err check
		if err != nil {
			return err
		}

		// Creating entry and storing it into result instance.
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})

		// err check
		if err != nil {
			return err
		}

		// Creating an entry and storing it into result instance
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})

		if err != nil {
			return err
		}

	// 	if arg.FromAccountID < arg.ToAccountID {
	// 		result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
	// 			ID:     arg.FromAccountID,
	// 			Amount: -arg.Amount,
	// 		})
	
	// 		if err != nil {
	// 			return err
	// 		}
	
	// 		result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
	// 			ID:     arg.ToAccountID,
	// 			Amount: arg.Amount,
	// 		})
	
	// 		if err != nil {
	// 			return err
	// 		}
	// 	} else {
	// 		result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
	// 			ID:     arg.ToAccountID,
	// 			Amount: arg.Amount,
	// 		})
	
	// 		if err != nil {
	// 			return err
	// 		}
	
	// 		result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
	// 			ID:     arg.FromAccountID,
	// 			Amount: -arg.Amount,
	// 		})
	
	// 		if err != nil {
	// 			return err
	// 		}
		
	// }
	if arg.FromAccountID < arg.ToAccountID {
		result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
	} else {
		result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
	}
		return nil
	})

	// Returning the created result instance and also the error
	return result, err
}

func addMoney (
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
)(account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID: accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID: accountID2,
		Amount: amount2,
	})

	if err != nil {
		return
	}

	return // Means return account1, account2, err
}
