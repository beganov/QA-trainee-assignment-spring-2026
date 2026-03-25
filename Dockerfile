FROM golang:1.21

RUN apt-get update && apt-get install -y wget unzip openjdk-17-jre-headless && rm -rf /var/lib/apt/lists/*

ENV ALLURE_VERSION=2.24.1
RUN wget -q https://github.com/allure-framework/allure2/releases/download/${ALLURE_VERSION}/allure-${ALLURE_VERSION}.zip \
    && unzip allure-${ALLURE_VERSION}.zip \
    && mv allure-${ALLURE_VERSION} /opt/allure \
    && ln -s /opt/allure/bin/allure /usr/local/bin/allure \
    && rm allure-${ALLURE_VERSION}.zip

WORKDIR /app
COPY . .
RUN go mod download

CMD ["sh", "-c", "go test ./... -v && allure generate allure-results --clean -o allure-report"]