GOOS=linux GOARCH=amd64 go build -o bin/app-linux
GOOS=darwin GOARCH=amd64 go build -o bin/app-macos
GOOS=windows GOARCH=amd64 go build -o bin/app-windows.exe

cp index.html bin/
touch bin/database.db
sqlite3 bin/database.db < create_tables.sql

zip -r Release.zip bin/