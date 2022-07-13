ALTER TABLE organizations
ALTER COLUMN address DROP NOT NULL, ALTER COLUMN address DROP DEFAULT '';
UPDATE organizations SET address = NULL WHERE address = '';
