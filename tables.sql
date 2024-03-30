create table quick_work_queue (
    queue_zone text not null,
    id text not null,
    payload text not null,
    priority int8,
    vesting_time timestamptz,
    lease_id text,

    primary key (queue_zone, id)
)
;

create index quick_work_queue_by_processing_order on quick_worker_queue(queue_zone, vesting_time, priority);


create table quick_top_level_queue (
    queue_zone text not null,
    vesting_time timestamptz,
    lease_id text,

    primary_key(queue_zone)
)
;

create index quick_top_level_queue_in_order on quick_top_level_quueue (vesting_time);


create table quick_top_level_queue_pointers (
    queue_zone text not null,
    vesting_time timestamptz,

    primary_key(queue_zone)
)
;