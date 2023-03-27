package main

import (
	"bytes"
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

func corsMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,UPDATE,OPTIONS")
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/go/long_polling", corsMiddleware(http.HandlerFunc(longPolling)))
	mux.Handle("/go/send", corsMiddleware(http.HandlerFunc(send)))
	mux.Handle("/go/posts/all", corsMiddleware(http.HandlerFunc(getPosts)))
	mux.Handle("/go/post/", corsMiddleware(http.HandlerFunc(getPostsByEmail)))

	port := "8000"
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), mux); err != nil {
		log.Panicln("Serve Error:", err)
	}
}

var msgCh = make(chan string)

func send(w http.ResponseWriter, r *http.Request) {
	// msgCh <- r.URL.Query().Get("sender")
	createPost(w, r)
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
	fmt.Println("start createPost")
	fmt.Println(r.Body)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println("Error reading request body:", err)
		return
	}
	defer r.Body.Close()

	fmt.Println("Request body:", string(body))
	decoder := json.NewDecoder(bytes.NewReader(body))
	fmt.Println(decoder)
	var p CreatePostInput
	if err := decoder.Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println("132", err.Error())
		return
	}

	byteArray, err := ioutil.ReadFile("post.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	var posts PostsDatabase
	if err := json.Unmarshal(byteArray, &posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("145", err.Error())
		return
	}

	byteUserArray, err := ioutil.ReadFile("user.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("152", err.Error())
		return
	}
	var users Users
	if err := json.Unmarshal(byteUserArray, &users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("158", err.Error())
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
		fmt.Println("171", err.Error())
		return
	}

	err = ioutil.WriteFile("post.json", updatedJSON, 0644)
	if err != nil {
		fmt.Println("177", err.Error())
		return
	}
	msgCh <- fmt.Sprintf("POST kizuku@mail.com %s %s", users["kizuku@mail.com"].Name, users["kizuku@mail.com"].Image)
}
