
-- Create an index on the 'id' column (although it's already the PRIMARY KEY)
CREATE INDEX idx_users_id ON users(id);

-- Create an index on the 'email' column for faster lookups
CREATE INDEX idx_users_email ON users(email);
