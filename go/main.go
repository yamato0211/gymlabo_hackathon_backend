package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type CreatePostInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Like struct {
	Email string `json:"email"`
}

type Post struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email"`
	User      User      `json:"user"`
	Likes     []Like    `json:"likes"`
}

type LikeRequest struct {
	PostId int    `json:"post_id"`
	Email  string `json:"email"`
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
	mux.Handle("/go/", corsMiddleware(http.HandlerFunc(hello)))
	mux.Handle("/go/yakisyamo", corsMiddleware(http.HandlerFunc(returnYakisyamo)))
	mux.Handle("/go/long_polling", corsMiddleware(http.HandlerFunc(longPolling)))
	mux.Handle("/go/send/post", corsMiddleware(http.HandlerFunc(sendPost)))
	mux.Handle("/go/posts/all", corsMiddleware(http.HandlerFunc(getPosts)))
	mux.Handle("/go/post/", corsMiddleware(http.HandlerFunc(getPostsByEmail)))
	mux.Handle("/go/like/create", corsMiddleware(http.HandlerFunc(createLike)))
	mux.Handle("/go/like/delete", corsMiddleware(http.HandlerFunc(deleteLike)))

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

func sendPost(w http.ResponseWriter, r *http.Request) {
	createPost(w, r)
}

func longPolling(w http.ResponseWriter, r *http.Request) {
	msg := <-msgCh
	w.Write([]byte(msg))
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT posts.id, posts.title, posts.content, posts.created_at, posts.email, users.name, users.image, COALESCE(json_agg(json_build_object('email', likes.email)) FILTER (WHERE likes.email IS NOT NULL), '[]'::json) AS likes FROM posts JOIN users ON users.email = posts.email LEFT JOIN likes ON likes.post_id = posts.id GROUP BY posts.id, users.id ORDER BY posts.created_at DESC;")
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
		var likes []Like
		var likesJSON []byte
		var createdAt string
		err = rows.Scan(&p.Id, &p.Title, &p.Content, &createdAt, &p.Email, &u.Name, &u.Image, &likesJSON)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(likesJSON, &likes)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.User = u
		p.Likes = likes
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

	query := fmt.Sprintf("SELECT posts.id, posts.title, posts.content, posts.created_at, posts.email, users.name, users.image, COALESCE(json_agg(json_build_object('email', likes.email)) FILTER (WHERE likes.email IS NOT NULL), '[]'::json) AS likes FROM posts JOIN users ON users.email = posts.email LEFT JOIN likes ON likes.post_id = posts.id where posts.email = '%s' GROUP BY posts.id, users.id ORDER BY posts.created_at DESC;", email)
	rows, err := db.Query(query)
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
		var likes []Like
		var likesJSON []byte
		err = rows.Scan(&p.Id, &p.Title, &p.Content, &createdAt, &p.Email, &u.Name, &u.Image, &likesJSON)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(likesJSON, &likes)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.User = u
		p.Likes = likes
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
		fmt.Println(err.Error())
		return
	}

	var post Post
	now := time.Now().In(time.UTC)
	query := fmt.Sprintf("insert into posts (title, content, created_at, email) values('%s','%s','%s','%s')", p.Title, p.Content, now.Format("2006-01-02 15:04:05"), "kizuku@mail.com")
	result, err := db.Exec(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query = fmt.Sprintf("SELECT * FROM posts WHERE id = '%d'", id)
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&post.Id, &post.Title, &post.Content, &post.CreatedAt, &post.Email)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(post)
	}
	var name string
	var image *string
	userQuery := fmt.Sprintf("select name,image from users where email = '%s'", "kizuku@mail.com")
	userRow, err := db.Query(userQuery)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for userRow.Next() {
		err = userRow.Scan(&name, &image)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	msgCh <- fmt.Sprintf("POST kizuku@mail.com %s %s", name, image)

	jsonString, err := json.Marshal(post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonString)
}

func createLike(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "this method not allowed", http.StatusBadRequest)
		return
	}
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
	var l LikeRequest
	if err := decoder.Decode(&l); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}

	query := fmt.Sprintf("select post_id, email from likes where post_id = '%d' and email = '%s'", l.PostId, l.Email)
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println(err.Error())
		fmt.Printf("0000")
		return
	}
	fmt.Println("rows", rows)
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println(err.Error())
		fmt.Printf("0001")
		return
	}
	if len(columns) > 0 {
		http.Error(w, "this user is already liked", http.StatusBadRequest)
		fmt.Println("column more than zero")
		return
	}
	fmt.Println("after rowsNext")

	query = fmt.Sprintf("insert into likes (post_id, email) values('%d', '%s')", l.PostId, l.Email)
	fmt.Println(query)
	result, err := db.Exec(query)
	fmt.Println("result", result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}

	if affected, err := result.RowsAffected(); err == nil && affected > 0 {
		query = fmt.Sprintf("select post_id, email from likes where post_id = '%d' and email = '%s'", l.PostId, l.Email)
		rows, err := db.Query(query)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var post_id int
		var email string
		var name string
		var image *string

		for rows.Next() {
			err = rows.Scan(&post_id, &email)
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		}

		query = fmt.Sprintf("select name,image from users where email = '%s'", l.Email)
		rows, err = db.Query(query)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for rows.Next() {
			err = rows.Scan(&name, &image)
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		msgCh <- fmt.Sprintf("LIKE %s %s %s %d", l.Email, name, image, l.PostId)
		w.Write([]byte("ok"))
	} else {
		fmt.Println("insert処理が正常に完了しませんでした")
		http.Error(w, "insert処理が正常に完了しませんでした", http.StatusInternalServerError)
		return
	}
}

func deleteLike(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "this method not allowed", http.StatusBadRequest)
		return
	}
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
	var l LikeRequest
	if err := decoder.Decode(&l); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}

	query := fmt.Sprintf("select post_id, email from likes where post_id = '%d' and email = '%s'", l.PostId, l.Email)
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println(err.Error())
		fmt.Printf("0000")
		return
	}
	fmt.Println("rows", rows)
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println(err.Error())
		fmt.Printf("0001")
		return
	}
	if len(columns) == 0 {
		http.Error(w, "this user is not liked", http.StatusBadRequest)
		fmt.Println("column zero")
		return
	}
	fmt.Println("after rowsNext")

	query = fmt.Sprintf("delete from likes where post_id = '%d' and email = '%s'", l.PostId, l.Email)
	fmt.Println(query)
	result, err := db.Exec(query)
	fmt.Println("result", result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}

	if affected, err := result.RowsAffected(); err == nil && affected > 0 {
		// msgCh <- fmt.Sprintf("LIKE %s %s %s %d", l.Email, name, image, l.PostId)
		w.Write([]byte("ok"))
	} else {
		fmt.Println("insert処理が正常に完了しませんでした")
		http.Error(w, "insert処理が正常に完了しませんでした", http.StatusInternalServerError)
		return
	}
}
