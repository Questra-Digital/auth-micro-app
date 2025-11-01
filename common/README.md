# 🐳 Service Stack Configuration

This repository contains the necessary configuration to run the core services (Database, Cache, and Message Queue) for the application using **Docker Compose**.

## 🚀 Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

You need to have **Docker** and **Docker Compose** installed on your system.

* [Install Docker](https://docs.docker.com/engine/install/)
* [Install Docker Compose](https://docs.docker.com/compose/install/)

### 🛠️ Installation and Setup

1.  **Clone the repository:**
    ```bash
    git clone git@github.com:Questra-Digital/auth-micro-app.git
    cd auth-micro-app/common
    ```

2.  **Configure Environment Variables:**
    The services rely on environment variables for configuration, particularly for sensitive data like passwords.

    * Copy the example environment file:
        ```bash
        cp .env.example .env
        ```
    * **Edit the `.env` file** and replace the default values with your desired secure credentials.

    > **Note:** The values in `.env` are crucial for the services to run correctly and securely.

3.  **Start all Services (Database, Cache, and Message Queue):**
    Use the following command to build and start all services defined in `docker-compose.yml`:
    ```bash
    docker compose up -d
    ```
    The `-d` flag runs the containers in detached mode (in the background).

### ⚙️ Running Specific Profiles

The configuration uses **Docker Compose Profiles** to allow running subsets of services.

| Profile Name | Services Included | Command to Run |
| :--- | :--- | :--- |
| `` | `db` | `docker compose up -d` |
| `all` | `db`, `cache`, `message_queue` | `docker compose --profile all up -d` (Default when running `up -d`) |
| `cache` | `db`, `cache` | `docker compose --profile cache up -d` |
| `queue` | `db`, `message_queue` | `docker compose --profile queue up -d` |