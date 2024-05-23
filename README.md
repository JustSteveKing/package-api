# Package API

A Go service to fetch and aggregate package information from packagist API, by vendor.

## Usage

```bash
go run main.go
```

Visit: `http://localhost:3000/?vendor=juststeveking`

It will use an ETag header so the second request will be quicker.
