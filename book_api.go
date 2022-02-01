package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "1234"
	dbname   = "book_api"
)

var db *sql.DB
var err error

func init() {
	psqlconn := fmt.Sprintf("host= %s port = %d user = %s password= %s dbname= %s sslmode=disable", host, port, user, password, dbname)
	db, err = sql.Open("postgres", psqlconn)
	Checkerror(err)

	err = db.Ping()

	if err != nil {
		Checkerror(err)
	}
	fmt.Println("Successfully connected!")
}
func main() {

	myRoute := mux.NewRouter()
	myRoute.HandleFunc("/create_book", createBookApi).Methods("POST")
	myRoute.HandleFunc("/get_book", getBookApi).Methods("GET")
	myRoute.HandleFunc("/delete_book", deleteBook).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", myRoute))
	defer db.Close()

}
func Checkerror(err error) {
	if err != nil {
		panic(err)
	}
}

type Book struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	NoOfChapters  int    `json:"no_of_chapters"`
	PublishedDate string `json:"published_date"`
}

func createBookApi(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var createBook Book
	err = json.Unmarshal(reqBody, &createBook)

	if err != nil {
		fmt.Fprint(w, "Something went wrong")
		return
	}

	insertDynStm := `insert into  "Book" ("title","author","no_of_chapters","published_date") values($1,$2,$3,$4)`
	_, err = db.Exec(insertDynStm, createBook.Title, createBook.Author, createBook.NoOfChapters, createBook.PublishedDate)
	Checkerror(err)
	fmt.Fprint(w, "Book Successfully Created")
}
func getBookApi(w http.ResponseWriter, r *http.Request) {
	var books []Book
	getDynStm := `SELECT * FROM  "Book"`
	getBook, err := db.Query(getDynStm)

	Checkerror(err)
	for getBook.Next() {
		var singleBook Book
		err := getBook.Scan(&singleBook.Title, &singleBook.Author, &singleBook.NoOfChapters, &singleBook.PublishedDate)
		Checkerror(err)
		books = append(books, singleBook)
	}

	json.NewEncoder(w).Encode(books)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var deleteBook Book
	err = json.Unmarshal(reqBody, &deleteBook)
	if err != nil {
		fmt.Fprint(w, "Something went wrong")
		return
	}

	insertDynStm := `delete from  "Book" where author=$1;`

	_, err = db.Exec(insertDynStm, deleteBook.Author)
	Checkerror(err)
	fmt.Fprint(w, "Deleted Sucessfully")
}
