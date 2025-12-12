# Simple Home Server Deployment

This guide outlines a simple process for deploying and updating the project on a home server using Git and Docker Compose.

## Prerequisites

Before you begin, ensure your home server has the following installed:
- **Git:** For cloning and pulling the project from GitHub.
- **Docker:** For running the application in containers.
- **Docker Compose:** For managing the multi-container application.

## Initial Setup

1.  **Clone the Repository:**
    Log into your home server via SSH and clone your project repository from GitHub.

    ```bash
    git clone <your-github-repository-url>
    cd camera-detection-project
    ```

2.  **Create Environment File:**
    Create a `.env` file for your production environment. You can start by copying the example file.

    ```bash
    cp .env.example .env
    ```

    Edit the `.env` file and fill in all the required variables for your setup (database passwords, camera URLs, etc.).

3.  **Initial Launch:**
    For the very first launch, run the deployment script. This will pull the latest code (which you already have), build the Docker images, and start all the services.

    ```bash
    bash bin/deploy.sh
    ```

    The application should now be running. You can check the status of the containers with `docker-compose ps`.

## Updating the Application

To update your application with the latest changes from your GitHub repository, simply run the deployment script again from your project directory on the home server.

```bash
bash bin/deploy.sh
```

This script will automatically:
1.  Pull the latest changes from the `main` branch.
2.  Stop the currently running containers.
3.  Rebuild the Docker images if there are any changes.
4.  Restart all services in detached mode.

---

### Optional: Automating with GitHub Actions

For a fully automated CI/CD process, you can set up a GitHub Action to automatically run the `deploy.sh` script on your server whenever you push changes to the `main` branch.

This typically involves:
1.  Creating a new workflow file (e.g., `.github/workflows/deploy.yml`).
2.  Using an action like `appleboy/ssh-action` to connect to your server.
3.  Storing your server's SSH credentials (`host`, `username`, `key`) as secrets in your GitHub repository settings.
4.  Configuring the workflow to run the `bash bin/deploy.sh` command on your server.
