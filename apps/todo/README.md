# Todo Sample Application

A sample todo application that showcases the how `glhf` and `tqla` can be used within
a basic rest api.

This example focuses on `glhf` and `tqla` not the best practices for building APIs in golang.

## Routes

This services exposes three api routes

- List: /todos
  - Get
  - Optional Query Parameters Limit ( uint8 ), after(base64(createdAt/todoID)), before(base64(createdAt/todoID))
- Create: /todos
  - Put
  - Body json list of TODOs
- Get: /todos/{id}
  - GET
  - URL Params Id

## Database

This applications leverages a local sqlite db for storage.

Below is the basic schema used to store Todos.

```sql
create table if not exists todos (
  id text primary key,
  title text not null,
  description text not null,
  completed boolean default 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Build and Run

`go build`

`./todo`
