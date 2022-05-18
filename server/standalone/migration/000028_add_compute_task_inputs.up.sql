SELECT execute($$

    CREATE TABLE compute_task_inputs (
        compute_task_key UUID NOT NULL REFERENCES compute_tasks(key),
        identifier varchar(100) NOT NULL,
        position integer NOT NULL,
        asset_key UUID,
        parent_task_key UUID REFERENCES compute_tasks(key),
        parent_task_output_identifier varchar(100),
        PRIMARY KEY (compute_task_key, position)
    );

    ALTER TABLE compute_task_inputs ADD CONSTRAINT co_compute_task_inputs_asset_key_parent_task_key_parent_task_output_identifier CHECK (
        (asset_key IS NULL AND parent_task_key IS NOT NULL AND parent_task_output_identifier IS NOT NULL)
        OR
        (asset_key IS NOT NULL AND parent_task_key IS NULL AND parent_task_output_identifier IS NULL)
    );

    /* Intentionally don't migrate existing data: a full migration is not doable in SQL-only */

$$) WHERE not table_exists('public', 'compute_task_inputs');
