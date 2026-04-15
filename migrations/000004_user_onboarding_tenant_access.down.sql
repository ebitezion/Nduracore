DROP TABLE IF EXISTS user_tenants;

ALTER TABLE users
    ALTER COLUMN status SET DEFAULT 'active';
