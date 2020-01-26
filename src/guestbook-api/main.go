package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/xyproto/simpleredis"
)

type version struct {
	Version string `json:"version"`
}

type health struct {
	Healthy bool `json:"healthy"`
}

type hits struct {
	HitsCounter int `json:"hitsCounter"`
}

type apiMessage struct {
	APIMessage interface{} `json:"APIMessage"`
}

type Message struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Name string `json:"name"`
	Date string `json:"date"`
}

var (
	db         DB
	masterPool *simpleredis.ConnectionPool
	slavePool  *simpleredis.ConnectionPool
)

const (
	ver           string = "1.0"
	hitsDbKey     string = "hitsDbKey"
	messagePrefixKey string = "/messages"
)

func newRouter() *mux.Router {
	log.Println("Creating a new Router")
	r := mux.NewRouter()
	r.HandleFunc("/", getIndex).Methods("GET")

	r.HandleFunc("/version", getVersion).Methods("GET")
	r.HandleFunc("/healthz", getHealthz).Methods("GET")
	r.HandleFunc("/hostname", getHostname).Methods("GET")

	r.HandleFunc("/hits", getHits).Methods("GET")
	r.HandleFunc("/hits", postHits).Methods("POST")
	r.HandleFunc("/hits", deleteHits).Methods("DELETE")

	r.HandleFunc("/message", postMessage).Methods("POST")
	r.HandleFunc("/message/{id}", getMessage).Methods("GET")
	r.HandleFunc("/message/{id}", deleteMessage).Methods("DELETE")

	r.HandleFunc("/messages", getMessages).Methods("GET")

	return r
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method + r.URL.Path)
	json.NewEncoder(w).Encode(apiMessage{APIMessage: "Guestbook API"})
}

func getVersion(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method + r.URL.Path)
	json.NewEncoder(w).Encode(version{Version: ver})
}

func getHealthz(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method + r.URL.Path)

	healty := true
	err := db.healthz()
	if err != nil {
		healty = false
	}
	json.NewEncoder(w).Encode(health{Healthy: healty})
}

func getHostname(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method + r.URL.Path)

	hostname, err := os.Hostname()
	if err != nil {
		json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
		return
	}
	msg := fmt.Sprintf("Hostname=%s", hostname)
	json.NewEncoder(w).Encode(apiMessage{APIMessage: msg})
}

func getHits(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method + r.URL.Path)

	redisData, err := db.get(hitsDbKey)
	if err != nil {
		json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
		return
	}
	objectsNr := redisData.(string)
	inc, err := strconv.Atoi(objectsNr)
	json.NewEncoder(w).Encode(hits{HitsCounter: inc})
}

func postHits(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method + r.URL.Path)

	redisData, err := db.get(hitsDbKey)
	if err != nil {
		json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
		return
	}

	hits := redisData.(string)
	inc, err := strconv.Atoi(hits)
	if err != nil {
		json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
		return
	}
	inc++
	hits = strconv.Itoa(inc)

	err = db.set(hitsDbKey, hits)
	if err != nil {
		json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
		return
	}

	json.NewEncoder(w).Encode(apiMessage{APIMessage: "The hits was increased"})
}

func deleteHits(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method + r.URL.Path)

	err := db.set(hitsDbKey, "0")
	if err != nil {
		json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
		return
	}
	json.NewEncoder(w).Encode(apiMessage{APIMessage: "The hits counter was reset."})
}

func postMessage(w http.ResponseWriter, r *http.Request) {

	var newMsg Message
	log.Println(r.Method + r.URL.Path)

	err := json.NewDecoder(r.Body).Decode(&newMsg)
	if err != nil {
		json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
		return
	}

	currentTime := time.Now()
	newMsg.Date = currentTime.Format("2006.01.02 15:04:05")
	newMsg.ID = uuid.New().String()
	
	// serialize User object to JSON
	newMsgJson, err := json.Marshal(newMsg)
	if err != nil {
		json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
		return
	}

	err = db.set(messagePrefixKey+"/"+newMsg.ID, string(newMsgJson))
	if err != nil {
		json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
		return
	}

	json.NewEncoder(w).Encode(apiMessage{APIMessage: "Message added successful"})
}

func deleteMessage(w http.ResponseWriter, r *http.Request) {

	log.Println(r.Method + r.URL.Path)

	msgID := strings.TrimPrefix(r.URL.Path, "/message/")
	key := messagePrefixKey + "/" + msgID

	db.delete(key)
	json.NewEncoder(w).Encode(apiMessage{APIMessage: "Message deleted successful."})
}

func getMessage(w http.ResponseWriter, r *http.Request) {

	log.Println(r.Method + r.URL.Path)

	msgID := strings.TrimPrefix(r.URL.Path, "/message/")
	msgJson, err := db.get(messagePrefixKey + "/" + msgID)

	if err != nil {
		json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
		return
	}

	msg := Message{}
	err = json.Unmarshal([]byte(msgJson.(string)), &msg)

	json.NewEncoder(w).Encode(msg)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method + r.URL.Path)

	keys, err := db.getKeys()
	if err != nil {
		json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
		return
	}

	results := []Message{}

	for _, key := range keys{
		msgJson, err := db.get(key)
		if err != nil {
			json.NewEncoder(w).Encode(apiMessage{APIMessage: err})
			return
		}
		msg := Message{}
		err = json.Unmarshal([]byte(msgJson.(string)), &msg)
		results = append(results, msg)
	}
	
	json.NewEncoder(w).Encode(results)
}

func initDB() {
	log.Println("Inititilize connection to DB")

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	if redisHost == "" || redisPort == "" {
		log.Fatal("REDIS_HOST and REDIS_PORT must be specified.")
	}

	dbAddr := redisHost + ":" + redisPort
	db = DB{}
	db.newClient(dbAddr)
	db.initKey(hitsDbKey, "0")
}

func main() {
	log.Println("Starting the app..")
	initDB()
	r := newRouter()
	http.ListenAndServe(":8080", r)
}
