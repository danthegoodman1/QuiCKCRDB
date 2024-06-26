create table quick_work_queue (
    queue_zone text not null,
    id int8 not null,
    payload text not null,
    priority int8,
    vesting_time timestamptz,
    lease_id text,

    primary key (queue_zone, id)
)
;

create index quick_work_queue_by_processing_order on quick_worker_queue(queue_zone, priority, vesting_time) where vesting_time is not null and priority is not null;


create table quick_top_level_queue (
    queue_zone text not null,
    vesting_time timestamptz not null,
    lease_id text,
    hash_token int8 not null,

    primary key(queue_zone)
)
;

create index quick_top_level_queue_in_order on quick_top_level_queue (hash_token, vesting_time);


create table quick_top_level_queue_pointers (
    queue_zone text not null,
    vesting_time timestamptz,
    hash_token int8 not null,

    primary key(queue_zone)
)
;
