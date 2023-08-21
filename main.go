package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	var err error
	toSpawn := runtime.NumCPU()
	if len(os.Args) > 1 {
		toSpawn, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	}

	// PG Setup
	dbString := fmt.Sprintf("user=%s password=%s sslmode=%s host=%s port=%s dbname=%s",
		"postgres",
		"postgres",
		"disable",
		"localhost",
		"5432",
		"example_db",
	)

	wg := sync.WaitGroup{}
	wg.Add(toSpawn)

	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < toSpawn; i++ {
		go func(i int) {
			defer wg.Done()
			spawn(ctx, dbString, i)
		}(i)
	}

	// Create a quit channel which carries os.Signal values.
	quit := make(chan os.Signal, 1)

	// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and
	// relay them to the quit channel.
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Read the signal from the quit channel. block until received
	sig := <-quit

	// Print the signal type.
	fmt.Println("Got signal:", sig)

	// Call the cancel function to cancel the context.
	cancel()

	// Wait for the goroutines to finish.
	wg.Wait()
}

const WorkerLockID = 8675309

var ScarySharedValue int

func spawn(ctx context.Context, dbString string, workerID int) {
	client, err := sql.Open("postgres", dbString)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
      fmt.Printf("Worker %d shutting down\n", workerID)
			return
		default:
      ok, err := tryObtainLock(client, WorkerLockID)
      if err != nil {
        fmt.Println("Error obtaining lock: ", err)
      }

			if !ok {
				fmt.Printf("Worker %d failed to obtain lock\n", workerID)
			} else {
				fmt.Printf("Worker %d obtained lock\n", workerID)
				ScarySharedValue++
				println(ScarySharedValue)
				time.Sleep(2 * time.Second)
				err := releaseLock(client, WorkerLockID)
				if err != nil {
					fmt.Println("Error releasing lock: ", err)
				}

				fmt.Printf("Worker %d released lock\n", workerID)
			}

			time.Sleep(3 * time.Second)
		}
	}
}

func releaseLock(db *sql.DB, lockID int) error {
	_, err := db.Exec(fmt.Sprintf("SELECT pg_advisory_unlock(%d)", lockID))
	return err
}

func tryObtainLock(client *sql.DB, lockID int) (bool, error) {
	var lockObtained bool
	err := client.QueryRow(fmt.Sprintf(`SELECT pg_try_advisory_lock(%d)`, lockID)).
		Scan(&lockObtained)
	if err != nil {
		return false, nil
	}

	return lockObtained, nil
}
