INSERT INTO compute_task_categories(category)
VALUES ('TASK_UNKNOWN')
ON CONFLICT DO NOTHING;
