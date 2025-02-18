CREATE TABLE IF NOT EXISTS url_table (
    id SERIAL PRIMARY KEY,          -- Unique ID for each URL
    long_url TEXT NOT NULL,         -- The original long URL
    short_url VARCHAR(7) UNIQUE NOT NULL, -- The shortened URL (e.g., "abc123")

    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    expires_at timestamp(0) with time zone NOT NULL 
);
