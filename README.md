# Plain rdiff

Simple implementation of rdiff functionalities.

## Build

```bash
go build -o plain-rdiff .
```

## Usage

```bash
plain-rdiff signature old-file signature-file
plain-rdiff delta signature-file new-file delta-file
plain-rdiff patch basis-file delta-file new-file
```


## Testing

### Unit tests
```bash
go test -v ./...
```

### e2e tests
Tests generate file with random content and second file with fixed amount of randomly changed bytes. Both files have equal lengths. Then both files are processed (signature->delta->patch) and SHA256 of respective files are compared.
```bash
go test -v -timeout 10m -run ^TestRdiffFile$ plain-rdiff -tags e2e
```

