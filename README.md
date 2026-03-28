# QRCode_Gen

A practice Go project for generating and managing QR codes with redirect support.

## Features

- Create a QR code from a URL
- Resolve a token back to the original URL
- Generate a QR code image for a token
- Update and delete QR code records
- Redirect by short token
- SQLite-backed storage

## Run

```bash
go run .
```

Default environment variables:

- `PORT=8080`
- `BASE_URL=http://localhost:8080`

## Main Endpoints

- `POST /v1/qr_code`
- `GET /v1/qr_code/{qr_token}`
- `GET /v1/qr_code_image/{qr_token}`
- `PUT /v1/qr_code/{qr_token}`
- `DELETE /v1/qr_code/{qr_token}`
- `GET /r/{qr_token}`
- `GET /health`
