package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type person struct {
	id         int
	first_name string
	last_name  string
	email      string
	ip_address string
}

func addPerson(db *sql.DB, newPerson person) {
	stmt, _ := db.Prepare("INSERT INTO people (id, first_name, last_name, email, ip_address) VALUES (?, ?, ?, ?, ?)")
	stmt.Exec(nil, newPerson.first_name, newPerson.last_name, newPerson.email, newPerson.ip_address)
	defer stmt.Close()

	fmt.Printf("Added %v %v \n", newPerson.first_name, newPerson.last_name)
}

func prepareAddPerson(db *sql.DB) {
	newPerson := person{
		first_name: "anban",
		last_name:  "gisak",
		email:      "anbangisak@gmail.com",
		ip_address: "127.0.0.1",
	}

	addPerson(db, newPerson)
}

func main() {
	db, err := sql.Open("sqlite3", "./auro.db")
	// db, err := sql.Open("sqlite3", ":memory:")
	checkErr(err)
	defer db.Close()

	//2. fail-fast if can't connect to DB
	checkErr(db.Ping())

	//3. create table
	_, err = db.Exec("create table USER (ID integer PRIMARY KEY, NAME string not null); delete from USER;")
	checkErr(err)

	//4. insert data
	//4.1 Begin transaction
	tx, err := db.Begin()
	checkErr(err)

	//4.2 Prepare insert stmt.
	stmt, err := tx.Prepare("insert into USER(ID, NAME) values(?, ?)")
	checkErr(err)
	defer stmt.Close()

	for i := 0; i < 10; i++ {
		_, err = stmt.Exec(i, fmt.Sprint("user-", i))
		checkErr(err)
	}

	//4.3 Commit transaction
	tx.Commit()

	//5. Query data
	rows, err := db.Query("select * from USER")
	checkErr(err)
	defer rows.Close()

	//5.1 Iterate through result set
	for rows.Next() {
		var name string
		var id int
		err := rows.Scan(&id, &name)
		checkErr(err)
		fmt.Printf("id=%d, name=%s\n", id, name)
	}

	//5.2 check error, if any, that were encountered during iteration
	err = rows.Err()
	checkErr(err)

	// prepareAddPerson(db)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world from GfG")
	})
	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})
	port := ":8200"
	fmt.Println("Server is running on port" + port)

	// Start server on port specified above
	log.Fatal(http.ListenAndServe(port, nil))
	fmt.Println("Hello World!")
}

func checkErr(err error, args ...string) {
	if err != nil {
		fmt.Println("Error")
		fmt.Println("%q: %s", err, args)
	}
}
