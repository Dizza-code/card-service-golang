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
✅ Card linking complete. Currently working on structuring card data responses and persisting them to MongoDB. 
✅ Completed Card Linking and Card Activate
✅ Completed Events Handling(Webhook Flow) and Card authorization
✅ Completed Get Card and Get Card by ID functions

## Getting Started
Clone and run:

```bash
go run cmd/server/main.go
