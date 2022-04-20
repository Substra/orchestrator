SELECT execute($$

    CREATE TABLE asset_kinds (
        kind varchar(50) PRIMARY KEY
    );

    INSERT INTO asset_kinds(kind)
    VALUES ('ASSET_NODE'),
           ('ASSET_METRIC'),
           ('ASSET_DATA_SAMPLE'),
           ('ASSET_DATA_MANAGER'),
           ('ASSET_ALGO'),
           ('ASSET_COMPUTE_TASK'),
           ('ASSET_COMPUTE_PLAN'),
           ('ASSET_MODEL'),
           ('ASSET_PERFORMANCE'),
           ('ASSET_FAILURE_REPORT');

    CREATE TABLE algo_inputs (
        algo_key UUID NOT NULL REFERENCES algos(key),
        identifier varchar(100) NOT NULL,
        kind varchar(50) NOT NULL REFERENCES asset_kinds(kind),
        multiple boolean NOT NULL,
        optional boolean NOT NULL,
        PRIMARY KEY(algo_key, identifier)
    );

    CREATE TABLE algo_outputs (
        algo_key UUID NOT NULL REFERENCES algos(key),
        identifier varchar(100) NOT NULL,
        kind varchar(50) NOT NULL REFERENCES asset_kinds(kind),
        multiple boolean NOT NULL,
        PRIMARY KEY(algo_key, identifier)
    );

$$) WHERE not table_exists('public', 'algo_inputs');
