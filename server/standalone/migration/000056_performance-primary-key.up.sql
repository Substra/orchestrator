SELECT execute($$
        ALTER TABLE performances
        DROP CONSTRAINT performances_pkey;

        ALTER TABLE performances
        DROP CONSTRAINT performances_compute_task_key_fkey;

        ALTER TABLE compute_task_output_assets
        DROP CONSTRAINT compute_task_output_assets_pkey;

        ALTER TABLE compute_task_output_assets
        DROP CONSTRAINT compute_task_output_assets_compute_task_key_compute_task_o_fkey;

        ALTER TABLE performances
        ADD COLUMN compute_task_output_identifier varchar(100);

        UPDATE compute_task_outputs cto
        SET identifier = f.name
        FROM functions f, events e
        WHERE e.asset_kind = 'ASSET_PERFORMANCE'
        AND f.key = (e.asset ->> 'metricKey')::uuid
        AND cto.compute_task_key = (e.asset ->> 'computeTaskKey')::uuid;

        UPDATE compute_task_output_assets ctoa
        SET compute_task_output_identifier = cto.identifier
        FROM compute_task_outputs cto, events e
        WHERE e.asset_kind = 'ASSET_PERFORMANCE'
        AND cto.compute_task_key=ctoa.compute_task_key
        AND cto.compute_task_key = (e.asset ->> 'computeTaskKey')::uuid;

        UPDATE performances p
        SET compute_task_output_identifier = cto.identifier
        FROM compute_task_outputs cto
        WHERE cto.compute_task_key=p.compute_task_key;

        ALTER TABLE performances
        ALTER COLUMN compute_task_output_identifier SET NOT NULL;

        ALTER TABLE performances
        ADD FOREIGN KEY (compute_task_key, compute_task_output_identifier)
        REFERENCES compute_task_outputs(compute_task_key, identifier);

        ALTER TABLE performances
        ADD PRIMARY KEY (compute_task_key, compute_task_output_identifier, function_key);

        ALTER TABLE compute_task_output_assets
        ADD FOREIGN KEY (compute_task_key, compute_task_output_identifier)
        REFERENCES compute_task_outputs(compute_task_key, identifier);

        ALTER TABLE compute_task_output_assets
        ADD PRIMARY KEY(compute_task_key, compute_task_output_identifier, position);

        UPDATE events e
        SET asset = jsonb_set(asset, '{computeTaskOutputIdentifier}', to_jsonb(cto.identifier))
        FROM compute_task_outputs cto
        WHERE asset_kind = 'ASSET_PERFORMANCE'
        AND NOT(asset ? 'computeTaskOutputIdentifier')
        AND cto.compute_task_key = (asset ->> 'computeTaskKey')::uuid;

$$) WHERE NOT column_exists('public', 'performances', 'compute_task_output_identifier');
