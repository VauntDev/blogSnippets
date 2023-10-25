package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/VauntDev/blogSnippets/apps/todo/pb"
	"github.com/VauntDev/glhf"
	"github.com/gorilla/mux"
)

type Handlers struct {
	service *TodoService
}

func (h *Handlers) GLHFListTodos(r *glhf.Request[glhf.EmptyBody], w *glhf.Response[[]*pb.Todo]) {

	after := r.URL().Query().Get("after")
	before := r.URL().Query().Get("before")
	limit := limit(r.URL().Query().Get("limit"))

	filter := &Filter{Limit: limit, After: after, Before: before}

	todos, err := h.service.List(filter)
	if err != nil {
		log.Println(err)
		w.SetStatus(http.StatusInternalServerError)
		return
	}

	w.SetBody(&todos)
	w.SetStatus(http.StatusOK)
}

func (h *Handlers) GLHFCreateTodo(r *glhf.Request[[]pb.Todo], w *glhf.Response[glhf.EmptyBody]) {
	if r.Body() == nil {
		w.SetStatus(http.StatusBadRequest)
		return
	}

	if err := h.service.Add(*r.Body()); err != nil {
		log.Println(err)
		w.SetStatus(http.StatusInternalServerError)
		return
	}

	w.SetStatus(http.StatusOK)
}

func (h *Handlers) GLHFLookupTodo(r *glhf.Request[glhf.EmptyBody], w *glhf.Response[pb.Todo]) {
	p := mux.Vars(r.HTTPRequest())

	id, ok := p["id"]
	if !ok {
		w.SetStatus(http.StatusInternalServerError)
		return
	}

	todo, err := h.service.Get(id)
	if err != nil {
		w.SetStatus(http.StatusNotFound)
		return
	}

	w.SetBody(todo)
	w.SetStatus(http.StatusOK)
}

// limit creates a uint8 limit from a string. If the supplied value
// is greater than maxLimit or invalid the defaultLimit is used.
func limit(limit string) uint8 {
	limitVal, err := strconv.Atoi(limit)
	if err != nil {
		return defaultLimit
	}
	if limitVal > maxLimit {
		return uint8(maxLimit)
	} else if limitVal >= defaultLimit {
		return uint8(limitVal)
	}
	return defaultLimit
}
