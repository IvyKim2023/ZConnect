package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	elastic "gopkg.in/olivere/elastic.v3"
)

// Update with your own Elastic Search URL installed on GCE,
// your bucket name, and your own project id and index
const (
	INDEX       = "zconnect"
	TYPE        = "post"
	ES_URL      = "http://34.45.28.179:9200"
	BUCKET_NAME = "post-images-429209"
	PROJECT_ID  = "around-429209"
)

type Post struct {
	User      string `json:"user"`
	Message   string `json:"message"`
	Url       string `json:"url"`
	Timestamp string `json:"timestamp"`
}

var SigningKey = []byte("secret secret key")

// Create a post
func handlerPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	user := r.Context().Value("user")
	claims := user.(*jwt.Token).Claims
	username := claims.(jwt.MapClaims)["username"]

	r.ParseMultipartForm(32 << 20)

	timestamp := time.Now().Format(time.RFC3339)

	p := &Post{
		User:      username.(string),
		Message:   r.FormValue("message"),
		Timestamp: timestamp,
	}

	id := uuid.New()

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "GCS is not setup", http.StatusInternalServerError)
		fmt.Printf("GCS is not setup %v\n", err)
		panic(err)
	}
	defer file.Close()

	ctx := context.Background()

	_, attrs, err := saveToGCS(ctx, file, BUCKET_NAME, id)
	if err != nil {
		http.Error(w, "GCS is not setup", http.StatusInternalServerError)
		fmt.Printf("GCS is not setup %v\n", err)
		panic(err)
	}

	p.Url = attrs.MediaLink

	saveToES(p, id)
}

// Save the image to the bucket
func saveToGCS(ctx context.Context, r io.Reader, bucketName, name string) (*storage.ObjectHandle, *storage.ObjectAttrs, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	if _, err := bucket.Attrs(ctx); err != nil {
		return nil, nil, err
	}

	obj := bucket.Object(name)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, r); err != nil {
		return nil, nil, err
	}
	if err := w.Close(); err != nil {
		return nil, nil, err
	}

	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return nil, nil, err
	}

	attrs, err := obj.Attrs(ctx)
	fmt.Printf("Post is saved to GCS: %s\n", attrs.MediaLink)
	return obj, attrs, err

}

// Save the post to elastic search
func saveToES(p *Post, id string) {
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}

	_, err = es_client.Index().
		Index(INDEX).
		Type(TYPE).
		Id(id).
		BodyJson(p).
		Refresh(true).
		Do()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Post is saved to index: %s\n", p.Message)
}

// Filter bad words in the post
func containsBadWords(s *string) bool {
	filteredWords := []string{
		"nigger",
		"nigero",
	}
	for _, word := range filteredWords {
		if strings.Contains(*s, word) {
			return true
		}
	}
	return false
}

// Get all the posts
func handlerSearchAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	user := r.Context().Value("user")
	claims := user.(*jwt.Token).Claims
	username := claims.(jwt.MapClaims)["username"].(string)

	friendsList, err := getFriends(username)
	if err != nil {
		http.Error(w, "Error retrieving friends list", http.StatusInternalServerError)
		return
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days == 0 {
		days = 10 // default to posts in 10 days
	}

	currentTime := time.Now().UTC()
	startTime := currentTime.AddDate(0, 0, -days).Format(time.RFC3339)

	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}

	boolQuery := elastic.NewBoolQuery()
	boolQuery = boolQuery.Must(elastic.NewRangeQuery("timestamp").Gte(startTime))
	boolQuery = boolQuery.Filter(elastic.NewTermsQuery("user", friendsList))

	searchResult, err := client.Search().
		Index(INDEX).
		Query(boolQuery).
		Sort("timestamp", false).
		Pretty(true).
		Do()
	if err != nil {
		panic(err)
	}

	var typ Post
	var ps []Post
	for _, item := range searchResult.Each(reflect.TypeOf(typ)) {
		p := item.(Post)
		fmt.Printf("Post by %s: %s at %s\n", p.User, p.Message, p.Timestamp)
		if !containsBadWords(&p.Message) {
			ps = append(ps, p)
		}
	}

	js, err := json.Marshal(ps)
	if err != nil {
		panic(err)
	}
	w.Write(js)
}

