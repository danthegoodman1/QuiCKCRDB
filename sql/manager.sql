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
    from quick_work_queue
      where quick_work_queue.queue_zone = $1
      and quick_work_queue.vesting_time <= now()
    order by priority, vesting_time
    limit $2
)
update quick_work_queue
set vesting_time = $3
from toupdate
where vesting_time <= now()
and quick_work_queue.queue_zone = toupdate.queue_zone
and quick_work_queue.id = toupdate.id
returning *
;

-- name: CheckQueueHasAtLeastOneItem :one
select coalesce((
    select 1
    from quick_work_queue
    where queue_zone = $1
    limit 1
), 0)::bool
;