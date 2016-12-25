package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const hostPort = "3000"

func main() {
	// Seed for fake skill score
	rand.Seed(time.Now().Unix())

	ds := &mockDataStore{}
	addFakeData(ds)
	http.Handle("/api/skills", httpHandlers(ds))
	log.Println("Listening on", hostPort)
	http.ListenAndServe(fmt.Sprintf(":%s", hostPort), nil)
}

func addFakeData(ds *mockDataStore) {
	user1 := make(map[string]skillScore)
	user1["Golang"] = skillScore{
		SkillName: "Golang",
		Score: 100,
		LastScored: time.Now().Add(-24 * time.Hour),
	}


	user1["Kafka"] = skillScore{
		SkillName: "Kafka",
		Score: 100,
		LastScored: time.Now().Add(-24 * time.Hour),
	}

	ds.data = make(map[string]map[string]skillScore)
	ds.data["user1"] = user1

}

func httpHandlers(ds *mockDataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getHandler(ds, w, r)
		} else if r.Method == http.MethodPost {
			postHandler(ds, w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}


func postHandler(ds *mockDataStore, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	reqBody := &requestBody{}
	e := decoder.Decode(reqBody)
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	for _, skill := range reqBody.Skills {
		processSkill(ds, reqBody.ID, skill)
	}

	w.WriteHeader(http.StatusOK)
}

func processSkill(ds *mockDataStore, userID string, skillName string) {
	// simulate long processing time
	time.Sleep(5 * time.Second)


	score := skillScore{
		SkillName: skillName,
		LastScored: time.Now(),
		Score: rand.Float32() * 100,
	}

	ds.WriteData(userID, score)
}

func getHandler(ds *mockDataStore, w http.ResponseWriter, r * http.Request) {
	userId := r.URL.Query().Get("userID")
	skill := r.URL.Query().Get("skill")

	if userId == "" || skill == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userSkills, ok := ds.ReadData(userId, skill)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		encoder.Encode(userSkills)
	}
}

