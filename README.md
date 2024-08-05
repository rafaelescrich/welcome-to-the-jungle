# Welcome to the Jungle

This project is a backend service for managing client data, inspired by the name of the company and the Guns N' Roses song "Welcome to the Jungle". It uses Golang with the Gin framework and Swagger for API documentation. The service connects to a PostgreSQL database and provides endpoints to manage client information.

## Features

- Load client data from a large CSV file into PostgreSQL
- API endpoints to retrieve client information by UID, filter clients by age range, and search clients by name
- Swagger documentation for the API

## Setup

### Prerequisites

- Go 1.22.2
- Docker
- Docker Compose
- PostgreSQL
- [godotenv](https://github.com/joho/godotenv)
- [gin-gonic](https://github.com/gin-gonic/gin)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/rafaelescrich/welcome-to-the-jungle.git
cd welcome-to-the-jungle
```

### Running the Application 
 
1. **Build and run the Docker containers** :

```bash
docker-compose up --build
```
 
2. **Access the API** :
Open your browser and navigate to `http://localhost:8080/swagger/index.html` to see the Swagger UI for your API.

## API Endpoints 
 
- `GET /info?uid={uid}`: Get client info by UID
 
- `GET /info/by-age?start={start}&end={end}`: Get clients by age range
 
- `GET /search?name={name}`: Search clients by name

## License 

This project is licensed under the MIT License.