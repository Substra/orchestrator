INSERT INTO compute_task_categories(category)
VALUES ('TASK_PREDICT')
ON CONFLICT DO NOTHING;
