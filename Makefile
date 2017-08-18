quickbundle:
	go build -o quickbundle

example: quickbundle
	./quickbundle -entry example/entry.js
