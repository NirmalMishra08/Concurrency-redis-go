package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Share Trading System with Redis ===")
	redisAddress := "localhost:6379"
	redisPassword := ""

	repo := NewRepository(redisAddress, redisPassword)

	ctx := context.Background()
	pong, err := repo.client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Printf("Connected to Redis: %s\n", pong)

	companyId := "GOOG"
	initialShares := 100

	err = repo.InitializeCompanyShares(ctx, companyId, initialShares)
	if err != nil {
		log.Fatalf("Failed to initialize company shares: %v", err)
	}

	// Display initial state
	shares, err := repo.GetCompanyShares(ctx, companyId)
	if err != nil {
		log.Fatalf("Failed to get company shares: %v", err)
	}
	fmt.Printf("\nInitial shares for %s: %d\n", companyId, shares)

	// Test 1: Single user buying shares
	// fmt.Println("\n--- Test 1: Single User Purchase ---")
	// var wg1 sync.WaitGroup
	// wg1.Add(1)

	// go func() {
	// 	err := repo.BuyShares(ctx, "user1", companyId, 10, &wg1)
	// 	if err != nil {
	// 		fmt.Printf("Purchase failed: %v\n", err)
	// 	}
	// }()

	// wg1.Wait()
	// time.Sleep(500 * time.Millisecond)

	finalShares, _ := repo.GetCompanyShares(ctx, companyId)
	fmt.Printf("\nFinal shares for %s: %d\n", companyId, finalShares)
	fmt.Println("Test completed successfully!")

	// Test 2: Multiple users buying shares concurrently
    fmt.Println("\n--- Test 2: Concurrent Purchases ---")
    var wg2 sync.WaitGroup
    users := []string{"user2", "user3", "user4", "user5"}
    
    for i, user := range users {
        wg2.Add(1)
        go func(u string, sharesToBuy int) {
           
            time.Sleep(time.Duration(i) * 100 * time.Millisecond)
            err := repo.BuyShares(ctx, u, companyId, sharesToBuy, &wg2)
            if err != nil {
                fmt.Printf("%s: Purchase failed - %v\n", u, err)
            }
        }(user, (i+1)*5) // Different amounts for each user
    }
    
    wg2.Wait()
    time.Sleep(1 * time.Second)

	// Test 3: Edge case - insufficient shares
    fmt.Println("\n--- Test 3: Insufficient Shares Test ---")
    var wg3 sync.WaitGroup
    wg3.Add(1)
    
    go func() {
        err := repo.BuyShares(ctx, "user6", companyId, 200, &wg3)
        if err != nil {
            fmt.Printf("Expected error for insufficient shares: %v\n", err)
        }
    }()
    
    wg3.Wait()


}
