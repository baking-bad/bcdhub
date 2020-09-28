windows:
	GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc CXX=i686-w64-mingw32-g++ go build -o bcd.dll -buildmode=c-shared .

linux:
	GOOS=linux CGO_ENABLED=1 go build -o bcd.so -buildmode=c-shared .
