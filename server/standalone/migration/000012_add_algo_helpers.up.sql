CREATE OR REPLACE FUNCTION build_addressable_jsonb(checksum text, storage_address text) RETURNS jsonb AS
$$
BEGIN
    RETURN JSONB_BUILD_OBJECT('checksum', checksum, 'storageAddress', storage_address);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION build_algo_jsonb(
    key uuid,
    name text,
    category text,
    description_checksum text,
    description_address text,
    algorithm_checksum text,
    algorithm_address text,
    permissions jsonb,
    owner text,
    creation_date timestamptz,
    metadata jsonb
) RETURNS jsonb AS
$$
BEGIN
    RETURN JSONB_BUILD_OBJECT(
            'key', key,
            'name', name,
            'category', category,
            'description', build_addressable_jsonb(description_checksum, description_address),
            'algorithm', build_addressable_jsonb(algorithm_checksum, algorithm_address),
            'permissions', permissions,
            'owner', owner,
            'creationDate', to_rfc_3339(creation_date),
            'metadata', metadata
        );
END;
$$ LANGUAGE plpgsql;
