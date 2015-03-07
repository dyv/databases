package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/sqs/mux"
)

var db *sql.DB
var store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))

func init() {
	log.SetFlags(log.Lshortfile)
	var err error
	db, err = sql.Open("mysql", "root:dylan@tcp(127.0.0.1:3306)/store")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	// create the tables if they don't already exist
	CreateCustomerTable(db)
	CreateItemTable(db)

	CreateInventoryTable(db)
	CreateReservationTable(db)
	CreateSoldTable(db)
}
func isSignedIn(s *sessions.Session) bool {
	v, ok := s.Values["signed-in"]
	if !ok {
		return false
	}
	return v.(bool)
}

func isCustomer(s *sessions.Session) bool {
	v, ok := s.Values["customer"]
	if !ok {
		return false
	}
	return v.(bool)
}

func isSupplier(s *sessions.Session) bool {
	v, ok := s.Values["supplier"]
	if !ok {
		return false
	}
	return v.(bool)
}

var templates = template.Must(
	template.ParseFiles(
		"signin.html",
		"signup.html",
		"home.html",
		"view.html",
		"404.html"))

func errorHandler(w http.ResponseWriter, r *http.Request, err string) {
	templates.ExecuteTemplate(w, "404.html", err)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "home.html", nil)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

type View struct {
	MyItems      []Item
	Reservations []Item
	Bought       []Item
	Inventory    []Item
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "default")
	if !isSignedIn(session) || !isCustomer(session) {
		signinPageHandler(w, r)
		return
	}
	name := session.Values["uname"].(string)
	tx, err := db.Begin()
	if err != nil {
		errorHandler(w, r, err.Error())
		return
	}
	defer tx.Rollback()
	var v View
	v.Inventory = GetInventory(tx)
	v.Bought = GetBought(tx, name)
	v.Reservations = GetReserved(tx, name)
	v.MyItems = GetMyItems(tx, name)
	err = tx.Commit()
	if err != nil {
		errorHandler(w, r, err.Error())
		return
	}
	templates.ExecuteTemplate(w, "view.html", v)
}

func reserveHandler(w http.ResponseWriter, r *http.Request) {
	// check if signed in

	session, _ := store.Get(r, "default")
	if !isSignedIn(session) || !isCustomer(session) {
		signinPageHandler(w, r)
		return
	}
	uname := session.Values["uname"].(string)
	item_id, err := strconv.Atoi(r.FormValue("item_id"))
	if err != nil {
		fmt.Fprintf(w, "bad item id")
		return
	}
	tx, err := db.Begin()
	if err != nil {
		errorHandler(w, r, err.Error())
		return
	}
	defer tx.Rollback()
	Reserve(tx, uname, item_id)
	tx.Commit()
	session.Save(r, w)
	http.Redirect(w, r, "/view", http.StatusFound)
}

func buyHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "default")
	if !isSignedIn(session) || !isCustomer(session) {
		signinPageHandler(w, r)
		return
	}
	uname := session.Values["uname"].(string)
	item_id, err := strconv.Atoi(r.FormValue("item_id"))
	if err != nil {
		fmt.Fprintf(w, "bad item id")
		return
	}
	tx, err := db.Begin()
	if err != nil {
		errorHandler(w, r, err.Error())
		return
	}
	defer tx.Rollback()
	Buy(tx, uname, item_id)
	tx.Commit()
	session.Save(r, w)
	http.Redirect(w, r, "/view", http.StatusFound)
}

func addPageHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "default")
	if !isSignedIn(session) || !isCustomer(session) {
		signinPageHandler(w, r)
		return
	}

}

func addHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "default")
	if !isSignedIn(session) || !isSupplier(session) {
		signinPageHandler(w, r)
		return
	}
	supplier, _ := session.Values["uname"].(string)
	name := r.PostFormValue("name")
	if name == "" {
		fmt.Fprintf(w, "invalid name")
	}
	price, err := strconv.ParseFloat(r.PostFormValue("price"), 64)
	if err != nil {
		fmt.Fprintf(w, "invalid price")
	}
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	AddItem(tx, name, price, supplier)
	tx.Commit()
	http.Redirect(w, r, "/view", http.StatusFound)
}

func signinPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("singin handler")
	templates.ExecuteTemplate(w, "signin.html", nil)
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "default")
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	name := r.PostFormValue("username")
	pwd := GetCustomer(tx, name)
	log.Println(name, pwd)
	if pwd != r.PostFormValue("password") {
		fmt.Fprintf(w, "wrong password")
		return
	}
	err = tx.Commit()
	if err != nil {
		errorHandler(w, r, err.Error())
	}
	session.Values["signed-in"] = true
	session.Values["uname"] = name
	session.Values["customer"] = true
	session.Values["supplier"] = true
	session.Save(r, w)
	http.Redirect(w, r, "/view", http.StatusFound)
}

func signupPageHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "signup.html", nil)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Signing Up")
	log.Println(r)
	session, _ := store.Get(r, "default")
	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintf(w, "failed to begin:", err)
		return
	}
	log.Println("session: ", session)
	defer tx.Rollback()
	err = r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "invalid form")
		return
	}
	name := r.PostFormValue("username")
	if name == "" {
		fmt.Fprintf(w, "no username")
	}
	log.Println("name: ", name)
	old_cust := GetCustomer(tx, name)
	if old_cust != "" {
		fmt.Fprintf(w, "username taken")
		return
	}
	pwd := r.PostFormValue("password")
	if len(pwd) < 4 {
		fmt.Fprintf(w, "password too short")
		return
	}
	if len(name) < 4 {
		fmt.Fprintf(w, "username too short")
		return
	}
	supplier := r.PostFormValue("supplier")
	supp := supplier == "on"
	NewCustomer(tx, name, 10000, pwd, supp)
	err = tx.Commit()
	if err != nil {
		errorHandler(w, r, err.Error())
	}
	log.Println("name: ", name)
	log.Println("pwd: ", pwd)
	session.Values["signed-in"] = true
	session.Values["uname"] = name
	session.Values["customer"] = true
	if supp {
		session.Values["supplier"] = true
	}
	err = session.Save(r, w)
	if err != nil {
		errorHandler(w, r, err.Error())
		return
	}
	log.Println("redirecting to view")
	http.Redirect(w, r, "/view", http.StatusFound)
}

func signoutHandler(w http.ResponseWriter, r *http.Request) {
	// redirect and delete cookie
	session, _ := store.Get(r, "default")
	session.Values["signed-in"] = false
	session.Values["uname"] = ""
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/static/", staticHandler)
	r.HandleFunc("/signin", signinPageHandler).Methods("GET")
	r.HandleFunc("/signin", signinHandler).Methods("POST")
	r.HandleFunc("/signup", signupPageHandler).Methods("GET")
	r.HandleFunc("/signup", signupHandler).Methods("POST")
	r.HandleFunc("/signout", signoutHandler)

	r.HandleFunc("/view", viewHandler)

	r.HandleFunc("/reserve", reserveHandler).Methods("POST")
	r.HandleFunc("/buy", buyHandler).Methods("POST")

	// for suppliers
	r.HandleFunc("/add", addPageHandler).Methods("GET")
	r.HandleFunc("/add", addHandler).Methods("POST")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
