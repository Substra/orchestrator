SELECT execute($$
    ALTER TABLE events
    ADD COLUMN asset jsonb;

    -- Algos
    UPDATE events e
    SET asset = build_algo_jsonb(
            a.key,
            a.name,
            a.category,
            a.description_checksum,
            a.description_address,
            a.algorithm_checksum,
            a.algorithm_address,
            a.permissions,
            a.owner,
            a.creation_date,
            a.metadata
        )
    FROM expanded_algos a
    WHERE e.asset_key = a.key::text AND e.asset_kind = 'ASSET_ALGO';

    -- Compute plans
    UPDATE events e
    SET asset = JSONB_BUILD_OBJECT(
            'key', cp.key,
            'owner', cp.owner,
            'deleteIntermediaryModels', cp.delete_intermediary_models,
            'creationDate', to_rfc_3339(cp.creation_date),
            'tag', cp.tag,
            'name', cp.name,
            'metadata', cp.metadata
        )
    FROM compute_plans cp
    WHERE e.asset_key = cp.key::text AND e.asset_kind = 'ASSET_COMPUTE_PLAN';

    -- Compute tasks
    UPDATE events e
    SET asset = JSONB_BUILD_OBJECT(
                'key', t.key,
                'category', t.category,
                'algo', build_algo_jsonb(
                        t.algo_key,
                        t.algo_name,
                        t.algo_category,
                        t.algo_description_checksum,
                        t.algo_description_address,
                        t.algo_algorithm_checksum,
                        t.algo_algorithm_address,
                        t.algo_permissions,
                        t.algo_owner,
                        t.algo_creation_date,
                        t.algo_metadata
                    ),
                'owner', t.owner,
                'computePlanKey', t.compute_plan_key,
                'parentTaskKeys', t.parent_task_keys,
                'rank', t.rank,
                'status', t.status,
                'worker', t.worker,
                'creationDate', to_rfc_3339(t.creation_date),
                'logsPermission', t.logs_permission,
                'metadata', t.metadata
            ) || t.task_data
    FROM expanded_compute_tasks t
    WHERE e.asset_key = t.key::text AND e.asset_kind = 'ASSET_COMPUTE_TASK';

    -- Datamanagers
    UPDATE events e
    SET asset = JSONB_BUILD_OBJECT(
            'key', dm.key,
            'name', dm.name,
            'owner', dm.owner,
            'permissions', dm.permissions,
            'description', build_addressable_jsonb(dm.description_checksum, dm.description_address),
            'opener', build_addressable_jsonb(dm.opener_checksum, dm.opener_address),
            'type', dm.type,
            'creationDate', to_rfc_3339(dm.creation_date),
            'logsPermission', dm.logs_permission,
            'metadata', dm.metadata
        )
    FROM expanded_datamanagers dm
    WHERE e.asset_key = dm.key::text AND e.asset_kind = 'ASSET_DATA_MANAGER';

    -- Datasamples
    UPDATE events e
    SET asset = JSONB_BUILD_OBJECT(
            'key', ds.key,
            'dataManagerKeys', ds.datamanager_keys,
            'owner', ds.owner,
            'testOnly', ds.test_only,
            'checksum', ds.checksum,
            'creationDate', to_rfc_3339(ds.creation_date)
        )
    FROM expanded_datasamples ds
    WHERE e.asset_key = ds.key::text AND e.asset_kind = 'ASSET_DATA_SAMPLE';

    -- Failure reports
    UPDATE events e
    SET asset = JSONB_BUILD_OBJECT(
            'computeTaskKey', r.compute_task_key,
            'errorType', r.error_type,
            'logsAddress', build_addressable_jsonb(r.logs_checksum, r.logs_address),
            'creationDate', to_rfc_3339(r.creation_date),
            'owner', r.owner
        )
    FROM expanded_failure_reports r
    WHERE e.asset_key = r.compute_task_key::text AND e.asset_kind = 'ASSET_FAILURE_REPORT';

    -- Models
    UPDATE events e
    SET asset = JSONB_BUILD_OBJECT(
            'key', m.key,
            'category', m.category,
            'computeTaskKey', m.compute_task_key,
            'address', build_addressable_jsonb(m.checksum, m.address),
            'permissions', m.permissions,
            'owner', m.owner,
            'creationDate', to_rfc_3339(m.creation_date)
        )
    FROM expanded_models m
    WHERE e.asset_key = m.key::text AND e.asset_kind = 'ASSET_MODEL';

    -- Nodes
    UPDATE events e
    SET asset = JSONB_BUILD_OBJECT(
            'id', n.id,
            'creationDate', to_rfc_3339(n.creation_date)
        )
    FROM nodes n
    WHERE e.asset_key = n.id AND e.asset_kind = 'ASSET_NODE';

    -- Performances
    UPDATE events e
    SET asset = JSONB_BUILD_OBJECT(
            'computeTaskKey', p.compute_task_key,
            'metricKey', p.algo_key,
            'performanceValue', p.performance_value,
            'creationDate', to_rfc_3339(p.creation_date)
        )
    FROM performances p
    WHERE e.asset_key = p.compute_task_key || '|' || p.algo_key AND e.asset_kind = 'ASSET_PERFORMANCE';

    ALTER TABLE events
    ALTER COLUMN asset SET NOT NULL;
$$) WHERE NOT column_exists('public', 'events', 'asset');
