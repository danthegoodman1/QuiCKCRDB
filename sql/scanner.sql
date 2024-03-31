-- name: SelectTopLevelQueues :many
select *
from quick_top_level_queue
where hash_token < $1
and vesting_time <= now()
;