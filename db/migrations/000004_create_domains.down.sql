-- Remove the FK from api_keys before dropping domains
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS fk_api_keys_domain_id;

DROP TABLE IF EXISTS domains;
