/* Delete obsolete properties from JSON */
UPDATE compute_tasks
SET task_data = task_data
    #- '{train,modelPermissions}'
    #- '{composite,headPermissions}'
    #- '{composite,trunkPermissions}'
    #- '{aggregate,modelPermissions}'
    #- '{predict,predictionPermissions}'
;

/* Delete obsolete properties from JSON */
UPDATE events
SET asset = asset
    #- '{train,modelPermissions}'
    #- '{composite,headPermissions}'
    #- '{composite,trunkPermissions}'
    #- '{aggregate,modelPermissions}'
    #- '{predict,predictionPermissions}'
WHERE asset_kind = 'ASSET_COMPUTE_TASK';