// Get all of my posts
func handlerSearchMy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	user := r.Context().Value("user")
	claims := user.(*jwt.Token).Claims
	username := claims.(jwt.MapClaims)["username"].(string)

	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}

	boolQuery := elastic.NewBoolQuery()
	boolQuery = boolQuery.Filter(elastic.NewTermsQuery("user", username))

	searchResult, err := client.Search().
		Index(INDEX).
		Query(boolQuery).
		Sort("timestamp", false).
		Pretty(true).
		Do()
	if err != nil {
		panic(err)
	}

	var typ Post
	var ps []Post
	for _, item := range searchResult.Each(reflect.TypeOf(typ)) {
		p := item.(Post)
		fmt.Printf("Post by %s: %s at %s\n", p.User, p.Message, p.Timestamp)
		if !containsBadWords(&p.Message) {
			ps = append(ps, p)
		}
	}

	js, err := json.Marshal(ps)
	if err != nil {
		panic(err)
	}

	w.Write(js)
}

// Get a certain user's posts
func handlerSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var reqData struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	username := reqData.Username
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}

	boolQuery := elastic.NewBoolQuery()
	boolQuery = boolQuery.Filter(elastic.NewTermsQuery("user", username))

	searchResult, err := client.Search().
		Index(INDEX).
		Query(boolQuery).
		Sort("timestamp", false).
		Pretty(true).
		Do()
	if err != nil {
		panic(err)
	}

	var typ Post
	var ps []Post
	for _, item := range searchResult.Each(reflect.TypeOf(typ)) {
		p := item.(Post)
		fmt.Printf("Post by %s: %s at %s\n", p.User, p.Message, p.Timestamp)
		if !containsBadWords(&p.Message) {
			ps = append(ps, p)
		}
	}

	js, err := json.Marshal(ps)
	if err != nil {
		panic(err)
	}

	w.Write(js)
}

// Get my friends
func getFriendsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight request
	if r.Method == "OPTIONS" {
		log.Println("OPTIONS request")
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Println("Actual request")
	w.Header().Set("Content-Type", "application/json")

	user := r.Context().Value("user")
	claims := user.(*jwt.Token).Claims
	username := claims.(jwt.MapClaims)["username"].(string)

	friends, err := getFriends(username)
	if err != nil {
		log.Printf("Error fetching friends for user %s: %v", username, err)
		if elastic.IsNotFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(friends)
}

func getFriends(username string) ([]string, error) {
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		return nil, err
	}

	res, err := client.Get().
		Index(INDEX).
		Type("user").
		Id(username).
		Do()
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(*res.Source, &user); err != nil {
		return nil, err
	}

	return user.Friends, nil
}

// Add a friend
func addFriendHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	user := r.Context().Value("user")
	claims := user.(*jwt.Token).Claims
	username := claims.(jwt.MapClaims)["username"].(string)

	decoder := json.NewDecoder(r.Body)
	var friendRequest struct {
		FriendUsername string `json:"friend_username"`
	}
	if err := decoder.Decode(&friendRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if friendRequest.FriendUsername == "" {
		http.Error(w, "Friend username is required", http.StatusBadRequest)
		return
	}

	if username == friendRequest.FriendUsername {
		http.Error(w, "You cannot add yourself as a friend", http.StatusBadRequest)
		return
	}

	exists, err := userExists(friendRequest.FriendUsername)
	if err != nil {
		http.Error(w, "Error checking friend existence", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	if err := addFriend(username, friendRequest.FriendUsername); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Friend added successfully"}
	json.NewEncoder(w).Encode(response)
}

func userExists(username string) (bool, error) {
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		return false, err
	}

	termQuery := elastic.NewTermQuery("username", username)
	queryResult, err := es_client.Search().
		Index(INDEX).
		Query(termQuery).
		Pretty(true).
		Do()
	if err != nil {
		return false, err
	}

	return queryResult.TotalHits() > 0, nil
}

func addFriend(username, friendUsername string) error {
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		return err
	}

	res, err := client.Get().
		Index(INDEX).
		Type("user").
		Id(username).
		Do()
	if err != nil {
		return err
	}

	var user User
	if err := json.Unmarshal(*res.Source, &user); err != nil {
		return err
	}

	user.Friends = append(user.Friends, friendUsername)

	_, err = client.Index().
		Index(INDEX).
		Type("user").
		Id(username).
		BodyJson(user).
		Refresh(true).
		Do()
	if err != nil {
		return err
	}

	return nil
}

