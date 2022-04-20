SELECT execute($$

    DROP TABLE algo_inputs;
    DROP TABLE algo_outputs;
    DROP TABLE asset_kinds;

$$) WHERE table_exists('public', 'algo_inputs');
