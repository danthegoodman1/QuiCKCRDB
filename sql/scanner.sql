-- name: SelectTopLevelQueues :many
select *
from quick_top_level_queue
where hash_token < $1
and vesting_time <= now()
;

-- name: ObtainTopLevelQueue :one
update quick_top_level_queue
set lease_id = $1
, vesting_time = $2
where queue_zone = $3
returning lease_id
;