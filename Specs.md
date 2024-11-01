# Budget Tracking API - Development Plan

## Introduction

This development plan is designed to help you build the **Budget Tracking API** systematically. It breaks the project into manageable phases, each with specific tasks and milestones, allowing you to track progress effectively.

## Phase 1: Project Setup and Core Infrastructure

**Goal**: Set up the basic infrastructure for the API, create the foundation for development, and secure the application.

1. **Project Initialization**
   - Set up a new Go module (`go mod init`).
   - Create the project directory structure (`controllers/`, `models/`, `routes/`, `middleware/`, etc.).
   - Initialize Git for version control.

2. **Docker Environment**
   - Write a `Dockerfile` to containerize the Go application.
   - Create a `docker-compose.yml` to manage PostgreSQL and API services.
   - Ensure that the app and database can run successfully using Docker Compose.

3. **Database Setup**
   - Set up PostgreSQL locally and within Docker.
   - Create the database schema with tables for `users`, `income`, `expenses`, `categories`, and `budgets`.
   - Use database migration tools like `golang-migrate` for schema changes.

   **Table Details**:

   - **users**:
     - **Description**: Stores user information.
     - **Columns**:
       - `id` (UUID, Primary Key)
       - `name` (String)
       - `email` (String, Unique)
       - `password_hash` (String)
       - `created_at` (Timestamp)
       - `deleted_at` (Timestamp, Nullable)
     - **Relationships**: One-to-Many relationship with `income`, `expenses`, `categories`, and `budgets` tables.

   - **income**:
     - **Description**: Stores user income records.
     - **Columns**:
       - `id` (UUID, Primary Key)
       - `user_id` (UUID, Foreign Key to users ON DELETE CASCADE)
       - `amount` (Decimal)
       - `currency` (String)
       - `source` (String)
       - `date` (Date)
       - `description` (String)
       - `created_at` (Timestamp)
       - `deleted_at` (Timestamp, Nullable)
     - **Relationships**: Many-to-One relationship with `users`, cascading delete.

   - **expenses**:
     - **Description**: Stores user expense records.
     - **Columns**:
       - `id` (UUID, Primary Key)
       - `user_id` (UUID, Foreign Key to users ON DELETE CASCADE)
       - `amount` (Decimal)
       - `currency` (String)
       - `category` (String)
       - `date` (Date)
       - `description` (String)
       - `created_at` (Timestamp)
       - `deleted_at` (Timestamp, Nullable)
     - **Relationships**: Many-to-One relationship with `users`, cascading delete.

   - **categories**:
     - **Description**: Stores budget categories created by users.
     - **Columns**:
       - `id` (UUID, Primary Key)
       - `user_id` (UUID, Foreign Key to users ON DELETE CASCADE)
       - `name` (String)
       - `created_at` (Timestamp)
       - `deleted_at` (Timestamp, Nullable)
     - **Relationships**: Many-to-One relationship with `users`, cascading delete. Categories can be linked to `expenses` and `budgets` for classification.

   - **budgets**:
     - **Description**: Stores user budget information.
     - **Columns**:
       - `id` (UUID, Primary Key)
       - `user_id` (UUID, Foreign Key to users ON DELETE CASCADE)
       - `amount` (Decimal)
       - `currency` (String)
       - `category` (String)
       - `type` (String: recurring or one-time)
       - `start_date` (Date)
       - `end_date` (Date)
       - `created_at` (Timestamp)
       - `deleted_at` (Timestamp, Nullable)
     - **Relationships**: Many-to-One relationship with `users`, cascading delete.

4. **User Authentication**
   - Implement user registration and login (`/register`, `/login` endpoints).
   - Encrypt passwords with `bcrypt`.
   - Implement JWT-based authentication and create middleware for securing routes.

**Milestone**: User registration and login features are implemented and tested successfully.

## Phase 2: Income, Expense, and Budget Management

**Goal**: Implement core features, including CRUD operations for income, expenses, and budget categories.

1. **Income Management**
   - Implement CRUD endpoints (`/income`, `/income/{id}`) for managing user income.
   - Connect income records to the authenticated user.

2. **Expense Management**
   - Implement CRUD endpoints (`/expenses`, `/expenses/{id}`) for managing user expenses.
   - Allow expenses to be categorized.

