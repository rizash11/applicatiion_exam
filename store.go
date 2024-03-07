package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

const STMT_PART1 = `SELECT ordered_product.*,
product.name,
product.shelf1 AS shelf1_id,
CASE 
	WHEN shelf1.quantity IS NULL THEN 0
	ELSE shelf1.quantity
END AS shelf1_quantity,
product.shelf2 AS shelf2_id,
CASE
	WHEN shelf2.quantity IS NULL THEN 0
	ELSE shelf2.quantity
END AS shelf2_quantity,
product.shelf3 AS shelf3_id,
CASE
	WHEN shelf3.quantity IS NULL THEN 0
	ELSE shelf3.quantity
END AS shelf3_quantity
FROM ordered_product
LEFT JOIN product ON ordered_product.product_id = product.product_id
LEFT JOIN shelf_product AS shelf1 ON product.shelf1 = shelf1.shelf_id AND ordered_product.product_id = shelf1.product_id
LEFT JOIN shelf_product AS shelf2 ON product.shelf2 = shelf2.shelf_id AND ordered_product.product_id = shelf2.product_id
LEFT JOIN shelf_product AS shelf3 ON product.shelf3 = shelf3.shelf_id AND ordered_product.product_id = shelf3.product_id
WHERE ordered_product.order_id IN (` // passing orders to a single placeholder (?) for some reason does not work well

const STMT_PART2 = `)
ORDER BY 5, 1;` // ordered first by shelf_id and then by product_id

func main() {
	username := flag.String("u", "oakward", "username")
	password := flag.String("p", "112233", "password")
	db := flag.String("db", "store", "database")
	host := flag.String("h", "localhost", "host")
	flag.Parse()
	dsn := *username + ":" + *password + "@tcp(" + *host + ")/" + *db

	store, err := openDB(dsn)
	if err != nil {
		errStmt := errors.New("error opening the database: " + err.Error())
		log.Fatalln(errStmt)
	}
	defer store.Close()

	var orders string
	if len(os.Args) == 2 {
		orders = os.Args[1]
	} else {
		log.Fatalln("provide a list of order id-s in a single string separated by commas as an argument, like this \"go run . 11,12,13,14\"")
	}

	rows, err := store.Query(STMT_PART1 + orders + STMT_PART2)
	if err != nil {
		errStmt := errors.New("illegal syntax for order id-s, it must cosist of integers: " + err.Error())
		log.Fatalln(errStmt)
	}
	defer rows.Close()

	err = printOrders(rows, orders)
	if err != nil {
		log.Fatalln("error printing results: " + err.Error())
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func printOrders(rows *sql.Rows, orders string) error {
	product_id, order_id, quantity := new(int), new(int), new(int)
	name, shelf1_id, shelf2_id, shelf3_id := new(string), new(string), new(string), new(string)
	shelf1_quantity, shelf2_quantity, shelf3_quantity := new(int), new(int), new(int)

	prev_shelf := ""
	shelfBlock := ""

	fmt.Println("=+=+=+=")
	fmt.Println("Страница сборки заказов " + orders + "\n")

	for rows.Next() {
		err := rows.Scan(product_id, order_id, quantity, name, shelf1_id, shelf1_quantity, shelf2_id, shelf2_quantity, shelf3_id, shelf3_quantity)
		if err != nil {
			return err
		}

		if prev_shelf != *shelf1_id {
			fmt.Printf("%s", shelfBlock)
			shelfBlock = ""
			shelfBlock = shelfBlock + "===Стеллаж " + *shelf1_id + "\n"
		}

		shelfBlock = shelfBlock + *name + " (id=" + strconv.Itoa(*product_id) + ")\n"
		shelfBlock = shelfBlock + "заказ " + strconv.Itoa(*order_id) + ", " + strconv.Itoa(*quantity) + "\n\n"

		prev_shelf = *shelf1_id
	}

	fmt.Printf("%s", shelfBlock) // printing the last block

	return rows.Err()
}
