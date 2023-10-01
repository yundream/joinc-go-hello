package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

type Counter struct {
	gorm.Model
	Name  string `gorm:"size:16"`
	Count int
}

func main() {
	_, err := GetDatabaseInstance()
	if err != nil {
		fmt.Println(">> ", err.Error())
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}

	r := mux.NewRouter()
	fmt.Println("Server start!!")
	r.HandleFunc("/api/count", getCounter).Methods("GET")
	r.HandleFunc("/api/count", addCounter).Methods("POST")
	http.ListenAndServe(":8000", r)
}

func addCounter(w http.ResponseWriter, r *http.Request) {
	tx := db.Exec("UPDATE counters SET count=count+1 WHERE name='count'")
	if tx.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error:%s", tx.Error.Error())
		return
	}
}

func getCounter(w http.ResponseWriter, r *http.Request) {
	readCounter := &Counter{}
	tx := db.Where("name=?", "count").First(&readCounter)
	if tx.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error:%s", tx.Error.Error())
		return
	}
	fmt.Fprintf(w, "%d", readCounter.Count)
}

func GetDatabaseInstance() (*gorm.DB, error) {
	var (
		retry    int
		database *gorm.DB
		err      error
	)
	retry = 0
	dbUser := os.Getenv("MYSQL_USER")
	dbPass := os.Getenv("MYSQL_PASSWORD")
	dbName := os.Getenv("MYSQL_DATABASE")
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")

	createDBDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
	)

	for retry < 3 {
		time.Sleep(10 * time.Second)
		database, err = gorm.Open(mysql.Open(createDBDsn), &gorm.Config{})
		if err == nil {
			break
		}
		retry++
	}
	if err != nil {
		fmt.Println(">> ", err.Error(), createDBDsn)
		return nil, err
	}

	tx := database.Exec("CREATE DATABASE IF NOT EXISTS " + dbName + ";")
	if tx.Error != nil {
		return nil, tx.Error
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
		dbName,
	)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	isTable := db.Migrator().HasTable(&Counter{})
	fmt.Println("isTable", isTable)
	err = db.AutoMigrate(&Counter{})
	if err == nil {
		if !isTable {
			db.Create(&Counter{Name: "count", Count: 0})
		}
	}
	return db, nil
}
