CREATE TABLE nodes (
       id varchar(100) PRIMARY KEY
);
CREATE TABLE objectives (
       id UUID PRIMARY KEY,
       asset JSONB NOT NULL
);
