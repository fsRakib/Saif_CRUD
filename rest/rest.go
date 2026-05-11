package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

type Zone struct {
	ID         int      `json:"id"`
	Name       *string  `json:"name"`
	CodeName   *string  `json:"code_name"`
	ZoneFactor *float64 `json:"zone_factor"`
}

var db *sql.DB

func main() {
	connStr := "host=localhost port=5432 user=rakib password=123 dbname=test_map_eta sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//Test DB connection

	err = db.Ping()
	if err != nil {
		log.Fatal("not able to connect to database: ", err)
	}

	log.Println("Connected to database successfully")
	// http.HandleFunc: The most common way to register a route. It maps a URL pattern to a function.
	http.HandleFunc("/zones", getZones)
	http.HandleFunc("/zone", getZoneByID)
	http.HandleFunc("/zones/create", createZone)
	http.HandleFunc("/zones/update", updateZone)
	http.HandleFunc("/zones/delete", deleteZone)

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getZones(w http.ResponseWriter, r *http.Request) {
	// Query to fetch zones from the database
	rows, err := db.Query("SELECT id, name, zone_factor FROM zone")

	if err != nil {
		http.Error(w, "Failed to fetch zones", http.StatusInternalServerError)
		return
	}
	// closes the database connection and frees resources associated with the query result.
	defer rows.Close()

	// Iterate through the rows and build a slice of zones
	var zones []Zone
	for rows.Next() {
		var z Zone
		// reads data from the current row and stores it into your Go variables
		err := rows.Scan(&z.ID, &z.Name, &z.ZoneFactor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		zones = append(zones, z)
	}
	// rows.Close() will run here automatically

	//return zones as JSON
	w.Header().Set("Content-Type", "application/json")

	// Converts Go data to JSON and sends it to the client
	json.NewEncoder(w).Encode(zones)

}

// w is where the JSON will be sent (your HTTP response)


func getZoneByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		http.Error(w, "Zone ID is required", http.StatusBadRequest)
		return
	}
	var z Zone
	err := db.QueryRow("SELECT id, name, zone_factor FROM zone WHERE id= $1", id).Scan(&z.ID, &z.Name, &z.ZoneFactor)

	if err == sql.ErrNoRows {
		http.Error(w, "Zone not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to fetch zone", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(z)
}

//=================================================================

func createZone(w http.ResponseWriter, r *http.Request) {

	var z Zone

	// Reads JSON from the HTTP request body and converts it into your Go struct
	err := json.NewDecoder(r.Body).Decode(&z)

	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = db.QueryRow("INSERT INTO zone (name, code_name, zone_factor) values ($1, $2, $3) RETURNING id",
		z.Name, z.CodeName, z.ZoneFactor).Scan(&z.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(z)
}

func updateZone(w http.ResponseWriter, r *http.Request) {
	var z Zone

	// Reads JSON from request body and converts it into your Go struct z
	err := json.NewDecoder(r.Body).Decode(&z)

	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if z.ID == 0 {
		http.Error(w, "Zone id required", http.StatusBadRequest)
		return
	}

	// The number = argument position
	result, err := db.Exec(`UPDATE zone SET name= $1, code_name= $2, zone_factor= $3 WHERE id= $4`, z.Name, z.CodeName, z.ZoneFactor, z.ID)

	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowAffected == 0 {
		http.Error(w, "Zone not found", http.StatusNotFound)
		return

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(z)
}

func deleteZone(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		http.Error(w, "Zone id need ", http.StatusBadRequest)
		return
	}

	res, err := db.Exec("delete from zone where id = $1", id)

	// w is just the delivery system (like a mailman)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowEffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowEffected == 0 {
		http.Error(w, "Zone not found", http.StatusNotFound)
	}
	//send an HTTP response status code to the client
	w.WriteHeader(http.StatusNoContent)
}
