OBJ:=guid

all: ${OBJ}

${OBJ}: *.go
	go build

start: ${OBJ}
	./${OBJ} start -idlen=6 

readme: ${OBJ}
	@cat intro.md > README.md
	@echo '```' >> README.md
	@./${OBJ} help 2>> README.md

bin:
	GOOS="windows" GOARCH="amd64" CGO_ENABLED="1" CC="x86_64-w64-mingw32-gcc" go build
	mv guid.exe guid.win64.exe
	env GOOS="linux" GOARCH="amd64" go build
	mv guid guid.linux64
	env GOOS="darwin" GOARCH="amd64" go build
	mv guid guid.darwin64