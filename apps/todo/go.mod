module github.com/VauntDev/blogSnippets/apps/todo

go 1.21.1

replace github.com/VauntDev/glhf => ../../../glhf
replace github.com/VauntDev/tqla => ../../../tqla

require (
	github.com/VauntDev/glhf v0.0.3
	github.com/VauntDev/tqla v0.0.0-20231016155458-a505d92ec874
	github.com/google/uuid v1.3.1
	github.com/gorilla/mux v1.8.0
	github.com/mattn/go-sqlite3 v1.14.17
	golang.org/x/sync v0.4.0
	golang.org/x/sys v0.13.0
	google.golang.org/protobuf v1.31.0
)
