package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Book struct {
	ID            bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title         string        `json:"title,omitempty" bson:"title,omitempty"`
	Author        string        `json:"author,omitempty" bson:"author,omitempty"`
	NoOfChapters  int           `json:"no_of_chapters,omitempty" bson:"no_of_chapters,omitempty"`
	PublishedDate string        `json:"published_date,omitempty" bson:"published_date,omitempty"`
}

type CustomResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var (
	client         *mongo.Client
	bookDatabase   *mongo.Database
	bookCollection *mongo.Collection
	initOnce       sync.Once
	initErr        error
)

func initMongo() error {
	initOnce.Do(func() {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)

		uri := os.Getenv("MONGODB_URI")
		opts := options.Client().
			ApplyURI(uri).
			SetServerAPIOptions(serverAPI)

		client, initErr = mongo.Connect(opts)
		if initErr != nil {
			return
		}

		initErr = client.Database("admin").
			RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).
			Err()
		if initErr != nil {
			return
		}

		bookDatabase = client.Database("Book")
		bookCollection = bookDatabase.Collection("Book")
	})

	return initErr
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var createBook Book
	if err := json.NewDecoder(r.Body).Decode(&createBook); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	bookResult, err := bookCollection.InsertOne(context.TODO(), createBook)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  true,
		"message": "Book successfully created",
		"id":      bookResult.InsertedID,
	})
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	var books []Book

	data, err := bookCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer data.Close(context.TODO())

	for data.Next(context.TODO()) {
		var book Book
		if err := data.Decode(&book); err != nil {
			http.Error(w, "decode error", http.StatusInternalServerError)
			return
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
	vars := mux.Vars(r)

	id, err := bson.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var book Book
	data := bookCollection.FindOne(context.TODO(), bson.M{"_id": id})
	if err := data.Decode(&book); err != nil {
		http.Error(w, "book not found", http.StatusNotFound)
		return
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

	id, err := bson.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	data, err := bookCollection.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if data.DeletedCount == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CustomResponse{
			Status:  false,
			Message: "Book not found",
			Data:    nil,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CustomResponse{
		Status:  true,
		Message: "Book deleted successfully",
		Data:    nil,
	})
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := bson.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var updateBookProps Book
	if err := json.NewDecoder(r.Body).Decode(&updateBookProps); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	_, err = bookCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": id},
		bson.D{{Key: "$set", Value: updateBookProps}},
	)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  true,
		"message": "Book updated successfully",
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if err := initMongo(); err != nil {
		http.Error(w, "failed to connect to database", http.StatusInternalServerError)
		return
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/book", createBook).Methods("POST")
	router.HandleFunc("/api/v1/books", getBooks).Methods("GET")
	router.HandleFunc("/api/v1/book/{id}", getBook).Methods("GET")
	router.HandleFunc("/api/v1/book/{id}", deleteBook).Methods("DELETE")
	router.HandleFunc("/api/v1/book/{id}", updateBook).Methods("PATCH")

	router.ServeHTTP(w, r)
}
