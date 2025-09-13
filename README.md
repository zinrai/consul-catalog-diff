# consul-catalog-diff

A tool to detect differences between Transaction API formatted JSON operations and the current state in Consul Catalog.

## Overview

`consul-catalog-diff` compares JSON-formatted operations (expected state) in Transaction API format against the current state in Consul Catalog and reports any differences.

- Automatically detects JSON format (NDJSON, JSON array)
- Only checks elements defined in the input JSON (ignores Consul-only elements)
- Detects differences for explicit delete operations when `"Verb": "delete"` is specified
- Designed for CI/CD pipelines with meaningful exit codes

## Diff detection logic

1. **Additions**: Elements in JSON with `"Verb": "set"` that don't exist in Consul
2. **Modifications**: Elements that exist in both JSON and Consul but have different values
3. **Deletions**: Only detected for elements with `"Verb": "delete"` in JSON

Elements that exist only in Consul (e.g., registered by Nomad) are ignored.

## Installation

```bash
$ go install github.com/example/consul-catalog-diff@latest
```

## Usage

```bash
$ consul-catalog-diff -file operations.json -consul-addr http://consul:8500
```

### Command-line options

- `-file PATH` (required): Path to JSON/NDJSON file containing expected operations
- `-consul-addr URL`: Consul HTTP address (default: `http://127.0.0.1:8500`)
- `-version`: Show version
- `-help`: Show help message

### Exit codes

- `0`: No differences found
- `1`: Differences found
- `2`: Error occurred

## Input formats

The tool automatically detects the following formats:

### NDJSON Transaction format

Each line is a complete JSON object representing a single operation:

```json
{"Node":{"Verb":"set","Node":{"Node":"web-001","Address":"10.0.0.1","Datacenter":"dc1"}}}
{"Service":{"Verb":"set","Node":"web-001","Service":{"ID":"nginx","Service":"nginx","Port":80}}}
{"Service":{"Verb":"delete","Node":"old-node","Service":{"ID":"deprecated-service"}}}
```

### JSON Transaction Array format

A JSON array containing operation objects:

```json
[
  {
    "Node": {
      "Verb": "set",
      "Node": {
        "Node": "web-001",
        "Address": "10.0.0.1",
        "Datacenter": "dc1"
      }
    }
  },
  {
    "Service": {
      "Verb": "set",
      "Node": "web-001",
      "Service": {
        "ID": "nginx",
        "Port": 80
      }
    }
  }
]
```

## Example output

```
=== Consul Catalog Diff Report ===
Total changes: 3

NODE CHANGES:
  Additions (1):
    + web-003 [10.0.0.3]
      Address: 10.0.0.3
      Datacenter: dc1
      Meta: map[type:web]
  Modifications (1):
    ~ web-001
      - Address: 10.0.0.1 -> 10.0.0.100
      - Meta.location: rack-1 -> rack-2

SERVICE CHANGES:
  Additions (1):
    + web-003/nginx port:80
      Port: 80
      Service: nginx
      Tags: [web, primary]
```

## License

This project is licensed under the [MIT License](./LICENSE).
