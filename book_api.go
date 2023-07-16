package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func getBooks(w http.ResponseWriter, r *http.Request) {
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
	id, _ := primitive.ObjectIDFromHex(idStr)
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
	id, _ := primitive.ObjectIDFromHex(idStr)
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
	id, _ := primitive.ObjectIDFromHex(idStr)
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
func main() {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://oyewalekehinde:Iam23yearsold@cluster0.cx7fyoz.mongodb.net/?retryWrites=true&w=majority").SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err = mongo.Connect(context.TODO(), opts)
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
	myRoute.HandleFunc("/book", createBook).Methods("POST")
	myRoute.HandleFunc("/book", getBooks).Methods("GET")
	myRoute.HandleFunc("/book/{id}", getBook).Methods("GET")
	myRoute.HandleFunc("/book/{id}", deleteBook).Methods("DELETE")
	myRoute.HandleFunc("/book/{id}", updateBook).Methods("PATCH")
	log.Fatal(http.ListenAndServe(":8080", myRoute))
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

}
