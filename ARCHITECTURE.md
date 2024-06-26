# QuiCKCRDB Architecture

QuiCKCRDB is designed to fail fast: if there are errors in operations against the database, the default operation is the fatal log (log, flush, and exit 1). This ensures that any incorrect state that could possible exist is immediately terminated, and the goroutines are not orphaned to never being able to process.

## Optional workers

QuiCKCRDB does not require you to run the Scanner, Manager, and Worker goroutines like QuiCK does. This means it can be used as a pull-queue for remote consumers.

To guarantee in-order processing in this mode, it is the responsibility of the consumer to request items sequentially.

## Incremental re-hashing

An outstanding issue between FoundationDB and CockroachDB is the ability to use different isolation levels within the same transaction on FoundationDB. The lack of this functionality suggests to increase contention in CockroachDB when obtaining leases on queue zones (Qc). To solve this, we use a combination of hash and range partitioning within CockroachDB to create higher-level queue-zones by hash.

While this may seem difficult scale due to re-hashing, QuiCK conveniently implements a native feature to support incremental re-hashing: The pointer index.

The pointer index, used to reduce contention during enqueue by providing a low-contention check to see if a top-level queue record already exists, can naturally also store a previously used hash token. This means that when we go to enqueue to a queue zone that was created before a hash ring change, we can use the previous hash ring size to ensure that we always hit the same index.

To optimize for incremental re-hashing, we do not use CRDB's native hash-partitioned indexes. Instead, we `ALTER TABLE ... SPLIT AT` to manually manage ranges. This allows us to directly communicate with a hash token across ring size changes.

This is analogous to multiple FoundationDB clusters in QuiCK.

## Caching the pointer index

CockroachDB is notable more sensitive to hot spots than FoundationDB, particularly around reading. In order to solve this, inserting nodes may use a in-memory cache for p, the pointer index to Qc to see if the queue zone exists in the top-level queue.

This is safe because Qc (and thus p) are lazily garbage collected, the pointer index is an optimization (and therefore can fail-through), and newly ingested records generally push the vesting time further back. The rule of if Vesting(p) >> Vesting(x) then update Vesting(p) and Vesting(Qc).

The cache duration must never be longer than the `min_inactive` duration (and generally should be less to account for local clock drift), as otherwise you risk orphaning items in the queue.

## Hash token walking

Like how QuiCK consumers walk multiple FoundationDB clusters, QuiCKCRDB consumers walk multiple hash tokens. Specifically, they walk them in order. If all nodes are started at the same time, this can introduce some initial increased contention. But over time they will spread out more evenly to cover the hash ring.

Likewise, the scanner algorithm has been adjusted to process per hash token. Additionally, it does not attempt to spin on a given hash token, but rather peeks each one once per step of the hash ring.

## Limitations

There are some closely related limitations to QuiCK on CRDB, both relating to transactions.

The first is connection pooling. Because transactions hold open a connection, the throughput will likely be limited by content on pool connections. FoundationDB uses optimistic transactions such that connections aren't held unless an operation is currently processing: They don't have 'sessions'.

The next issue is that because there are the extra delays between write operation round trips due to the session-style transactions, this increases the time window when conflicts can occur. The pointer index and hash ring should both reduce this greatly, but it's still likely to be higher than FoundationDB's conflict rate.

Another issue is a single transaction can't use multiple isolation levels, so if you want to use a read committed isolation level to read the pointer index, then you'll need to use another connection. This further exacerbates the problem of connection pool content, and compounds the issue with more time for serialization conflicts to occur.