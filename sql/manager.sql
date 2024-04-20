-- name: ObtainTopLevelQueue :one
update quick_top_level_queue
set lease_id = @new_lease
  , vesting_time = @vesting_time
where queue_zone = @queue_zone
and lease_id = @known_lease -- ensure it's still how we last saw it
    returning lease_id
;