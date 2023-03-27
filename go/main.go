package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	Name  string  `json:"name"`
	Image *string `json:"image"`
}

type Users map[string]User

// type Post struct {
// 	Title   string   `json:"title"`
// 	Content string   `json:"content"`
// 	Likes   []string `json:"likes"`
// }

// type UserPosts struct {
// 	Posts map[string]Post `json:"posts"`
// }

// type PostsDatabase map[string]UserPosts

type CreatePostInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Post struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email"`
	User      User      `json:"user"`
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

var (
	db *sql.DB
)

func main() {
	// PostgreSQLへの接続情報
	const (
		host     = "db"
		port     = 5432
		user     = "user"
		password = "password"
		dbname   = "test"
	)

	// PostgreSQLに接続
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("接続完了")
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	if err := db.Ping(); err != nil {
		fmt.Println(err.Error())
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/go/", http.HandlerFunc(hello))
	mux.Handle("/go/yakisyamo", http.HandlerFunc(returnYakisyamo))
	mux.Handle("/go/long_polling", corsMiddleware(http.HandlerFunc(longPolling)))
	//mux.Handle("/go/send", corsMiddleware(http.HandlerFunc(send)))
	mux.Handle("/go/posts/all", corsMiddleware(http.HandlerFunc(getPosts)))
	mux.Handle("/go/post/", corsMiddleware(http.HandlerFunc(getPostsByEmail)))

	apiPort := "8000"
	log.Printf("Listening on port %s", apiPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", apiPort), mux); err != nil {
		log.Panicln("Serve Error:", err)
	}
}

var msgCh = make(chan string)

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

func returnYakisyamo(w http.ResponseWriter, r *http.Request) {
	data := "{\"image_url\" : \"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAD6AAABMaCAYAAAD1UCWrAAAACXBIWXMAACTpAAAk6QFQJOf4AAAfGGlUWHRYTUw6Y29tLmFkb2JlLnhtcAAAAAAAPD94cGFja2V0IGJlZ2luPSLvu78iIGlkPSJXNU0wTXBDZWhpSHpyZVN6TlRjemtjOWQiPz4gPHg6eG1wbWV0YSB4bWxuczp4PSJhZG9iZTpuczptZXRhLyIgeDp4bXB0az0iQWRvYmUgWE1QIENvcmUgOS4wLWMwMDEgNzkuMTRlY2I0MmYyYywgMjAyMy8wMS8xMy0xMjoyNTo0NCAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczphdXg9Imh0dHA6Ly9ucy5hZG9iZS5jb20vZXhpZi8xLjAvYXV4LyIgeG1sbnM6ZXhpZkVYPSJodHRwOi8vY2lwYS5qcC9leGlmLzEuMC8iIHhtbG5zOnBob3Rvc2hvcD0iaHR0cDovL25zLmFkb2JlLmNvbS9waG90b3Nob3AvMS4wLyIgeG1sbnM6eG1wTU09Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9tbS8iIHhtbG5zOnN0RXZ0PSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VFdmVudCMiIHhtbG5zOnN0UmVmPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VSZWYjIiB4bWxuczpkYz0iaHR0cDovL3B1cmwub3JnL2RjL2VsZW1lb\"}"
	var url interface{}
	if err := json.Unmarshal([]byte(data), &url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	jsonString, err := json.Marshal(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

// func send(w http.ResponseWriter, r *http.Request) {
// 	// msgCh <- r.URL.Query().Get("sender")
// 	createPost(w, r)
// }

func longPolling(w http.ResponseWriter, r *http.Request) {
	// msgChへ値が送信されるまで処理をブロック
	msg := <-msgCh
	w.Write([]byte(msg))
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	// データを取得する例
	rows, err := db.Query("select posts.id, posts.title, posts.content, posts.created_at, posts.email, users.name, users.image FROM posts inner join users on posts.email = users.email order by created_at desc")
	fmt.Println(rows)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	fmt.Println(rows)
	var posts []Post
	for rows.Next() {
		var p Post
		var u User
		var createdAt string
		err = rows.Scan(&p.Id, &p.Title, &p.Content, &createdAt, &p.Email, &u.Name, &u.Image)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.User = u
		t, _ := time.Parse("2006-01-02 15:04:05", createdAt)
		p.CreatedAt = t
		posts = append(posts, p)
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
	rows, err := db.Query("select posts.id, posts.title, posts.content, posts.created_at, posts.email, users.name, users.image FROM posts inner join users on posts.email = users.email where posts.email = :email order by created_at desc;", sql.Named("email", email))
	fmt.Println(rows)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var p Post
		var u User
		var createdAt string
		err = rows.Scan(&p.Id, &p.Title, &p.Content, &createdAt, &p.Email, &u.Name, &u.Image)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.User = u
		t, _ := time.Parse("2006-01-02 15:04:05", createdAt)
		p.CreatedAt = t
		posts = append(posts, p)
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

// func createPost(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("start createPost")
// 	fmt.Println(r.Body)
// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		fmt.Println("Error reading request body:", err)
// 		return
// 	}
// 	defer r.Body.Close()

// 	fmt.Println("Request body:", string(body))
// 	decoder := json.NewDecoder(bytes.NewReader(body))
// 	fmt.Println(decoder)
// 	var p CreatePostInput
// 	if err := decoder.Decode(&p); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		fmt.Println("132", err.Error())
// 		return
// 	}

// 	byteArray, err := ioutil.ReadFile("post.json")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		fmt.Println(err.Error())
// 		return
// 	}
// 	var posts PostsDatabase
// 	if err := json.Unmarshal(byteArray, &posts); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		fmt.Println("145", err.Error())
// 		return
// 	}

// 	byteUserArray, err := ioutil.ReadFile("user.json")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		fmt.Println("152", err.Error())
// 		return
// 	}
// 	var users Users
// 	if err := json.Unmarshal(byteUserArray, &users); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		fmt.Println("158", err.Error())
// 		return
// 	}

// 	new := Post{
// 		Title:   p.Title,
// 		Content: p.Content,
// 		Likes:   []string{},
// 	}
// 	newPostIndex := strconv.Itoa(len(posts["kizuku@mail.com"].Posts) + 1)
// 	posts["kizuku@mail.com"].Posts[newPostIndex] = new
// 	updatedJSON, err := json.Marshal(posts)
// 	if err != nil {
// 		fmt.Println("171", err.Error())
// 		return
// 	}

// 	err = ioutil.WriteFile("post.json", updatedJSON, 0644)
// 	if err != nil {
// 		fmt.Println("177", err.Error())
// 		return
// 	}
// 	msgCh <- fmt.Sprintf("POST kizuku@mail.com %s %s", users["kizuku@mail.com"].Name, users["kizuku@mail.com"].Image)
// }
