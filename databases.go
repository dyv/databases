package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type Item struct {
	ID       int64
	Name     string
	Price    float64
	Supplier string
}

func GetInventory(tx *sql.Tx) []Item {
	var items []Item
	stmt, err := tx.Prepare(
		"SELECT item.item_id, item.name, item.price, item.supplier " +
			"FROM inventory INNER JOIN item " +
			"ON item.item_id = inventory.item_id")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Supplier)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Read in Item: ", item)
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return items
}

func GetReserved(tx *sql.Tx, customer string) []Item {
	var items []Item
	stmt, err := tx.Prepare(
		"SELECT item.item_id, item.name, item.price, item.supplier " +
			"FROM item INNER JOIN (SELECT * FROM reservation WHERE customer=?) myreserves WHERE item.item_id = myreserves.item_id")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := stmt.Query(customer)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Supplier)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return items
}

func GetBought(tx *sql.Tx, customer string) []Item {
	var items []Item
	stmt, err := tx.Prepare(
		"SELECT item.item_id, item.name, item.price, item.supplier " +
			"FROM item INNER JOIN (SELECT * FROM sold WHERE customer=?) myreserves ON item.item_id = myreserves.item_id")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := stmt.Query(customer)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Supplier)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return items
}

func GetMyItems(tx *sql.Tx, supplier string) []Item {
	var items []Item
	stmt, err := tx.Prepare(
		"SELECT item.item_id, item.name, item.price, item.supplier " +
			"FROM item WHERE item.supplier = ?")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := stmt.Query(supplier)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Supplier)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return items
}

func Reserve(tx *sql.Tx, uname string, item_id int) {
	log.Println("moving into reservation")
	stmt, err := tx.Prepare("INSERT INTO reservation(customer,item_id) VALUES(?,?)")
	if err != nil {
		log.Fatal(err)
	}
	stmt.Exec(uname, item_id)

	log.Println("deleting from inventory")
	stmt, err = tx.Prepare("DELETE FROM inventory WHERE item_id=?")
	if err != nil {
		log.Fatal(err)
	}
	stmt.Exec(item_id)

	tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func Buy(tx *sql.Tx, uname string, item_id int) {
	stmt, err := tx.Prepare("INSERT INTO sold(customer,item_id) VALUES(?,?)")
	if err != nil {
		log.Fatal(err)
	}
	stmt.Exec(uname, item_id)

	stmt, err = tx.Prepare("DELETE FROM reservation WHERE item_id=?")
	if err != nil {
		log.Fatal(err)
	}
	stmt.Exec(item_id)
}

func AddItem(tx *sql.Tx, name string, price float64, supplier string) {
	stmt, err := tx.Prepare("INSERT INTO item(name,price,supplier) VALUES(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(name, price, supplier)
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
	stmt, err = tx.Prepare("INSERT INTO inventory(item_id) VALUES(?)")
	if err != nil {
		log.Fatal(err)
	}
	res, err = stmt.Exec(lastId)
	if err != nil {
		log.Fatal(err)
	}
}

func GetCustomer(tx *sql.Tx, name string) string {
	stmt, err := tx.Prepare("select pwd from customer where uname=?")
	rows, err := stmt.Query(name)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var pwd string
		err := rows.Scan(&pwd)
		if err != nil {
			log.Fatal(err)
		}
		return pwd
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return ""
}

func NewCustomer(tx *sql.Tx, name string, money float64, pwd string, supp bool) {
	// check if the customer exists
	stmt, err := tx.Prepare("insert into customer(uname,money,pwd,supplier) values(?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(name, money, pwd, supp)
	if err != nil {
		log.Println(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("NewCustomer: ID = %d, affected = %d\n", lastId, rowCnt)
	return
}
