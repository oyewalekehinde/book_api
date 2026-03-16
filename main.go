package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var client *mongo.Client
var err error
var bookDatabase *mongo.Database
var bookCollection *mongo.Collection

func createBook(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var createBook Book
	err := json.Unmarshal(reqBody, &createBook)

	if err != nil {
		fmt.Fprint(w, "Something went wrong")
		return
	}

	bookResult, err := bookCollection.InsertOne(context.TODO(), createBook)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(bookResult.InsertedID)
	fmt.Fprint(w, "Book Successfully Created")
}
func getBooks(w http.ResponseWriter, _ *http.Request) {
	var books []Book

	data, err := bookCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer data.Close(context.TODO())
	for data.Next(context.TODO()) {
		var book Book

		if err := data.Decode(&book); err != nil {
			log.Fatal(err)

		}
		books = append(books, book)
	}
	response := CustomResponse{
		Status:  true,
		Message: "Books fetched successfully",
		Data:    books,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func getBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, _ := bson.ObjectIDFromHex(idStr)
	data := bookCollection.FindOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		log.Fatal(err)
	}

	if err := data.Decode(&book); err != nil {
		log.Fatal(err)

	}

	response := CustomResponse{
		Status:  true,
		Message: "Book fetched successfully",
		Data:    book,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func deleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, _ := bson.ObjectIDFromHex(idStr)
	data, err := bookCollection.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, "DataBase Error", http.StatusInternalServerError)
	}
	if data.DeletedCount == 0 {
		response := CustomResponse{
			Status:  false,
			Message: "Book Not Found",
			Data:    nil,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CustomResponse{
		Status:  true,
		Message: "Books deleted successfully",
		Data:    nil,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	reqBody, _ := ioutil.ReadAll(r.Body)
	var updateBookProps Book
	err := json.Unmarshal(reqBody, &updateBookProps)
	if err != nil {
		fmt.Fprint(w, "Something went wrong")
		return
	}
	id, _ := bson.ObjectIDFromHex(idStr)
	data, err := bookCollection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.D{
		{Key: "$set", Value: updateBookProps},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(data.MatchedCount)
	response := map[string]interface{}{
		"status":  true,
		"message": "Book updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// func Handler(w http.ResponseWriter, r *http.Request) {
// 	if err := initMongo(); err != nil {
// 		http.Error(w, "failed to connect to database", http.StatusInternalServerError)
// 		return
// 	}

// 	router := mux.NewRouter()
// 	router.HandleFunc("/api/v1/book", createBook).Methods("POST")
// 	router.HandleFunc("/api/v1/books", getBooks).Methods("GET")
// 	router.HandleFunc("/api/v1/book/{id}", getBook).Methods("GET")
// 	router.HandleFunc("/api/v1/book/{id}", deleteBook).Methods("DELETE")
// 	router.HandleFunc("/api/v1/book/{id}", updateBook).Methods("PATCH")

//		router.ServeHTTP(w, r)
//	}
func Handler(w http.ResponseWriter, r *http.Request) {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://oyewalekehinde:Iam23yearsold@cluster0.cx7fyoz.mongodb.net/?appName=Cluster0").SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err = mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	bookDatabase = client.Database("Book")
	bookCollection = bookDatabase.Collection("Book")
	myRoute := mux.NewRouter()
	myRoute.HandleFunc("/api/v1/book", createBook).Methods("POST")
	myRoute.HandleFunc("/api/v1/books", getBooks).Methods("GET")
	myRoute.HandleFunc("/api/v1/book/{id}", getBook).Methods("GET")
	myRoute.HandleFunc("/api/v1/book/{id}", deleteBook).Methods("DELETE")
	myRoute.HandleFunc("/api/v1/book/{id}", updateBook).Methods("PATCH")
	myRoute.ServeHTTP(w, r)
	// defer func() {
	// 	if err = client.Disconnect(context.TODO()); err != nil {
	// 		panic(err)
	// 	}
	// }()

}
