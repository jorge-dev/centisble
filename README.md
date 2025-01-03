# Project Centsible

Centsible is a Budget Tracking API designed to help users manage their finances by tracking income, expenses, and budgets. The API provides endpoints for user authentication, financial record management, and budget tracking, with features for generating financial summaries and insights.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

- Go 1.16+
- Docker
- Docker Compose

### Installation

1. **Clone the repository:**

    ```bash
    git clone https://github.com/yourusername/centsible.git
    cd centsible
    ```

2. **Initialize the Go module:**

    ```bash
    go mod init
    go mod tidy
    ```

3. **Set up environment variables:**

    Create a `.env` file in the root directory and add the following variables:

    ```env
    APP_ENV=development
    PORT=8080
    CENTSIBLE_DB_HOST=psql_bp
    CENTSIBLE_DB_PORT=5432
    CENTSIBLE_DB_DATABASE=centsible
    CENTSIBLE_DB_USERNAME=yourusername
    CENTSIBLE_DB_PASSWORD=yourpassword
    CENTSIBLE_DB_SCHEMA=public
    ```

4. **Build and run the application using Docker Compose:**

    ```bash
    docker-compose up --build
    ```

## MakeFile

The Makefile provides various commands to build, run, test, and manage the application. Below are the available targets:

- **Build the application:**

    ```bash
    make build
    ```

- **Run the application:**

    ```bash
    make run
    ```

- **Run the application in a Docker container:**

    ```bash
    make d-run
    ```

- **Stop the application running in a Docker container:**

    ```bash
    make d-down
    ```

- **Build the application in a Docker container:**

    ```bash
    make d-build
    ```

- **Clean the application running in a Docker container:**

    ```bash
    make d-clean
    ```

- **Run the tests:**

    ```bash
    make test
    ```

- **Run the integration tests:**

    ```bash
    make itest
    ```

- **Clean up binary from the last build:**

    ```bash
    make clean
    ```

- **Live reload the application:**

    ```bash
    make watch
    ```

## API Documentation

The API is documented using OpenAPI 3.1.0. You can view the API documentation by running the application and navigating to `/swagger` endpoint.

### OpenAPI Specification

The OpenAPI specification file is located at `openApi.yaml`. It defines the endpoints, request/response schemas, and security schemes for the API.

### Example Endpoints

- **Register a new user:**

    ```http
    POST /register
    ```

- **Authenticate user and return JWT token:**

    ```http
    POST /login
    ```

- **Add a new income record:**

    ```http
    POST /income
    ```

- **Get all income records:**

    ```http
    GET /income
    ```

- **Add a new expense record:**

    ```http
    POST /expenses
    ```

- **Get all expense records:**

    ```http
    GET /expenses
    ```

- **Create a new budget:**

    ```http
    POST /budgets
    ```

- **Get all budgets:**

    ```http
    GET /budgets
    ```

- **Get a monthly financial summary:**

    ```http
    GET /summary/monthly
    ```

- **Get a yearly financial summary:**

    ```http
    GET /summary/yearly
    ```

For more details, refer to the `openApi.yaml` file.

## Development Plan

For a detailed development plan, refer to the [Specs.md](./Specs.md) file.

## Development Progress Checklist

### Phase 1: Project Setup and Core Infrastructure

- [X] Project Initialization
  - [X] Set up Go module
  - [X] Create project structure
  - [X] Initialize Git
- [X] Docker Environment
  - [X] Create Dockerfile
  - [X] Create docker-compose.yml
  - [X] Test containerized setup
- [X] Database Setup
  - [X] PostgreSQL setup
  - [X] Create database schema
  - [X] Implement migrations
  - [X] Create tables:
    - [X] users
    - [X] income
    - [X] expenses
    - [X] categories
    - [X] budgets
- [X] User Authentication
  - [X] Implement user registration
  - [X] Implement user login
  - [X] Setup JWT authentication
  - [X] Create auth middleware

### Phase 2: Income, Expense, and Budget Management

- [X] Income Management
  - [X] CRUD endpoints for income
  - [X] User-income association
- [X] Expense Management
  - [X] CRUD endpoints for expenses
  - [X] Expense categorization
- [X] Budget Management
  - [X] CRUD endpoints for budgets
  - [X] Budget limits and tracking
  - [X] Budget alerts
  - [X] Recurring/one-time budgets
- [X] Input Validation
  - [X] Implement request validation
  - [x] Data sanitation

### Phase 3: Budget Tracking and Summaries

- [X] Budget Tracking
  - [X] Compare expenses vs limits
  - [X] Budget adjustment endpoints
- [X] Financial Summaries
  - [X] Monthly summary endpoint
  - [X] Yearly summary endpoint
  - [X] Currency handling
- [X] Reports and Insights
  - [X] Spending trends
  - [X] Budget progress tracking

### Phase 4: Documentation, Testing, and Security

- [X] API Documentation
  - [X] Swagger UI integration
  - [X] OpenAPI specifications
- [X] Testing
  - [X] Unit tests
  - [X] Integration tests
  - [X] Mock implementations
- [X] Security
  - [X] Rate limiting
  - [X] Security headers
  - [X] Input validation security

### Phase 5: Deployment and Monitoring

- [ ] CI/CD Pipeline
  - [X] GitHub Actions setup
  - [ ] Docker registry integration
- [ ] Cloud Deployment
  - [ ] Cloud provider setup
  - [ ] Database deployment
  - [ ] Container orchestration
- [ ] Monitoring
  - [ ] Logging setup
  - [ ] Metrics collection
  - [ ] Alerting system

### Phase 6: Future Enhancements

- [ ] Recurring Transactions
  - [ ] Automated recurring expenses
  - [ ] Automated recurring income
- [ ] Advanced Features
  - [ ] ML-based insights
  - [ ] Multi-currency support
  - [ ] Banking API integration

## Contributing

Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## License

This project is licensed under the MIT License - see the [LICENSE.md](./LICENSE.md) file for details.
