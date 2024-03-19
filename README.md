# Argus

<p align="center">
    <img src="argusdb.png" alt="Project Image" width="50%">
</p>

## Description

A scalable rate limiter service written in Golang.

## Exmaple Usage

```
curl -X POST -H "Content-Type: application/json" -d '{
    "key": "my_key",
    "capacity": 1,
    "interval": 5,
    "unit": "s"
}' http://localhost:8123/api/v1/limit
```

## Installation

1. Clone the repository.
2. Install the dependencies.
3. Run the project.

## Usage

How to use the project.

## Contributing

Contributing guidelines.

## License

This project is licensed under the [Apache License](LICENSE).