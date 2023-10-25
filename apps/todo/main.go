package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"

	"github.com/VauntDev/blogSnippets/apps/todo/pb"
	"github.com/VauntDev/glhf"

	_ "github.com/mattn/go-sqlite3"
)

const dbName = "example.db"
const todoSchema = `create table if not exists todos (
	id text primary key,
	title text not null,
	description text not null,
	completed boolean default 0,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP 
);
`

func main() {

	log.Println("connecting to db...")

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("creating todo table if it does not exist...")

	if _, err := db.Exec(todoSchema); err != nil {
		log.Fatal(err)
	}

	TodoService := &TodoService{
		db: db,
	}

	h := &Handlers{service: TodoService}

	mux := mux.NewRouter()

	mux.HandleFunc("/todos", glhf.Get(h.GLHFListTodos)).Methods("GET")
	mux.HandleFunc("/todos", glhf.Put(h.GLHFCreateTodo)).Methods("PUT")
	mux.HandleFunc("/todos/{id}", glhf.Get(h.GLHFLookupTodo))

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		log.Println("starting server")
		if err := server.ListenAndServe(); err != nil {
			return nil
		}
		return nil
	})

	g.Go(func() error {
		sigs := make(chan os.Signal, 1)
		// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
		// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
		signal.Notify(sigs, os.Interrupt, unix.SIGTERM)
		// block until we receive our signal or context is done
		select {
		case <-ctx.Done():
			log.Println("ctx done, shutting down server")
		case <-sigs:
			log.Println("caught sig, shutting down server")
		}
		// Create a deadline to wait for cleanup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// Doesn't block if no connections, but will otherwise wait
		// until the timeout deadline.
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("error in server shutdown: %w", err)
		}
		return nil
	})

	// ----- Client Code ----- //

	// wait for server to start
	time.Sleep(time.Second * 1)

	client := http.DefaultClient

	// lookupId is used to store a single todo for easy lookup post creation
	var lookupId string
	todos := make([]pb.Todo, 0)
	var created time.Time
	//create 10 example todos
	for i := 0; i < 10; i++ {
		createdAt := time.Now()
		id := uuid.NewString()
		if i == 0 {
			// store first todo for lookup later on
			lookupId = id
			created = createdAt
		}
		todos = append(todos, pb.Todo{
			Id:          id,
			Title:       fmt.Sprintf("example-%d", i),
			Description: fmt.Sprintf("awesome example-%d todo", i),
			Complete:    false,
			CreatedAt:   createdAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	b, err := json.Marshal(todos)
	if err != nil {
		log.Fatal("failed to marshal proto")
	}

	putRequest, err := http.NewRequest("PUT", "http://localhost:8080/todos", bytes.NewBuffer(b))
	if err != nil {
		log.Fatal("failed to create put request")
	}

	putRequest.Header.Add("Content-Type", "application/json") // send protobuff

	log.Println("sending put request to create todo")

	putResp, err := client.Do(putRequest)
	if err != nil {
		log.Fatal("failed to do put request", err)
	}

	if putResp.StatusCode != http.StatusOK {
		log.Fatal("put request failed with", putResp.StatusCode)
	}

	getRequest, err := http.NewRequest("GET", "http://localhost:8080/todos/"+lookupId, nil)
	if err != nil {
		log.Fatal("failed to create get request")
	}

	getRequest.Header.Add("Accept", "application/json") // get json

	log.Println("sending get request to lookup todo")
	getResp, err := client.Do(getRequest)
	if err != nil {
		log.Fatal("failed to do get request", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		log.Fatal("get failed with", getResp.StatusCode)
	}

	body, err := io.ReadAll(getResp.Body)
	if err != nil {
		log.Fatal("failed to read response body", err)
	}

	log.Println("todo:", string(body))
	log.Println("sending get request to list todos ")

	timeformat := created.Format("2006-01-02T15:04:05Z")
	log.Println(timeformat)

	cursor := timeformat + "/" + lookupId

	encodedCursor := base64.StdEncoding.EncodeToString([]byte(cursor))

	ListRequest, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/todos?limit=%d&after=%s", 3, encodedCursor), nil)
	if err != nil {
		log.Fatal("failed to create get request")
	}

	ListRequest.Header.Add("Accept", "application/json") // get json

	listResp, err := client.Do(ListRequest)
	if err != nil {
		log.Fatal("failed to do get request", err)
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != http.StatusOK {
		log.Fatal("get failed with", listResp.StatusCode)
	}

	listBody, err := io.ReadAll(listResp.Body)
	if err != nil {
		log.Fatal("failed to read response body", err)
	}

	log.Println("todos list:", string(listBody))
}
