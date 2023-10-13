package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
)

const (
	dbHost     = "postgresql-db"
	dbPort     = 5432
	dbUser     = "asha"
	dbPassword = "okidoki"
	dbName     = "mydatabase"
)

var db *sql.DB

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/items", getItems).Methods("GET")
	r.HandleFunc("/items/{id:[0-9]+}", getItem).Methods("GET")
	r.HandleFunc("/items", createItem).Methods("POST")

	http.Handle("/", r)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("El servidor está escuchando en el puerto 8080")
	log.Fatal(server.ListenAndServe())
}

func initDB() {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	
	// Crear la tabla "items" si no existe
        createTableQuery := `
        CREATE TABLE IF NOT EXISTS items (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255)
        );
        `
        
        _, err = db.Exec(createTableQuery)
        if err != nil {
            log.Fatal(err)
        }

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "¡Bienvenido a la API en Go con PostgreSQL!")
}

func getItems(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name FROM items")
	if err != nil {
		log.Fatal(err)
		http.Error(w, "No se pudieron recuperar los elementos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			log.Fatal(err)
			http.Error(w, "No se pudieron recuperar los elementos", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func getItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["id"]

	var item Item
	err := db.QueryRow("SELECT id, name FROM items WHERE id = $1", itemID).Scan(&item.ID, &item.Name)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Elemento no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		log.Fatal(err)
		http.Error(w, "Datos de solicitud no válidos", http.StatusBadRequest)
		return
	}

	var newItem Item
	err := db.QueryRow("INSERT INTO items (name) VALUES ($1) RETURNING id", item.Name).Scan(&newItem.ID)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Error al crear el elemento", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newItem)
}

