SELECT execute($$

    ALTER TABLE compute_plans
    DROP COLUMN name;

    DROP VIEW expanded_compute_plans;

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
    LEFT JOIN compute_tasks t ON cp.key = t.compute_plan_key
    GROUP BY cp.key;

$$) WHERE column_exists('public', 'compute_plans', 'name');