// Delete a friend
func deleteFriendHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	user := r.Context().Value("user")
	claims := user.(*jwt.Token).Claims
	username := claims.(jwt.MapClaims)["username"].(string)

	decoder := json.NewDecoder(r.Body)
	var friendRequest struct {
		FriendUsername string `json:"friend_username"`
	}
	if err := decoder.Decode(&friendRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := deleteFriend(username, friendRequest.FriendUsername); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Friend deleted successfully"}
	json.NewEncoder(w).Encode(response)
}

func deleteFriend(username, friendUsername string) error {
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		return err
	}

	res, err := client.Get().
		Index(INDEX).
		Type("user").
		Id(username).
		Do()
	if err != nil {
		return err
	}

	var user User
	if err := json.Unmarshal(*res.Source, &user); err != nil {
		return err
	}

	updatedFriends := []string{}
	for _, friend := range user.Friends {
		if friend != friendUsername {
			updatedFriends = append(updatedFriends, friend)
		}
	}
	user.Friends = updatedFriends

	_, err = client.Index().
		Index(INDEX).
		Type("user").
		Id(username).
		BodyJson(user).
		Refresh(true).
		Do()
	if err != nil {
		return err
	}

	return nil
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	message := r.FormValue("message")
	timestamp := r.FormValue("timestamp")

	if username == "" || message == "" || timestamp == "" {
		http.Error(w, "Username, message, and timestamp are required", http.StatusBadRequest)
		return
	}

	success, err := deletePostByContent(username, message, timestamp)
	if err != nil {
		http.Error(w, "Could not delete post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !success {
		http.Error(w, "Post not found or could not be deleted", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Post deleted"))
}

func deletePostByContent(username, message, timestamp string) (bool, error) {
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		return false, err
	}

	boolQuery := elastic.NewBoolQuery().
		Must(elastic.NewTermQuery("user", username)).
		Must(elastic.NewTermQuery("message", message)).
		Must(elastic.NewTermQuery("timestamp", timestamp))

	searchResult, err := es_client.Search().
		Index(INDEX).
		Query(boolQuery).
		Pretty(true).
		Do()
	if err != nil {
		return false, err
	}

	if searchResult.TotalHits() == 0 {
		return false, nil
	}

	for _, hit := range searchResult.Hits.Hits {
		_, err := es_client.Delete().
			Index(INDEX).
			Type(TYPE).
			Id(hit.Id).
			Do()
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Create a client
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(INDEX).Do()
	if err != nil {
		panic(err)
	}
	if !exists {
		// Create a new index
		mapping := `{
					"mappings":{
							"post":{
									"properties":{
											"timestamp":{
												"type":"date",
												"format":"strict_date_optional_time||epoch_millis"
											}
									}
							}
					}
			}`
		_, err := client.CreateIndex(INDEX).Body(mapping).Do()
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("service-started")

	r := mux.NewRouter()

	var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return SigningKey, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	r.Handle("/post", jwtMiddleware.Handler(http.HandlerFunc(handlerPost))).Methods("POST")
	r.Handle("/search", jwtMiddleware.Handler(http.HandlerFunc(handlerSearch))).Methods("GET")
	r.Handle("/searchall", jwtMiddleware.Handler(http.HandlerFunc(handlerSearchAll))).Methods("GET")
	r.Handle("/searchmy", jwtMiddleware.Handler(http.HandlerFunc(handlerSearchMy))).Methods("GET")
	r.Handle("/login", http.HandlerFunc(loginHandler)).Methods("POST")
	r.Handle("/signup", http.HandlerFunc(signupHandler)).Methods("POST")
	r.Handle("/addfriend", jwtMiddleware.Handler(http.HandlerFunc(addFriendHandler))).Methods("POST")
	r.Handle("/deletefriend", jwtMiddleware.Handler(http.HandlerFunc(deleteFriendHandler))).Methods("POST")
	r.Handle("/getfriends", jwtMiddleware.Handler(http.HandlerFunc(getFriendsHandler))).Methods("GET")
	r.Handle("/deletepost", jwtMiddleware.Handler(http.HandlerFunc(deletePostHandler))).Methods("DELETE")

	http.Handle("/", corsMiddleware(r))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
