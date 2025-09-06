## Write A Log Package

### A Write-Ahead Log (WAL) is a technique used in databases and distributed systems to ensure data durability and consistency

- `Record:` the data stored in our log.
- `Store:` the file we store records in.
- `Index:` the file we store index entries in.
- `Segment:` the abstraction that ties a store and an index together.
- `Log:` the abstraction that ties all the segments together.