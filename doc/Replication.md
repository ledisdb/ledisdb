At first, LedisDB uses BinLog (like MySQL BinLog) to support replication. Slave syncs logs from Master with specified BinLog filename and position. It is simple but not suitable for some cases. 

Let's assume below scenario: A -> B and A -> C, here A is master, B and C are slaves. -> means "replicates to". If master A failed, we must select B or C as the new master. Usually, we must choose the one which has most up to date from A, but it is not easy to check which one is it.

MySQL has the same problem for this, so from MySQL 5.6, it introduces GTID (Global Transaction ID) to solve it. GTID is very powerful but a little complex, I just want to a simple and easy solution.

Before GTID, Google has supplied a solution called [Global Transaction IDs](https://code.google.com/p/google-mysql-tools/wiki/GlobalTransactionIds) which uses a monotonically increasing group id to represent an unique transaction event in BinLog. Although it has some limitations for MySQL hierarchical replication, I still think using a integer id like group id for log event is simple and suitable for LedisDB.

Another implementation influencing me is [Raft](http://raftconsensus.github.io/), a consensus algorithm based on the replicated log. Leader must ensure that some followers receive the replicated log before executing the commands in log. The log has an unique log id (like group id above), if the leader failed, the candidate which has the up to date log (checked by log id) will be elected a new leader. 

Refer above, I supply a simple solution for LedisDB's replication.

## Keyword

+ LogID: a monotonically increasing integer for a log
+ FirstLogID: the oldest log id for a server, all the logs before this id have been purged.
+ LastLogID: the newest log id for a server.
+ CommitID: the last log committed to execute. If LastLogID is 10 and CommitID is 5, server needs to commit logs from 6 - 10 to catch the up to date status.

## Sync Flow

For a master, every write changes will be handled below:

1. Logs the changes to disk, it will calculate a new LogID based on LastLogID.
2. Sends this log to slaves and waits the ACK from slaves or timeout.
3. Commits to execute the changes.
4. Updates the CommitID to the LogID.

For a slave:

1. Connects to master and tells it which log to sync by LogID, it may have below cases:
    
    + The LogID is less than master's FirstLogID, master will tell slave log has been purged, the slave must do a full sync from master first.
    + The master has this log and will send it to slave.
    + The master has not this log (The slave has up to date with master). Slave will wait for some time or timeout then to start a new sync.

2. After slave receiving a log (eg. LogID 10), it will save this log to disk and notice the replication thread to handle it.
3. Slave will start a new sync with LogID 11.


## Full Sync Flow

If slave syncs a log but master has purged it, slave has to start a full sync.

+ Master generates a snapshot with current LastLogID and dumps to a file.
+ Slave discards all old data and replicated logs, then loads the dump file and updates CommitID with LastLogID in dump file.
+ Slave starts to sync with LogID = CommitID + 1.

## ReadOnly

Slave is always read only, which means that any write operations will be denied except `FlushAll` and replication.

For a master, if it first writes log OK but commits or updates CommitID error, it will also turn into read only mode until replication thread executes this log correctly.

## Strong Consensus Replication

For the sync flow, we see that master will wait some slaves to return an ACK telling it has received the log, this mechanism implements strong consensus replication. If master failed, we can choose a slave which has up to date data with the master. 

You must notice that this feature has a big influence on the performance. Use your own risk!

## Use 

Using replication is very simple for LedisDB, only using `slaveof` command.

+ Use `slaveof host port` to enable replication from master at "host:port".
+ Use `slaveof no one` to stop replication and change the slave to master. 

If a slave first syncs from a master A, then uses `slaveof` to sync from master B, it will sync with the LogID = LastLogID + 1. If you want to start over from B, you must use `slaveof host port restart` which will start a full sync first. 

## Limitation

+ Multi-Master is not supported.
+ Replication can not store log less than current LastLogID.
+ Circular replication is not supported.
+ Master and slave must set `use_replication` to true to support replication.

