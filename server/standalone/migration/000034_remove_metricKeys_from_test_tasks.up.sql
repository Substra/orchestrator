/*
 * Test tasks:
 * - Set algo key to metricsKeys[0] (first item of metricsKeys)
 * - Delete metricsKeys
*/

UPDATE compute_tasks
SET algo_key = (task_data->'test'->'metricKeys'->>0)::uuid,
    task_data = task_data #- '{test,metricKeys}'
WHERE category = 'TASK_TEST'
    AND task_data->'test'?'metricKeys'; /* task has a "metricKeys" key */

UPDATE events
/* Only update algo.key. Don't update the other algo.* fields (too complex) */
SET asset = jsonb_set(asset, '{algo,key}', asset->'test'->'metricKeys'->0) #- '{test,metricKeys}'
WHERE asset_kind = 'ASSET_COMPUTE_TASK'
    AND asset->'test'?'metricKeys';
