SELECT execute($$

    CREATE TABLE compute_task_outputs (
        compute_task_key UUID NOT NULL REFERENCES compute_tasks(key),
        identifier varchar(100) NOT NULL,
        permissions jsonb  NOT NULL,
        PRIMARY KEY (compute_task_key, identifier)
    );

    /* Intentionally don't migrate existing data to be consistent with migration 28 */

$$) WHERE not table_exists('public', 'compute_task_outputs');
