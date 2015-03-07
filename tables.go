package main

import (
	"database/sql"
	"fmt"
	"log"
)

// item: [item_id, name, price, supplier]
// inventory: [item_id]
// reservation: [item_id, customer]
// sold: [item_id, customer]
// customer: [uname, money, pwd, supplier]
func CreateItemTable(db *sql.DB) {
	itemTable, err := db.Prepare("create table item (item_id int NOT NULL AUTO_INCREMENT, name VARCHAR(12), price DOUBLE, supplier CHAR(20), PRIMARY KEY(item_id))")
	if err != nil {
		log.Fatal(err)
	}

	res, err := itemTable.Exec()
	if err != nil {
		log.Println(err)
	}
	log.Printf("Created item Table")
	fmt.Println(res)
}

func CreateInventoryTable(db *sql.DB) {
	itemTable, err := db.Prepare("create table inventory (item_id int, PRIMARY KEY(item_id))")
	if err != nil {
		log.Fatal(err)
	}
	res, err := itemTable.Exec()
	if err != nil {
		log.Println(err)
	}
	log.Printf("Created inventory Table")
	fmt.Println(res)
}

func CreateReservationTable(db *sql.DB) {
	itemTable, err := db.Prepare("create table reservation (item_id int, customer CHAR(20), PRIMARY KEY(item_id))")
	if err != nil {
		log.Fatal(err)
	}
	res, err := itemTable.Exec()
	if err != nil {
		log.Print(err)
	}
	log.Printf("Created inventory Table")
	fmt.Println(res)
}

func CreateSoldTable(db *sql.DB) {
	itemTable, err := db.Prepare("create table sold (item_id int, customer VARCHAR(20), PRIMARY KEY(item_id))")
	if err != nil {
		log.Fatal(err)
	}
	res, err := itemTable.Exec()
	if err != nil {
		log.Print(err)
	}
	log.Printf("Created inventory Table")
	fmt.Println(res)
}

func CreateCustomerTable(db *sql.DB) {
	itemTable, err := db.Prepare("create table customer (uname VARCHAR(20) NOT NULL,  money DOUBLE, pwd VARCHAR(20), supplier bool, PRIMARY KEY(uname))")
	if err != nil {
		log.Fatal(err)
	}
	res, err := itemTable.Exec()
	if err != nil {
		log.Print(err)
	}
	log.Printf("Created customer Table")
	fmt.Println(res)
}
