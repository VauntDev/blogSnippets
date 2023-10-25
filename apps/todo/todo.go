package main

import (
	"database/sql"
	"encoding/base64"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/VauntDev/blogSnippets/apps/todo/pb"
	"github.com/VauntDev/tqla"
)

type TodoService struct {
	db *sql.DB
}

type Filter struct {
	Limit  uint8
	After  string
	Before string
}

type Page struct {
	Forward     bool
	CreatedDttm time.Time
	Id          string
	Limit       uint8
}

const (
	maxLimit     = 100
	defaultLimit = 1
)

// tqla can be expanded through custom functions.
var exampleFuncs = template.FuncMap{
	"sub": func(a int, b int) int {
		return a - b
	},
	// parseCursor creates a Pagination object from a base64 encoded cursor.
	// Uses after if before is empty other wise uses before. If both after and before are set
	// then the Pagination object returned in nil.
	"parseCursor": func(after string, before string, limit uint8) *Page {
		cursor := ""
		forward := false
		createdDttm := ""
		id := ""

		if len(after) > 0 && len(before) == 0 {
			forward = true
			cursor = after

		} else if len(before) > 0 && len(after) == 0 {
			forward = false
			cursor = before
		}

		if len(cursor) > 0 {
			data, err := base64.StdEncoding.DecodeString(cursor)
			if err != nil {
				return nil
			}
			values := strings.Split(string(data), "/")
			if len(values) == 2 {
				createdDttm = values[0]
				id = values[1]
			}
		}

		if limit > maxLimit {
			limit = maxLimit
		}

		t, _ := time.Parse("2006-01-02T15:04:05Z", createdDttm)
		return &Page{
			Forward:     forward,
			CreatedDttm: t,
			Id:          id,
			Limit:       limit,
		}
	},
}

func (ts *TodoService) List(filter *Filter) ([]*pb.Todo, error) {

	t, err := tqla.New(tqla.WithPlaceHolder(tqla.Dollar), tqla.WithFuncMap(exampleFuncs))
	if err != nil {
		return nil, err
	}

	selectStmt, selectArgs, err := t.Compile(`
	SELECT id, title, description, completed, created_at 
	FROM todos 
	{{ $page := ( parseCursor .After .Before .Limit ) }}
	{{ if $page.Forward }}
		WHERE (created_at,id) > ({{ $page.CreatedDttm }},{{ $page.Id }})
	{{ else }}
		WHERE (created_at,id) < ({{ $page.CreatedDttm }}, {{ $page.Id }})
	{{ end }}
	ORDER BY created_at
	LIMIT {{ $page.Limit }}
	`, filter)
	if err != nil {
		return nil, err
	}

	rows, err := ts.db.Query(selectStmt, selectArgs...)
	if err != nil {
		return nil, err
	}

	todos := make([]*pb.Todo, 0)
	for rows.Next() {
		todo := &pb.Todo{}
		if err := rows.Scan(&todo.Id, &todo.Title, &todo.Description, &todo.Complete, &todo.CreatedAt); err != nil {
			return nil, err
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

func (ts *TodoService) Add(todos []pb.Todo) error {

	// Init TQLA with the DB placeholder to use and additional functions that can
	// be called by the template.
	t, err := tqla.New(tqla.WithPlaceHolder(tqla.Dollar), tqla.WithFuncMap(exampleFuncs))
	if err != nil {
		return err
	}

	insertStmt, insertArgs, err := t.Compile(`
	{{ $length := sub ( len . ) 1 }}
	INSERT INTO 'todos' ('id', 'title', 'description', 'completed', 'created_at') 
	VALUES {{ range $i, $v := . }}
    	( {{$v.Id}}, {{$v.Title}}, {{$v.Description}}, {{ $v.Complete }}, {{ $v.CreatedAt }} ){{if lt $i $length}}, {{else}}; {{end -}}
	{{end}}`, todos)
	if err != nil {
		return err
	}

	if _, err := ts.db.Exec(insertStmt, insertArgs...); err != nil {
		return err
	}

	return nil
}

func (ts *TodoService) Get(id string) (*pb.Todo, error) {

	t, err := tqla.New(tqla.WithPlaceHolder(tqla.Dollar))
	if err != nil {
		return nil, err
	}

	selectStmt, selectArgs, err := t.Compile(`select id,title, description, completed, created_at from todos where id={{ . }}`, id)
	if err != nil {
		return nil, err
	}

	todo := &pb.Todo{}
	row := ts.db.QueryRow(selectStmt, selectArgs...)
	if err := row.Scan(&todo.Id, &todo.Title, &todo.Description, &todo.Complete, &todo.CreatedAt); err != nil {
		log.Println(err)
		return nil, err
	}

	return todo, nil
}
