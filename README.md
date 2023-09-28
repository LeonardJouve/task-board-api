# Task board API

Go API for [task-board-webapp](https://github.com/LeonardJouve/task-board-webapp)

## Usage
`cp .env.example .env`

Fill `.env`


`go run main.go` or using [air](github.com/cosmtrek/air) for hot reloading (`go install github.com/cosmtrek/air@latest`) `air`

## TODO
- store userId in Connections map
- send messages only to board users
- websocket event on CRUD
- REDIS persistance
- store token available since in db
- docker image
- TLS