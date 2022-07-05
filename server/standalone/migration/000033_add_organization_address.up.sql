ALTER TABLE organizations
ADD COLUMN IF NOT EXISTS address varchar(200); 
