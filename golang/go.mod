module github.com/ideamans/libnextimage/golang

go 1.24.7

replace github.com/ideamans/libnextimage/golang/cwebp => ./cwebp
replace github.com/ideamans/libnextimage/golang/dwebp => ./dwebp
replace github.com/ideamans/libnextimage/golang/avifenc => ./avifenc
replace github.com/ideamans/libnextimage/golang/avifdec => ./avifdec
replace github.com/ideamans/libnextimage/golang/gif2webp => ./gif2webp
replace github.com/ideamans/libnextimage/golang/webp2gif => ./webp2gif

require (
	github.com/ideamans/libnextimage/golang/cwebp v0.0.0
	github.com/ideamans/libnextimage/golang/dwebp v0.0.0
	github.com/ideamans/libnextimage/golang/avifenc v0.0.0
)
