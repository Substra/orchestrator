UPDATE organizations SET address = '' WHERE address IS NULL;
ALTER TABLE organizations
ALTER COLUMN address SET NOT NULL, ALTER COLUMN address SET DEFAULT '';
