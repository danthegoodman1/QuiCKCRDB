# QuiCKCRDB Architecture

## Incremental re-hashing

An outstanding issue between FoundationDB and CockroachDB is the ability to use different isolation levels within the same transaction on FoundationDB. The lack of this functionality suggests to increase contention in CockroachDB when obtaining leases on queue zones (Qc). To solve this, we use a combination of hash and range partitioning within CockroachDB to create higher-level queue-zones by hash.

While this may seem difficult scale due to re-hashing, QuiCK conveniently implements a native feature to support incremental re-hashing: The pointer index.

The pointer index, used to reduce contention during enqueue by providing a low-contention check to see if a top-level queue record already exists, can naturally also store a ring size (or token count). This means that when we go to enqueue to a queue zone that was created before a hash ring change, we can use the previous hash ring size to ensure that we always hit the same index.

To optimize for incremental re-hashing, we do not use CRDB's native hash-partitioned indexes. Instead, we `ALTER TABLE ... SPLIT AT` to manually manage ranges. This allows us to directly communicate with a hash token across ring size changes.

## READ COMMITTED peeking

Due to CRDB's inability to change isolation levels mid-transaction, we instead peek in a separate transaction using `READ COMMITTED` isolation.