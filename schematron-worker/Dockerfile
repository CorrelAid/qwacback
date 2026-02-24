# Build stage
FROM gradle:8.12-jdk21 AS builder
WORKDIR /app
COPY schematron-worker/build.gradle.kts schematron-worker/settings.gradle.kts ./
RUN gradle dependencies --no-daemon || true
COPY schematron-worker/src/ src/
COPY schematron/ /app/schematron/
COPY xml/ /app/xml/
RUN gradle shadowJar --no-daemon

# Run stage
FROM eclipse-temurin:21-jre-alpine
RUN adduser -D -u 1001 appuser
WORKDIR /app
COPY --from=builder /app/build/libs/*-all.jar /app/worker.jar
COPY --from=builder /app/schematron /app/schematron
COPY --from=builder /app/xml /app/xml
RUN chown -R appuser:appuser /app

USER appuser
ENTRYPOINT ["java", "-jar", "/app/worker.jar"]
