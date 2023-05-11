SELECT execute($$
        /* Delete constraints regarding the function key */
        ALTER TABLE performances
        DROP CONSTRAINT performances_pkey;

        /* Create an identifier column on performances */
        ALTER TABLE performances
        DROP COLUMN function_key;


        /* Set constraints back */
        ALTER TABLE performances
        ADD PRIMARY KEY (compute_task_key, compute_task_output_identifier);

        /* Update asset */
        UPDATE events e
        SET asset = asset::jsonb - 'functionKey'
        WHERE asset_kind = 'ASSET_PERFORMANCE';

$$) WHERE column_exists('public', 'performances', 'function_key');
