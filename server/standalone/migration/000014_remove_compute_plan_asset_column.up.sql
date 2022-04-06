SELECT execute($$
    ALTER TABLE compute_plans
    ADD COLUMN owner                      varchar(100),
    ADD COLUMN delete_intermediary_models bool,
    ADD COLUMN creation_date              timestamptz,
    ADD COLUMN tag                        varchar(100),
    ADD COLUMN metadata                   jsonb,
    ADD CONSTRAINT compute_plans_owner_channel_fkey FOREIGN KEY (owner, channel) REFERENCES nodes (id, channel);

    UPDATE compute_plans
    SET owner = asset ->> 'owner',
        delete_intermediary_models = (asset ->> 'deleteIntermediaryModels')::bool,
        creation_date = (asset ->> 'creationDate')::timestamptz,
        tag = asset ->> 'tag',
        metadata = COALESCE(asset -> 'metadata', '{}'::jsonb);

    ALTER TABLE compute_plans
    ALTER COLUMN owner SET NOT NULL,
    ALTER COLUMN delete_intermediary_models SET NOT NULL,
    ALTER COLUMN creation_date SET NOT NULL,
    ALTER COLUMN tag SET NOT NULL,
    ALTER COLUMN tag SET DEFAULT '',
    ALTER COLUMN metadata SET NOT NULL,
    ALTER COLUMN metadata SET DEFAULT '{}'::jsonb;

    DROP INDEX ix_compute_plans_creation;
    DROP INDEX ix_compute_plans_owner;

    CREATE INDEX ix_compute_plans_creation_date ON compute_plans (creation_date);
    CREATE INDEX ix_compute_plans_owner ON compute_plans (owner);

    ALTER TABLE compute_plans
    RENAME COLUMN id to key;

    ALTER TABLE compute_plans
    DROP COLUMN asset;

    CREATE VIEW expanded_compute_plans AS
    SELECT cp.key                                               AS key,
           cp.channel                                           AS channel,
           cp.owner                                             AS owner,
           cp.delete_intermediary_models                        AS delete_intermediary_models,
           cp.creation_date                                     AS creation_date,
           cp.tag                                               AS tag,
           cp.metadata                                          AS metadata,
           COUNT(1)                                             AS task_count,
           COUNT(1) FILTER (WHERE t.status = 'STATUS_WAITING')  AS waiting_count,
           COUNT(1) FILTER (WHERE t.status = 'STATUS_TODO')     AS todo_count,
           COUNT(1) FILTER (WHERE t.status = 'STATUS_DOING')    AS doing_count,
           COUNT(1) FILTER (WHERE t.status = 'STATUS_CANCELED') AS canceled_count,
           COUNT(1) FILTER (WHERE t.status = 'STATUS_FAILED')   AS failed_count,
           COUNT(1) FILTER (WHERE t.status = 'STATUS_DONE')     AS done_count
    FROM compute_plans cp
    LEFT JOIN compute_tasks t ON key = t.compute_plan_id
    GROUP BY cp.key;
$$) WHERE NOT view_exists('public', 'expanded_compute_plans');
