# repo-migration

Migrate list of GitHub repositories to Gitea.

## Build

Go >= 1.17 required.

```bash
go build -a -o app .
```

## Configuration

Create `application.yml` with main application configuration.
Use `application.sample.yml` for reference.

Create `repos.txt` with list of GitHub repositories to import, one in line. For example:

```
https://github.com/n-at/box
https://github.com/n-at/dockerfiles
```

## Run

```bash
./app
```
