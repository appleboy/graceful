module example03

go 1.21

toolchain go1.24.3

require (
	github.com/appleboy/graceful v0.0.2-0.20220102161147-760cecbcf493
	github.com/rs/zerolog v1.33.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

replace github.com/appleboy/graceful => ../../
