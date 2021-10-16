clean:
	echo "Removing dist/ dir"
	rm -rf dist

build:
	mkdir dist
	echo "Building for Linux/macOS/Windows"
	mkdir dist/shortcut-linux-amd64 dist/shortcut-windows-amd64 dist/shortcut-darwin-amd64 dist/shortcut-darwin-arm64
	cd src && \
	GOOS=linux GOARCH=amd64 go build -o ../dist/shortcut-linux-amd64/shortcut-linux-amd64 . && \
	GOOS=windows GOARCH=amd64 go build -o ../dist/shortcut-windows-amd64/shortcut-windows-amd64.exe . && \
	GOOS=darwin GOARCH=amd64 go build -o ../dist/shortcut-darwin-amd64/shortcut-darwin-amd64 . && \
	GOOS=darwin GOARCH=arm64 go build -o ../dist/shortcut-darwin-arm64/shortcut-darwin-arm64 .
	cp -r src/public dist/shortcut-linux-amd64
	cp -r src/public dist/shortcut-windows-amd64
	cp -r src/public dist/shortcut-darwin-amd64
	cp -r src/public dist/shortcut-darwin-arm64
	cp -r src/config.json dist/shortcut-linux-amd64
	cp -r src/config.json dist/shortcut-windows-amd64
	cp -r src/config.json dist/shortcut-darwin-amd64
	cp -r src/config.json dist/shortcut-darwin-arm64
	zip -r dist/shortcut-linux-amd64.zip dist/shortcut-linux-amd64 
	zip -r dist/shortcut-windows-amd64.zip dist/shortcut-windows-amd64
	zip -r dist/shortcut-darwin-amd64.zip dist/shortcut-darwin-amd64
	zip -r dist/shortcut-darwin-arm64.zip dist/shortcut-darwin-arm64

run:
	cd src && go run .
