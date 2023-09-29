# Task board API

Go API for [task-board-webapp](https://github.com/LeonardJouve/task-board-webapp)

## Usage
`cp .env.example .env`

Fill `.env`


`go run main.go` or using [air](github.com/cosmtrek/air) for hot reloading (`go install github.com/cosmtrek/air@latest`) `air`

## TODO
- check if user is allowed to join a board channel
- send hookMessages in board channels
- send join / leave messages in board channels
- store token available since in db
- REDIS persistance
- TLS
- docker image