3. **Budget Management**
   - Implement budget creation and CRUD endpoints (`/budgets`, `/budgets/{id}`).
   - Add budget limits for specific categories and track actual spending against these limits.
   - Implement alerts when users approach or exceed their budget limits.
   - Add support for creating recurring or one-time budgets, allowing users to set budgets that recur monthly or are a one-time allocation.

4. **Basic Input Validation**
   - Add validation for incoming requests (e.g., validating that amount fields are positive numbers).

**Milestone**: All core financial management features are functional, with basic validations in place.

## Phase 3: Budget Tracking and Summaries

**Goal**: Allow users to manage budgets and access monthly and yearly summaries.

1. **Budget Tracking**
   - Track budget usage by comparing expenses with budget limits.
   - Implement endpoints for updating and adjusting budgets as needed.

2. **Monthly and Yearly Summaries**
   - Add endpoints for generating monthly (`/summary/monthly`) and yearly summaries (`/summary/yearly`).
   - Include total income, expenses, savings, and top spending categories.
   - Include currency information for accurate financial reporting.

3. **Reports and Insights**
   - Implement spending trend insights (e.g., identify top spending categories).
   - Track and display progress towards budget goals.

**Milestone**: Users can manage budgets, view financial summaries, and receive basic spending insights.

## Phase 4: Documentation, Testing, and Security Enhancements

**Goal**: Ensure the application is well-documented, thoroughly tested, and secure.

1. **API Documentation**
   - Integrate Swagger UI to provide OpenAPI documentation for all endpoints.
   - Write OpenAPI 3.1.0 specifications for endpoints, request/response bodies, and validation rules.

2. **Unit and Integration Testing**
   - Write unit tests for service logic (e.g., income, expense, and budget management).
   - Write integration tests to validate the interaction between routes and database.
   - Use mocks or testcontainers to simulate the database during testing.

3. **Security Improvements**
   - Implement rate limiting to prevent abuse of the API.
   - Use middleware to add HTTP security headers.
   - Ensure that all input validation is robust to prevent SQL injection or other security vulnerabilities.

**Milestone**: API documentation is complete, and all major features are covered by tests, with security best practices implemented.

## Phase 5: Deployment and Monitoring

**Goal**: Deploy the application to the cloud and set up monitoring for production use.

1. **CI/CD Pipeline**
   - Set up GitHub Actions or another CI/CD tool to automate building, testing, and deployment.
   - Build and push Docker images to Docker Hub or a private registry.

2. **Cloud Deployment**
   - Deploy the application using a cloud provider (e.g., AWS, GCP, Azure).
   - Use managed PostgreSQL as the database.
   - Use Kubernetes for container orchestration if desired.

3. **Monitoring and Logging**
   - Set up structured logging with a tool like `logrus`.
   - Use Prometheus for monitoring metrics like API response times and error rates.
   - Set up alerts for important metrics (e.g., high error rates).

**Milestone**: The application is live, with monitoring and logging set up for production use.

## Phase 6: Future Enhancements

**Goal**: Continue to add features and improvements to provide more value to users.

1. **Recurring Expense Automation**
   - Add functionality to automate recurring income or expenses.

2. **Machine Learning for Insights**
   - Implement basic ML models to analyze spending patterns and make personalized suggestions.

3. **Multi-Currency Support**
   - Add support for users to track income and expenses in different currencies with real-time exchange rates.

4. **Integration with Banking APIs**
   - Integrate with banking APIs to automatically sync user transactions.

**Milestone**: Continuous improvement of features and user experience.

### Development Timeline Summary

1. **Phase 1**: 1-2 weeks
2. **Phase 2**: 2-3 weeks
3. **Phase 3**: 1-2 weeks
4. **Phase 4**: 1-2 weeks
5. **Phase 5**: 2-3 weeks
6. **Phase 6**: Ongoing based on user feedback and needs

### Notes

- Focus on testing and validation throughout development to ensure robust and secure implementation.
- Prioritize Docker and cloud deployment for scalability and ease of use.
- Iteratively add value by gathering feedback from early users and focusing on solving their problems.

This development plan should help guide the process of creating a complete and production-ready Budget Tracking API. Let me know if you need further details or adjustments to this plan!
