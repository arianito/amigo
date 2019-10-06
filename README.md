# amigo
GOLANG sql migrate cli


# Usage
first set environment variables
```go

export DB_DRIVER="mysql"
export DB_QUERY="hello:123@tcp(localhost:3306)/mydb"
mkdir -p ./migrations
```

```bash
# to create migrationn
amigo create migration_record


# to run migrations
amigo up

# to downgrade migrations
amigo down

# to rollback migrations with step
amigo rollback 1

```


cli arguments:
```bash
amigo --path=./migrations
```
