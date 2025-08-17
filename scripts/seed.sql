INSERT INTO users (id, name, email, created_at, updated_at) VALUES
  (gen_random_uuid(), 'Alice', 'alice@example.com', now(), now()),
  (gen_random_uuid(), 'Bob', 'bob@example.com', now(), now())
ON CONFLICT DO NOTHING;
