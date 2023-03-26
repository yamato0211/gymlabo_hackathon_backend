package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type User struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type Users map[string]User

type Post struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Likes   []string `json:"likes"`
}

type UserPosts struct {
	Posts map[string]Post `json:"posts"`
}

type PostsDatabase map[string]UserPosts

type CreatePostInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func main() {
	http.HandleFunc("/long_polling", longPolling)
	http.HandleFunc("/send", send)
	http.HandleFunc("/posts/all", getPosts)
	http.HandleFunc("/post/", getPostsByEmail)

	port := "8000"
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		log.Panicln("Serve Error:", err)
	}
}

var msgCh = make(chan string)

func send(w http.ResponseWriter, r *http.Request) {
	// msgCh <- r.URL.Query().Get("sender")
	createPost(w, r)
	w.WriteHeader(http.StatusOK)
}

func longPolling(w http.ResponseWriter, r *http.Request) {
	// msgChへ値が送信されるまで処理をブロック
	msg := <-msgCh
	w.Write([]byte(msg))
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	byteArray, err := ioutil.ReadFile("post.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	var posts PostsDatabase
	if err := json.Unmarshal(byteArray, &posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	jsonString, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

func getPostsByEmail(w http.ResponseWriter, r *http.Request) {
	sub := strings.TrimPrefix(r.URL.Path, "/post")
	_, email := filepath.Split(sub)
	byteArray, err := ioutil.ReadFile("post.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	var posts PostsDatabase
	if err := json.Unmarshal(byteArray, &posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	post := posts[email]
	jsonString, err := json.Marshal(post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var p CreatePostInput
	if err := decoder.Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	byteArray, err := ioutil.ReadFile("post.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	var posts PostsDatabase
	if err := json.Unmarshal(byteArray, &posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	byteUserArray, err := ioutil.ReadFile("user.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	var users Users
	if err := json.Unmarshal(byteUserArray, &users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	new := Post{
		Title:   p.Title,
		Content: p.Content,
		Likes:   []string{},
	}
	newPostIndex := strconv.Itoa(len(posts["kizuku@mail.com"].Posts) + 1)
	posts["kizuku@mail.com"].Posts[newPostIndex] = new
	updatedJSON, err := json.Marshal(posts)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = ioutil.WriteFile("post.json", updatedJSON, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	msgCh <- fmt.Sprintf("POST kizuku@mail.com %s %s", users["kizuku@mail.com"].Name, users["kizuku@mail.com"].Image)
}
