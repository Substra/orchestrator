SELECT execute($$
    ALTER TABLE compute_plans
    ADD COLUMN failure_date timestamptz;

    WITH failures as (
        SELECT
            (asset->>'computePlanKey')::uuid as compute_plan_key,
            min(timestamp) as timestamp
        FROM events
        WHERE event_kind = 'EVENT_ASSET_UPDATED'
          AND asset_kind = 'ASSET_COMPUTE_TASK'
          AND asset->>'status' = 'STATUS_FAILED'
        GROUP BY (asset->>'computePlanKey')::uuid
    )
    UPDATE compute_plans
    SET failure_date = failures.timestamp
    FROM failures
    WHERE compute_plans.key = failures.compute_plan_key;

    DROP VIEW expanded_compute_plans;

    UPDATE events
    SET asset = asset - 'status'
    WHERE asset_kind = 'ASSET_COMPUTE_PLAN';
$$) WHERE NOT column_exists('public', 'compute_plans', 'failure_date');
