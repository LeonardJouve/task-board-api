# Task board API

Go API for [task-board-webapp](https://github.com/LeonardJouve/task-board-webapp)

## Usage
`cp .env.example .env`

Fill `.env`


`go run main.go` or using [air](github.com/cosmtrek/air) for hot reloading (`go install github.com/cosmtrek/air@latest`) `air`

## TODO
- refresh should not require access token cookie
- intl error messages
- Cookies configuration
- handle register duplicate email
- REDIS persistance
- docker image