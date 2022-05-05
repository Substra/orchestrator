SELECT execute($$

    DELETE FROM algo_categories
    WHERE category = 'ALGO_PREDICT';

$$) WHERE 'ALGO_PREDICT' IN (
    SELECT category
    FROM algo_categories
);
