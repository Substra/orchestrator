SELECT execute($$

    INSERT INTO asset_kinds(kind) VALUES ('ASSET_COMPUTE_TASK_OUTPUT_ASSET');

    CREATE TABLE compute_task_output_assets (
        compute_task_key UUID NOT NULL,
        compute_task_output_identifier varchar(100) NOT NULL,
        position integer NOT NULL,
        asset_kind varchar(50) REFERENCES asset_kinds(kind),
        asset_key varchar(200) NOT NULL,
        FOREIGN KEY (compute_task_key, compute_task_output_identifier) REFERENCES compute_task_outputs(compute_task_key, identifier),
        PRIMARY KEY(compute_task_key, compute_task_output_identifier, position),
        UNIQUE(asset_key)
    );

    -- insert head models
    INSERT INTO compute_task_output_assets (compute_task_key, compute_task_output_identifier, position, asset_kind, asset_key)
    SELECT compute_task_key, 'local', 0, 'ASSET_MODEL', key
    FROM models WHERE category = 'MODEL_HEAD';

    -- insert trunk models
    INSERT INTO compute_task_output_assets (compute_task_key, compute_task_output_identifier, position, asset_kind, asset_key)
    SELECT m.compute_task_key, 'shared', 0, 'ASSET_MODEL', m.key
    FROM models m
    LEFT JOIN compute_tasks t ON m.compute_task_key = t.key
    WHERE m.category = 'MODEL_SIMPLE' AND t.category = 'TASK_COMPOSITE';

    -- insert train/aggregate models
    INSERT INTO compute_task_output_assets (compute_task_key, compute_task_output_identifier, position, asset_kind, asset_key)
    SELECT m.compute_task_key, 'model', 0, 'ASSET_MODEL', m.key
    FROM models m
    LEFT JOIN compute_tasks t ON m.compute_task_key = t.key
    WHERE m.category = 'MODEL_SIMPLE' AND (t.category = 'TASK_AGGREGATE' OR t.category = 'TASK_TRAIN');

    -- insert predictions artifacts
    INSERT INTO compute_task_output_assets (compute_task_key, compute_task_output_identifier, position, asset_kind, asset_key)
    SELECT m.compute_task_key, 'predictions', 0, 'ASSET_MODEL', m.key
    FROM models m
    LEFT JOIN compute_tasks t ON m.compute_task_key = t.key
    WHERE m.category = 'MODEL_SIMPLE' AND t.category = 'TASK_PREDICT';

    -- insert performances
    INSERT INTO compute_task_output_assets (compute_task_key, compute_task_output_identifier, position, asset_kind, asset_key)
    SELECT compute_task_key, 'performance', 0, 'ASSET_PERFORMANCE', CONCAT(compute_task_key, '|', algo_key)
    FROM performances;

    -- Create events for ComputeTaskOutputAsset entities
    CREATE TEMPORARY SEQUENCE seq_tmp_events_offset AS bigint MINVALUE 1;

    -- locking the table to prevent any sequence issue
    LOCK TABLE events IN SHARE ROW EXCLUSIVE MODE;
    INSERT INTO events (id, position, asset_key, channel, asset_kind, event_kind, timestamp, asset)
    SELECT
        gen_random_uuid(),
        (SELECT MAX(position) FROM events) + NEXTVAL('seq_tmp_events_offset'),
        o.compute_task_key,
        t.channel,
        'ASSET_COMPUTE_TASK_OUTPUT_ASSET',
        'EVENT_ASSET_CREATED',
        now(),
        JSONB_BUILD_OBJECT(
            'assetKey', o.compute_task_key,
            'assetKind', o.asset_kind,
            'computeTaskKey', o.compute_task_key,
            'computeTaskOutputIdentifier', o.compute_task_output_identifier
        )
    FROM compute_task_output_assets o
    LEFT JOIN compute_tasks t ON o.compute_task_key = t.key;

    DROP SEQUENCE seq_tmp_events_offset;

$$) WHERE NOT table_exists('public', 'compute_task_output_assets');
