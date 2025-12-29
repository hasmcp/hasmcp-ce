FROM node:25-alpine3.22 AS frontendbuild

# Set the working directory inside the container
WORKDIR /app

# Copy package.json
COPY ./frontend/package.json ./

# Install project dependencies
RUN npm install

# Copy the rest of the application code
COPY ./frontend .

# Build the front-end app to copy to public
RUN npm run build

RUN apk add brotli

# RUN apt-get update && apt-get install -y --no-install-recommends brotli

# Compress files in the root directory (public/)
RUN	find dist -maxdepth 1 -type f -not -name "*.gz" -not -name "*.br" -exec gzip -9 -f -k {} +
RUN find dist -maxdepth 1 -type f -not -name "*.gz" -not -name "*.br" -exec brotli -9 -f -k {} +

# Compress files in the assets directory (public/assets/)
RUN find dist -maxdepth 1 -type f -not -name "*.gz" -not -name "*.br" -exec gzip -9 -f -k {} +
RUN	find dist -maxdepth 1 -type f -not -name "*.gz" -not -name "*.br" -exec brotli -9 -f -k {} +

FROM golang:1.25 AS build

WORKDIR /app
RUN apt-get update && \
  apt-get install -y gcc && \
  rm -rf /var/lib/apt/lists/*

# RUN apt-get install -y ca-certificates
COPY backend/go.mod backend/go.sum ./
RUN GO111MODULE=on go mod download

COPY backend/internal internal
COPY backend/cmd cmd

# Remove .env file
RUN rm -f cmd/server/.env
RUN rm -f cmd/server/.env.development
RUN rm -f cmd/server/.env.staging
RUN rm -f cmd/server/.env.production
RUN rm -f cmd/server/.env.example

# Clean old public files
RUN rm -rf cmd/server/public/*

# Copy new public files
COPY --from=frontendbuild /app/dist cmd/server/public
#COPY LICENSE cmd/server/public
COPY COMMERCIAL_LICENSE cmd/server/public
RUN ls -ltrh cmd/server/public

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ENV CGO_ENABLED=0

RUN go build -v -o hasmcp cmd/server/main.go

# Create a "nobody" non-root user for the next image by crafting an /etc/passwd
# file that the next image can copy in. This is necessary since the next image
# is based on scratch, which doesn't have adduser, cat, echo, or even sh.
RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd

# No need extra files
FROM scratch

COPY --from=build /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=build /app/hasmcp /
COPY --from=build /etc_passwd /etc/passwd
COPY --from=build /app/cmd/server/public /public
COPY --from=build /app/cmd/server/_config /_config
COPY --from=build /app/cmd/server/_storage/README.md /_storage/README.md

USER nobody

CMD ["/hasmcp"]
