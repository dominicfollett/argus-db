# Argus

<p align="center">
    <img src="argusdb.png" alt="Project Image" width="50%">
</p>

## Description

A simple rate limiter service written in Golang. This was built as a quick exercise to revise concurrency in Go.
The rate limiter uses a thread-safe BST to store records. A 'shadow' AVL tree is periodically swapped with the BST
to provide eventual log(n) guarantees for tree accesses.

## Known Issues

Occassionally under heavy concurrent requests, a Read Lock is not being correctly released leading to starvation
of the switchover go routine that handle tree swapping.

## Todos

- Resolve the deadlock issue.
- Perform profiling and implement optimizations (e.g. custom json decoding among others) 
- Capture some performance benchmarks
- Implement an alternate DB engine using perhaps a thread-safe hash table.
- And lastly, there's a bit of cleanup/refactoring needed

## Performance Benchmarks

Incomming

## Example Usage
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
2. Set environment variables
3. `make all`
4. `./bin/argus`

Alternately use the provided Dockerfile to deploy a containerized version. 

```sh
$ docker-compose up --build
```

## Contributing

Contributions are welcome. If you would like to contribute please follow these steps:

1. Fork the repository
2. Create a new branch
3. Make your changes
4. Submit a pull request

## License

This project is licensed under the [Apache License](LICENSE).