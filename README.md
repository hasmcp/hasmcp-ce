# HasMCP-CE

HasMCP is a tool that converts API endpoints into MCP Server without writing a single line of code. It is practical alternative for developers and SaaS owner who prefers not to develop anad maintain MCP spec changes over time.

HasMCP-CE shares the some logic for the commercial
versions of the `HasMCP-Pro` and `HasMCP-Enterprise`.

NOTE: `HasMCP-Pro` and `HasMCP-Enterprise` versions are maintained in private repositories and source codes are not available. HasMCP Cloud service itself uses the `HasMCP-Enterprise` edition on
production.

## HasMCP-CE Features

- Automated MCP server creation using OpenAPI Spec v3+ and Swagger

- Oauth2 authentication

- Manual MCP from API endpoints

- Toggle endpoints per MCP Server

- Proxy headers (optional per MCP Server) to actual API endpoints

- Long term, short-term authentication tokens per MCP Server

- Live tail MCP Server tool call logs

- Optional automated SSL with Let's encrypt

## HasMCP Cloud Features

Cloud version is available with Hobby(with monthly free-tier) and Pro subscriptions at [hasmcp.com](https://hasmcp.com). It has some additonal feaatures compared to the Community Edition:

- Request/response payload optimization per tool with a minimum coding (interceptors). Up to 98% token reduction on MCP tool responses (depending on what is needed by the described tool).
- Per tool call, per user usage analytics
- Users, teams management (Pro and Enterprise only)
- Audit logs (Enterprise only)

**Why does it have cloud version?**

- To support futher development
- Help developers and companies those are not intersted in running a server and maintain it
- Entrepreneurship

## Roadmap

### Highest priority

[ ] Bug fixes (Highest priority)
[ ] Protocol updates

### HasMCP-CE (including Pro and Enterprise) short term roadmap

**Observability and analytics**

[ ] Live MCP server analytics (In progress)

**Functionality and token optimizations**

[ ] MCP composition with Search/Add/Remove by LLMs directly (ETA: January 2026)
[ ] Toon format on responses (ETA: January 2026)

**Extended protocol support**

[ ] GRPC support (ETA: February 2026)

### Long term road map

[ ] Github integration (Git as source of truth)
[ ] Your requirements

## Development

HasMCP-CE is using monorepo approach that hosts both frontend and backend codes in the same
repository.

### Directory Structure

```
hasmcp-ce
|- backend            # backend for hasmcp-ce
|- frontend           # front-end for hasmcp-ce
|- COMMERCIAL_LICENSE # Commercial license
|- Dockerfile         # Dockerfile
|- LICENSE            # License file
|- Makefile           # Handy Make commands to help developers
|- README.md          # The current file that you are reading now
```

Development environment requires 2 services up and running at the same time:

1. Frontend

```
cd frontend
npm run dev # this will open port 5173
```

2. Backend

```
cd backend/cmd/server
go run main.go
```

3. Database

By default hasmcp-ce uses sqlite so you don't need to do anything. It is more than enough for single user experience.

To enable Postgres instead of Sqlite, you can add the following to the your `.env` file

```
POSTGRES_ENABLED=false # set true to use postgres as production db

# below are the your db connection details for postgres
POSTGRES_TIMEZONE=UTC
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DBNAME=postgres
```

## Running the HasMCP-CE with Docker (recommended)

The recommended way of running HasMCP-CE version using docker. Please do not confuse using docker does not mean that
using the latest version. The following example is using the latest version, please adjust the version and changes based
on your needs!

### Setup

Apply both `One time setup` and `Configurations` for the first run then you can use only `Configurations` section.

**One time setup**

```
mkdir hasmcp
cd hasmcp
mkdir -p _certs _storage # creates folders
chmod 0777 _certs
chmod 0777 _storage
wget https://github.com/hasmcp/hasmcp-ce/blob/main/backend/cmd/server/.env.example -O .env
```

**.env file**

`.env` overrides the configurations that to ensure you do not need to store.

```
cd hasmcp
wget .env.example .env
```

After downloading the example .env file, edit the content as you wish.

### Run

Latest version:

```
docker stop hasmcp-ce || true; \
	docker rm hasmcp-ce || true; \
	docker image prune -f; \
	docker pull hasmcp/hasmcp-ce:latest; \
	docker run --env-file .env -p 80:80 -p 443:443 --name hasmcp-ce \
    -v ./_certs:/_certs \
    -v ./_storage:/_storage \
    -d --restart always hasmcp/hasmcp-ce:latest
```

Other known versions for example v0.2.0:

```
HASMCP_VERSION=v0.2.0 \
  docker stop hasmcp-ce || true; \
  docker rm hasmcp-ce || true; \
  docker image prune -f; \
  docker pull hasmcp/hasmcp-ce:$HASMCP_VERSION; \
  docker run --env-file .env -p 80:80 -p 443:443 \
    --name hasmcp-ce \
    -v ./_certs:/_certs \
    -v ./_storage:/_storage \
    -d \
    --restart always \
    hasmcp/hasmcp-ce:$HASMCP_VERSION
```

## Documentation

Documentation for tutorials and terminology are available at [docs.hasmcp.com](https://docs.hasmcp.com).

## Licenses

- [GPLV3](./LICENSE)
- [Commercial License](./COMMERCIAL_LICENSE) is available that removes the restrictions of GPLV3 with license purchase.

(c) 2026 Contextual, Inc.
[hasmcp.com](https://hasmcp.com)

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

### Licenses of Dependencies

#### Frontend

All front-end dependencies are ensured to have either MIT and Apache 2.0 licenses. In the future if any of them switches
to another license, it might be evaluated again and replaced if needed.

For all frontend dependencies please see the [package.json](./frontend/package.json) file.

#### Backend

For all backend dependencies please see the [go.mod](./backend/go.mod) file.

##### MIT

- github.com/glebarez/sqlite
- github.com/gofiber/contrib/fiberzerolog
- github.com/gofiber/fiber/v2
- github.com/golang-jwt/jwt/v5
- github.com/kaptinlin/jsonschema
- github.com/robfig/cron/v3
- github.com/rs/zerolog
- github.com/valyala/fasthttp
- gorm.io/driver/postgres
- gorm.io/gorm

##### Apache 2.0

- github.com/mustafaturan/monoflake
- gopkg.in/yaml.v3

##### Other

- [golang.org/x/crypto](https://cs.opensource.google/go/x/crypto/+/master:LICENSE)
