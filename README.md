# Task board API

Go API for [task-board-webapp](https://github.com/LeonardJouve/task-board-webapp)

## Usage
`cp .env.example .env`

Fill `.env`


`go run main.go` or using [air](github.com/cosmtrek/air) for hot reloading (`go install github.com/cosmtrek/air@latest`) `air`

## TODO
- sort is not working with multiple boards
- intl error messages
- websocket / webserver Allowed Origins
- Cookies configuration
- REDIS persistance
- docker image