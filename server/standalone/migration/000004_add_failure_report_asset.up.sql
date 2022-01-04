CREATE TABLE failure_reports (
    compute_task_id UUID PRIMARY KEY REFERENCES compute_tasks (id),
    channel varchar(100) NOT NULL,
    asset JSONB NOT NULL
);
