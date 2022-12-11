# Azula a registry manipulator

Basically, I created this repo to play with a clean architecture design pattern.  
But nevertheless, there are some useful functions like deleting tags from the remote registry by picking them interactively :]

## Local registry with auth

```shell
# Start local registry
./_scripts/registry.sh up

# Stop and delete container
./_scripts/registry.sh down
```

## Usage example

```shell
# Note, you should be logged in with `docker login` into manipulated registry first
MAN_REGISTRY="https://your.registry.com" go run ./cmd/azula/main.go img del -l name_part
```
