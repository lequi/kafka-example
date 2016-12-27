package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	hostPort = "3000"
	topicName = "kafka-example-topic"
)

func main() {
	// Seed for fake skill score
	rand.Seed(time.Now().Unix())

	producer, err := createKafkaProducer("127.0.0.1:9092")

	if err != nil {
		log.Fatal("Failed to connect to Kafka")
	}
	
	ds := &mockDataStore{}
	addFakeData(ds)
	http.Handle("/api/skills", httpHandlers(ds, producer.Input()))
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

func httpHandlers(ds *mockDataStore, c chan <- *sarama.ProducerMessage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getHandler(ds, w, r)
		} else if r.Method == http.MethodPost {
			postHandler(c, w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}


func postHandler(c chan <- *sarama.ProducerMessage, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	reqBody := &requestBody{}
	e := decoder.Decode(reqBody)
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	for _, skill := range reqBody.Skills {
		processSkill(c, reqBody.ID, skill)
	}

	w.WriteHeader(http.StatusOK)
}

func processSkill(c chan <- *sarama.ProducerMessage, userID string, skillName string) {
	body := skillScoreMessage{
		SkillName: skillName,
		ProfileID: userID,
	}

	bodyBytes, _ := json.Marshal(body)

	msg := &sarama.ProducerMessage{
		Topic: topicName,
		Key: sarama.StringEncoder(userID),
		Timestamp: time.Now(),
		Value: sarama.ByteEncoder(bodyBytes),
	}

	c <- msg
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

