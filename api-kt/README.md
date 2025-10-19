# Kotlin API

The backend built with Kotlin, Spring Boot, SQLite, JOOQ and Redis.

## Getting Started

The Spring profile in this project is controlled by the `SPRING_PROFILES_ACTIVE` environment variable, which can be
set to `dev` (default), `prod`, or `test`.

Before running the application, ensure that the Redis service is up.

### Run in development

1. Generate JOOQ Code

```bash
./mvnw jooq-codegen:generate
```

2. Start the Application

```bash
PEBBLE_PASSWORD=xxx ./mvnw spring-boot:run
```

### Run in production

1. Build with Maven

```bash
./mvnw clean package
```

2. Run the JAR

```bash
PEBBLE_PASSWORD=xxx SPRING_PROFILES_ACTIVE=prod java -jar target/pebble-1.0.0.jar
```

The `PEBBLE_PASSWORD` variable is used for login. Ensure it is complex and securely stored in production.

### Database Migration

To create the sqlite database and tables if missing:

```bash
./mvnw flyway:migrate -Dflyway.url=jdbc:sqlite:../data/app-dev.db
```
