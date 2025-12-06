package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

// Repository struct
type Repository struct {
	client *goredislib.Client
	mutex  *redsync.Mutex
}

func NewRepository(address, password string) Repository {
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     address,
		Password: password,
		DB:       0, // ← Using database index 0
	})

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)
	mutexname := "my-global-mutex"
	mutex := rs.NewMutex(mutexname, redsync.WithExpiry(10*time.Second))

	return Repository{
		client: client,
		mutex:  mutex,
	}
}

func BuildCompanySharesKey(companyId string) string {
	return fmt.Sprintf("shares:%s", companyId)
}

func BuildUserSharesKey(userId, companyId string) string {
	return fmt.Sprintf("user:%s:shares:%s", userId, companyId)
}

func (r *Repository) BuyShares(ctx context.Context, userId, companyId string, numShares int, wg *sync.WaitGroup) error {
	defer wg.Done()

	fmt.Printf("User %s attempting to buy %d shares of company %s\n", userId, numShares, companyId)

	if err := r.mutex.Lock(); err != nil {
		fmt.Println("Failed to acquire lock:", err)
		return err
	}

	defer func() {
		if ok, err := r.mutex.Unlock(); !ok || err != nil {
			fmt.Printf("warning during unlock: %v\n", err)
		}
	}()

	currentShares, err := r.client.Get(ctx, BuildCompanySharesKey(companyId)).Int()
	if err != nil {
		if err == goredislib.Nil {
			fmt.Printf("No shares available for company %s\n", companyId)
			return err
		} else {
			fmt.Println("Error getting current shares:", err)
			return err
		}

	}

	if currentShares < numShares {
		fmt.Printf("Not enough shares available for company %s. Requested: %d, Available: %d\n", companyId, numShares, currentShares)
		return errors.New("Not enought shares")
	}

	currentShares -= numShares

	err = r.client.Set(ctx, BuildCompanySharesKey(companyId), currentShares, 0).Err()
	if err != nil {
		fmt.Printf("error updating company shares: %v\n", err)
		return err
	}

	userKey := BuildUserSharesKey(userId, companyId)
	currentUserShares, _ := r.client.Get(ctx, userKey).Int()
	newUserShares := currentUserShares + numShares

	err = r.client.Set(ctx, userId, newUserShares, 0).Err()
	if err != nil {
		fmt.Printf("Not able to save new value for userId : %d", userId)
		return err
	}

	fmt.Printf("Success! User %s bought %d shares of %s. Remaining: %d\n",
		userId, numShares, companyId, currentShares)

	return nil

}

// InitializeCompanyShares sets initial shares for a company
func (r *Repository) InitializeCompanyShares(ctx context.Context, companyId string, initialShares int) error {
    key := BuildCompanySharesKey(companyId)
    err := r.client.Set(ctx, key, initialShares, 0).Err()
    if err != nil {
        return fmt.Errorf("failed to initialize company shares: %v", err)
    }
    fmt.Printf("Initialized company %s with %d shares\n", companyId, initialShares)
    return nil
}
