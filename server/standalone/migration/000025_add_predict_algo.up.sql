SELECT execute($$

    INSERT INTO algo_categories(category)
    VALUES ('ALGO_PREDICT');

$$) WHERE 'ALGO_PREDICT' NOT IN (
    SELECT category
    FROM algo_categories
);
