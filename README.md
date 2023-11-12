# Uptime Checker

Monitoring background jobs such as backups, cron jobs, weekly reports. 
Uptime Checker notifies when these tasks don't run on time. 

Swagger documentation file: [docs/swagger.yaml](docs/swagger.yaml). 

**How to set up monitoring for your job:**
1. Sign up and sign in to application (**PUT /v1/auth/signup** and **PUT /v1/auth/signin**).
2. Add some channels where you want to get notifications (**POST /v1/channels**).
3. Create "check" for your job with specified name, description, channels, interval and grace timeout values (**POST /v1/checks**).
The endpoint returns the id of the created "check".
4. Then modify your job's code by adding requests to ping endpoints (**PUT /v1/pings/{checkId}** and/or **PUT /v1/pings/{checkId}/fail** and/or **PUT /v1/pings/{checkId}/start**).
5. Now you will receive notifications about delayed or failed jobs. You can also check statistics of your job's state changes (**GET /v1/checks/{id}/flips**).

### Architecture
Go application consists of 3 components:
- **Server (REST API).**
- **Poller.** Checks jobs that failed or delayed and sends notification requests to Rabbit MQ queue.
- **Notifier.** Consumes Rabbit MQ queue and sends notifications to corresponding channels (webhook/email).

Databases: 
- **PostgreSQL** stores almost all the data (users, checks, channels, ping...).
- **Redis** stores user sessions.
- **Rabbit MQ** used as notifications queue.

### How to run in dev mode
1. Install [Task](https://taskfile.dev/installation/) if it is not installed.
2. Create .env file and copy .env.example contents into it. Set MAILJET* variables if you want to test how the notifier works.
3. `task services:run`
4. `task run` (Runs the server) / `task poller:run` (Runs the poller) / `task notifier:run` (Runs the notifier).

### How to run tests
1. Install [Task](https://taskfile.dev/installation/) if it is not installed.
2. `task test:run`