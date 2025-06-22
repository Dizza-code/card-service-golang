# card-service-golang

A backend system for managing customers, cards, and transaction authorization in a fintech-like environment.

Built with:
- Go (Gin framework)
- MongoDB
- Integration with an external card issuing API
- Ngrok for local webhook testing

## Features

- Customer creation
- Account setup and linking
- External API integration
- Secure webhook setup for transaction authorization

## Status
🚧 Currently implementing card linking logic and preparing to store linked card response into MongoDB.

## Getting Started
Clone and run:

```bash
go run cmd/server/main.go
