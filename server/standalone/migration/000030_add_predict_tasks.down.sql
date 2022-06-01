DELETE FROM compute_tasks
WHERE category = 'TASK_PREDICT';

DELETE FROM compute_task_categories
WHERE category = 'TASK_PREDICT';	
