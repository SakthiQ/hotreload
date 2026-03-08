.PHONY: demo clean testserver

demo:
	@echo "Building testserver..."
	cd testserver && go mod tidy && cd ..
	@echo "Testing hotreload phase 1..."
	# (legacy comments removed)
	@echo "Building hotreload binary..."
	cd hotreload && go build -o ../hotreload.exe . && cd ..
	@echo "Running demo (hotreload watching testserver)..."
	@./hotreload.exe --root ./testserver \
		--build "go build -o server.exe ./testserver" \
		--exec "./server.exe"
	@echo "✅ Demo complete!"

testserver:
	cd testserver && go mod tidy && go build -o server .

clean:
	rm -f server testserver/server
	rm -f server testserver/server hotreload.exe
