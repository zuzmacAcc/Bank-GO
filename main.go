package main

import (
	"flag"
	"fmt"
	"log"
)

func seedAccount(store Storage, first_name, last_name, password string) *Account {
	acc, err := NewAccount(first_name, last_name, password)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateAccount(acc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("new account => ", acc.Number)

	return acc
}

func seedAccounts(s Storage) {
	seedAccount(s, "Jack", "Reacher", "password123")
}

func main() {
	seed := flag.Bool("seed", false, "seed the db with some accounts")
	flag.Parse()

	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	if *seed {
		fmt.Println("seeding the database")
		seedAccounts(store)
	}
	server := NewAPIServer(":3000", store)
	server.Run()
}
