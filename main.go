package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID        string    `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string    `json:"name,omitempty" bson:"name,omitempty"`
	Lastname  string    `json:"lastname,omitempty" bson:"lastname,omitempty"`
	Birthdate time.Time `json:"birthdate,omitempty" bson:"birthdate,omitempty"`
	Email     string    `json:"email,omitempty" bson:"email,omitempty"`
}

var client *mongo.Client

func main() {
	fmt.Println("Starting the application...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	router := mux.NewRouter()
	router.HandleFunc("/users", createUser).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func createUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var user User
	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	collection := client.Database("test").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	sendEmail(user.Email, "Başarıyla kayıt oldunuz. Sitemize hoş geldiniz!")

	responseData := struct {
		Message string `json:"message"`
	}{
		Message: "Başarıyla kayıt oldunuz. Sitemize hoş geldiniz!",
	}

	json.NewEncoder(response).Encode(responseData)
}

func sendEmail(to, message string) error {
	from := "your-email@gmail.com" // Buraya e-posta adresinizi yazın
	password := "your-password"    // Buraya e-posta şifrenizi yazın

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(message))
	if err != nil {
		log.Println("Failed to send email:", err)
		return err
	}

	return nil
}
