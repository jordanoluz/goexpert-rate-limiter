# Go Expert Rate Limiter

This is a simple tool designed to limit the number of requests made using an api token or ip address, ensuring fair resource usage and preventing abuse. It is token-prioritized, meaning token thresholds override ip thresholds if both are provided.

## How to Run the Project Locally

### 1. Clone the repository

Clone this repository to your local machine:

```
git clone https://github.com/jordanoluz/goexpert-rate-limiter.git
```

### 2. Navigate to the project directory

Change into the project directory:

```
cd goexpert-rate-limiter
```

### 3. Configure Environment Variables

Set up your configuration by updating the **.env** file located in the root folder.

### 4. Run Docker Compose

To run the project using Docker, use Docker Compose to build and start the containers:

```
docker compose up -d --build
```

This will build the Docker images and start the application containers.

### 5. Test the Rate Limiter

Once the application is running, test the rate limiter making concurrent requests:

#### 5.1 Request with an **API_KEY** in [http://localhost:8080/](http://localhost:8080/)

Expected behavior:

- If the token has not exceeded its limit, the response will be 200 OK.
- If the token has exceeded its limit, the response will be 429 Too Many Requests.

#### 5.2 Request without an **API_KEY** (ip based limiting) in [http://localhost:8080/](http://localhost:8080/)

Expected behavior:

- If the ip address has not exceeded its limit, the response will be 200 OK.
- If the ip address has exceeded its limit, the response will be 429 Too Many Requests.

## How It Works

### 1. Token-First Priority

If an api token is provided in the API_KEY header, the rate limiter enforces the token's threshold. The ip threshold is ignored in this case.

### 2. IP Fallback

If no token is provided, the rate limiter enforces rate limits based on the client’s ip address.

### 3. Blocking

Tokens or ips that exceed their thresholds are automatically blocked for a configurable duration. Blocked tokens or ips are denied further requests until the block period expires.

### 4. Persistence

The rate limiter uses redis to store request counts and block statuses. This ensures durability and scalability across distributed systems.