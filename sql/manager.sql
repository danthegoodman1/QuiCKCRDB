-- name: ObtainTopLevelQueue :one
update quick_top_level_queue
set lease_id = @new_lease
  , vesting_time = @vesting_time
where queue_zone = @queue_zone
and lease_id = @known_lease -- ensure it's still how we last saw it
    returning lease_id
;

-- name: DequeueItems :many
with toupdate as (
    select *
    from quick_top_level_queue
    where quick_top_level_queue.hash_token < $1
      and quick_top_level_queue.vesting_time <= now()
      and quick_top_level_queue.queue_zone = $2
    limit $3
)
update quick_work_queue
set vesting_time = $4
from toupdate
where vesting_time <= now()
and quick_work_queue.queue_zone = toupdate.queue_zone
and quick_work_queue.id = toupdate.id
returning *
;