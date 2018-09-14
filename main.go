package main

import (
	"Gopher/Model"
	"database/sql"
	"encoding/json"
	"net/http"
	
	//mysql db 
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	//mongo db 
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var items []item.Inventory

//HandleRequest : handle for all Request
func HandleRequest() {
	r := mux.NewRouter()
	r.HandleFunc("/", getIndex).Methods("GET")
	r.HandleFunc("/items", getIndex).Methods("GET")
	r.HandleFunc("/items/{id}", getItem).Methods("GET")
	r.HandleFunc("/items", saveItem).Methods("POST")
	r.HandleFunc("/items/{id}", deleteItem).Methods("DELETE")
	r.HandleFunc("/items/{id}", updateItem).Methods("PUT")
	http.ListenAndServe(":6000", r)
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&items)
}

func saveItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var createItem item.Inventory
	valid := true
	json.NewDecoder(r.Body).Decode(&createItem)
	for _, i := range items {
		if i.Barcode == createItem.Barcode {
			valid = false
		}
	}
	if valid {
		items = append(items, createItem)
		db := ConfigDb()
		//mysql command insert
		db.Prepare("insert into item(barcode,itemname,amt) values(?,?,?)")
		db.Exec(createItem.Barcode, createItem.Itemname, createItem.Price)
		
		//mongo command insert
		dbmgo.C("items").Insert(createItem)
		
		defer db.Close()
	}

	json.NewEncoder(w).Encode(&items)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, i := range items {
		if i.Barcode == params["id"] {
			items = append(items[:index], items[index+1:]...)
			db := ConfigDb()
			//mysql command
			db.Prepare("delete from item barcode=?")
			//mongo command
			dbmgo.C("items").Remove(bson.M{"barcode": i.Barcode})
			db.Exec(i.Barcode)
			defer db.Close()
		}
	}
	json.NewEncoder(w).Encode(&items)
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var createItem item.Inventory
	json.NewDecoder(r.Body).Decode(&createItem)
	for index, i := range items {
		if i.Barcode == params["id"] {
			createItem.Barcode = i.Barcode
			items = append(items[:index], items[index+1:]...)
			items = append(items, createItem)
			db := ConfigDb()
			//mysql command
			db.Prepare("update set itemname=?,amt=? where barcode=?")
			db.Exec(createItem.Itemname, createItem.Price, createItem.Barcode)
			//mongodb
			dbmgo.C("items").Update(bson.M{"Barcode": i.Barcode}, item.Inventory{Barcode: createItem.Barcode, Itemname: createItem.Itemname, Price: createItem.Price})

			defer db.Close()

		}
	}

	json.NewEncoder(w).Encode(&items)
}

func getItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, i := range items {
		if params["id"] == i.Barcode {
			json.NewEncoder(w).Encode(&i)
		}

	}
}

var dbmgo *mgo.Database

//ConfigDb : import items
func ConfigDb() *sql.DB {

	db, err := sql.Open("mysql", "root:pass@tcp(localhost:3309)/dbname") //mysql user:password@tcp(address:port)/dbname
	if err != nil {
		panic(err)
	}

	session, err := mgo.Dial("localhost") //mongo address
	if err != nil {
		panic(err)
	}
	dbmgo = session.DB("sampledatabase") //mongo database
	return db

}

//Index : Intiate Database
func Index() {
	db := ConfigDb()
	//mysql command
	qry, err := db.Query("select barcode,itemname,amt from item where itemid")

	if err != nil {
		panic(err.Error())
	}
	//mysql qry return collection
	for qry.Next() {
		var iBarcode string
		var iItemname string
		var iPrice float64
		err = qry.Scan(&iBarcode, &iItemname, &iPrice)
		
		items = append(items, item.Inventory{Barcode: iBarcode, Itemname: iItemname, Price: iPrice})//just adding it in array
		//for mongo db command insert transferring data to mongodb from mysql
		dbmgo.C("items").Insert(item.Inventory{Barcode: iBarcode, Itemname: iItemname, Price: iPrice})
	}
	defer db.Close()
}
func main() {
	Index()
	HandleRequest()
}
