# Server

Server uses the internal components to run the application server that hosts both backend and frontend.

Directory structure

```
server
|- _certs   # ssl certificates
|- _config  # app configurations
|- _storage # app storage
|- public   # static files that hosts front-end
|- .env     # .env file that provides environment variables and app secrets which overrides _config/*
|- main.go  # main app to run
```
